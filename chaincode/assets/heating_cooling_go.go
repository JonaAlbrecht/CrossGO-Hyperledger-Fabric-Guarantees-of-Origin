package assets

// HeatingCoolingGO is the public (world-state) representation of a heating/cooling guarantee of origin.
// ADR-015 extension: Extends the multi-carrier model with heating and cooling support per RED III.
// ADR-012: Fields align with CEN-EN 16325 GO data attributes.
type HeatingCoolingGO struct {
	AssetID               string `json:"AssetID"`
	CreationDateTime      int64  `json:"CreationDateTime"`
	GOType                string `json:"GOType"`                         // always "HeatingCooling"
	Status                string `json:"Status"`                         // ADR-007: "active", "cancelled", "transferred"
	QuantityCommitment    string `json:"QuantityCommitment"`             // ADR-009: SHA-256(AmountMWh || salt)
	CountryOfOrigin       string `json:"CountryOfOrigin,omitempty" metadata:",optional"`      // ADR-012: ISO 3166-1 alpha-2
	GridConnectionPoint   string `json:"GridConnectionPoint,omitempty" metadata:",optional"`  // ADR-012: EIC code
	SupportScheme         string `json:"SupportScheme,omitempty" metadata:",optional"`        // ADR-012: "none", "FIT", "FIP", "quota"
	EnergySource          string `json:"EnergySource,omitempty" metadata:",optional"`         // ADR-012: EN 16325 source code
	ProductionPeriodStart int64  `json:"ProductionPeriodStart,omitempty" metadata:",optional"` // ADR-012: UNIX timestamp
	ProductionPeriodEnd   int64  `json:"ProductionPeriodEnd,omitempty" metadata:",optional"`   // ADR-012: UNIX timestamp
}

// HeatingCoolingGOPrivateDetails holds the confidential fields for a heating/cooling GO.
type HeatingCoolingGOPrivateDetails struct {
	AssetID                       string   `json:"AssetID"`
	OwnerID                       string   `json:"OwnerID"`
	CreationDateTime              int64    `json:"CreationDateTime"`
	AmountMWh                     float64  `json:"AmountMWh"`                     // Thermal energy in MWh
	Emissions                     float64  `json:"Emissions"`                     // gCO2eq
	HeatingCoolingProductionMethod string  `json:"HeatingCoolingProductionMethod"` // "heat_pump", "solar_thermal", "geothermal", "biomass_boiler", "district_heating", "absorption_chiller"
	SupplyTemperature             float64  `json:"SupplyTemperature,omitempty"`   // °C — distinguishes heating vs cooling
	ConsumptionDeclarations       []string `json:"ConsumptionDeclarations"`
	DeviceID                      string   `json:"DeviceID,omitempty"`
	CommitmentSalt                string   `json:"CommitmentSalt,omitempty"` // ADR-009
}

// CancellationStatementHeatingCooling records the cancellation of one or more heating/cooling GOs.
type CancellationStatementHeatingCooling struct {
	HCCancellationKey              string  `json:"hcCancellationKey"`
	CancellationTime               int64   `json:"CancellationTime"`
	OwnerID                        string  `json:"OwnerID"`
	AmountMWh                      float64 `json:"AmountMWh"`
	Emissions                      float64 `json:"Emissions"`
	HeatingCoolingProductionMethod string  `json:"HeatingCoolingProductionMethod"`
}
