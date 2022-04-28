package networking

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/NethermindEth/posmoni/internal/utils"
	log "github.com/sirupsen/logrus"
)

// BeaconClient : Struct BeaconAPI interface implementation
type BeaconClient struct {
	// Beacon node endpoint to connect to. Probably needs to change when support for several endpoints is added
	Endpoint string
	// Time between retries when a request fails
	RetryDuration time.Duration
}

/*
SetEndpoints :
Set the endpoints for the beacon client implementation.

params :-
a. endpoints []string
Endpoints to set for the beacon client

returns :-
none
*/
func (bc *BeaconClient) SetEndpoints(endpoints []string) {
	// TODO: Update when support for several endpoints is made
	// notest
	// ^^ not test covered for now
	bc.Endpoint = endpoints[0]
}

/*
ValidatorBalances :
Get the validator balances for the given checkpoint.

params :-
a. stateID string
Blockchain state ID from when to get the balances
b. validatorIdxs []string
Validator indexes to get the balances for

returns :-
a. []ValidatorBalance
Validator balances fetched from the beacon node
b. error
Error if any
*/
func (bc *BeaconClient) ValidatorBalances(stateID string, validatorIdxs []string) ([]ValidatorBalance, error) {
	// notest
	idxs := strings.Join(validatorIdxs, ",")
	// http://<endpoint>/eth/v1/beacon/states/<stateID>/validator_balances?id=1,2,3
	url := fmt.Sprintf("%s%s%s%s?id=%s", bc.Endpoint, "/eth/v1/beacon/states/", stateID, "/validator_balances", idxs)

	resp, err := utils.GetRequest(url, bc.RetryDuration)
	if err != nil {
		return nil, fmt.Errorf(RequestFailedError, url, err)
	}

	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(ReadBodyError, err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(BadResponseError, url, resp.StatusCode, string(contents))
	}

	var balances ValidatorBalanceList
	balances, err = unmarshalData(contents, balances)
	if err != nil {
		return nil, err
	}

	return balances.Data, nil
}

/*
Health :
Health check to the given endpoints using the API method '/eth/v1/beacon/health'.

params :-
a. endpoints []string
Endpoints to check

returns :-
a. []HealthResponse
Health responses from the given endpoints
*/
func (bc *BeaconClient) Health(endpoints []string) []HealthResponse {
	if len(endpoints) == 0 {
		log.Warn("No endpoints provided for health check")
		return nil
	}

	ch := make(chan HealthResponse, len(endpoints))
	defer close(ch)

	for _, endpoint := range endpoints {
		go func(endpoint string) {
			url := fmt.Sprintf("%s%s", endpoint, "/eth/v1/beacon/health")
			resp, err := utils.GetRequest(url, bc.RetryDuration)
			if err != nil {
				ch <- HealthResponse{Endpoint: endpoint, Healthy: false, Error: err}
				return
			}

			switch resp.StatusCode {
			case 200:
				ch <- HealthResponse{Endpoint: endpoint, Healthy: true, Error: nil}
			case 206:
				ch <- HealthResponse{Endpoint: endpoint, Healthy: false, Error: fmt.Errorf(BadResponseError, url, resp.StatusCode, "Node is syncing but can serve incomplete data")}
			case 503:
				ch <- HealthResponse{Endpoint: endpoint, Healthy: false, Error: fmt.Errorf(BadResponseError, url, resp.StatusCode, "Node not initialized or having issues")}
			default:
				ch <- HealthResponse{Endpoint: endpoint, Healthy: false, Error: fmt.Errorf(BadResponseError, url, resp.StatusCode, "")}
			}

		}(endpoint)
	}

	responses := make([]HealthResponse, 0)
	for i := 0; i < len(endpoints); i++ {
		responses = append(responses, <-ch)
	}

	return responses
}

/*
SyncStatus :
Check sync status of the given endpoints using the API method '/eth/v1/node/syncing'.

params :-
a. endpoints []string
Endpoints to check

returns :-
a. []HealthResponse
Health responses from the given endpoints
*/
func (bc *BeaconClient) SyncStatus(endpoints []string) []SyncingStatus {
	if len(endpoints) == 0 {
		log.Warn("No endpoints provided for health check")
		return nil
	}

	type status struct {
		SyncingStatus
		err error
	}

	ch := make(chan status, len(endpoints))
	defer close(ch)

	for _, endpoint := range endpoints {
		go func(endpoint string) {
			url := fmt.Sprintf("%s%s", endpoint, "/eth/v1/node/syncing")
			resp, err := utils.GetRequest(url, bc.RetryDuration)
			if err != nil {
				ch <- status{err: err}
				return
			}

			defer resp.Body.Close()
			contents, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ch <- status{err: fmt.Errorf(ReadBodyError, err)}
				return
			}

			if resp.StatusCode != 200 {
				ch <- status{err: fmt.Errorf(BadResponseError, url, resp.StatusCode, string(contents))}
				return
			}

			var ssr SyncingStatusResponse
			ssr, err = unmarshalData(contents, ssr)
			if err != nil {
				ch <- status{err: err}
				return
			}

			ch <- status{SyncingStatus: ssr.Data, err: nil}

		}(endpoint)
	}

	responses := make([]SyncingStatus, 0)
	for i := 0; i < len(endpoints); i++ {
		resp := <-ch
		ss := SyncingStatus{}

		if resp.err != nil {
			ss.Error = resp.err
		} else {
			ss = resp.SyncingStatus
		}

		responses = append(responses, ss)
	}

	return responses
}
