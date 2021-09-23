package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultConfigName = "." + appName
	defaultConfigType = "yaml"
)

// initConfig reads in config file and ENV variables if set.
func initConfig() (string, error) {
	if flags.Config != "" {
		// Use config file from the flag.
		viper.SetConfigFile(flags.Config)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory.
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.SetConfigType(defaultConfigType)
		viper.SetConfigName(defaultConfigName)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()

	return viper.ConfigFileUsed(), err
}
