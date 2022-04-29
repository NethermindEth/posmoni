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

// notest
import (
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"

	"github.com/NethermindEth/posmoni/configs"
	"github.com/NethermindEth/posmoni/pkg/eth2"
	log "github.com/sirupsen/logrus"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "posmoni [flags]",
	Short: "A tool to monitor validators in a PoS network of chain, using official HTTP APIs.",
	Long:  `A tool to monitor validators in a PoS network of chain, using official HTTP APIs.`,
	Run: func(cmd *cobra.Command, args []string) {
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
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(configs.InitConfig)

	rootCmd.PersistentFlags().StringVar(&configs.CfgFile, "config", "", "config file (default is $HOME/.posmoni.yaml)")
}
