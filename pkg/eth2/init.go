package eth2

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

/*
Init :
Get monitor configuration data from config file or enviroment variables.

params :-
none

returns :-
a. eth2Config
Monitor configuration data
b. error
Error if any
*/
func Init() (cfg eth2Config, err error) {
	fmt.Println("Initializing configuration")
	viper.SetEnvPrefix("PGM")
	viper.BindEnv(validators)
	viper.BindEnv(consensus)

	// DEV: Should we validate validators provided? in that case it should be a valid validator index or an address
	v, err := checkVariable(validators, NoValidatorsFoundError)
	if err != nil {
		return cfg, err
	}

	// DEV: Should we validate consensus nodes provided? in that case it should be a valid endpoint
	c, err := checkVariable(consensus, NoConsensusFoundError)
	if err != nil {
		return cfg, err
	}

	cfg.Validators = v
	cfg.Consensus = c

	return
}

/*
checkVariable :
Check if a variable is set in config file or enviroment variables. Variables values can be a yaml list or a list in form of a string.

params :-
a. key string
Variable name
b. errMsg string
Error message to be returned if variable is not set

returns :-
a. data []string
Variable value
b. error
Error if any
*/
func checkVariable(key, errMsg string) (data []string, err error) {
	var ok bool
	tmp := viper.Get(key)

	if _, ok = tmp.(string); ok {
		data = strings.Split(viper.GetString(key), ",")
	} else {
		data = viper.GetStringSlice(key)
	}

	if len(data) == 0 {
		return data, fmt.Errorf(errMsg)
	}

	return
}
