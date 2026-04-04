package assets

// GreenHydrogenGO is the public (world-state) representation of a hydrogen guarantee of origin.
// ADR-012: Fields align with CEN-EN 16325 GO data attributes.
type GreenHydrogenGO struct {
	AssetID              string `json:"AssetID"`
	CreationDateTime     int64  `json:"CreationDateTime"`
	GOType               string `json:"GOType"`             // always "Hydrogen"
	Status               string `json:"Status"`             // ADR-007: "active", "cancelled", "transferred"
	QuantityCommitment   string `json:"QuantityCommitment"` // ADR-009: SHA-256(Kilosproduced || salt)
	CountryOfOrigin      string `json:"CountryOfOrigin,omitempty"`      // ADR-012: ISO 3166-1 alpha-2
	GridConnectionPoint  string `json:"GridConnectionPoint,omitempty"`  // ADR-012: EIC code
	SupportScheme        string `json:"SupportScheme,omitempty"`        // ADR-012: "none", "FIT", "FIP", "quota" etc.
	EnergySource         string `json:"EnergySource,omitempty"`         // ADR-012: EN 16325 source code
	ProductionPeriodStart int64  `json:"ProductionPeriodStart,omitempty"` // ADR-012: UNIX timestamp
	ProductionPeriodEnd   int64  `json:"ProductionPeriodEnd,omitempty"`   // ADR-012: UNIX timestamp
}

// GreenHydrogenGOPrivateDetails holds the confidential fields for a hydrogen GO.
type GreenHydrogenGOPrivateDetails struct {
	AssetID                     string   `json:"AssetID"`
	OwnerID                     string   `json:"OwnerID"`
	CreationDateTime            int64    `json:"CreationDateTime"`
	Kilosproduced               float64  `json:"Kilosproduced"`
	EmissionsHydrogen           float64  `json:"Emissions"`
	HydrogenProductionMethod    string   `json:"HydrogenProductionMethod"`
	InputEmissions              float64  `json:"InputEmissions"`
	UsedMWh                     float64  `json:"UsedMWh"`
	ElectricityProductionMethod []string `json:"ElectricityProductionMethod"`
	ConsumptionDeclarations     []string `json:"ConsumptionDeclarations"`
	DeviceID                    string   `json:"DeviceID,omitempty"`
	CommitmentSalt              string   `json:"CommitmentSalt,omitempty"` // ADR-009
}

// GreenHydrogenGOBacklog is the public marker for a pending hydrogen conversion backlog.
type GreenHydrogenGOBacklog struct {
	Backlogkey string `json:"Backlogkey"`
	GOType     string `json:"GOType"` // always "Hydrogen"
}

// GreenHydrogenGOBacklogPrivateDetails holds the confidential backlog data.
type GreenHydrogenGOBacklogPrivateDetails struct {
	Backlogkey               string  `json:"Backlogkey"`
	OwnerID                  string  `json:"OwnerID"`
	Kilosproduced            float64 `json:"Kilosproduced"`
	EmissionsHydrogen        float64 `json:"Emissions"`
	HydrogenProductionMethod string  `json:"HydrogenProductionMethod"`
	UsedMWh                  float64 `json:"UsedMWh"`
}
