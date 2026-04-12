package assets

// BiogasGO is the public (world-state) representation of a biogas guarantee of origin.
// ADR-015: Extends the multi-carrier model with biogas support per RED III.
// ADR-012: Fields align with CEN-EN 16325 GO data attributes.
type BiogasGO struct {
	AssetID               string `json:"AssetID"`
	CreationDateTime      int64  `json:"CreationDateTime"`
	GOType                string `json:"GOType"`                         // always "Biogas"
	Status                string `json:"Status"`                         // ADR-007: "active", "cancelled", "transferred"
	QuantityCommitment    string `json:"QuantityCommitment"`             // ADR-009: SHA-256(VolumeNm3 || salt)
	CountryOfOrigin       string `json:"CountryOfOrigin,omitempty" metadata:",optional"`      // ADR-012: ISO 3166-1 alpha-2
	GridConnectionPoint   string `json:"GridConnectionPoint,omitempty" metadata:",optional"`  // ADR-012: EIC code
	SupportScheme         string `json:"SupportScheme,omitempty" metadata:",optional"`        // ADR-012: "none", "FIT", "FIP", "quota"
	EnergySource          string `json:"EnergySource,omitempty" metadata:",optional"`         // ADR-012: EN 16325 source code (e.g. "F04010100" = biogas)
	ProductionPeriodStart int64  `json:"ProductionPeriodStart,omitempty" metadata:",optional"` // ADR-012: UNIX timestamp
	ProductionPeriodEnd   int64  `json:"ProductionPeriodEnd,omitempty" metadata:",optional"`   // ADR-012: UNIX timestamp
}

// BiogasGOPrivateDetails holds the confidential fields for a biogas GO.
type BiogasGOPrivateDetails struct {
	AssetID                 string   `json:"AssetID"`
	OwnerID                 string   `json:"OwnerID"`
	CreationDateTime        int64    `json:"CreationDateTime"`
	VolumeNm3              float64  `json:"VolumeNm3"`              // Volume in normal cubic metres
	EnergyContentMWh       float64  `json:"EnergyContentMWh"`       // Calorific value converted to MWh
	Emissions              float64  `json:"Emissions"`              // gCO2eq
	BiogasProductionMethod string   `json:"BiogasProductionMethod"` // "anaerobic_digestion", "landfill_gas", etc.
	FeedstockType          string   `json:"FeedstockType"`          // "agricultural_waste", "sewage", "energy_crops"
	ConsumptionDeclarations []string `json:"ConsumptionDeclarations"`
	DeviceID               string   `json:"DeviceID,omitempty"`
	CommitmentSalt         string   `json:"CommitmentSalt,omitempty"` // ADR-009
}

// CancellationStatementBiogas records the cancellation of one or more biogas GOs.
type CancellationStatementBiogas struct {
	BCancellationkey       string  `json:"bCancellationkey"`
	CancellationTime       int64   `json:"CancellationTime"`
	OwnerID                string  `json:"OwnerID"`
	VolumeNm3             float64 `json:"VolumeNm3"`
	EnergyContentMWh      float64 `json:"EnergyContentMWh"`
	Emissions             float64 `json:"Emissions"`
	BiogasProductionMethod string  `json:"BiogasProductionMethod"`
	FeedstockType         string  `json:"FeedstockType"`
}
