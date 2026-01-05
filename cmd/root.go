/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"ffiii-tui/internal/firefly"
	"ffiii-tui/internal/logging"
	"ffiii-tui/internal/ui"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ffiii-tui",
	Short: "A TUI for Firefly III personal finance manager",
	Long: `ffiii-tui is a terminal user interface (TUI) application that connects to your Firefly III personal finance manager via its API:
It allows you to view and manage your financial data directly from the terminal.

Prerequisites:
  - A running instance of Firefly III with API access enabled.
  - An API key generated from your Firefly III user settings. `,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initializeConfig(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		debug := viper.GetBool("logging.debug")
		logFile := viper.GetString("logging.file")

		if debug {
			fmt.Println("Debug logging is enabled")
		}

		logger, cleanup, err := logging.New(debug, logFile)
		if err != nil {
			return fmt.Errorf("failed to init logger: %w", err)
		}
		defer cleanup()

		zap.ReplaceGlobals(logger)

		apiKey := viper.GetString("firefly.api_key")
		if apiKey == "" {
			return fmt.Errorf("firefly API key is not set")
		}

		apiUrl := viper.GetString("firefly.api_url")
		if apiUrl == "" {
			return fmt.Errorf("firefly API URL is not set")
		}

		ff, err := firefly.NewApi(firefly.ApiConfig{
			ApiKey:         apiKey,
			ApiUrl:         apiUrl,
			TimeoutSeconds: 10,
		})
		if err != nil {
			return fmt.Errorf("failed to connect to Firefly III: %w", err)
		}

		logger.Info("Connected to Firefly III", zap.String("api_url", apiUrl), zap.String("user", ff.User.Email))

		ui.Show(ff)

		viper.Set("logging.debug", false)

		return viper.WriteConfigAs(viper.ConfigFileUsed())
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
	rootCmd.Flags().BoolP("logging.debug", "d", false, "Enable debug logging")
	rootCmd.Flags().StringP("logging.file", "l", "messages.log", "Log file path (if empty, logs to stdout)")

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
			return fmt.Errorf("Config file not found, %s", err.Error())
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
