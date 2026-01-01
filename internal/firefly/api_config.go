/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

// ApiConfig holds configuration for the Firefly III API.
type ApiConfig struct {
	// ApiKey is the API key used for authentication.
	ApiKey string
	// ApiUrl is the base URL of the Firefly III API.
	ApiUrl string
	// TimeoutSeconds specifies the timeout for API requests in seconds.
	TimeoutSeconds int
}
