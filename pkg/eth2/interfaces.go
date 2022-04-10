package eth2

type Subscriber interface {
	listen(url string, ch chan<- checkpoint)
}
