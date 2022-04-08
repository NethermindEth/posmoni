package eth2

import (
	"fmt"
	"os"

	"github.com/NethermindEth/posgonitor/configs"

	sse "github.com/r3labs/sse/v2"
	log "github.com/sirupsen/logrus"
)

func Monitor(handleCfg bool) {
	if handleCfg {
		configs.InitConfig()
	}

	cfg, err := Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	configs.InitLogging()

	log.Info(cfg)
	chkps := subscribe(cfg.Consensus)

	//TODO: Init and setup DB

	go getValidatorBalance(chkps)
	go setupAlerts(chkps)
}

func subscribe(endpoints []string) <-chan checkpoint {
	c := make(chan checkpoint)

	go func() {
		for _, endpoint := range endpoints {
			url := endpoint + "/eth/v1/events?topics=finalized_checkpoint"
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
					c <- chkp
				}
			})
		}
	}()

	return c
}

func getValidatorBalance(<-chan checkpoint) {

}

func setupAlerts(<-chan checkpoint) {

}
