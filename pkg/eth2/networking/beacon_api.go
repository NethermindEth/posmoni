package networking

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/NethermindEth/posmoni/internal/utils"
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
