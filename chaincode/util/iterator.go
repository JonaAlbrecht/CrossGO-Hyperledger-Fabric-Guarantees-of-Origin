// Package util provides shared helper functions for the GO chaincode contracts.
package util

import (
	"encoding/json"

	"github.com/JonaAlbrecht/HLF-GOconversionissuance-JA-MA/chaincode/assets"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// ConstructEGOsFromIterator reads all ElectricityGO entries from a state query iterator.
func ConstructEGOsFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*assets.ElectricityGO, error) {
	var eGOs []*assets.ElectricityGO
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var eGO assets.ElectricityGO
		err = json.Unmarshal(queryResult.Value, &eGO)
		if err != nil {
			return nil, err
		}
		eGOs = append(eGOs, &eGO)
	}
	return eGOs, nil
}

// ConstructHGOsFromIterator reads all GreenHydrogenGO entries from a state query iterator.
func ConstructHGOsFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*assets.GreenHydrogenGO, error) {
	var hGOs []*assets.GreenHydrogenGO
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var hGO assets.GreenHydrogenGO
		err = json.Unmarshal(queryResult.Value, &hGO)
		if err != nil {
			return nil, err
		}
		hGOs = append(hGOs, &hGO)
	}
	return hGOs, nil
}
