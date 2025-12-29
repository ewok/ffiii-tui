/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const accountsEndpoint = "%s/accounts?page=%d&type=%s"

type Account struct {
	ID           string
	Name         string
	CurrencyCode string
	Balance      float64
	Type         string
	Spent        float64
	Earned       float64
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

func (api *Api) CreateAccount(name string, accountType string, currencyCode string) error {
	endpoint := fmt.Sprintf("%s/accounts", api.Config.ApiUrl)

	var payload map[string]any
	if accountType == "asset" {
		payload = map[string]any{
			"name":              name,
			"type":              accountType,
			"currency_code":     strings.ToUpper(currencyCode),
			"include_net_worth": true,
			"active":            true,
			"account_role":      "defaultAsset",
		}

	} else {
		payload = map[string]any{
			"name": name,
			"type": accountType,
		}
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Config.ApiKey))

	client := &http.Client{Timeout: time.Duration(api.Config.TimeoutSeconds) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// TODO: Make nice error handling
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	// if returned .data.id is present then is good
	var response map[string]any
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		message, ok := response["message"].(string)
		if ok && message != "" {
			return fmt.Errorf("API error: %s", message)
		}
		return fmt.Errorf("failed status code : %d", resp.StatusCode)
	}

	data, ok := response["data"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid response format: missing data field")
	}

	id, ok := data["id"].(string)
	if !ok || id == "" {
		return fmt.Errorf("invalid response format: missing transaction id")
	}

	return nil
}

func (api *Api) UpdateExpenseInsights() error {

	// TODO: Need error reporting
	tSpent := make(map[string]float64)
	spentInsights, err := api.GetInsights("expense/expense")
	if err == nil {
		for _, item := range spentInsights {
			tSpent[item.ID] = (-1) * item.DifferenceFloat
		}
	}

	expenses := api.Accounts["expense"]
	for i, account := range expenses {
		if val, ok := tSpent[account.ID]; ok {
			expenses[i].Spent = val
		} else {
			expenses[i].Spent = 0
		}
	}
	api.Accounts["expense"] = expenses

	return nil
}

func (api *Api) UpdateRevenueInsights() error {

	tEarned := make(map[string]float64)
	earnedInsights, err := api.GetInsights("income/revenue")
	if err == nil {
		for _, item := range earnedInsights {
			tEarned[item.ID] = item.DifferenceFloat
		}
	}

	revenues := api.Accounts["revenue"]
	for i, account := range revenues {
		if val, ok := tEarned[account.ID]; ok {
			revenues[i].Earned = val
		} else {
			revenues[i].Earned = 0
		}
	}
	api.Accounts["revenue"] = revenues

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
		})
	}

	maps.Copy(api.Accounts, accs)

	if accType == "expense" {
		api.UpdateExpenseInsights()
	}
	if accType == "revenue" {
		api.UpdateRevenueInsights()
	}

	return nil
}

func (api *Api) ListAccounts(accountType string) ([]apiAccount, error) {
	allAccounts := []apiAccount{}
	page := 1

	for {
		accounts, err := api.listAccounts(accountType, page)
		if err != nil {
			return nil, err
		}

		if len(accounts) == 0 {
			break
		}

		allAccounts = append(allAccounts, accounts...)
		page++
	}

	return allAccounts, nil
}

func (api *Api) listAccounts(accountType string, page int) ([]apiAccount, error) {

	endpoint := fmt.Sprintf(accountsEndpoint, api.Config.ApiUrl, page, accountType)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Config.ApiKey))

	client := &http.Client{Timeout: time.Duration(api.Config.TimeoutSeconds) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var rawJson map[string]any
	if err := json.Unmarshal(body, &rawJson); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		message, ok := rawJson["message"].(string)
		if ok && message != "" {
			return nil, fmt.Errorf("API error: %s", message)
		}
		return nil, fmt.Errorf("failed status code : %d", resp.StatusCode)
	}

	data, ok := rawJson["data"].([]any)
	if !ok {
		return nil, fmt.Errorf("invalid data format in response")
	}

	accounts := make([]apiAccount, 0, len(data))
	for _, item := range data {
		itemBytes, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal item: %v", err)
		}

		var account apiAccount
		if err := json.Unmarshal(itemBytes, &account); err != nil {
			return nil, fmt.Errorf("failed to unmarshal account: %v", err)
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

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
