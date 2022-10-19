package eth2

import (
	"fmt"
	"strconv"
	"time"

	"github.com/NethermindEth/posmoni/configs"
	"github.com/NethermindEth/posmoni/pkg/eth2/db"
	net "github.com/NethermindEth/posmoni/pkg/eth2/networking"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

// Middleware for ETH2 validators monitoring
type eth2Monitor struct {
	// Interface for data access
	repository db.Repository
	// Interface for Beacon chain API interaction
	beaconClient net.BeaconAPI
	// Interface for ETH1 json-rpc API interaction
	executionClient net.ExecutionAPI
	// Configuration options for events subscriber
	subscriberOpts net.SubscribeOpts
	// Configuration data for eth2Monitor
	config eth2Config
}

/*
DefaultEth2Monitor :
Factory for eth2Monitor with recommended settings.

params :-
a. opts ConfigOpts
Monitor configuration options

returns :-
a. *eth2Monitor
Monitor middleware intialized with default settings
b. error
Error if any
*/
func DefaultEth2Monitor(opts ConfigOpts) (*eth2Monitor, error) {
	// notest
	// Setup database
	ormdb, err := gorm.Open(sqlite.Open("eth2_monitor.db"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf(SQLiteCreationError, err)
	}

	monitor := &eth2Monitor{
		repository:      &db.SQLiteRepository{DB: ormdb},
		beaconClient:    &net.BeaconClient{RetryDuration: time.Minute},
		executionClient: &net.ExecutionClient{RetryDuration: time.Minute},
		subscriberOpts: net.SubscribeOpts{
			StreamURL:  net.FinalizedCkptTopic,
			Subscriber: &net.SSESubscriber{},
		},
	}

	opts = ConfigOpts{
		HandleCfg: false,
		Checkers: []CfgChecker{
			{Key: Execution, ErrMsg: NoExecutionFoundError},
			{Key: Consensus, ErrMsg: NoConsensusFoundError},
			{Key: Validators, ErrMsg: NoValidatorsFoundError, Optional: true},
			{Key: ValidatorsExternalHttp, Optional: true},
		},
	}

	err = monitor.setup(opts)
	if err != nil {
		return nil, fmt.Errorf(SetupError, err)
	}

	return monitor, nil
}

/*
NewEth2Monitor :
Factory for eth2Monitor.

params :-
a. r db.Repository
Interface implementation for data access
b. bc networking.BeaconAPI
Interface implementation for Beacon chain API interaction
c. so networking.SubscribeOpts
Configuration options for events subscriber. Should include implementation for Subscriber interface.
d. opts ConfigOpts
Monitor configuration options

returns :-
a. *eth2Monitor
Monitor middleware intialized with desired settings
*/
func NewEth2Monitor(r db.Repository, bc net.BeaconAPI, ex net.ExecutionAPI, so net.SubscribeOpts, opts ConfigOpts) (*eth2Monitor, error) {
	monitor := &eth2Monitor{
		repository:      r,
		beaconClient:    bc,
		executionClient: ex,
		subscriberOpts:  so,
	}

	err := monitor.setup(opts)
	if err != nil {
		return nil, fmt.Errorf(SetupError, err)
	}

	return monitor, nil
}

/*
setup :
Handle eth2Monitor configuration.
params :-
a. handleCfg bool
True if configuration setup (configuration file setup or enviroment variables setup) should be handled
b. config *eth2Config
Configuration data. Should be used when is not desired to use config file or enviroment variables to get configuration data.

returns :-
a. error
Error if any
*/
func (e *eth2Monitor) setup(opts ConfigOpts) error {
	if opts.HandleCfg {
		configs.InitConfig()
	}

	// TODO: Handle empty opts for uses cases like TrackSync only
	cfg, err := Init(opts.Checkers)
	if err != nil {
		fmt.Println(err)
		return err
	}
	e.config = cfg

	// setup beacon nodes endpoints
	e.subscriberOpts.Endpoints = e.config.consensus
	e.beaconClient.SetEndpoints(e.config.consensus)

	if opts.handleLogs {
		// setup logger
		configs.InitLogging()
	}

	log.Debugf("Configuration object: %+v", e.config)

	if err := e.repository.Migrate(); err != nil {
		return fmt.Errorf(MigrationError, err)
	}

	return nil
}

/*
Monitor :
Pipeline and entrypoint for validator monitoring.

params :-
a. handleCfg bool
True if configuration setup (configuration file setup or enviroment variables setup) should be handled

returns :-
a. []chan struct{}
List of channels to be closed when monitoring is done
b. error
Error if any
*/
func (e *eth2Monitor) Monitor() ([]chan struct{}, error) {
	subDone := make(chan struct{})
	chkps := net.Subscribe(subDone, e.subscriberOpts)

	go e.getValidatorBalance(chkps, e.config.validators)
	go e.setupAlerts(chkps)

	return []chan struct{}{subDone}, nil
}

/*
getValidatorBalance :
Track validator balance and performance.

params :-
a. chkps <-chan networking.Checkpoint
Channel to get new checkpoints from
b. validatorsIdxs []string
List of validator indexes to track

returns :-
none
*/
func (e *eth2Monitor) getValidatorBalance(chkps <-chan net.Checkpoint, validatorsIdxs []string) {
	logFields := log.Fields{configs.Component: "ETH2 Monitor", "Method": "getValidatorBalance"}

	for c := range chkps {
		log.WithFields(logFields).Infof("Got Checkpoint: %+v", c)

		// New finalized checkpoint. Fetch validator balances
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

			metricsValidatorBalance.With(prometheus.Labels{ValidatorLabel: vb.Index}).Set(float64(newBalance))

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

func (e *eth2Monitor) TrackSync(done <-chan struct{}, beaconEndpoints, executionEndpoints []string, wait time.Duration) <-chan EndpointSyncStatus {
	logFields := log.Fields{configs.Component: "ETH2 Monitor", "Method": "TrackSync"}
	c := make(chan EndpointSyncStatus, len(executionEndpoints)+len(beaconEndpoints))
	var w time.Duration

	go func() {
		for {
			select {
			case <-done:
				close(c)
				return
			case <-time.After(w):
				if w == 0 {
					// Don't wait the first time
					w = wait
				}
				// TODO: Benchmark this and check what happens if the processing is longer than the wait
				// Check sync progress of beacon nodes
				log.WithFields(logFields).Info("Tracking sync progress of consensus nodes...")
				bStatus := e.beaconClient.SyncStatus(beaconEndpoints)
				for _, s := range bStatus {
					if s.Error != nil {
						log.WithFields(logFields).Errorf(CheckingSyncStatusError, s.Endpoint, s.Error)
						c <- EndpointSyncStatus{Endpoint: s.Endpoint, Error: s.Error}
					} else {
						if s.IsSyncing {
							log.WithFields(logFields).Infof("Endpoint %s is syncing", s.Endpoint)
						} else {
							log.WithFields(logFields).Infof("Endpoint %s is synced", s.Endpoint)
						}
						c <- EndpointSyncStatus{Endpoint: s.Endpoint, Synced: !s.IsSyncing}
					}
				}

				// Check sync progress of execution nodes. Rule of Three not acomplished yet, so no harm in repetition :)
				log.WithFields(logFields).Info("Tracking sync progress of execution nodes...")
				eStatus := e.executionClient.SyncStatus(executionEndpoints)
				for _, s := range eStatus {
					if s.Error != nil {
						log.WithFields(logFields).Errorf(CheckingSyncStatusError, s.Endpoint, s.Error)
						c <- EndpointSyncStatus{Endpoint: s.Endpoint, Error: s.Error}
					} else {
						if s.IsSyncing {
							log.WithFields(logFields).Infof("Endpoint %s is syncing", s.Endpoint)
						} else {
							log.WithFields(logFields).Infof("Endpoint %s is synced", s.Endpoint)
						}
						c <- EndpointSyncStatus{Endpoint: s.Endpoint, Synced: !s.IsSyncing}
					}
				}
			}
		}
	}()

	return c
}
