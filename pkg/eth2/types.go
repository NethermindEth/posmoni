package eth2

type eth2Config struct {
	Validators []string
	Consensus  []string
}

type checkpoint struct {
	Block string `json:"block"`
	State string `json:"state"`
	Epoch string `json:"epoch"`
}
