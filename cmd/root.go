// Package cmd implements the CLI for booker using cobra and viper.
package cmd

import (
	"fmt"
	"os"

	"booker/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// envKeys are the environment variables that config.Default() reads via os.Getenv.
// We push viper values into the OS environment so the config package picks them up.
var envKeys = []string{
	config.EnvKiwiAPIKey,
	config.EnvSerpAPIKey,
	config.EnvOpenAIAPIKey,
	config.EnvAnumaAPIKey,
}

var rootCmd = &cobra.Command{
	Use:   "booker",
	Short: "Flight search agent with multi-city stopover support",
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	// Silently ignore if .env doesn't exist.
	_ = viper.ReadInConfig()

	// Push .env values into OS environment so config.Default() can read them.
	for _, key := range envKeys {
		if val := viper.GetString(key); val != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}
