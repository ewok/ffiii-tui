/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

// ApiConfig holds configuration for the Firefly III API.
type Api struct {
	// Config contains the API configuration details.
	Config ApiConfig
}

// NewApi creates a new Api instance with the provided configuration.
// Parameters:
//   - config: an ApiConfig struct containing the API configuration details.
//
// Returns:
//   - A pointer to an Api struct initialized with the provided configuration.
func NewApi(config ApiConfig) *Api {
	return &Api{
		Config: config,
	}
}

