package access

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ABAC attribute helpers — extract X.509 certificate attributes from the client identity.
// These are used during device-based creation (SmartMeter / OutputMeter) as a secondary
// check alongside the on-chain Device registry.

// GetAttribute reads a named attribute from the client's X.509 certificate.
func GetAttribute(ctx contractapi.TransactionContextInterface, attrName string) (string, error) {
	val, found, err := ctx.GetClientIdentity().GetAttributeValue(attrName)
	if err != nil {
		return "", fmt.Errorf("failed to read attribute %q: %v", attrName, err)
	}
	if !found {
		return "", fmt.Errorf("certificate does not have attribute %q", attrName)
	}
	return val, nil
}

// AssertAttribute checks that a certificate attribute equals an expected value.
func AssertAttribute(ctx contractapi.TransactionContextInterface, attrName, expected string) error {
	return ctx.GetClientIdentity().AssertAttributeValue(attrName, expected)
}

// GetClientMSPID returns the MSP ID of the submitting client.
// Bug fix #7: Use GetMSPID() exclusively (not the "organization" attribute) for collection names.
func GetClientMSPID(ctx contractapi.TransactionContextInterface) (string, error) {
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get client MSP ID: %v", err)
	}
	return mspID, nil
}
