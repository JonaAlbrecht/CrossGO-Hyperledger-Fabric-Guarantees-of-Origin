package assets

// CancellationStatementElectricity is an electronic, non-transferrable receipt providing
// evidence of the cancellation of one or more electricity GOs for disclosure purposes (EN16325 4.9.2).
type CancellationStatementElectricity struct {
	ECancellationkey            string   `json:"eCancellationkey"`
	CancellationTime            int64    `json:"CancellationTime"`
	OwnerID                     string   `json:"OwnerID"`
	AmountMWh                   float64  `json:"AmountMWh"`
	Emissions                   float64  `json:"Emissions"`
	ElectricityProductionMethod string   `json:"ElectricityProductionMethod"`
	ConsumptionDeclarations     []string `json:"ConsumptionDeclarations"`
}

// CancellationStatementHydrogen is the hydrogen equivalent of a cancellation statement.
type CancellationStatementHydrogen struct {
	HCancellationkey            string   `json:"hCancellationkey"`
	CancellationTime            int64    `json:"CancellationTime"`
	OwnerID                     string   `json:"OwnerID"`
	Kilosproduced               float64  `json:"Kilosproduced"`
	EmissionsHydrogen           float64  `json:"Emissions"`
	HydrogenProductionMethod    string   `json:"HydrogenProductionMethod"`
	InputEmissions              float64  `json:"InputEmissions"`
	ElectricityProductionMethod []string `json:"ElectricityProductionMethod"`
	UsedMWh                     float64  `json:"UsedMWh"`
	ConsumptionDeclarations     []string `json:"ConsumptionDeclarations"`
}

// ConsumptionDeclarationElectricity records the amount of energy input per carrier during
// a period and the identity and relevant attributes of the GOs cancelled (EN16325 4.5.5.1.2 c).
type ConsumptionDeclarationElectricity struct {
	Consumptionkey              string   `json:"Consumptionkey"`
	CancelledGOID               string   `json:"CancelledGOID"`
	ConsumptionDateTime         int64    `json:"ConsumptionDateTime"`
	AmountMWh                   float64  `json:"AmountMWh"`
	Emissions                   float64  `json:"Emissions"`
	ElectricityProductionMethod string   `json:"ElectricityProductionMethod"`
	ConsumptionDeclarations     []string `json:"ConsumptionDeclarations"`
}

// ConsumptionDeclarationHydrogen records hydrogen energy consumption.
// Bug fix #10: ConsumptionDateTime unified to int64 (was string in original).
type ConsumptionDeclarationHydrogen struct {
	Consumptionkey           string   `json:"Consumptionkey"`
	CancelledGOID            string   `json:"CancelledGOID"`
	ConsumptionDateTime      int64    `json:"ConsumptionDateTime"` // fixed: was string
	Kilosproduced            float64  `json:"Kilosproduced"`
	EmissionsHydrogen        float64  `json:"Emissions"`
	HydrogenProductionMethod string   `json:"HydrogenProductionMethod"`
	ConsumptionDeclarations  []string `json:"ConsumptionDeclarations"`
}
