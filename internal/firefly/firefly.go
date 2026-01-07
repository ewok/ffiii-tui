/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import (
	"fmt"
	"time"
)

// ApiConfig holds configuration for the Firefly III API.
type Api struct {
	// Config contains the API configuration details.
	Config ApiConfig

	Accounts        map[string][]Account
	cashAccount     Account

	expenseInsights map[string]accountInsight
	revenueInsights map[string]accountInsight

	// Categories holds the list of categories.
	Categories       []Category
	categoryInsights map[string]categoryInsight

	// Currencies
	Currencies []Currency

	// User
	User User

	// Date range
	StartDate time.Time
	EndDate   time.Time
}

// NewApi creates a new Api instance with the provided configuration.
// Parameters:
//   - config: an ApiConfig struct containing the API configuration details.
//
// Returns:
//   - A pointer to an Api struct initialized with the provided configuration.
func NewApi(config ApiConfig) (*Api, error) {
	api := &Api{Config: config}

	api.StartDate = time.Now().AddDate(0, 0, -time.Now().Day()+1)
	api.EndDate = time.Now().AddDate(0, 1, -time.Now().Day())

	// Test connection and get current user
	userEmail, err := api.GetCurrentUser()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Firefly III: %w", err)
	}
	api.User = User{
		Email: userEmail,
	}

	api.Accounts = make(map[string][]Account, 0)

	api.UpdateAccounts("special")
	api.UpdateCurrencies()

	return api, nil
}

func (api *Api) PreviousPeriod() {
	api.StartDate = time.Date(api.StartDate.Year(), api.StartDate.Month()-1, 1, 0, 0, 0, 0, api.StartDate.Location())
	api.EndDate = api.StartDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
}

func (api *Api) NextPeriod() {
	api.StartDate = time.Date(api.StartDate.Year(), api.StartDate.Month()+1, 1, 0, 0, 0, 0, api.StartDate.Location())
	api.EndDate = api.StartDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
}
