/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import (
	"fmt"
	"maps"
	"strconv"
	"strings"
)

const accountsEndpoint = "%s/accounts?page=%d&type=%s"

type Account struct {
	ID           string
	Name         string
	CurrencyCode string
	Balance      float64
	Type         string
}

type apiAccount struct {
	ID         string         `json:"id"`
	Attributes apiAccountAttr `json:"attributes"`
}

type apiAccountAttr struct {
	Active         bool   `json:"active"`
	Name           string `json:"name"`
	CurrencyCode   string `json:"currency_code"`
	CurrentBalance string `json:"current_balance"`
	Type           string `json:"type"`
}

type NewLiability struct {
	Name         string `json:"name"`
	CurrencyCode string `json:"currency_code"`
	Type         string `json:"liability_type"`
	Direction    string `json:"liability_direction"`
}

func (api *Api) CreateAssetAccount(name string, currencyCode string) error {
	return api.createAccount(map[string]any{
		"name":              name,
		"type":              "asset",
		"currency_code":     strings.ToUpper(currencyCode),
		"include_net_worth": true,
		"active":            true,
		"account_role":      "defaultAsset",
	})
}

func (api *Api) CreateExpenseAccount(name string) error {
	return api.createAccount(map[string]any{
		"name": name,
		"type": "expense",
	})
}

func (api *Api) CreateRevenueAccount(name string) error {
	return api.createAccount(map[string]any{
		"name": name,
		"type": "revenue",
	})
}

func (api *Api) CreateLiabilityAccount(nl NewLiability) error {
	return api.createAccount(map[string]any{
		"name":                nl.Name,
		"type":                "liability",
		"currency_code":       strings.ToUpper(nl.CurrencyCode),
		"liability_type":      nl.Type,
		"liability_direction": nl.Direction,
	})
}

func (api *Api) createAccount(payload map[string]any) error {
	endpoint := fmt.Sprintf("%s/accounts", api.Config.ApiUrl)
	response, err := api.postRequest(endpoint, payload)
	if err != nil {
		return err
	}

	data, ok := response.Data.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid response format: missing data field")
	}
	id, ok := data["id"].(string)
	if !ok || id == "" {
		return fmt.Errorf("invalid response format: missing transaction id")
	}

	return nil
}

func (api *Api) GetExpenseDiff(ID string) float64 {
	if insight, ok := api.expenseInsights[ID]; ok {
		return insight.Diff
	}
	return 0
}

func (api *Api) GetRevenueDiff(ID string) float64 {
	if insight, ok := api.revenueInsights[ID]; ok {
		return insight.Diff
	}
	return 0
}

func (api *Api) UpdateExpenseInsights() error {
	// TODO: Need error reporting
	insights := make(map[string]accountInsight)
	spentInsights, err := api.GetInsights("expense/expense")
	if err == nil {
		for _, item := range spentInsights {
			insights[item.ID] = accountInsight{
				Diff: (-1) * item.DifferenceFloat,
			}
		}
	}
	api.expenseInsights = insights

	return nil
}

func (api *Api) UpdateRevenueInsights() error {
	insights := make(map[string]accountInsight)
	earnedInsights, err := api.GetInsights("income/revenue")
	if err == nil {
		for _, item := range earnedInsights {
			insights[item.ID] = accountInsight{
				Diff: item.DifferenceFloat,
			}
		}
	}

	api.revenueInsights = insights

	return nil
}

func (api *Api) UpdateAccounts(accType string) error {
	accounts, err := api.ListAccounts(accType)
	if err != nil {
		return err
	}

	accs := make(map[string][]Account, 0)

	for _, account := range accounts {
		balance, err := strconv.ParseFloat(account.Attributes.CurrentBalance, 64)
		if err != nil {
			balance = 0.0
		}

		accs[account.Attributes.Type] = append(accs[account.Attributes.Type], Account{
			ID:           account.ID,
			Name:         account.Attributes.Name,
			CurrencyCode: account.Attributes.CurrencyCode,
			Balance:      balance,
			Type:         account.Attributes.Type,
		})
	}

	maps.Copy(api.Accounts, accs)

	switch accType {
	case "expense":
		api.UpdateExpenseInsights()
	case "revenue":
		api.UpdateRevenueInsights()
	case "all":
		api.UpdateExpenseInsights()
		api.UpdateRevenueInsights()
	}

	return nil
}

func (api *Api) ListAccounts(accountType string) ([]apiAccount, error) {
	allData, err := api.fetchPaginated("%s/accounts?type=%s&page=%d",
		api.Config.ApiUrl,
		accountType)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch paginated accounts: %v", err)
	}
	accs, err := unmarshalItems[apiAccount](allData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal accounts: %v", err)
	}
	return accs, nil
}

// TODO: Optimize search with a map
func (api *Api) GetAccountByID(ID string) Account {
	for _, groups := range api.Accounts {
		for _, account := range groups {
			if account.ID == ID {
				return account
			}
		}
	}
	return Account{}
}
