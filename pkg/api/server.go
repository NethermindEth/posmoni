package api

import (
	"encoding/json"
	"github.com/NethermindEth/posmoni/pkg/eth2"
	"github.com/NethermindEth/posmoni/pkg/eth2/db"
	net "github.com/NethermindEth/posmoni/pkg/eth2/networking"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var (
	upgrader = websocket.Upgrader{}
)

type Server struct {
	Router *mux.Router
}

// Initialize the Server
func (s *Server) Initialize() {
	s.Router = mux.NewRouter()
	s.initializeRoutes()
}

func (s *Server) Run(address string) {
	log.Fatal(http.ListenAndServe(address, s.Router))
}

type info struct {
	ConsensusUrls []string      `json:"consensus_urls"`
	ExecutionUrls []string      `json:"execution_urls"`
	Wait          time.Duration `json:"wait"`
}

// trackSync handle incoming request to track sync
func (s *Server) trackSync(w http.ResponseWriter, r *http.Request) {
	//// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade failed: ", err)
		return
	}
	defer conn.Close()

	_, message, err := conn.ReadMessage()
	if err != nil {
		return
	}
	var infoVal info
	err = json.Unmarshal(message, &infoVal)
	if err != nil {
		return
	}
	monitor, err := eth2.NewEth2Monitor(
		db.EmptyRepository{},
		&net.BeaconClient{RetryDuration: time.Second},
		&net.ExecutionClient{RetryDuration: time.Second},
		net.SubscribeOpts{},
		eth2.ConfigOpts{
			HandleCfg: false,
			Checkers: []eth2.CfgChecker{
				{Key: eth2.Execution, ErrMsg: eth2.NoExecutionFoundError, Data: infoVal.ExecutionUrls},
				{Key: eth2.Consensus, ErrMsg: eth2.NoConsensusFoundError, Data: infoVal.ConsensusUrls},
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan struct{})
	results := monitor.TrackSync(done, infoVal.ConsensusUrls, infoVal.ExecutionUrls, infoVal.Wait)
	go func() {
		messageType, _, _ := conn.ReadMessage()
		switch messageType {
		case websocket.CloseMessage:
			close(done)
		}
	}()
	for r := range results {
		if r.Error != nil {
			log.Errorf("Endpoint %s returned an error. Error: %v", r.Endpoint, r.Error)
		}
		msg, err := json.Marshal(r)
		if err != nil {
			return
		}
		err = conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}

func (s *Server) initializeRoutes() {
	s.Router.HandleFunc("/trackSync", s.trackSync)
}
