/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const accountsEndpoint = "%s/accounts?page=%d&type=%s"

type Account struct {
	ID           string
	Name         string
	CurrencyCode string
	Balance      float64
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
}

func (api *Api) UpdateAssets() error {
	assets, err := api.ListAccounts("asset")
	if err != nil {
		return err
	}

	api.Assets = make([]Account, 0)
	for _, account := range assets {
		balance, err := strconv.ParseFloat(account.Attributes.CurrentBalance, 64)
		if err != nil {
			balance = 0.0
		}

		api.Assets = append(api.Assets, Account{
			ID:           account.ID,
			Name:         account.Attributes.Name,
			CurrencyCode: account.Attributes.CurrencyCode,
			Balance:      balance,
		})
	}

	return nil
}

func (api *Api) UpdateExpenses() error {
	expenses, err := api.ListAccounts("expense")
	if err != nil {
		return err
	}

	api.Expenses = make([]Account, 0)
	for _, account := range expenses {
		api.Expenses = append(api.Expenses, Account{
			ID:   account.ID,
			Name: account.Attributes.Name,
		})
	}

	return nil
}

func (api *Api) UpdateLiabilities() error {
	liabilities, err := api.ListAccounts("liability")
	if err != nil {
		return err
	}

	api.Liabilities = make([]Account, 0)
	for _, account := range liabilities {
		api.Liabilities = append(api.Liabilities, Account{
			ID:   account.ID,
			Name: account.Attributes.Name,
		})
	}

	return nil
}

func (api *Api) UpdateRevenues() error {
	revenues, err := api.ListAccounts("revenue")
	if err != nil {
		return err
	}

	api.Revenues = make([]Account, 0)
	for _, account := range revenues {
		api.Revenues = append(api.Revenues, Account{
			ID:   account.ID,
			Name: account.Attributes.Name,
		})
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

	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Config.ApiKey))

	client := &http.Client{Timeout: time.Duration(api.Config.TimeoutSeconds) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed status code : %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var rawJson map[string]any
	if err := json.Unmarshal(body, &rawJson); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
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
