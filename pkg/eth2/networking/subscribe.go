package networking

import (
	"net/http"

	"github.com/NethermindEth/posmoni/configs"
	"github.com/r3labs/sse/v2"
	log "github.com/sirupsen/logrus"
)

// SSESubscriber : Struct Subscriber interface implementation
type SSESubscriber struct {
}

/*
Listen :
Subscribe to beacon chain SSE events and listen for new beacon chain checkpoints.

params :-
a. url string
URL to subscribe to
b. ch chan<- Checkpoint
Channel to send new checkpoints to

returns :-
none
*/
func (s SSESubscriber) Listen(url string, ch chan<- Checkpoint) {
	// notest
	logFields := log.Fields{configs.Component: "SSESubscriber", "Method": "Listen"}
	log.WithFields(logFields).Info("Subscribing to: ", url)

	// smoke test
	_, err := http.Get(url)
	if err != nil {
		log.WithFields(logFields).Errorf(RequestFailedError, url, err)
	}

	client := sse.NewClient(url)
	err = client.SubscribeRaw(func(msg *sse.Event) {
		if len(msg.Data) == 0 {
			log.WithFields(logFields).Debug("Got empty event")
			return
		}

		log.WithFields(logFields).Infof("Got event data: %v", string(msg.Data))

		chkp, err := unmarshalData(msg.Data, Checkpoint{})
		if err != nil {
			log.WithFields(logFields).Errorf(parseDataError, err)
		} else {
			ch <- chkp
		}
	})
	if err != nil {
		log.WithFields(logFields).Errorf(SSESubscribeError, err)
	}
}

/*
Subscribe :
Setup subscriptions to beacon chain events using several beacon node endpoints.

params :-
a. done <- chan struct{}
Channel to get stop listening signal from
b. sub SubscribeOpts
Subscription data and handlers

returns :-
a. <-chan Checkpoint
Channel to get new checkpoints from
*/
func Subscribe(done <-chan struct{}, sub SubscribeOpts) <-chan Checkpoint {
	logFields := log.Fields{"Method": "Subscribe"}
	c := make(chan Checkpoint)

	go func() {
		//TODO: Add support for multiple endpoints. This only works well for one endpoint. Probably consistency checks are needed.
		for _, endpoint := range sub.Endpoints {
			url := endpoint + sub.StreamURL
			go sub.Subscriber.Listen(url, c)
		}

		<-done
		log.WithFields(logFields).Info("Subscription to ", sub.StreamURL, " ended")
		close(c)
	}()

	return c
}
