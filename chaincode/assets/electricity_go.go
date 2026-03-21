// Package assets defines all on-chain asset types for the GO lifecycle platform.
package assets

// ElectricityGO is the public (world-state) representation of an electricity guarantee of origin.
type ElectricityGO struct {
	AssetID          string `json:"AssetID"`
	CreationDateTime int64  `json:"CreationDateTime"`
	GOType           string `json:"GOType"` // always "Electricity"
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
	DeviceID                    string   `json:"DeviceID,omitempty"` // NEW: links to registering device
}
