package networking

type Subscriber interface {
	listen(url string, ch chan<- Checkpoint)
}

type BeaconAPI interface {
	ValidatorBalances(stateID string, validatorIdxs []string) ([]ValidatorBalance, error)
}
