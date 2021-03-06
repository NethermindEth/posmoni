package eth2

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/NethermindEth/posmoni/internal/utils"
	"github.com/NethermindEth/posmoni/pkg/eth2/db"
	net "github.com/NethermindEth/posmoni/pkg/eth2/networking"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type bcSyncStatusInfo struct {
	returnData [][]net.BeaconSyncingStatus
	current    int
}

type validatorBalanceInfo struct {
	returnData [][]net.ValidatorBalance
	current    int
}

// Mock of BeaconAPI
type TestBeaconClient struct {
	endpoints []string
	vbCall    validatorBalanceInfo
	ssCall    bcSyncStatusInfo
}

func (tbc *TestBeaconClient) SetEndpoints(endpoints []string) {
	tbc.endpoints = endpoints
}

func (tbc *TestBeaconClient) ValidatorBalances(stateID string, validatorIdxs []string) ([]net.ValidatorBalance, error) {
	// Simulates an iterator
	if tbc.vbCall.current >= len(tbc.vbCall.returnData) {
		return nil, fmt.Errorf("No more data")
	}

	if tbc.vbCall.returnData[tbc.vbCall.current] == nil {
		return nil, fmt.Errorf("Intentional error")
	}

	tbc.vbCall.current++
	return tbc.vbCall.returnData[tbc.vbCall.current-1], nil
}

func (tbc *TestBeaconClient) Health(endpoints []string) []net.HealthResponse {
	return nil
}

func (tbc *TestBeaconClient) SyncStatus(endpoints []string) []net.BeaconSyncingStatus {
	if tbc.ssCall.current >= len(tbc.ssCall.returnData) {
		return nil
	}

	tbc.ssCall.current++
	return tbc.ssCall.returnData[tbc.ssCall.current-1]
}

type exSyncStatusInfo struct {
	returnData [][]net.ExecutionSyncingStatus
	current    int
}

// Mock of ExecutionAPI
type TestExecutionClient struct {
	ssCall exSyncStatusInfo
}

func (tec *TestExecutionClient) Call(endpoint, method string, params ...any) (json.RawMessage, error) {
	return nil, nil
}

func (tec *TestExecutionClient) SyncStatus(endpoints []string) []net.ExecutionSyncingStatus {
	if tec.ssCall.current >= len(tec.ssCall.returnData) {
		return nil
	}

	tec.ssCall.current++
	return tec.ssCall.returnData[tec.ssCall.current-1]
}

func newTestExecutionClient(ssData [][]net.ExecutionSyncingStatus) *TestExecutionClient {
	return &TestExecutionClient{
		ssCall: exSyncStatusInfo{
			returnData: ssData,
			current:    0,
		},
	}
}

func fillChannel(data []net.Checkpoint) <-chan net.Checkpoint {
	ch := make(chan net.Checkpoint, len(data))
	go func() {
		for _, d := range data {
			ch <- d
		}
		close(ch)
	}()
	return ch
}

func populateDb(r db.Repository, existingData []db.Validator) error {
	for _, d := range existingData {
		_, err := r.FirstOrCreate(d)
		if err != nil {
			return err
		}
	}
	return nil
}

func newTestBeaconClient(vbData [][]net.ValidatorBalance, ssData [][]net.BeaconSyncingStatus) *TestBeaconClient {
	return &TestBeaconClient{
		vbCall: validatorBalanceInfo{
			returnData: vbData,
			current:    0,
		},
		ssCall: bcSyncStatusInfo{
			returnData: ssData,
			current:    0,
		},
	}
}

func setup(data [][]net.ValidatorBalance, opts net.SubscribeOpts, cfgOpts ConfigOpts) (*eth2Monitor, error) {
	ormdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("In memory sqlite creation failed. Error '%v'", err)
	}

	ormdb.AutoMigrate(&db.ValidatorORM{})

	monitor, err := NewEth2Monitor(&db.SQLiteRepository{DB: ormdb}, newTestBeaconClient(data, nil), &net.ExecutionClient{}, opts, cfgOpts)
	if err != nil {
		return nil, fmt.Errorf("Monitor creation failed. Error '%v'", err)
	}

	return monitor, nil
}

