package assets

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Counter key constants for on-chain persistent counters.
// Bug fix #1: Replaces in-memory sync.Mutex counters that reset on peer restart.
// Bug fix #2: Eliminates race condition — Fabric deterministic execution guarantees
// single-writer semantics per key within a transaction.
const (
	CounterKeyEGO          = "counter_eGO"
	CounterKeyHGO          = "counter_hGO"
	CounterKeyECancellation = "counter_eCancellation"
	CounterKeyHCancellation = "counter_hCancellation"
	CounterKeyEConsumption  = "counter_eConsumption"
	CounterKeyHConsumption  = "counter_hConsumption"
	CounterKeyDevice        = "counter_device"
)

// GetNextID atomically reads the current counter from the ledger, increments it,
// writes it back, and returns the new value. This is safe because Fabric executes
// transactions deterministically — there are no concurrent writes to the same key
// within a single transaction.
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
