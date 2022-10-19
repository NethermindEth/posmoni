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
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/NethermindEth/posmoni/configs"
)

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "posmoni",
	Short: "A tool to monitor validators in a PoS network of chain, using official HTTP APIs.",
	Long:  `A tool to monitor validators in a PoS network of chain, using official HTTP APIs. By default runs Ethereum Beacon Chain monitor, equivalent to 'posmoni ethereum'. Run 'posmoni ethereum --help' for more description about command usage.`,
	Run: func(cmd *cobra.Command, args []string) {
		ExecuteEthMonitor()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(configs.InitConfig)

	RootCmd.PersistentFlags().StringVar(&configs.CfgFile, "config", "", "config file (default is $HOME/.posmoni.yaml)")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":2112", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()
}
