// Package assets defines all on-chain asset types for the GO lifecycle platform.
package assets

// ElectricityGO is the public (world-state) representation of an electricity guarantee of origin.
// ADR-012: Fields align with CEN-EN 16325 GO data attributes.
type ElectricityGO struct {
	AssetID              string `json:"AssetID"`
	CreationDateTime     int64  `json:"CreationDateTime"`
	GOType               string `json:"GOType"`             // always "Electricity"
	Status               string `json:"Status"`             // ADR-007: "active", "cancelled", "transferred"
	QuantityCommitment   string `json:"QuantityCommitment"` // ADR-009: SHA-256(AmountMWh || salt) for selective disclosure
	CountryOfOrigin      string `json:"CountryOfOrigin,omitempty" metadata:",optional"`      // ADR-012: ISO 3166-1 alpha-2 (e.g. "DE", "NL")
	GridConnectionPoint  string `json:"GridConnectionPoint,omitempty" metadata:",optional"`  // ADR-012: EIC code of the grid connection
	SupportScheme        string `json:"SupportScheme,omitempty" metadata:",optional"`        // ADR-012: "none", "FIT", "FIP", "quota" etc.
	EnergySource         string `json:"EnergySource,omitempty" metadata:",optional"`         // ADR-012: EN 16325 source code (e.g. "F01010100" = solar)
	ProductionPeriodStart int64  `json:"ProductionPeriodStart,omitempty" metadata:",optional"` // ADR-012: UNIX timestamp
	ProductionPeriodEnd   int64  `json:"ProductionPeriodEnd,omitempty" metadata:",optional"`   // ADR-012: UNIX timestamp
}

// ElectricityGOPrivateDetails holds the confidential fields stored in a private data collection.
type ElectricityGOPrivateDetails struct {
	AssetID                     string   `json:"AssetID"`
	OwnerID                     string   `json:"OwnerID"`
	CreationDateTime            int64    `json:"CreationDateTime"`
	AmountMWh                   float64  `json:"AmountMWh"`
	Emissions                   float64  `json:"Emissions"`
	ElectricityProductionMethod string   `json:"ElectricityProductionMethod"`
	ConsumptionDeclarations     []string `json:"ConsumptionDeclarations"`
	DeviceID                    string   `json:"DeviceID,omitempty"`
	CommitmentSalt              string   `json:"CommitmentSalt,omitempty"` // ADR-009: salt for quantity commitment
}
