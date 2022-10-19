/*
Copyright © 2022 Nethermind

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package eth

import (
	"os"
	"os/signal"
	"time"

	"github.com/NethermindEth/posmoni/pkg/eth2"
	"github.com/NethermindEth/posmoni/pkg/eth2/db"
	net "github.com/NethermindEth/posmoni/pkg/eth2/networking"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	executionEndp []string
	consensusEndp []string
	cron          int
)

// TrackSyncCmd represents the TrackSync command
var TrackSyncCmd = &cobra.Command{
	Use:   "trackSync",
	Short: "Track sync progress of Ethereum nodes",
	Long:  `Track sync progress of Ethereum's execution and Ethereum2 consensus nodes. You need to provide a list of execution and consensus nodes endpoints or put them in a configuration file or environment variables. Check the project's README for more information.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		monitor, err := eth2.NewEth2Monitor(
			db.EmptyRepository{},
			&net.BeaconClient{RetryDuration: time.Second},
			&net.ExecutionClient{RetryDuration: time.Second},
			net.SubscribeOpts{},
			eth2.ConfigOpts{
				HandleCfg: false,
				Checkers: []eth2.CfgChecker{
					{Key: eth2.Execution, ErrMsg: eth2.NoExecutionFoundError, Data: executionEndp},
					{Key: eth2.Consensus, ErrMsg: eth2.NoConsensusFoundError, Data: consensusEndp},
				},
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		done := make(chan struct{})
		// listen for SIGINT
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)

		results := monitor.TrackSync(done, consensusEndp, executionEndp, time.Duration(cron)*time.Second)

		go func() {
			for r := range results {
				if r.Error != nil {
					log.Errorf("Endpoint %s returned an error. Error: %v", r.Endpoint, r.Error)
				}
			}
		}()

		for range sigChan {
			log.Info("Received SIGINT, exiting...")
			close(done)
			time.Sleep(time.Second)
			os.Exit(0)
		}
	},
}

func init() {
	//ethereumCmd.AddCommand(trackSyncCmd)

	// Flags
	TrackSyncCmd.Flags().StringSliceVar(&executionEndp, "execution", []string{}, "Execution endpoints to which track sync progress. Example: 'posmoni ethereum --execution=<endpoint1>,<endpoint2>'")
	TrackSyncCmd.Flags().StringSliceVar(&consensusEndp, "consensus", []string{}, "Consensus endpoints to which track sync progress. Example: 'posmoni ethereum --consensus=<endpoint1>,<endpoint2>'")
	TrackSyncCmd.Flags().IntVarP(&cron, "cron", "c", 60, "Wait time in seconds between sync progress checks")
}
