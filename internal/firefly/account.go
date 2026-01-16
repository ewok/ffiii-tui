/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import (
	"fmt"
	"maps"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Account struct {
	ID           string
	Name         string
	CurrencyCode string
	Type         string
}

type apiAccount struct {
	ID         string         `json:"id"`
	Attributes apiAccountAttr `json:"attributes"`
}

type apiAccountAttr struct {
	Active         bool    `json:"active"`
	Name           string  `json:"name"`
	CurrencyCode   string  `json:"currency_code"`
	CurrentBalance float64 `json:"current_balance,string"`
	Type           string  `json:"type"`
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

func (api *Api) GetTotalExpenseDiff() float64 {
	total := 0.0
	for _, insight := range api.expenseInsights {
		total += insight.Diff
	}
	return total
}

// GetTotalExpenseDiff2 gets total expense difference per currency
// This is an alternative to GetTotalExpenseDiff which provides per-currency totals
// instead of a single aggregated total
// Note: This function fetches insights directly from the API each time it is called
// and does not use cached insights
// which may have performance implications
func (api *Api) GetTotalExpenseDiff2() (totals []struct {
	CurrencyCode string
	Diff         float64
},
) {
	spentInsights, err := api.GetInsights("expense/total")
	if err == nil {
		for _, item := range spentInsights {
			totals = append(totals, struct {
				CurrencyCode string
				Diff         float64
			}{
				CurrencyCode: item.CurrencyCode,
				Diff:         (-1) * item.DifferenceFloat,
			})
			zap.S().Debugf("Expense total insight: diff=%f, currency=%s", item.DifferenceFloat, item.CurrencyCode)
		}
	}
	return
}

func (api *Api) GetRevenueDiff(ID string) float64 {
	if insight, ok := api.revenueInsights[ID]; ok {
		return insight.Diff
	}
	return 0
}

func (api *Api) GetTotalRevenueDiff() float64 {
	total := 0.0
	for _, insight := range api.revenueInsights {
		total += insight.Diff
	}
	return total
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
		api.accountBalances[account.ID] = account.Attributes.CurrentBalance
		accs[account.Attributes.Type] = append(accs[account.Attributes.Type], Account{
			ID:           account.ID,
			Name:         account.Attributes.Name,
			CurrencyCode: account.Attributes.CurrencyCode,
			Type:         account.Attributes.Type,
		})
	}

	maps.Copy(api.Accounts, accs)

	switch accType {
	case "expense":
		api.Accounts["expense"] = append(api.Accounts["expense"], api.CashAccount())
		err := api.UpdateExpenseInsights()
		if err != nil {
			return fmt.Errorf("failed to update expense insights: %v", err)
		}
	case "revenue":
		err := api.UpdateRevenueInsights()
		if err != nil {
			return fmt.Errorf("failed to update revenue insights: %v", err)
		}
	case "all":
		api.Accounts["expense"] = append(api.Accounts["expense"], api.CashAccount())
		errs := []error{}
		err1 := api.UpdateExpenseInsights()
		if err1 != nil {
			errs = append(errs, fmt.Errorf("failed to update expense insights: %v", err1))
		}
		err2 := api.UpdateRevenueInsights()
		if err2 != nil {
			errs = append(errs, fmt.Errorf("failed to update revenue insights: %v", err2))
		}
		if len(errs) > 0 {
			return fmt.Errorf("multiple errors: %v", errs)
		}
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
	const retryLimit = 10
	const retryDelay = 1 * time.Second

	var account Account
	for attempt := 1; attempt <= retryLimit; attempt++ {
		for _, groups := range api.Accounts {
			for _, acc := range groups {
				if acc.ID == ID {
					return acc
				}
			}
		}

		if attempt < retryLimit {
			time.Sleep(retryDelay)
		}
	}

	return account
}

func (api *Api) CashAccount() Account {
	if api.cashAccount != (Account{}) {
		return api.cashAccount
	}

	accounts, err := api.ListAccounts("special")
	if err != nil {
		zap.S().Errorf("Failed to fetch special accounts for cash account: %v", err)
		return Account{}
	}

	for _, account := range accounts {
		if account.Attributes.Type == "cash" {
			cash := Account{
				ID:   account.ID,
				Name: account.Attributes.Name,
				Type: account.Attributes.Type,
			}
			api.cashAccount = cash
			zap.S().Debugf("Using cash account: %s (%s)", cash.Name, cash.ID)
			return cash
		}
	}
	zap.S().Error("No asset accounts available to use as cash account")
	return Account{}
}

// AccountsByType returns the cached accounts for the given type.
// It returns a copy of the slice to avoid accidental mutation by callers.
func (api *Api) AccountsByType(accountType string) []Account {
	accounts := api.Accounts[accountType]
	return append([]Account(nil), accounts...)
}

// AccountBalance returns the cached balance for the given account ID.
func (api *Api) AccountBalance(accountID string) float64 {
	if balance, ok := api.accountBalances[accountID]; ok {
		return balance
	}
	return 0
}

func (a *Account) GetBalance(api *Api) float64 {
	return api.AccountBalance(a.ID)
}

func (a *Account) IsEmpty() bool {
	return *a == Account{}
}

func (a Account) GetName() string {
	return a.Name
}
