package assets

// GreenHydrogenGO is the public (world-state) representation of a hydrogen guarantee of origin.
type GreenHydrogenGO struct {
	AssetID          string `json:"AssetID"`
	CreationDateTime int64  `json:"CreationDateTime"`
	GOType           string `json:"GOType"` // always "Hydrogen"
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
