package contracts

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// OracleContract implements a carrier-agnostic external market data oracle (ADR-029, v10.1+).
//
// Any trusted issuer may publish reference data from any external source. The design uses a
// single generic OracleRecord structure extensible to arbitrary carrier types and data sources,
// meaning new carriers (e.g. natural gas, green ammonia, compressed air) require no chaincode
// upgrade — only a new CarrierType string and appropriate Attributes entries.
//
// Supported data sources include (non-exhaustive):
//
//	Electricity:     ENTSO-E Transparency Platform (generation, emission factors, bidding zones)
//	Natural Gas:     ENTSOG Transparency Platform (gas flows, capacity, gas quality)
//	Hydrogen:        EHB (European Hydrogen Backbone), H2Global, HyWay27, national H2 registries
//	Biogas/Biomethane: EBA statistical database, BiogasRegister-DE, TIGF (France), Vertogas (NL),
//	                 Green Gas Certification Scheme (GGCS, UK)
//	Heating/Cooling: Euroheat & Power, VEKS/HOFOR (Denmark), Energiföretagen Sverige, Finnish Energy
//	Custom:          Any data source, identified by a non-empty CarrierType string
type OracleContract struct {
	contractapi.Contract
}

// OracleRecord is the universal market/network data record published by a trusted issuer.
// It replaces the four carrier-specific record types from v10.0 (GridGenerationRecord,
// HydrogenProductionRecord, BiogasProductionRecord, HeatingCoolingNetworkRecord).
//
// Carrier-specific metadata that does not fit the common fields should be placed in Attributes:
//
//	electricity:     GenerationMW, BiddingZone (= Zone), EECS source code (= ProductionMethod)
//	natural_gas:     FlowType, GasQuality, PressureBar, NetworkPoint (= Zone)
//	hydrogen:        InputEnergyMWh, ElectrolyserType
//	biogas:          FeedstockType, VolumeNm3, MethaneContent
//	heating_cooling: AverageSupplyTempC, DeliveryNetwork
type OracleRecord struct {
	RecordID         string            `json:"recordId"`
	CarrierType      string            `json:"carrierType"`      // e.g. "electricity", "natural_gas", "hydrogen", "biogas", "heating_cooling"
	Zone             string            `json:"zone"`             // Bidding zone, network point, region, delivery area — carrier-agnostic
	PeriodStart      int64             `json:"periodStart"`      // UNIX timestamp (inclusive)
	PeriodEnd        int64             `json:"periodEnd"`        // UNIX timestamp (exclusive)
	ProductionMethod string            `json:"productionMethod"` // EECS source code, flow type, production or generation method
	EnergyUnit       string            `json:"energyUnit"`       // "MWh", "GWh", "kg", "Nm3", "MJ"
	Quantity         float64           `json:"quantity"`         // Total production/flow for the period in EnergyUnit
	EmissionFactor   float64           `json:"emissionFactor"`   // gCO2eq per EnergyUnit (0 = not reported)
	DataSource       string            `json:"dataSource"`       // e.g. "ENTSO-E-TP", "ENTSOG-TP", "EBA", "BiogasRegister-DE", "Euroheat"
	Attributes       map[string]string `json:"attributes"`       // Carrier-specific extras: see struct doc above
	PublishedBy      string            `json:"publishedBy"`      // Issuer MSP
	PublishedAt      int64             `json:"publishedAt"`      // UNIX timestamp of on-chain publication
}

// OracleCrossReference is written on-chain when CrossReferenceGO is invoked.
// It provides a permanent, auditable attestation of the consistency check result.
type OracleCrossReference struct {
	CrossRefID     string `json:"crossRefId"`
	GOAssetID      string `json:"goAssetId"`
	OracleRecordID string `json:"oracleRecordId"`
	IsConsistent   bool   `json:"isConsistent"`
	CheckedBy      string `json:"checkedBy"`
	CheckedAt      int64  `json:"checkedAt"`
}

