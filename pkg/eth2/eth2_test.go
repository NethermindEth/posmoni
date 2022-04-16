package eth2

import (
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

// Mock of BeaconAPI.
type TestBeaconClient struct {
	data    [][]net.ValidatorBalance
	current int
}

func (tbc *TestBeaconClient) SetEndpoints(endpoints []string) {
	// do nothing
}

func (tbc *TestBeaconClient) ValidatorBalances(stateID string, validatorIdxs []string) ([]net.ValidatorBalance, error) {
	// Simulates an iterator
	if tbc.current >= len(tbc.data) {
		return nil, fmt.Errorf("No more data")
	}

	if tbc.data[tbc.current] == nil {
		return nil, fmt.Errorf("Intentional error")
	}

	tbc.current++
	return tbc.data[tbc.current-1], nil
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

func newTestBeaconClient(data [][]net.ValidatorBalance) *TestBeaconClient {
	return &TestBeaconClient{
		data:    data,
		current: 0,
	}
}

func setup(data [][]net.ValidatorBalance, opts net.SubscribeOpts) (*eth2Monitor, error) {
	ormdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("In memory sqlite creation failed. Error '%v'", err)
	}

	ormdb.AutoMigrate(&db.ValidatorORM{})

	return NewEth2Monitor(&db.SQLiteRepository{DB: ormdb}, newTestBeaconClient(data), opts), nil
}

func cleanup(r db.Repository) error {
	rorm := r.(*db.SQLiteRepository)
	sqlDB, err := rorm.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
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
			monitor, err := setup(tc.requestData, net.SubscribeOpts{})
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

func TestMonitor(t *testing.T) {
	// DEV: Current tests assume:
	// - Support for only one beacon node endpoint
	// - setupAlerts not implemented
	// - ValidatorBalances within getValidatorBalance fetch 'head' instead of slot

	type setupArgs struct {
		requestData  [][]net.ValidatorBalance
		opts         net.SubscribeOpts
		handleCfg    bool
		env          map[string]string
		migrateError bool
	}

	setupMonitor := func(args setupArgs) (*eth2Monitor, error) {
		m, err := setup(args.requestData, args.opts)
		if err != nil {
			return nil, err
		}

		if args.migrateError {
			err = cleanup(m.repository)
			if err != nil {
				return nil, err
			}
		}

		if !args.handleCfg {
			viper.AutomaticEnv()
		}

		for k, v := range args.env {
			t.Setenv(k, v)
		}

		return m, nil
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
			name: "Test case 1, config error, handleCfg",
			args: setupArgs{
				requestData: [][]net.ValidatorBalance{},
				opts:        net.SubscribeOpts{},
				handleCfg:   true,
				env:         map[string]string{},
			},
			want:  []db.Validator{},
			isErr: true,
			sleep: time.Millisecond,
		},
		{
			name: "Test case 2, config error, !handleCfg",
			args: setupArgs{
				requestData: [][]net.ValidatorBalance{},
				opts:        net.SubscribeOpts{},
				handleCfg:   false,
				env:         map[string]string{},
			},
			want:  []db.Validator{},
			isErr: true,
			sleep: time.Millisecond,
		},
		{
			name: "Test case 3, migration error",
			args: setupArgs{
				requestData:  [][]net.ValidatorBalance{},
				opts:         net.SubscribeOpts{},
				handleCfg:    false,
				env:          map[string]string{},
				migrateError: true,
			},
			want:  []db.Validator{},
			isErr: true,
			sleep: time.Millisecond,
		},
		{
			name: "Test case 4, normal workflow",
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
				handleCfg: true,
				env: map[string]string{
					"PGM_VALIDATORS": "1,2,3",
					"PGM_CONSENSUS":  "Endpoint1",
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
			monitor, err := setupMonitor(tc.args)
			if err != nil {
				t.Fatal(err)
			}

			err = populateDb(monitor.repository, tc.existingData)
			if err != nil {
				t.Fatalf("Populate db failed. Error %v", err)
			}

			descr := fmt.Sprintf("Monitor() with args %+v", tc.args)
			_, err = monitor.Monitor(tc.args.handleCfg)
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
