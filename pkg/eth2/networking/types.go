package networking

// Checkpoint : Struct Represent event data from beacon chain
type Checkpoint struct {
	Block string `json:"block"`
	State string `json:"state"`
	Epoch string `json:"epoch"`
}

// SubscribeOpts : Struct Represent subscription data and handlers
type SubscribeOpts struct {
	// Endpoints exposing beacon chain API
	Endpoints []string
	// URL and topic to subscribe to within an endpoint
	StreamURL string
	// Interface with Listen implementation to subscribe to beacon chain events
	Subscriber Subscriber
}

// ValidatorBalanceList : Struct Represent response data from 'http://<endpoint>/eth/v1/beacon/states/<stateID>/validator_balances' API call
type ValidatorBalanceList struct {
	Data []ValidatorBalance `json:"data"`
}

// ValidatorBalance : Struct Represent a single entry of response data from 'http://<endpoint>/eth/v1/beacon/states/<stateID>/validator_balances' API call
type ValidatorBalance struct {
	Index   string `json:"index"`
	Balance string `json:"balance"`
}

// HealthResponse : Struct Represent response information from 'http://<endpoint>/eth/v1/beacon/health' API call
type HealthResponse struct {
	Endpoint string
	Healthy  bool
	Error    error
}

// SyncingStatusResponse : Struct Represent response body from 'http://<endpoint>/eth/v1/node/syncing' API call
type SyncingStatusResponse struct {
	Data SyncingStatus `json:"data"`
}

// SyncingStatus : Struct Represent response data from 'http://<endpoint>/eth/v1/node/syncing' API call
type SyncingStatus struct {
	HeadSlot     string `json:"head_slot"`
	SyncDistance string `json:"sync_distance"`
	IsSyncing    bool   `json:"is_syncing"`
	Error        error
}