const (
	PrefixOracleBase     = "oracle_"
	RangeEndOracleBase   = "oracle_~"
	PrefixOracleCrossRef = "xref_"
)

// oracleKeyRange returns the CouchDB key range for a given carrier type.
// An empty carrierType covers all oracle records.
func oracleKeyRange(carrierType string) (prefix, rangeEnd string) {
	if carrierType == "" {
		return PrefixOracleBase, RangeEndOracleBase
	}
	// Normalise: lowercase, spaces → underscores — matches how PublishOracleData stores the key.
	ct := strings.ToLower(strings.ReplaceAll(carrierType, " ", "_"))
	prefix = PrefixOracleBase + ct + "_"
	rangeEnd = prefix + "~"
	return
}

// normaliseCarrierType lowercases and replaces spaces with underscores so that
// "Natural Gas" and "natural_gas" produce the same key prefix.
func normaliseCarrierType(ct string) string {
	return strings.ToLower(strings.ReplaceAll(ct, " ", "_"))
}

// ============================================================================
// Publish
// ============================================================================

// PublishOracleData publishes a market/network data record on-chain.
// Only issuers may call this function.
//
// Transient key: "OracleData"
//
//	CarrierType      string            — Carrier type (case-insensitive, e.g. "electricity", "natural_gas")
//	Zone             string            — Bidding zone, network point, or delivery region
//	PeriodStart      int64             — UNIX timestamp (inclusive)
//	PeriodEnd        int64             — UNIX timestamp (exclusive)
//	ProductionMethod string            — Source code, flow type, or production method (optional)
//	EnergyUnit       string            — Unit of Quantity ("MWh", "kg", "Nm3", "GWh", etc.)
//	Quantity         float64           — Total production/flow in EnergyUnit (>= 0)
//	EmissionFactor   float64           — gCO2eq per EnergyUnit (0 if not reported)
//	DataSource       string            — Data source identifier (e.g. "ENTSO-E-TP", "ENTSOG-TP")
//	Attributes       map[string]string — Carrier-specific extras (optional, pass {} if none)
func (c *OracleContract) PublishOracleData(ctx contractapi.TransactionContextInterface) (*OracleRecord, error) {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return nil, fmt.Errorf("only issuers can publish oracle data: %v", err)
	}

	type oracleInput struct {
		CarrierType      string            `json:"CarrierType"`
		Zone             string            `json:"Zone"`
		PeriodStart      int64             `json:"PeriodStart"`
		PeriodEnd        int64             `json:"PeriodEnd"`
		ProductionMethod string            `json:"ProductionMethod"`
		EnergyUnit       string            `json:"EnergyUnit"`
		Quantity         float64           `json:"Quantity"`
		EmissionFactor   float64           `json:"EmissionFactor"`
		DataSource       string            `json:"DataSource"`
		Attributes       map[string]string `json:"Attributes"`
	}

	var input oracleInput
	if err := util.UnmarshalTransient(ctx, "OracleData", &input); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("CarrierType", input.CarrierType); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("Zone", input.Zone); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("EnergyUnit", input.EnergyUnit); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("DataSource", input.DataSource); err != nil {
		return nil, err
	}
	if input.PeriodEnd <= input.PeriodStart {
		return nil, fmt.Errorf("PeriodEnd must be after PeriodStart")
	}
	if input.Quantity < 0 {
		return nil, fmt.Errorf("Quantity must be non-negative")
	}

	normCarrier := normaliseCarrierType(input.CarrierType)
	prefix, _ := oracleKeyRange(normCarrier)

	recordID, err := assets.GenerateID(ctx, prefix, 0)
	if err != nil {
		return nil, fmt.Errorf("error generating oracle record ID: %v", err)
	}

	now, err := util.GetTimestamp(ctx)
	if err != nil {
		return nil, err
	}
	issuerMSP, err := access.GetClientMSPID(ctx)
	if err != nil {
		return nil, err
	}

	if input.Attributes == nil {
		input.Attributes = make(map[string]string)
	}

	record := &OracleRecord{
		RecordID:         recordID,
		CarrierType:      normCarrier,
		Zone:             input.Zone,
		PeriodStart:      input.PeriodStart,
		PeriodEnd:        input.PeriodEnd,
		ProductionMethod: input.ProductionMethod,
		EnergyUnit:       input.EnergyUnit,
		Quantity:         input.Quantity,
		EmissionFactor:   input.EmissionFactor,
		DataSource:       input.DataSource,
		Attributes:       input.Attributes,
		PublishedBy:      issuerMSP,
		PublishedAt:      now,
	}

	recordBytes, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal oracle record: %v", err)
	}
	if err := ctx.GetStub().PutState(recordID, recordBytes); err != nil {
		return nil, fmt.Errorf("failed to write oracle record: %v", err)
	}

	_ = util.EmitLifecycleEvent(ctx, util.LifecycleEvent{
		EventType: "ORACLE_DATA_PUBLISHED",
		AssetID:   recordID,
		Initiator: issuerMSP,
		Timestamp: now,
		Details: map[string]string{
			"carrierType": normCarrier,
			"zone":        input.Zone,
			"dataSource":  input.DataSource,
			"energyUnit":  input.EnergyUnit,
		},
	})

	return record, nil
}

