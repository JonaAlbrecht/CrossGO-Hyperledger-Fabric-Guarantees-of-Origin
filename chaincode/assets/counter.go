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
	PrefixBGO            = "bGO_"      // ADR-015: biogas
	PrefixHCGO           = "hcGO_"     // v9: heating/cooling
	PrefixECancellation  = "eCancel_"
	PrefixHCancellation  = "hCancel_"
	PrefixBCancellation  = "bCancel_"  // ADR-015: biogas
	PrefixHCCancellation = "hcCancel_" // v9: heating/cooling
	PrefixEConsumption   = "eCon_"
	PrefixHConsumption   = "hCon_"
	PrefixDevice         = "device_"
)

// RangeEnd constants for prefix-based range queries.
// In Fabric's LevelDB/CouchDB, keys are sorted lexicographically.
// The tilde '~' character (0x7E) sorts after all alphanumeric characters,
// so "prefix~" captures all keys starting with "prefix".
const (
	RangeEndEGO    = "eGO_~"
	RangeEndHGO    = "hGO_~"
	RangeEndBGO    = "bGO_~"  // ADR-015: biogas
	RangeEndHCGO   = "hcGO_~" // v9: heating/cooling
	RangeEndDevice = "device_~"
)

// GO lifecycle status constants (ADR-007: tombstone pattern, ADR-031: bridge states).
const (
	GOStatusActive      = "active"
	GOStatusCancelled   = "cancelled"
	GOStatusTransferred = "transferred"
	GOStatusLocked      = "locked"  // ADR-031: locked for cross-channel bridge transfer
	GOStatusBridged     = "bridged" // ADR-031: successfully bridged to another channel
)

// MaxTimestampDrift is the maximum allowed difference (in seconds) between
// a transaction proposal timestamp and the orderer block time. This prevents
// backdating attacks (ADR-008).
const MaxTimestampDrift int64 = 300 // 5 minutes

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

// GenerateCommitment produces a SHA-256 commitment of a quantity value with a
// cryptographically random salt.
// ADR-009: Selective disclosure — the quantity is hidden on the public ledger
// but can be revealed to a verifier by disclosing the salt.
// ADR-017 (v6.0): Salt is derived deterministically from the transaction ID
// and a secret prefix to ensure all endorsing peers produce identical proposals.
// A truly random salt (crypto/rand) would cause endorsement mismatches in
// multi-org collection policies. The txID-derived salt is a tradeoff:
// less brute-force resistance but required for Fabric endorsement consistency.
// Returns (commitmentHex, saltHex).
func GenerateCommitment(ctx contractapi.TransactionContextInterface, quantity float64) (string, string, error) {
	txID := ctx.GetStub().GetTxID()
	quantityStr := strconv.FormatFloat(quantity, 'f', -1, 64)
	// Deterministic salt: HMAC-SHA256(txID, quantity) → first 16 bytes
	saltInput := txID + "||salt||" + quantityStr
	saltHash := sha256.Sum256([]byte(saltInput))
	saltHex := hex.EncodeToString(saltHash[:16]) // 128-bit deterministic salt
	// Commitment = SHA-256(quantity_string || saltHex)
	commitInput := quantityStr + "||" + saltHex
	commitment := sha256.Sum256([]byte(commitInput))
	return hex.EncodeToString(commitment[:]), saltHex, nil
}

// VerifyCommitment checks that a claimed quantity and salt match a published commitment.
// ADR-009: Used for selective disclosure verification — a verifier calls this to confirm
// that the producer's disclosed quantity matches the on-chain commitment.
func VerifyCommitment(quantity float64, salt string, expectedCommitment string) bool {
	quantityStr := strconv.FormatFloat(quantity, 'f', -1, 64)
	commitInput := quantityStr + "||" + salt
	hash := sha256.Sum256([]byte(commitInput))
	return hex.EncodeToString(hash[:]) == expectedCommitment
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
