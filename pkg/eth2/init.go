package eth2

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// CfgChecker : Struct Handle data loading for a given configuration key
type CfgChecker struct {
	// Configuration key
	Key string
	// Error message to throw in case the checks fail
	ErrMsg string
	// Configuration data. Should be pre-populated when is not desired to use config file or enviroment variables to get configuration data for 'Key'.
	Data []string
}

/*
checker :
Handle configuration data setup for a given configuration key. If 'Data' field is already filled, then no data is get from configuration file or environment variables is done.

params :-
none

returns :-
a. error
Error if any
*/
func (cc *CfgChecker) checker() error {
	if len(cc.Data) == 0 {
		d, err := checkVariable(cc.Key, cc.ErrMsg)
		if err != nil {
			return err
		}
		cc.Data = d
	}
	return nil
}

/*
Init :
Get monitor configuration data from config file or enviroment variables. If a configuration struct is given with data on it, then this data won't be overrided, it will look for missing data.

params :-
a. checkers []CfgChecker
Configuration data to search for

returns :-
a. eth2Config
Monitor configuration data
b. error
Error if any
*/
func Init(checkers []CfgChecker) (cfg eth2Config, err error) {
	fmt.Println("Initializing configuration")
	viper.SetEnvPrefix("PM")

	for _, c := range checkers {
		viper.BindEnv(c.Key)
		if err := c.checker(); err != nil {
			return cfg, err
		}

		switch c.Key {
		case Execution:
			cfg.execution = c.Data
		case Consensus:
			cfg.consensus = c.Data
		case Validators:
			cfg.validators = c.Data
		default:
			// execution should never go here, checker() should fail if an invalid key was provided
			return cfg, fmt.Errorf(InvalidConfigKeyError, c.Key, []string{Execution, Consensus, Validators})
		}
	}

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