// ============================================================================
// Read
// ============================================================================

// GetOracleData reads an oracle record by its record ID.
// Any channel member may call this.
func (c *OracleContract) GetOracleData(ctx contractapi.TransactionContextInterface, recordID string) (*OracleRecord, error) {
	recordBytes, err := ctx.GetStub().GetState(recordID)
	if err != nil {
		return nil, fmt.Errorf("failed to read oracle record: %v", err)
	}
	if recordBytes == nil {
		return nil, fmt.Errorf("oracle record %s does not exist", recordID)
	}
	var record OracleRecord
	if err := json.Unmarshal(recordBytes, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal oracle record: %v", err)
	}
	return &record, nil
}

// ListOracleDataPaginated returns a paginated list of oracle records.
// Pass carrierType="" to list records for all carriers.
// Pass carrierType="electricity" (or "natural_gas", "hydrogen", etc.) to filter by carrier.
// Any channel member may call this.
func (c *OracleContract) ListOracleDataPaginated(ctx contractapi.TransactionContextInterface, carrierType string, pageSize int32, bookmark string) (string, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	prefix, rangeEnd := oracleKeyRange(carrierType)

	iter, meta, err := ctx.GetStub().GetStateByRangeWithPagination(prefix, rangeEnd, pageSize, bookmark)
	if err != nil {
		return "", fmt.Errorf("error querying oracle records: %v", err)
	}
	defer iter.Close()

	var records []*OracleRecord
	for iter.HasNext() {
		qr, err := iter.Next()
		if err != nil {
			return "", err
		}
		var record OracleRecord
		if err := json.Unmarshal(qr.Value, &record); err != nil {
			return "", err
		}
		records = append(records, &record)
	}

	result := struct {
		Records     []*OracleRecord `json:"records"`
		Bookmark    string          `json:"bookmark"`
		Count       int32           `json:"count"`
		CarrierType string          `json:"carrierType,omitempty"`
	}{
		Records:     records,
		Bookmark:    meta.GetBookmark(),
		Count:       meta.GetFetchedRecordsCount(),
		CarrierType: carrierType,
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %v", err)
	}
	return string(resultBytes), nil
}

// ============================================================================
// Cross-Reference
// ============================================================================

