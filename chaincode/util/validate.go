package util

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Bug fix #6: All input validation is centralized here to prevent panics from nil/zero/empty values.

// GetTransientBytes extracts a named key from the transient map and validates it is non-empty.
func GetTransientBytes(ctx contractapi.TransactionContextInterface, key string) ([]byte, error) {
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return nil, fmt.Errorf("error getting transient: %v", err)
	}
	data, ok := transientMap[key]
	if !ok {
		return nil, fmt.Errorf("%q must be a key in the transient map", key)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("%q value in transient map must be non-empty", key)
	}
	return data, nil
}

// UnmarshalTransient extracts a transient map key and unmarshals it into the target struct.
func UnmarshalTransient(ctx contractapi.TransactionContextInterface, key string, target interface{}) error {
	data, err := GetTransientBytes(ctx, key)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("failed to decode JSON from transient key %q: %v", key, err)
	}
	return nil
}

// ValidatePositive returns an error if any of the named float values are not positive.
func ValidatePositive(values map[string]float64) error {
	for name, val := range values {
		if val <= 0 {
			return fmt.Errorf("%s must be a positive number, got %f", name, val)
		}
	}
	return nil
}

// ValidateNonEmpty returns an error if the string is empty.
func ValidateNonEmpty(name, value string) error {
	if value == "" {
		return fmt.Errorf("%s must be a non-empty string", name)
	}
	return nil
}

// GetTimestamp returns the current transaction timestamp in seconds.
func GetTimestamp(ctx contractapi.TransactionContextInterface) (int64, error) {
	ts, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return 0, fmt.Errorf("error getting timestamp: %v", err)
	}
	return ts.GetSeconds(), nil
}
