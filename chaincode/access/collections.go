package access

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Collection name prefix — all private data collections follow this convention.
const collectionPrefix = "privateDetails-"

// Public collection name.
const PublicCollection = "publicGOcollection"

// GetOwnCollection returns the private data collection for the caller's organization.
// Bug fix #7: Always derives from GetMSPID(), never from the "organization" X.509 attribute.
func GetOwnCollection(ctx contractapi.TransactionContextInterface) (string, error) {
	mspID, err := GetClientMSPID(ctx)
	if err != nil {
		return "", err
	}
	return collectionPrefix + mspID, nil
}

// GetCollectionForOrg returns the private data collection for a specific organization MSP ID.
func GetCollectionForOrg(orgMSP string) string {
	return collectionPrefix + orgMSP
}

// ValidateCollectionAccess checks that the caller has permission to read from the given collection.
// Issuers can read any collection; producers and buyers can only read their own.
func ValidateCollectionAccess(ctx contractapi.TransactionContextInterface, collection string) error {
	mspID, err := GetClientMSPID(ctx)
	if err != nil {
		return err
	}
	ownCollection := collectionPrefix + mspID

	// If accessing own collection, always allowed
	if collection == ownCollection {
		return nil
	}

	// Issuers can access any collection
	isIssuer, err := IsIssuer(ctx)
	if err != nil {
		return err
	}
	if isIssuer {
		return nil
	}

	return fmt.Errorf("access denied: %s cannot read collection %s", mspID, collection)
}
