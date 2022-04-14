package networking

import (
	"github.com/NethermindEth/posgonitor/configs"
	sse "github.com/r3labs/sse/v2"
	log "github.com/sirupsen/logrus"
)

type SSESubscriber struct {
}

func (s SSESubscriber) Listen(url string, ch chan<- Checkpoint) {
	// notest
	log.Info("Subscribing to: ", url)

	client := sse.NewClient(url)
	client.SubscribeRaw(func(msg *sse.Event) {
		if len(msg.Data) == 0 {
			log.WithField(configs.Component, "ETH2").Debug("Got empty event")
			return
		}

		log.WithField(configs.Component, "ETH2").Infof("Got event data: %v", string(msg.Data))

		chkp, err := unmarshalData(msg.Data, Checkpoint{})
		if err != nil {
			log.WithField(configs.Component, "ETH2").Errorf(parseDataError, err)
		} else {
			ch <- chkp
		}
	})
}

func Subscribe(done <-chan struct{}, sub SubscribeOpts) <-chan Checkpoint {
	c := make(chan Checkpoint)

	go func() {
		//TODO: Add support for multiple endpoints. This only works well for one endpoint
		for _, endpoint := range sub.Endpoints {
			url := endpoint + sub.StreamURL
			go sub.Subscriber.Listen(url, c)
		}

		<-done
		log.Info("Subscription to ", sub.StreamURL, " ended")
		close(c)
	}()

	return c
}
