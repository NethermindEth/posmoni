package eth2

// Eth2Config : Struct Represent monitor configuration data
type eth2Config struct {
	// List of validator addresses or public index to monitor
	validators []string
	// List of consensus nodes from which to interact with Beacon chain API
	consensus []string
	// List of execution nodes from which to interact with Ethereum json-rpc API
	execution []string
}

// ConfigOpts : Struct Represent monitor setup options
type ConfigOpts struct {
	// True if configuration setup (configuration file setup or enviroment variables setup) should be handled
	HandleCfg bool
	// True if logging configuration should be handled
	handleLogs bool
	// Handle how configuration data should be loaded
	Checkers []CfgChecker
}

// EndpointSyncStatus : Struct Represent sync status of an endpoint
type EndpointSyncStatus struct {
	Endpoint string
	Synced   bool
	Error    error
}
