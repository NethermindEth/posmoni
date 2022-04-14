package networking

type Subscriber interface {
	Listen(url string, ch chan<- Checkpoint)
}

type BeaconAPI interface {
	SetEndpoints(endpoints []string)
	ValidatorBalances(stateID string, validatorIdxs []string) ([]ValidatorBalance, error)
}
