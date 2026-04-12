// Package access provides role-based and attribute-based access control for the GO platform.
package access

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Organization role constants.
const (
	RoleIssuer   = "issuer"
	RoleProducer = "producer"
	RoleBuyer    = "buyer"
)

// OrgRole state key prefix for the on-chain org→role registry.
const orgRolePrefix = "orgRole_"

// GetOrgRole looks up the role for the caller's MSP from the on-chain org registry.
func GetOrgRole(ctx contractapi.TransactionContextInterface) (string, error) {
	mspID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get MSPID: %v", err)
	}
	role, err := ctx.GetStub().GetState(orgRolePrefix + mspID)
	if err != nil {
		return "", fmt.Errorf("failed to read org role for %s: %v", mspID, err)
	}
	if role == nil {
		return "", fmt.Errorf("organization %s is not registered in the network", mspID)
	}
	return string(role), nil
}

// RegisterOrgRole writes an org→role mapping to the ledger. Only issuers may call this.
func RegisterOrgRole(ctx contractapi.TransactionContextInterface, mspID, role string) error {
	if role != RoleIssuer && role != RoleProducer && role != RoleBuyer {
		return fmt.Errorf("invalid role %q: must be issuer, producer, or buyer", role)
	}
	return ctx.GetStub().PutState(orgRolePrefix+mspID, []byte(role))
}

// RequireRole asserts that the caller's organization has the specified role.
func RequireRole(ctx contractapi.TransactionContextInterface, requiredRole string) error {
	role, err := GetOrgRole(ctx)
	if err != nil {
		return err
	}
	if role != requiredRole {
		return fmt.Errorf("access denied: requires role %s, caller has role %s", requiredRole, role)
	}
	return nil
}

// RequireAnyRole asserts that the caller's organization has one of the specified roles.
func RequireAnyRole(ctx contractapi.TransactionContextInterface, roles ...string) error {
	role, err := GetOrgRole(ctx)
	if err != nil {
		return err
	}
	for _, r := range roles {
		if role == r {
			return nil
		}
	}
	return fmt.Errorf("access denied: requires one of %v, caller has role %s", roles, role)
}

// IsIssuer returns true if the caller's organization is an issuer.
func IsIssuer(ctx contractapi.TransactionContextInterface) (bool, error) {
	role, err := GetOrgRole(ctx)
	if err != nil {
		return false, err
	}
	return role == RoleIssuer, nil
}

// IsProducer returns true if the caller's organization is a producer.
func IsProducer(ctx contractapi.TransactionContextInterface) (bool, error) {
	role, err := GetOrgRole(ctx)
	if err != nil {
		return false, err
	}
	return role == RoleProducer, nil
}
