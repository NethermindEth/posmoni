package eth2

import (
	"github.com/NethermindEth/posgonitor/configs"
	sse "github.com/r3labs/sse/v2"
	log "github.com/sirupsen/logrus"
)

type sseSubscriber struct {
}

func (s sseSubscriber) listen(url string, ch chan<- checkpoint) {
	// notest
	log.Info("Subscribing to: ", url)

	client := sse.NewClient(url)
	client.SubscribeRaw(func(msg *sse.Event) {
		if len(msg.Data) == 0 {
			log.WithField(configs.Component, "ETH2").Debug("Got empty event")
			return
		}

		log.WithField(configs.Component, "ETH2").Infof("Got event data: %v", string(msg.Data))

		chkp, err := parseEventData(msg.Data)
		if err != nil {
			log.WithField(configs.Component, "ETH2").Errorf(ParseDataError, err)
		} else {
			ch <- chkp
		}
	})
}

func subscribe(done <-chan struct{}, sub SubscribeOpts) <-chan checkpoint {
	c := make(chan checkpoint)

	go func() {
		//TODO: Add support for multiple endpoints. This only works well for one endpoint
		for _, endpoint := range sub.endpoints {
			url := endpoint + sub.streamURL
			go sub.subscriber.listen(url, c)
		}

		<-done
		close(c)
	}()

	return c
}
