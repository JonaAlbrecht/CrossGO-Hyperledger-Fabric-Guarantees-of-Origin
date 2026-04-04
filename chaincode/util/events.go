package util

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// LifecycleEventType enumerates the lifecycle stages that emit chaincode events.
// ADR-016: Each event is consumed by an off-chain CQRS indexer to build
// query-optimised projections without CouchDB rich-query overhead.
type LifecycleEventType string

const (
	EventGOCreated    LifecycleEventType = "GO_CREATED"
	EventGOTransferred LifecycleEventType = "GO_TRANSFERRED"
	EventGOCancelled  LifecycleEventType = "GO_CANCELLED"
	EventGOConverted  LifecycleEventType = "GO_CONVERTED"
	EventGOSplit      LifecycleEventType = "GO_SPLIT"
	EventDeviceRegistered LifecycleEventType = "DEVICE_REGISTERED"
	EventDeviceRevoked    LifecycleEventType = "DEVICE_REVOKED"
)

// LifecycleEvent is the payload attached to a Fabric chaincode event.
// Off-chain listeners (e.g. an event-sourced read-model) consume these events
// to build query-optimised projections.
type LifecycleEvent struct {
	EventType LifecycleEventType `json:"eventType"`
	AssetID   string             `json:"assetId"`
	GOType    string             `json:"goType,omitempty"`    // "Electricity", "Hydrogen", "Biogas"
	Initiator string             `json:"initiator,omitempty"` // MSP ID of the transaction submitter
	TxID      string             `json:"txId"`
	Timestamp int64              `json:"timestamp,omitempty"`
	Details   map[string]string  `json:"details,omitempty"` // arbitrary key-value metadata
}

// EmitLifecycleEvent sets a Fabric chaincode event on the current transaction.
// The event name format is the EventType string (e.g. "GO_CREATED").
// Off-chain listeners filter on this event name and parse the JSON payload.
//
// Fabric allows only one event per transaction. If a transaction involves
// multiple logical events (e.g. a split producing two new GOs), the caller
// should combine them into a single event with appropriate Details.
func EmitLifecycleEvent(ctx contractapi.TransactionContextInterface, event LifecycleEvent) error {
	event.TxID = ctx.GetStub().GetTxID()
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal lifecycle event: %v", err)
	}
	return ctx.GetStub().SetEvent(string(event.EventType), payload)
}
