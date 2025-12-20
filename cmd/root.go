/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"ffiii-tui/internal/ui"
	"ffiii-tui/internal/firefly"
)

var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ffiii-tui",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		apiKey := viper.GetString("firefly.api_key")
		if apiKey == "" {
			return fmt.Errorf("firefly API key is not set")
		}

		apiUrl := viper.GetString("firefly.api_url")
		if apiUrl == "" {
			return fmt.Errorf("firefly API URL is not set")
		}

		fireflyApi := firefly.NewApi(firefly.ApiConfig{
			ApiKey:         apiKey,
			ApiUrl:         apiUrl,
			TimeoutSeconds: 10,
		})

        transactions := []firefly.Transaction{}
        page := 1
        for {
            // txs, err := fireflyApi.SearchTransactions(page, 20, "taxi")
            txs, err := fireflyApi.ListTransactions(page, 20, "", "")
            if err != nil {
                return err
            }
            if len(txs) == 0 {
                break
            }
            transactions = append(transactions, txs...)
            page++
        }

		ui.Show(transactions, fireflyApi)

		return nil
	},
}

var initConfigCmd = &cobra.Command{
	Use:   "init-config",
	Short: "Generate a default configuration file",
	Long:  `Generate a default configuration file for ffiii-tui.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		initViper := viper.New()
		initViper.Set("firefly.api_key", viper.GetString("firefly.api_key"))
		initViper.Set("firefly.api_url", viper.GetString("firefly.api_url"))

		initViper.AddConfigPath(".")
		initViper.SetConfigName("config")
		initViper.SetConfigType("yaml")
		initViper.SetConfigFile("./config.yaml")

		err := initViper.SafeWriteConfig()
		if err != nil {
			var configFileAlreadyExistsError viper.ConfigFileAlreadyExistsError
			if errors.As(err, &configFileAlreadyExistsError) {
				return err
			}
		}

		fmt.Println("Configuration file created at:", initViper.ConfigFileUsed())
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/ffiii-tui/config)")
	rootCmd.PersistentFlags().StringP("firefly.api_key", "k", "your_firefly_api_key_here", "Firefly III API key")
	rootCmd.PersistentFlags().StringP("firefly.api_url", "u", "https://your-firefly-iii-instance.com/api/v1", "Firefly III API URL")

	rootCmd.AddCommand(initConfigCmd)
}

func initializeConfig(cmd *cobra.Command) error {

	viper.SetEnvPrefix("FFIII_TUI")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "*", "-", "*"))
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(".")
		viper.AddConfigPath(home + "/.config/ffiii-tui")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}
	} else {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}

	return nil
}
