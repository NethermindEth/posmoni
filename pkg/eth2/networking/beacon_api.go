package networking

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/NethermindEth/posgonitor/internal/utils"
)

type BeaconClient struct {
	Endpoint      string
	RetryDuration time.Duration
}

func (bc *BeaconClient) SetEndpoints(endpoints []string) {
	// TODO: Update when support for several endpoints is made
	// notest
	// ^^ not test covered for now
	bc.Endpoint = endpoints[0]
}

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
