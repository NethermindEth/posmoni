package eth2

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func Init() (cfg eth2Config, err error) {
	fmt.Println("Initializing configuration")
	viper.SetEnvPrefix("PGM")
	viper.BindEnv(validators)
	viper.BindEnv(consensus)

	v, err := checkVariable(validators, NoValidatorsFoundError)
	if err != nil {
		return cfg, err
	}

	c, err := checkVariable(consensus, NoConsensusFoundError)
	if err != nil {
		return cfg, err
	}

	cfg.Validators = v
	cfg.Consensus = c

	return
}

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
