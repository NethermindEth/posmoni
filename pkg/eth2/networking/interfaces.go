package networking

// Subscriber : Interface Represents a subscriber for a given topic
type Subscriber interface {
	Listen(url string, ch chan<- Checkpoint)
}

// BeaconAPI : Interface for Beacon chain HTTP API
type BeaconAPI interface {
	SetEndpoints(endpoints []string)
	ValidatorBalances(stateID string, validatorIdxs []string) ([]ValidatorBalance, error)
	Health(endpoints []string) []HealthResponse
	SyncStatus(endpoints []string) []SyncingStatus
}
