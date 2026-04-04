package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Counter key constants for on-chain persistent counters.
// DEPRECATED for ID generation: counters are retained for backward-compatible
// reads but new IDs are generated via GenerateID (transaction-ID based).
const (
	CounterKeyEGO           = "counter_eGO"
	CounterKeyHGO           = "counter_hGO"
	CounterKeyECancellation = "counter_eCancellation"
	CounterKeyHCancellation = "counter_hCancellation"
	CounterKeyEConsumption  = "counter_eConsumption"
	CounterKeyHConsumption  = "counter_hConsumption"
	CounterKeyDevice        = "counter_device"
)

// ID prefix constants used by GenerateID and range queries.
const (
	PrefixEGO           = "eGO_"
	PrefixHGO           = "hGO_"
	PrefixECancellation = "eCancel_"
	PrefixHCancellation = "hCancel_"
	PrefixEConsumption  = "eCon_"
	PrefixHConsumption  = "hCon_"
	PrefixDevice        = "device_"
)

// RangeEnd constants for prefix-based range queries.
// In Fabric's LevelDB/CouchDB, keys are sorted lexicographically.
// The tilde '~' character (0x7E) sorts after all alphanumeric characters,
// so "prefix~" captures all keys starting with "prefix".
const (
	RangeEndEGO    = "eGO_~"
	RangeEndHGO    = "hGO_~"
	RangeEndDevice = "device_~"
)

// GenerateID creates a deterministic, unique asset ID from the transaction ID.
// The ID format is: <prefix><short_hash> where short_hash is the first 16 hex
// characters of SHA-256(txID + suffix). The suffix parameter distinguishes
// multiple IDs generated within the same transaction (e.g., splits creating
// both a taken part and a remainder).
//
// This eliminates the MVCC_READ_CONFLICT bottleneck of the old GetNextID
// pattern, because no shared counter key is read or written.
func GenerateID(ctx contractapi.TransactionContextInterface, prefix string, suffix int) (string, error) {
	txID := ctx.GetStub().GetTxID()
	if txID == "" {
		return "", fmt.Errorf("transaction ID is empty")
	}
	input := txID + "_" + strconv.Itoa(suffix)
	hash := sha256.Sum256([]byte(input))
	shortHash := hex.EncodeToString(hash[:8]) // 16 hex chars = 8 bytes
	return prefix + shortHash, nil
}

// GetNextID atomically reads the current counter from the ledger, increments it,
// writes it back, and returns the new value. This is safe because Fabric executes
// transactions deterministically — there are no concurrent writes to the same key
// within a single transaction.
//
// DEPRECATED: Use GenerateID instead. GetNextID creates an MVCC hot-spot when
// multiple transactions in the same block read-modify-write the same counter key.
// Retained for backward compatibility with legacy IDs (eGO1, device1, etc.).
func GetNextID(ctx contractapi.TransactionContextInterface, counterKey string) (int, error) {
	data, err := ctx.GetStub().GetState(counterKey)
	if err != nil {
		return 0, fmt.Errorf("failed to read counter %s: %v", counterKey, err)
	}
	current := 0
	if data != nil {
		current, err = strconv.Atoi(string(data))
		if err != nil {
			return 0, fmt.Errorf("failed to parse counter %s: %v", counterKey, err)
		}
	}
	next := current + 1
	err = ctx.GetStub().PutState(counterKey, []byte(strconv.Itoa(next)))
	if err != nil {
		return 0, fmt.Errorf("failed to write counter %s: %v", counterKey, err)
	}
	return next, nil
}