// CrossReferenceGO validates a GO's production period against an oracle record and writes
// an OracleCrossReference attestation on-chain.
//
// Returns true when:
//  1. The GO's ProductionPeriodStart/End overlaps with the oracle record's PeriodStart/End, AND
//  2. The GO's EnergySource field (if set) does not conflict with the record's ProductionMethod.
//
// The cross-reference record is written regardless of the result, providing a permanent audit trail.
// Any channel member may call this.
func (c *OracleContract) CrossReferenceGO(ctx contractapi.TransactionContextInterface, goAssetID string, oracleRecordID string) (bool, error) {
	goJSON, err := ctx.GetStub().GetState(goAssetID)
	if err != nil {
		return false, fmt.Errorf("failed to read GO: %v", err)
	}
	if goJSON == nil {
		return false, fmt.Errorf("GO %s does not exist", goAssetID)
	}

	recordBytes, err := ctx.GetStub().GetState(oracleRecordID)
	if err != nil {
		return false, fmt.Errorf("failed to read oracle record: %v", err)
	}
	if recordBytes == nil {
		return false, fmt.Errorf("oracle record %s does not exist", oracleRecordID)
	}

	var record OracleRecord
	if err := json.Unmarshal(recordBytes, &record); err != nil {
		return false, fmt.Errorf("failed to unmarshal oracle record: %v", err)
	}

	var goData map[string]interface{}
	if err := json.Unmarshal(goJSON, &goData); err != nil {
		return false, fmt.Errorf("failed to unmarshal GO: %v", err)
	}

	// Temporal overlap is mandatory
	goPeriodStart, _ := goData["ProductionPeriodStart"].(float64)
	goPeriodEnd, _ := goData["ProductionPeriodEnd"].(float64)
	if goPeriodStart == 0 || goPeriodEnd == 0 {
		return false, fmt.Errorf("GO %s does not have production period timestamps", goAssetID)
	}
	if int64(goPeriodStart) >= record.PeriodEnd || int64(goPeriodEnd) <= record.PeriodStart {
		return c.writeXRef(ctx, goAssetID, oracleRecordID, false)
	}

	// Soft-match on energy source / production method (mismatch → false, not error)
	goSource, _ := goData["EnergySource"].(string)
	if goSource != "" && record.ProductionMethod != "" && goSource != record.ProductionMethod {
		return c.writeXRef(ctx, goAssetID, oracleRecordID, false)
	}

	return c.writeXRef(ctx, goAssetID, oracleRecordID, true)
}

// writeXRef persists an OracleCrossReference record and returns the consistency flag.
func (c *OracleContract) writeXRef(ctx contractapi.TransactionContextInterface, goAssetID, oracleRecordID string, consistent bool) (bool, error) {
	now, _ := util.GetTimestamp(ctx)
	checkerMSP, _ := access.GetClientMSPID(ctx)

	xrefID := fmt.Sprintf("%s%s_%s", PrefixOracleCrossRef, goAssetID, oracleRecordID)
	xref := OracleCrossReference{
		CrossRefID:     xrefID,
		GOAssetID:      goAssetID,
		OracleRecordID: oracleRecordID,
		IsConsistent:   consistent,
		CheckedBy:      checkerMSP,
		CheckedAt:      now,
	}
	xrefBytes, err := json.Marshal(xref)
	if err != nil {
		return consistent, fmt.Errorf("failed to marshal cross-reference: %v", err)
	}
	if err := ctx.GetStub().PutState(xrefID, xrefBytes); err != nil {
		return consistent, fmt.Errorf("failed to write cross-reference: %v", err)
	}
	return consistent, nil
}

// GetCrossReference reads a previously written cross-reference attestation.
// xrefID format: "xref_<goAssetID>_<oracleRecordID>"
func (c *OracleContract) GetCrossReference(ctx contractapi.TransactionContextInterface, goAssetID string, oracleRecordID string) (*OracleCrossReference, error) {
	xrefID := fmt.Sprintf("%s%s_%s", PrefixOracleCrossRef, goAssetID, oracleRecordID)
	xrefBytes, err := ctx.GetStub().GetState(xrefID)
	if err != nil {
		return nil, fmt.Errorf("failed to read cross-reference: %v", err)
	}
	if xrefBytes == nil {
		return nil, fmt.Errorf("cross-reference for GO %s / oracle record %s does not exist", goAssetID, oracleRecordID)
	}
	var xref OracleCrossReference
	if err := json.Unmarshal(xrefBytes, &xref); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cross-reference: %v", err)
	}
	return &xref, nil
}
