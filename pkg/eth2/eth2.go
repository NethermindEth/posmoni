package eth2

import (
	"fmt"
	"strconv"
	"time"

	"github.com/NethermindEth/posgonitor/configs"
	"github.com/NethermindEth/posgonitor/pkg/eth2/db"
	net "github.com/NethermindEth/posgonitor/pkg/eth2/networking"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

// Middleware
type eth2Monitor struct {
	repository     db.Repository
	beaconClient   net.BeaconAPI
	subscriberOpts net.SubscribeOpts
}

func DefaultEth2Monitor() (*eth2Monitor, error) {
	ormdb, err := gorm.Open(sqlite.Open("eth2_monitor.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf(SQLiteCreationError, err)
	}

	return &eth2Monitor{
		repository:   &db.SQLiteRepository{DB: ormdb},
		beaconClient: &net.BeaconClient{RetryDuration: time.Minute},
		subscriberOpts: net.SubscribeOpts{
			StreamURL:  net.FinalizedCkptTopic,
			Subscriber: &net.SSESubscriber{},
		},
	}, nil
}

func NewEth2Monitor(r db.Repository, bc net.BeaconAPI, so net.SubscribeOpts) *eth2Monitor {
	return &eth2Monitor{
		repository:     r,
		beaconClient:   bc,
		subscriberOpts: so,
	}
}

func (e *eth2Monitor) Monitor(handleCfg bool) ([]chan struct{}, error) {
	if handleCfg {
		configs.InitConfig()
	}

	cfg, err := Init()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// setup beacon nodes endpoints
	e.subscriberOpts.Endpoints = cfg.Consensus
	e.beaconClient.SetEndpoints(cfg.Consensus)

	// setup logger
	configs.InitLogging()

	log.Debugf("Configuration object: %+v", cfg)

	if err = e.repository.Migrate(); err != nil {
		return nil, fmt.Errorf(MigrationError, err)
	}

	subDone := make(chan struct{})
	chkps := net.Subscribe(subDone, e.subscriberOpts)

	go e.getValidatorBalance(chkps, cfg.Validators)
	go e.setupAlerts(chkps)

	return []chan struct{}{subDone}, nil
}

func (e *eth2Monitor) getValidatorBalance(chkps <-chan net.Checkpoint, validatorsIdxs []string) {
	logFields := log.Fields{configs.Component: "ETH2 Monitor", "Method": "getValidatorBalance"}

	for c := range chkps {
		log.WithFields(logFields).Infof("Got Checkpoint: %+v", c)

		// Hardcoding head state for now
		vbs, err := e.beaconClient.ValidatorBalances("head", validatorsIdxs)
		if err != nil {
			log.WithFields(logFields).Errorf(ValidatorBalancesError, err)
			continue
		}

		for _, vb := range vbs {
			log.WithFields(logFields).Debugf("Validator Balance fetched: %+v", vb)

			// Get validator index from response data
			idx, err := parseUint(vb.Index)
			if err != nil {
				log.WithFields(logFields).Errorf(ParseUintError, err)
				continue
			}

			// Get validator balance from response data
			newBalance, err := strconv.ParseUint(vb.Balance, 10, 64)
			if err != nil {
				log.WithFields(logFields).Errorf(ParseUintError, err)
				continue
			}

			// Get validator from db
			v, err := e.repository.FirstOrCreate(db.Validator{Idx: idx, Balance: newBalance})
			if err != nil {
				log.WithFields(logFields).Errorf(ValidatorNotFoundError, err)
				continue
			}

			if newBalance < v.Balance {
				log.WithFields(logFields).Warnf("Attestation has been missed by %d, count: %d", v.Idx, v.MissedAtts+1)
				e.repository.Update(db.Validator{
					Idx:             v.Idx,
					Balance:         newBalance,
					MissedAtts:      v.MissedAtts + 1,
					MissedAttsTotal: v.MissedAttsTotal + 1,
				})
			} else {
				e.repository.Update(db.Validator{
					Idx:             v.Idx,
					Balance:         newBalance,
					MissedAtts:      0,
					MissedAttsTotal: v.MissedAttsTotal,
				})
			}
		}
	}
}

func (e *eth2Monitor) setupAlerts(<-chan net.Checkpoint) {

}
