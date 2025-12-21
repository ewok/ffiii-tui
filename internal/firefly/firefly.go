/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import "fmt"

// ApiConfig holds configuration for the Firefly III API.
type Api struct {
	// Config contains the API configuration details.
	Config ApiConfig

	// Assets holds the list of asset accounts.
	Assets []Asset

	// Expenses holds the list of expense accounts.
	Expenses []Expense

	// Liabilities holds the list of liability accounts.
	Liabilities []Liability

	// Revenues holds the list of revenue accounts.
	Revenues []Revenue

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
	if err := api.UpdateAssets(); err != nil {
		fmt.Println("Error updating assets:", err)
	}
	if err := api.UpdateExpenses(); err != nil {
		fmt.Println("Error updating expenses:", err)
	}
	if err := api.UpdateLiabilities(); err != nil {
		fmt.Println("Error updating liabilities:", err)
	}
	if err := api.UpdateRevenues(); err != nil {
		fmt.Println("Error updating revenues:", err)
	}
	if err := api.UpdateCategories(); err != nil {
		fmt.Println("Error updating categories:", err)
	}
	if err := api.UpdateCurrencies(); err != nil {
		fmt.Println("Error updating currencies:", err)
	}

	return api
}
