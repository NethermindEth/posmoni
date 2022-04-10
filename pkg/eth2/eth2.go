package eth2

import (
	"fmt"
	"os"

	"github.com/NethermindEth/posgonitor/configs"

	log "github.com/sirupsen/logrus"
)

func Monitor(handleCfg bool) []chan struct{} {
	if handleCfg {
		configs.InitConfig()
	}

	cfg, err := Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	configs.InitLogging()

	log.Debugf("Configuration object: %+v", cfg)

	subDone := make(chan struct{})
	sub := SubscribeOpts{
		endpoints:  cfg.Consensus,
		streamURL:  finalizedCkptTopic,
		subscriber: &sseSubscriber{},
	}
	chkps := subscribe(subDone, sub)

	//TODO: Init and setup DB

	go getValidatorBalance(chkps)
	go setupAlerts(chkps)

	return []chan struct{}{subDone}
}

func getValidatorBalance(chkps <-chan checkpoint) {
	for c := range chkps {
		log.WithField(configs.Component, "ETH2").Infof("Got checkpoint: %+v", c)
	}
}

func setupAlerts(<-chan checkpoint) {

}
