package contracts

import (
	"encoding/json"
	"fmt"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/access"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// OracleContract implements external data oracle integration (ADR-029, v7.0).
// Trusted issuer organisations publish grid generation data from ENTSO-E
// (European Network of Transmission System Operators for Electricity) onto
// the ledger. This data is cross-referenced during GO issuance to validate
// that a producer's claimed generation is consistent with the grid's actual
// generation mix for that area and time period.
type OracleContract struct {
	contractapi.Contract
}

// GridGenerationRecord stores a snapshot of grid generation data from ENTSO-E.
type GridGenerationRecord struct {
	RecordID        string  `json:"recordId"`
	BiddingZone     string  `json:"biddingZone"`     // ENTSO-E bidding zone (e.g. "DE-LU", "NL")
	PeriodStart     int64   `json:"periodStart"`     // UNIX timestamp
	PeriodEnd       int64   `json:"periodEnd"`       // UNIX timestamp
	EnergySource    string  `json:"energySource"`    // EECS code (e.g. "F01010100")
	GenerationMW    float64 `json:"generationMW"`    // Total generation in MW for the period
	EmissionFactor  float64 `json:"emissionFactor"`  // gCO2eq/kWh for the source
	DataSource      string  `json:"dataSource"`      // "ENTSO-E-TP", "ENTSOG", etc.
	PublishedBy     string  `json:"publishedBy"`     // Issuer MSP that uploaded the record
	PublishedAt     int64   `json:"publishedAt"`
}

// Oracle key prefix.
const (
	PrefixOracle  = "oracle_"
	RangeEndOracle = "oracle_~"
)

// PublishGridData publishes an ENTSO-E grid generation record on-chain.
// Only issuers can publish oracle data (they act as trusted data feeds).
// Transient key: "GridData" containing BiddingZone, PeriodStart, PeriodEnd,
// EnergySource, GenerationMW, EmissionFactor, DataSource.
func (c *OracleContract) PublishGridData(ctx contractapi.TransactionContextInterface) (*GridGenerationRecord, error) {
	if err := access.RequireRole(ctx, access.RoleIssuer); err != nil {
		return nil, fmt.Errorf("only issuers can publish oracle data: %v", err)
	}

	type gridInput struct {
		BiddingZone    string  `json:"BiddingZone"`
		PeriodStart    int64   `json:"PeriodStart"`
		PeriodEnd      int64   `json:"PeriodEnd"`
		EnergySource   string  `json:"EnergySource"`
		GenerationMW   float64 `json:"GenerationMW"`
		EmissionFactor float64 `json:"EmissionFactor"`
		DataSource     string  `json:"DataSource"`
	}

	var input gridInput
	if err := util.UnmarshalTransient(ctx, "GridData", &input); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("BiddingZone", input.BiddingZone); err != nil {
		return nil, err
	}
	if err := util.ValidateNonEmpty("DataSource", input.DataSource); err != nil {
		return nil, err
	}
	if err := util.ValidateEnergySource(input.EnergySource); err != nil {
		return nil, err
	}
	if input.PeriodEnd <= input.PeriodStart {
		return nil, fmt.Errorf("PeriodEnd must be after PeriodStart")
	}

	recordID, err := assets.GenerateID(ctx, PrefixOracle, 0)
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

	record := &GridGenerationRecord{
		RecordID:       recordID,
		BiddingZone:    input.BiddingZone,
		PeriodStart:    input.PeriodStart,
		PeriodEnd:      input.PeriodEnd,
		EnergySource:   input.EnergySource,
		GenerationMW:   input.GenerationMW,
		EmissionFactor: input.EmissionFactor,
		DataSource:     input.DataSource,
		PublishedBy:    issuerMSP,
		PublishedAt:    now,
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
			"biddingZone":  input.BiddingZone,
			"energySource": input.EnergySource,
			"dataSource":   input.DataSource,
		},
	})

	return record, nil
}

// GetGridData reads an oracle record by ID.
func (c *OracleContract) GetGridData(ctx contractapi.TransactionContextInterface, recordID string) (*GridGenerationRecord, error) {
	recordBytes, err := ctx.GetStub().GetState(recordID)
	if err != nil {
		return nil, fmt.Errorf("failed to read oracle record: %v", err)
	}
	if recordBytes == nil {
		return nil, fmt.Errorf("oracle record %s does not exist", recordID)
	}

	var record GridGenerationRecord
	if err := json.Unmarshal(recordBytes, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal oracle record: %v", err)
	}
	return &record, nil
}

// ListGridDataPaginated returns paginated oracle records.
func (c *OracleContract) ListGridDataPaginated(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) (string, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	resultsIterator, metadata, err := ctx.GetStub().GetStateByRangeWithPagination(PrefixOracle, RangeEndOracle, pageSize, bookmark)
	if err != nil {
		return "", fmt.Errorf("error querying oracle records: %v", err)
	}
	defer resultsIterator.Close()

	var records []*GridGenerationRecord
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return "", err
		}
		var record GridGenerationRecord
		if err := json.Unmarshal(queryResult.Value, &record); err != nil {
			return "", err
		}
		records = append(records, &record)
	}

	result := struct {
		Records  []*GridGenerationRecord `json:"records"`
		Bookmark string                  `json:"bookmark"`
		Count    int32                   `json:"count"`
	}{
		Records:  records,
		Bookmark: metadata.GetBookmark(),
		Count:    metadata.GetFetchedRecordsCount(),
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %v", err)
	}
	return string(resultBytes), nil
}

// CrossReferenceGO validates a GO's generation data against ENTSO-E oracle records.
// Returns true if the GO's production data is consistent with grid generation data.
func (c *OracleContract) CrossReferenceGO(ctx contractapi.TransactionContextInterface, goAssetID string, oracleRecordID string) (bool, error) {
	// Read the GO
	goJSON, err := ctx.GetStub().GetState(goAssetID)
	if err != nil {
		return false, fmt.Errorf("failed to read GO: %v", err)
	}
	if goJSON == nil {
		return false, fmt.Errorf("GO %s does not exist", goAssetID)
	}

	// Read the oracle record
	recordBytes, err := ctx.GetStub().GetState(oracleRecordID)
	if err != nil {
		return false, fmt.Errorf("failed to read oracle record: %v", err)
	}
	if recordBytes == nil {
		return false, fmt.Errorf("oracle record %s does not exist", oracleRecordID)
	}

	var record GridGenerationRecord
	if err := json.Unmarshal(recordBytes, &record); err != nil {
		return false, fmt.Errorf("failed to unmarshal oracle record: %v", err)
	}

	// Parse GO to check production period overlap
	var goData map[string]interface{}
	if err := json.Unmarshal(goJSON, &goData); err != nil {
		return false, fmt.Errorf("failed to unmarshal GO: %v", err)
	}

	// Validate that the GO's production period falls within the oracle period
	goPeriodStart, _ := goData["ProductionPeriodStart"].(float64)
	goPeriodEnd, _ := goData["ProductionPeriodEnd"].(float64)

	if goPeriodStart == 0 || goPeriodEnd == 0 {
		// GO doesn't have production period set — cannot cross-reference
		return false, fmt.Errorf("GO %s does not have production period timestamps", goAssetID)
	}

	// Check period overlap
	if int64(goPeriodStart) >= record.PeriodEnd || int64(goPeriodEnd) <= record.PeriodStart {
		return false, nil // No overlap
	}

	// Check energy source match
	goEnergy, _ := goData["EnergySource"].(string)
	if goEnergy != "" && record.EnergySource != "" && goEnergy != record.EnergySource {
		return false, nil // Energy source mismatch
	}

	return true, nil
}
