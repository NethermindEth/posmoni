package eth2

// eth2Config : Struct Represent monitor configuration data
type eth2Config struct {
	// List of validator addresses or public index to monitor
	Validators []string
	// List of consensus nodes from which to interact with Beacon chain API
	Consensus []string
}

// ConfigOpts : Struct Represent monitor setup options
type ConfigOpts struct {
	// True if configuration setup (configuration file setup or enviroment variables setup) should be handled
	HandleCfg bool
	// Configuration data. Should be used when is not desired to use config file or enviroment variables to get configuration data.
	Config *eth2Config
}
