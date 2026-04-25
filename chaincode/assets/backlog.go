package assets

// ================================================================================
// v10.0: Universal Backlog System for All Energy Carriers
// ================================================================================
// Each energy carrier maintains a backlog that accumulates metering data before
// GO issuance. This enables batch processing and carrier-to-carrier conversions.

// CarrierBacklog is the public (world-state) marker for a pending GO backlog.
// The backlog key format: "{carrierType}_backlog_{ownerMSP}"
type CarrierBacklog struct {
	BacklogKey string `json:"BacklogKey"`
	CarrierType string `json:"CarrierType"` // "Electricity", "Hydrogen", "Biogas", "HeatingCooling"
	OwnerMSP    string `json:"OwnerMSP"`
}

// ElectricityBacklogPrivateDetails holds confidential electricity metering data awaiting GO issuance.
type ElectricityBacklogPrivateDetails struct {
	BacklogKey                  string  `json:"BacklogKey"`
	OwnerMSP                    string  `json:"OwnerMSP"`
	AccumulatedMWh              float64 `json:"AccumulatedMWh"`
	AccumulatedEmissions        float64 `json:"AccumulatedEmissions"`
	ElectricityProductionMethod string  `json:"ElectricityProductionMethod"`
	DeviceID                    string  `json:"DeviceID,omitempty"`
	FirstMeteringTimestamp      int64   `json:"FirstMeteringTimestamp"` // earliest data point
	LastMeteringTimestamp       int64   `json:"LastMeteringTimestamp"`  // most recent data point
}

// HydrogenBacklogPrivateDetails holds confidential hydrogen production data awaiting GO issuance.
// This replaces the v9 GreenHydrogenGOBacklogPrivateDetails with a more generic structure.
type HydrogenBacklogPrivateDetails struct {
	BacklogKey               string  `json:"BacklogKey"`
	OwnerMSP                 string  `json:"OwnerMSP"`
	AccumulatedKilosProduced float64 `json:"AccumulatedKilosProduced"`
	AccumulatedEmissions     float64 `json:"AccumulatedEmissions"`
	HydrogenProductionMethod string  `json:"HydrogenProductionMethod"`
	AccumulatedInputMWh      float64 `json:"AccumulatedInputMWh"` // electricity used (for e→h conversion)
	DeviceID                 string  `json:"DeviceID,omitempty"`
	FirstMeteringTimestamp   int64   `json:"FirstMeteringTimestamp"`
	LastMeteringTimestamp    int64   `json:"LastMeteringTimestamp"`
}

// BiogasBacklogPrivateDetails holds confidential biogas production data awaiting GO issuance.
type BiogasBacklogPrivateDetails struct {
	BacklogKey                 string  `json:"BacklogKey"`
	OwnerMSP                   string  `json:"OwnerMSP"`
	AccumulatedVolumeNm3       float64 `json:"AccumulatedVolumeNm3"`
	AccumulatedEnergyContentMWh float64 `json:"AccumulatedEnergyContentMWh"`
	AccumulatedEmissions       float64 `json:"AccumulatedEmissions"`
	BiogasProductionMethod     string  `json:"BiogasProductionMethod"`
	FeedstockType              string  `json:"FeedstockType"`
	DeviceID                   string  `json:"DeviceID,omitempty"`
	FirstMeteringTimestamp     int64   `json:"FirstMeteringTimestamp"`
	LastMeteringTimestamp      int64   `json:"LastMeteringTimestamp"`
}

// HeatingCoolingBacklogPrivateDetails holds confidential heating/cooling production data awaiting GO issuance.
type HeatingCoolingBacklogPrivateDetails struct {
	BacklogKey                        string  `json:"BacklogKey"`
	OwnerMSP                          string  `json:"OwnerMSP"`
	AccumulatedAmountMWh              float64 `json:"AccumulatedAmountMWh"`
	AccumulatedEmissions              float64 `json:"AccumulatedEmissions"`
	HeatingCoolingProductionMethod    string  `json:"HeatingCoolingProductionMethod"`
	AverageSupplyTemperature          float64 `json:"AverageSupplyTemperature,omitempty"` // weighted average
	DeviceID                          string  `json:"DeviceID,omitempty"`
	FirstMeteringTimestamp            int64   `json:"FirstMeteringTimestamp"`
	LastMeteringTimestamp             int64   `json:"LastMeteringTimestamp"`
}

// Backlog key prefixes for each carrier type
const (
	BacklogKeyElectricity    = "electricity_backlog"
	BacklogKeyHydrogen       = "hydrogen_backlog"
	BacklogKeyBiogas         = "biogas_backlog"
	BacklogKeyHeatingCooling = "heating_cooling_backlog"
)
