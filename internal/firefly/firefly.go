/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import (
	"fmt"
	"sync"
)

// ApiConfig holds configuration for the Firefly III API.
type Api struct {
	// Config contains the API configuration details.
	Config ApiConfig

	// Assets holds the list of asset accounts.
	Assets []Account

	// Expenses holds the list of expense accounts.
	Expenses []Account

	// Liabilities holds the list of liability accounts.
	Liabilities []Account

	// Revenues holds the list of revenue accounts.
	Revenues []Account

	// Categories holds the list of categories.
	Categories []Category

	// Currencies
	Currencies []Currency
}

// NewApi creates a new Api instance with the provided configuration.
// Parameters:
//   - config: an ApiConfig struct containing the API configuration details.
//
// Returns:
//   - A pointer to an Api struct initialized with the provided configuration.
func NewApi(config ApiConfig) *Api {
	api := &Api{Config: config}

	// Initial data fetch
	var wg sync.WaitGroup
	updateFuncs := []func() error{
		api.UpdateAssets,
		api.UpdateExpenses,
		api.UpdateLiabilities,
		api.UpdateRevenues,
		api.UpdateCategories,
		api.UpdateCurrencies,
	}

	wg.Add(len(updateFuncs))

	for _, updateFunc := range updateFuncs {
		go func(f func() error) {
			defer wg.Done()
			if err := f(); err != nil {
				fmt.Println("Error during update:", err)
			}
		}(updateFunc)
	}

	wg.Wait()

	return api
}
