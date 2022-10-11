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
package cli

import (
	"os"
	"os/signal"
	"time"

	"github.com/NethermindEth/posmoni/cli/eth"
	"github.com/NethermindEth/posmoni/pkg/eth2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ethereumCmd represents the ethereum command
var ethereumCmd = &cobra.Command{
	Use:   "ethereum",
	Short: "Monitor validators of Ethereum Beacon Chain",
	Long: `
Monitor Ethereum Beacon Chain validator's balance changes and missed attestations using Beacon Chain official HTTP API.

Needs a consensus client endpoint to interacts with Beacon Chain API and a set of validator addresses or public indexes to monitor. All of these endpoints should be provided using a configuration file or environment variables.

The configuration file must be a .yaml. By default posmoni searches for a .posmoni.yaml file at the HOME directory. Example of configuration file:

validators: [269870, 0xb3456c17df6d9bddab9dedfcc590bbebccd24eca811099ad4b10f0fcd7583c91e160848713d4bb5c23ab1eeae9c9b3c0]
consensus: "http://111.111.111.111:5052"

logs:
logLevel: debug

Example of environment variables:
"PM_VALIDATORS": "269870,0xb3456c17df6d9bddab9dedfcc590bbebccd24eca811099ad4b10f0fcd7583c91e160848713d4bb5c23ab1eeae9c9b3c0",
"PM_CONSENSUS":  "http://111.111.111.111:5052"
  `,
	Run: func(cmd *cobra.Command, args []string) {
		ExecuteEthMonitor()
	},
}

func init() {
	RootCmd.AddCommand(ethereumCmd)
	ethereumCmd.AddCommand(eth.TrackSyncCmd)
}

func ExecuteEthMonitor() {
	// listen for SIGINT
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	monitor, err := eth2.DefaultEth2Monitor(eth2.ConfigOpts{HandleCfg: false})
	if err != nil {
		log.Fatal(err)
	}
	doneChans, err := monitor.Monitor()
	if err != nil {
		log.Fatal(err)
	}

	for range sigChan {
		log.Info("Received SIGINT, exiting...")
		for _, done := range doneChans {
			close(done)
		}
		time.Sleep(time.Second)
		os.Exit(0)
	}
}
