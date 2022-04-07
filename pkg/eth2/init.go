package eth2

import (
	"fmt"

	"github.com/spf13/viper"
)

func Init() (cfg eth2Config, err error) {
	viper.SetEnvPrefix("PGM")

	cfg.Validators = viper.GetStringSlice(validators)
	if len(cfg.Validators) == 0 {
		return cfg, fmt.Errorf(NoValidatorsFoundError)
	}
	cfg.Consensus = viper.GetStringSlice(consensus)
	if len(cfg.Consensus) == 0 {
		return cfg, fmt.Errorf(NoConsensusFoundError)
	}

	return
}
