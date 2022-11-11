package cli

import (
	"github.com/NethermindEth/posmoni/pkg/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
)

var (
	port int64
)

// serverCmd represents the api command
var serverCmd = &cobra.Command{
	Use:   "api",
	Short: "Starts a api to monitor validators of Ethereum Beacon Chain",
	Long: `
Starts a api to monitor Ethereum Beacon Chain validator's balance changes and missed attestations using Beacon Chain official HTTP API.
`,
	Run: func(cmd *cobra.Command, args []string) {
		ExecuteServer()
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)

	serverCmd.Flags().Int64Var(&port, "port", 8080, "Port to listen to")
}

func ExecuteServer() {
	// Execute api
	server := api.Server{}
	server.Initialize()
	log.Info("Starting server")
	server.Run(":" + strconv.FormatInt(port, 10))
}
