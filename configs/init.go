package configs

// notest
import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

var CfgFile string

// initConfig reads in config file and ENV variables if set.
func InitConfig() {
	if CfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(CfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".posmoni" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".posmoni")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	InitLogging()
}

/*
InitLogging :
This function is responsible for
initializing the logging configurations

params :-
none

returns :-
none
*/
func InitLogging() {
	var config logConfig

	log.SetFormatter(&nested.Formatter{
		HideKeys:        true,
		FieldsOrder:     []string{Component},
		TimestampFormat: "2006-01-02 15:04:05 --",
	})

	err := viper.UnmarshalKey("logs", &config)
	if err != nil {
		log.WithField(Component, "Logger Init").Errorf("Unable to decode into struct, %v", err)
		return
	}
	log.WithField(Component, "Logger Init").Infof("Logging configuration: %+v", config)

	level, err := log.ParseLevel(strings.ToLower(config.Level))
	if err != nil {
		log.WithField(Component, "Logger Init").Error(err)
		return
	}
	log.SetLevel(level)
}
