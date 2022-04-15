package eth2

// eth2Config : Struct Represent monitor configuration data
type eth2Config struct {
	// List of validator addresses or public index to monitor
	Validators []string
	// List of consensus nodes from which to interact with Beacon chain API
	Consensus []string
}