func cleanup(r db.Repository) error {
	rorm := r.(*db.SQLiteRepository)
	sqlDB, err := rorm.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

type testSubscriber struct {
	data map[string][]net.Checkpoint
}

func (s testSubscriber) Listen(url string, ch chan<- net.Checkpoint) {
	for _, data := range s.data[url] {
		ch <- data
		//sleep to simulate a delay
		time.Sleep(time.Millisecond * 50)
	}
}

func TestGetValidatorBalance(t *testing.T) {
	tcs := []struct {
		name string
		// input data
		subscriptionData []net.Checkpoint
		// data that should be in the db before the test
		existingData []db.Validator
		// ordered data returned by the beacon client
		requestData [][]net.ValidatorBalance
		// data to validate in test db
		want []db.Validator
	}{
		{
			name:             "Test case 1, Empty and closed channel, nothing should happen",
			subscriptionData: []net.Checkpoint{},
			existingData:     []db.Validator{},
			requestData:      [][]net.ValidatorBalance{},
			want:             []db.Validator{},
		},
		{
			name: "Test case 2, One entry in channel, one validator to insert",
			subscriptionData: []net.Checkpoint{
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"}},
			existingData: []db.Validator{},
			requestData: [][]net.ValidatorBalance{
				{
					{Index: "1", Balance: "32000136946"},
				},
			},
			want: []db.Validator{
				{Idx: 1, Balance: 32000136946, MissedAtts: 0, MissedAttsTotal: 0},
			},
		},
		{
			name: "Test case 3, One entry in channel, one validator to update, positive balance change",
			subscriptionData: []net.Checkpoint{
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"}},
			existingData: []db.Validator{
				{Idx: 1, Balance: 32000136946, MissedAtts: 0, MissedAttsTotal: 0},
			},
			requestData: [][]net.ValidatorBalance{
				{
					{Index: "1", Balance: "34000136946"},
				},
			},
			want: []db.Validator{
				{Idx: 1, Balance: 34000136946, MissedAtts: 0, MissedAttsTotal: 0},
			},
		},
		{
			name: "Test case 4, One entry in channel, one validator to update, negative balance change",
			subscriptionData: []net.Checkpoint{
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"}},
			existingData: []db.Validator{
				{Idx: 1, Balance: 34000136946, MissedAtts: 0, MissedAttsTotal: 0},
			},
			requestData: [][]net.ValidatorBalance{
				{
					{Index: "1", Balance: "32000136946"},
				},
			},
			want: []db.Validator{
				{Idx: 1, Balance: 32000136946, MissedAtts: 1, MissedAttsTotal: 1},
			},
		},
		{
			name: "Test case 5, One entry in channel, one validator to update, negative balance change, existing missed atts",
			subscriptionData: []net.Checkpoint{
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"}},
			existingData: []db.Validator{
				{Idx: 1, Balance: 34000136946, MissedAtts: 1, MissedAttsTotal: 1},
			},
			requestData: [][]net.ValidatorBalance{
				{
					{Index: "1", Balance: "32000136946"},
				},
			},
			want: []db.Validator{
				{Idx: 1, Balance: 32000136946, MissedAtts: 2, MissedAttsTotal: 2},
			},
		},
		{
			name: "Test case 6, One entry in channel, one validator to update, positive balance change, existing missed atts",
			subscriptionData: []net.Checkpoint{
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"}},
			existingData: []db.Validator{
				{Idx: 1, Balance: 32000136946, MissedAtts: 2, MissedAttsTotal: 4},
			},
			requestData: [][]net.ValidatorBalance{
				{
					{Index: "1", Balance: "34000136946"},
				},
			},
			want: []db.Validator{
				{Idx: 1, Balance: 34000136946, MissedAtts: 0, MissedAttsTotal: 4},
			},
		},
		{
			name: "Test case 7, One entry in channel, negative index in request",
			subscriptionData: []net.Checkpoint{
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"}},
			existingData: []db.Validator{
				{Idx: 1, Balance: 32000136946, MissedAtts: 2, MissedAttsTotal: 4},
			},
			requestData: [][]net.ValidatorBalance{
				{
					{Index: "-1", Balance: "34000136946"},
				},
			},
			want: []db.Validator{
				{Idx: 1, Balance: 32000136946, MissedAtts: 2, MissedAttsTotal: 4},
			},
		},
		{
			name: "Test case 8, One entry in channel, incorrect balance in request",
			subscriptionData: []net.Checkpoint{
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"}},
			existingData: []db.Validator{
				{Idx: 1, Balance: 32000136946, MissedAtts: 2, MissedAttsTotal: 4},
			},
			requestData: [][]net.ValidatorBalance{
				{
					{Index: "1", Balance: "aaaaa"},
				},
			},
			want: []db.Validator{
				{Idx: 1, Balance: 32000136946, MissedAtts: 2, MissedAttsTotal: 4},
			},
		},
		{
			name: "Test case 9, One entry in channel, bad request, no data",
			subscriptionData: []net.Checkpoint{
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"}},
			existingData: []db.Validator{},
			requestData: [][]net.ValidatorBalance{
				{
					{Index: "aaaaa", Balance: "aaaaa"},
				},
			},
			want: []db.Validator{},
		},
		{
			name: "Test case 10, One entry in channel, error from ValidatorBalance, no data",
			subscriptionData: []net.Checkpoint{
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"}},
			existingData: []db.Validator{},
			requestData:  [][]net.ValidatorBalance{},
			want:         []db.Validator{},
		},
		{
			name: "Test case 11, several entries in channel, mixed behavior",
			subscriptionData: []net.Checkpoint{
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"},
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"},
				{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"}},
			existingData: []db.Validator{
				{Idx: 1, Balance: 32000136946, MissedAtts: 6, MissedAttsTotal: 400},
				{Idx: 2, Balance: 33000136946, MissedAtts: 2, MissedAttsTotal: 30},
				{Idx: 3, Balance: 35000136946, MissedAtts: 0, MissedAttsTotal: 0},
			},
			requestData: [][]net.ValidatorBalance{
				{
					{Index: "1", Balance: "33000136946"},
				},
				{
					{Index: "2", Balance: "31000136946"},
					{Index: "3", Balance: "36000136946"},
				},
				{
					{Index: "2", Balance: "30000136946"},
				},
			},
			want: []db.Validator{
				{Idx: 1, Balance: 33000136946, MissedAtts: 0, MissedAttsTotal: 400},
				{Idx: 2, Balance: 30000136946, MissedAtts: 4, MissedAttsTotal: 32},
				{Idx: 3, Balance: 36000136946, MissedAtts: 0, MissedAttsTotal: 0},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			monitor, err := setup(tc.requestData, net.SubscribeOpts{}, ConfigOpts{Checkers: []CfgChecker{{Key: Consensus, ErrMsg: NoConsensusFoundError, Data: []string{"1", "2", "3"}}}})
			if err != nil {
				t.Fatalf("Setup failed. Error %v", err)
			}

			err = populateDb(monitor.repository, tc.existingData)
			if err != nil {
				t.Fatalf("Populate db failed. Error %v", err)
			}
			input := fillChannel(tc.subscriptionData)
			monitor.getValidatorBalance(input, []string{})

			for _, want := range tc.want {
				got, err := monitor.repository.Validator(want.Idx)
				if err != nil {
					t.Fatalf("Validator %v not found in db", want.Idx)
				}
				assert.Equal(t, want, got, "Validator in db with index %v is not equal to the wanted one", want.Idx)
			}

			if err = cleanup(monitor.repository); err != nil {
				t.Fatalf("Cleanup failed. Error %v", err)
			}
		})
	}
}

type repositoryMock struct {
	migrationCalled        int
	expectedMigrationCalls int
	migrationError         bool
}

func (rm *repositoryMock) FirstOrCreate(val db.Validator) (v db.Validator, err error) {
	return
}

func (rm *repositoryMock) Update(val db.Validator) (err error) {
	return
}

func (rm *repositoryMock) Validator(idx uint) (v db.Validator, err error) {
	return
}

func (rm *repositoryMock) Migrate() error {
	rm.migrationCalled++

	if rm.migrationError {
		return errors.New("migration error")
	}

	return nil
}

func TestSetup(t *testing.T) {
	tcs := []struct {
		name  string
		opts  ConfigOpts
		want  *eth2Monitor
		env   map[string]string
		isErr bool
		mock  repositoryMock
	}{
		{
			name: "Test case 1, valid config, prepared configuration data",
			opts: ConfigOpts{
				HandleCfg: false,
				Checkers: []CfgChecker{
					{Key: Validators, ErrMsg: NoValidatorsFoundError, Data: []string{"1", "2", "3"}},
					{Key: Consensus, ErrMsg: NoConsensusFoundError, Data: []string{"Endpoint1"}},
				},
			},
			want: &eth2Monitor{
				config: eth2Config{
					validators: []string{"1", "2", "3"},
					consensus:  []string{"Endpoint1"},
				},
				subscriberOpts: net.SubscribeOpts{Endpoints: []string{"Endpoint1"}},
			},
			isErr: false,
			mock:  repositoryMock{migrationError: false, expectedMigrationCalls: 1},
		},
		{
			name: "Test case 2, valid config, load configuration data from env, config setup not handled",
			opts: ConfigOpts{
				HandleCfg: false,
				Checkers: []CfgChecker{
					{Key: Validators, ErrMsg: NoValidatorsFoundError},
					{Key: Consensus, ErrMsg: NoConsensusFoundError},
				},
			},
			want: &eth2Monitor{
				config: eth2Config{
					validators: []string{"1", "2", "3"},
					consensus:  []string{"Endpoint1"},
				},
				subscriberOpts: net.SubscribeOpts{Endpoints: []string{"Endpoint1"}},
			},
			env: map[string]string{
				"PM_VALIDATORS": "1,2,3",
				"PM_CONSENSUS":  "Endpoint1",
			},
			isErr: false,
			mock:  repositoryMock{migrationError: false, expectedMigrationCalls: 1},
		},
		{
			name: "Test case 3, valid config, load configuration data from env, config setup handled",
			opts: ConfigOpts{
				HandleCfg: true,
				Checkers: []CfgChecker{
					{Key: Validators, ErrMsg: NoValidatorsFoundError},
					{Key: Consensus, ErrMsg: NoConsensusFoundError},
				},
			},
			want: &eth2Monitor{
				config: eth2Config{
					validators: []string{"1", "2", "3"},
					consensus:  []string{"Endpoint1"},
				},
				subscriberOpts: net.SubscribeOpts{Endpoints: []string{"Endpoint1"}},
			},
			env: map[string]string{
				"PM_VALIDATORS": "1,2,3",
				"PM_CONSENSUS":  "Endpoint1",
			},
			isErr: false,
			mock:  repositoryMock{migrationError: false, expectedMigrationCalls: 1},
		},
		{
			name: "Test case 4, invalid config (not consensus provided), prepared configuration data",
			opts: ConfigOpts{
				HandleCfg: false,
				Checkers: []CfgChecker{
					{Key: Validators, ErrMsg: NoValidatorsFoundError, Data: []string{"1", "2", "3"}},
					{Key: Consensus, ErrMsg: NoConsensusFoundError},
				},
			},
			want:  &eth2Monitor{},
			isErr: true,
			mock:  repositoryMock{migrationError: false, expectedMigrationCalls: 0},
		},
		{
			name: "Test case 5, invalid config, load configuration data from env, config setup not handled",
			opts: ConfigOpts{
				HandleCfg: false,
				Checkers: []CfgChecker{
					{Key: Validators, ErrMsg: NoValidatorsFoundError},
					{Key: Consensus, ErrMsg: NoConsensusFoundError},
				},
			},
			want: &eth2Monitor{},
			env: map[string]string{
				"PM_VALIDATORS": "1,2,3",
			},
			isErr: true,
			mock:  repositoryMock{migrationError: false, expectedMigrationCalls: 0},
		},
		{
			name: "Test case 6, invalid config, load configuration data from env, config setup handled",
			opts: ConfigOpts{
				HandleCfg: true,
				Checkers: []CfgChecker{
					{Key: Validators, ErrMsg: NoValidatorsFoundError},
					{Key: Consensus, ErrMsg: NoConsensusFoundError},
				},
			},
			want: &eth2Monitor{},
			env: map[string]string{
				"PM_VALIDATORS": "1,2,3",
			},
			isErr: true,
			mock:  repositoryMock{migrationError: false, expectedMigrationCalls: 0},
		},
		{
			name: "Test case 7, valid config, migration error",
			opts: ConfigOpts{
				HandleCfg: false,
				Checkers: []CfgChecker{
					{Key: Validators, ErrMsg: NoValidatorsFoundError, Data: []string{"1", "2", "3"}},
					{Key: Consensus, ErrMsg: NoConsensusFoundError, Data: []string{"Endpoint1"}},
				},
			},
			want: &eth2Monitor{
				config: eth2Config{
					validators: []string{"1", "2", "3"},
					consensus:  []string{"Endpoint1"},
				},
				subscriberOpts: net.SubscribeOpts{Endpoints: []string{"Endpoint1"}},
			},
			isErr: true,
			mock:  repositoryMock{migrationError: true, expectedMigrationCalls: 1},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			// setup
			if !tc.opts.HandleCfg {
				viper.AutomaticEnv()
			}

			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			monitor := &eth2Monitor{
				repository:     &tc.mock,
				beaconClient:   newTestBeaconClient([][]net.ValidatorBalance{}, nil),
				subscriberOpts: net.SubscribeOpts{},
			}

			// execute
			err := monitor.setup(tc.opts)

			descr := fmt.Sprintf("setup(%+v)", tc.opts)
			if err = utils.CheckErr(descr, tc.isErr, err); err != nil {
				t.Error(err)
			}

			assert.Equal(t, monitor.subscriberOpts, tc.want.subscriberOpts, descr+" gave a misconfigured monitor")
			assert.Equal(t, monitor.config, tc.want.config, descr+" gave a misconfigured monitor")

			if tc.mock.expectedMigrationCalls != tc.mock.migrationCalled {
				t.Errorf("Expected %d migration calls, got %d", tc.mock.expectedMigrationCalls, tc.mock.migrationCalled)
			}
		})
	}
}

func TestMonitor(t *testing.T) {
	// DEV: Current tests assume:
	// - Support for only one beacon node endpoint
	// - setupAlerts not implemented
	// - ValidatorBalances within getValidatorBalance fetch 'head' instead of slot

	type setupArgs struct {
		requestData [][]net.ValidatorBalance
		opts        net.SubscribeOpts
		cfgOpts     ConfigOpts
	}

	tcs := []struct {
		name         string
		args         setupArgs
		want         []db.Validator
		isErr        bool
		sleep        time.Duration
		existingData []db.Validator
	}{
		{
			name: "Test case 1, normal workflow",
			args: setupArgs{
				requestData: [][]net.ValidatorBalance{
					{
						{Index: "1", Balance: "33000136946"},
					},
					{
						{Index: "2", Balance: "31000136946"},
						{Index: "3", Balance: "36000136946"},
					},
					{
						{Index: "2", Balance: "30000136946"},
					},
				},
				opts: net.SubscribeOpts{Subscriber: testSubscriber{
					data: map[string][]net.Checkpoint{
						"Endpoint1": {
							{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"},
							{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"},
							{Block: "0x9a2fefd2fdb57f74993c7780ea5b9030d2897b615b89f808011ca5aebed54eaf", State: "0x600e852a08c1200654ddf11025f1ceacb3c2e74bdd5c630cde0838b2591b69f9", Epoch: "2"},
						},
					},
				},
				},
				cfgOpts: ConfigOpts{
					HandleCfg: false,
					Checkers: []CfgChecker{
						{Key: Validators, ErrMsg: NoValidatorsFoundError, Data: []string{"1", "2", "3"}},
						{Key: Consensus, ErrMsg: NoConsensusFoundError, Data: []string{"Endpoint1"}},
					},
				},
			},
			want: []db.Validator{
				{Idx: 1, Balance: 33000136946, MissedAtts: 0, MissedAttsTotal: 400},
				{Idx: 2, Balance: 30000136946, MissedAtts: 4, MissedAttsTotal: 32},
				{Idx: 3, Balance: 36000136946, MissedAtts: 0, MissedAttsTotal: 0},
			},
			isErr: false,
			sleep: time.Second,
			existingData: []db.Validator{
				{Idx: 1, Balance: 32000136946, MissedAtts: 6, MissedAttsTotal: 400},
				{Idx: 2, Balance: 33000136946, MissedAtts: 2, MissedAttsTotal: 30},
				{Idx: 3, Balance: 35000136946, MissedAtts: 0, MissedAttsTotal: 0},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			monitor, err := setup(tc.args.requestData, tc.args.opts, tc.args.cfgOpts)
			if err != nil {
				t.Fatal(err)
			}

			err = populateDb(monitor.repository, tc.existingData)
			if err != nil {
				t.Fatalf("Populate db failed. Error %v", err)
			}

			descr := fmt.Sprintf("Monitor() with args %+v", tc.args)
			_, err = monitor.Monitor()
			if err = utils.CheckErr(descr, tc.isErr, err); err != nil {
				t.Error(err)
			}

			// Wait for goroutines to work
			time.Sleep(time.Second)

			for _, want := range tc.want {
				got, err := monitor.repository.Validator(want.Idx)
				if err != nil {
					t.Fatalf("Validator %v not found in db", want.Idx)
				}
				assert.Equal(t, want, got, "Validator in db with index %v is not equal to the wanted one", want.Idx)
			}

			if err = cleanup(monitor.repository); err != nil {
				t.Fatalf("Cleanup failed. Error %v", err)
			}
		})
	}
}

func TestTrackSync(t *testing.T) {
	t.Parallel()

	type opts struct {
		bcEndpoints []string
		bcData      [][]net.BeaconSyncingStatus
		exEndpoints []string
		exData      [][]net.ExecutionSyncingStatus
		wait        time.Duration
	}

	tcs := []struct {
		name      string
		setupOpts opts
		want      []EndpointSyncStatus
	}{
		{
			"Test case 1, one consensus node, synced",
			opts{
				bcEndpoints: []string{"1"},
				bcData: [][]net.BeaconSyncingStatus{
					{
						net.BeaconSyncingStatus{
							Endpoint:  "1",
							IsSyncing: false,
						},
					},
				},
				wait: time.Millisecond,
			},
			[]EndpointSyncStatus{{Endpoint: "1", Synced: true}},
		},
		{
			"Test case 2, one consensus node, not synced",
			opts{
				bcEndpoints: []string{"1"},
				bcData: [][]net.BeaconSyncingStatus{
					{
						net.BeaconSyncingStatus{
							Endpoint:  "1",
							IsSyncing: true,
						},
					},
				},
				wait: time.Millisecond,
			},
			[]EndpointSyncStatus{{Endpoint: "1", Synced: false}},
		},
		{
			"Test case 3, one execution node, synced",
			opts{
				exEndpoints: []string{"1"},
				exData: [][]net.ExecutionSyncingStatus{
					{
						net.ExecutionSyncingStatus{
							Endpoint:  "1",
							IsSyncing: false,
						},
					},
				},
				wait: time.Millisecond,
			},
			[]EndpointSyncStatus{{Endpoint: "1", Synced: true}},
		},
		{
			"Test case 4, one execution node, not synced",
			opts{
				exEndpoints: []string{"1"},
				exData: [][]net.ExecutionSyncingStatus{
					{
						net.ExecutionSyncingStatus{
							Endpoint:  "1",
							IsSyncing: true,
						},
					},
				},
				wait: time.Millisecond,
			},
			[]EndpointSyncStatus{{Endpoint: "1", Synced: false}},
		},
		{
			"Test case 5, two execution nodes, mixed sync status",
			opts{
				exEndpoints: []string{"1", "2"},
				exData: [][]net.ExecutionSyncingStatus{
					{
						net.ExecutionSyncingStatus{
							Endpoint:  "1",
							IsSyncing: true,
						},
						net.ExecutionSyncingStatus{
							Endpoint:  "2",
							IsSyncing: false,
						},
					},
				},
				wait: time.Millisecond,
			},
			[]EndpointSyncStatus{{Endpoint: "1", Synced: false}, {Endpoint: "2", Synced: true}},
		},
		{
			"Test case 6, two mixed nodes, mixed sync status",
			opts{
				bcEndpoints: []string{"2"},
				bcData: [][]net.BeaconSyncingStatus{
					{
						net.BeaconSyncingStatus{
							Endpoint:  "2",
							IsSyncing: true,
						},
					},
				},
				exEndpoints: []string{"1"},
				exData: [][]net.ExecutionSyncingStatus{
					{
						net.ExecutionSyncingStatus{
							Endpoint:  "1",
							IsSyncing: false,
						},
					},
				},
				wait: time.Millisecond,
			},
			[]EndpointSyncStatus{{Endpoint: "2", Synced: false}, {Endpoint: "1", Synced: true}},
		},
		{
			"Test case 7, two mixed nodes, mixed sync status, one wait",
			opts{
				bcEndpoints: []string{"2"},
				bcData: [][]net.BeaconSyncingStatus{
					{
						net.BeaconSyncingStatus{
							Endpoint:  "2",
							IsSyncing: false,
						},
					},
					{
						net.BeaconSyncingStatus{
							Endpoint:  "2",
							IsSyncing: false,
						},
					},
				},
				exEndpoints: []string{"1"},
				exData: [][]net.ExecutionSyncingStatus{
					{
						net.ExecutionSyncingStatus{
							Endpoint:  "1",
							IsSyncing: true,
						},
					},
					{
						net.ExecutionSyncingStatus{
							Endpoint:  "1",
							IsSyncing: false,
						},
					},
				},
				wait: time.Millisecond,
			},
			[]EndpointSyncStatus{{Endpoint: "2", Synced: true}, {Endpoint: "1", Synced: false}, {Endpoint: "2", Synced: true}, {Endpoint: "1", Synced: true}},
		},
		{
			"Test case 8, one node, no response",
			opts{
				bcEndpoints: []string{"1"},
				bcData:      nil,
				wait:        time.Millisecond,
			},
			[]EndpointSyncStatus{},
		},
		{
			"Test case 9, one consensus node, error",
			opts{
				bcEndpoints: []string{"1"},
				bcData: [][]net.BeaconSyncingStatus{
					{
						net.BeaconSyncingStatus{
							Endpoint:  "1",
							IsSyncing: false,
							Error:     errors.New(""),
						},
					},
				},
				wait: time.Second,
			},
			[]EndpointSyncStatus{{Endpoint: "1", Error: errors.New("")}},
		},
		{
			"Test case 10, one execution node, error",
			opts{
				exEndpoints: []string{"1"},
				exData: [][]net.ExecutionSyncingStatus{
					{
						net.ExecutionSyncingStatus{
							Endpoint:  "1",
							IsSyncing: false,
							Error:     errors.New(""),
						},
					},
				},
				wait: time.Second,
			},
			[]EndpointSyncStatus{{Endpoint: "1", Error: errors.New("")}},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			monitor := eth2Monitor{
				beaconClient:    newTestBeaconClient(nil, tc.setupOpts.bcData),
				executionClient: newTestExecutionClient(tc.setupOpts.exData),
			}

			done := make(chan struct{})
			doneList := make(chan struct{})
			defer close(done)
			defer close(doneList)
			result := monitor.TrackSync(done, tc.setupOpts.bcEndpoints, tc.setupOpts.exEndpoints, tc.setupOpts.wait)

			got := make([]EndpointSyncStatus, 0)
			if len(tc.want) > 0 {
				go func() {
					for {
						for r := range result {
							got = append(got, r)
							if len(got) == len(tc.want) {
								doneList <- struct{}{}
								return
							}
						}
					}
				}()

				<-doneList
			}
			done <- struct{}{}

			assert.Equalf(t, tc.want, got, "TrackSync(..., %+v, %+v, %v) gave wrong results", tc.setupOpts.bcEndpoints, tc.setupOpts.exEndpoints, tc.setupOpts.wait)
		})
	}
}

// TODO: Test NewEth2Monitor
