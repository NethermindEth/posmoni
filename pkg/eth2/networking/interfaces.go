package networking

import "encoding/json"

// Subscriber : Interface Represents a subscriber for a given topic
type Subscriber interface {
	Listen(url string, ch chan<- Checkpoint)
}

// BeaconAPI : Interface for Beacon chain HTTP API
type BeaconAPI interface {
	SetEndpoints(endpoints []string)
	ValidatorBalances(stateID string, validatorIdxs []string) ([]ValidatorBalance, error)
	Health(endpoints []string) []HealthResponse
	SyncStatus(endpoints []string) []BeaconSyncingStatus
}

// ExecutionAPI : Interface for ETH1 JSON RPC API
type ExecutionAPI interface {
	Call(endpoint, method string, params ...any) (json.RawMessage, error)
	SyncStatus(endpoints []string) []ExecutionSyncingStatus
}
