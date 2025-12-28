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
	"time"
)

const transactionsEndpoint = "%s/transactions?page=%d&limit=%d&start=%s&end=%s"
const searchTransactionsEndpoint = "%s/search/transactions?page=%d&limit=%d&query=%s"

type Transaction struct {
	TransactionID   string
	Type            string
	Date            string
	Source          Account
	Destination     Account
	Category        Category
	Currency        Currency
	ForeignCurrency Currency
	Amount          float64
	ForeignAmount   float64
	Description     string
}

type ApiTransaction struct {
	Type       string                   `json:"type"`
	ID         string                   `json:"id"`
	Attributes apiTransactionAttributes `json:"attributes"`
}

type apiTransactionAttributes struct {
	CreatedAt    string              `json:"created_at"`
	UpdatedAt    string              `json:"updated_at"`
	User         string              `json:"user"`
	GroupTitle   string              `json:"group_title"`
	Transactions []apiSubTransaction `json:"transactions"`
}

type apiSubTransaction struct {
	User                         string  `json:"user"`
	TransactionJournalID         string  `json:"transaction_journal_id"`
	Type                         string  `json:"type"`
	Date                         string  `json:"date"`
	Order                        int     `json:"order"`
	ObjectHasCurrencySetting     bool    `json:"object_has_currency_setting"`
	CurrencyID                   string  `json:"currency_id"`
	CurrencyCode                 string  `json:"currency_code"`
	CurrencySymbol               string  `json:"currency_symbol"`
	CurrencyName                 string  `json:"currency_name"`
	CurrencyDecimalPlaces        int     `json:"currency_decimal_places"`
	ForeignCurrencyID            string  `json:"foreign_currency_id"`
	ForeignCurrencyCode          string  `json:"foreign_currency_code"`
	ForeignCurrencySymbol        string  `json:"foreign_currency_symbol"`
	ForeignCurrencyDecimalPlaces int     `json:"foreign_currency_decimal_places"`
	PrimaryCurrencyID            string  `json:"primary_currency_id"`
	PrimaryCurrencyCode          string  `json:"primary_currency_code"`
	PrimaryCurrencySymbol        string  `json:"primary_currency_symbol"`
	PrimaryCurrencyDecimalPlaces int     `json:"primary_currency_decimal_places"`
	Amount                       string  `json:"amount"`
	PCAmount                     string  `json:"pc_amount"`
	ForeignAmount                string  `json:"foreign_amount"`
	PCForeignAmount              string  `json:"pc_foreign_amount"`
	SourceBalanceAfter           string  `json:"source_balance_after"`
	PCSourceBalanceAfter         string  `json:"pc_source_balance_after"`
	DestinationBalanceAfter      string  `json:"destination_balance_after"`
	PCDestinationBalanceAfter    string  `json:"pc_destination_balance_after"`
	Description                  string  `json:"description"`
	SourceID                     string  `json:"source_id"`
	SourceName                   string  `json:"source_name"`
	SourceIBAN                   string  `json:"source_iban"`
	SourceType                   string  `json:"source_type"`
	DestinationID                string  `json:"destination_id"`
	DestinationName              string  `json:"destination_name"`
	DestinationIBAN              string  `json:"destination_iban"`
	DestinationType              string  `json:"destination_type"`
	BudgetID                     string  `json:"budget_id"`
	BudgetName                   string  `json:"budget_name"`
	CategoryID                   string  `json:"category_id"`
	CategoryName                 string  `json:"category_name"`
	BillID                       string  `json:"bill_id"`
	BillName                     string  `json:"bill_name"`
	SubscriptionID               string  `json:"subscription_id"`
	SubscriptionName             string  `json:"subscription_name"`
	Reconciled                   bool    `json:"reconciled"`
	Notes                        string  `json:"notes"`
	Tags                         any     `json:"tags"`
	InternalReference            string  `json:"internal_reference"`
	ExternalID                   string  `json:"external_id"`
	ExternalURL                  string  `json:"external_url"`
	OriginalSource               string  `json:"original_source"`
	RecurrenceID                 string  `json:"recurrence_id"`
	RecurrenceTotal              int     `json:"recurrence_total"`
	RecurrenceCount              int     `json:"recurrence_count"`
	ImportHashV2                 string  `json:"import_hash_v2"`
	SepaCC                       string  `json:"sepa_cc"`
	SepaCTOp                     string  `json:"sepa_ct_op"`
	SepaCTID                     string  `json:"sepa_ct_id"`
	SepaDB                       string  `json:"sepa_db"`
	SepaCountry                  string  `json:"sepa_country"`
	SepaEP                       string  `json:"sepa_ep"`
	SepaCI                       string  `json:"sepa_ci"`
	SepaBatchID                  string  `json:"sepa_batch_id"`
	InterestDate                 string  `json:"interest_date"`
	BookDate                     string  `json:"book_date"`
	ProcessDate                  string  `json:"process_date"`
	DueDate                      string  `json:"due_date"`
	PaymentDate                  string  `json:"payment_date"`
	InvoiceDate                  string  `json:"invoice_date"`
	Latitude                     float64 `json:"latitude"`
	Longitude                    float64 `json:"longitude"`
	ZoomLevel                    int     `json:"zoom_level"`
	HasAttachments               bool    `json:"has_attachments"`
}

func (api *Api) ListTransactions(start, end, account string) ([]Transaction, error) {
	transactions := []Transaction{}
	page := 1
	for {
		txs, err := api.listTransactions(page, 50, start, end)
		if err != nil {
			return nil, fmt.Errorf("failed to list transactions: %v", err)
		}
		if len(txs) == 0 {
			break
		}
		for _, t := range txs {
			for _, subTx := range t.Attributes.Transactions {
				transaction := Transaction{
					TransactionID:   t.ID,
					Type:            subTx.Type,
					Date:            subTx.Date,
					Source:          subTx.SourceName,
					Destination:     subTx.DestinationName,
					Category:        subTx.CategoryName,
					Currency:        subTx.CurrencyCode,
					ForeignCurrency: subTx.ForeignCurrencyCode,
					Amount:          subTx.Amount,
					ForeignAmount:   subTx.ForeignAmount,
					Description:     subTx.Description,
				}
				if account != "" {
					if subTx.SourceName != account && subTx.DestinationName != account {
						continue
					}
				}
				transactions = append(transactions, transaction)
			}
		}
		page++
	}
	return transactions, nil
}

func (api *Api) listTransactions(page, limit int, start, end string) ([]ApiTransaction, error) {

	if start == "" {
		// First day of the current month
		start = time.Now().AddDate(0, 0, -time.Now().Day()+1).Format("2006-01-02")
	}
	if end == "" {
		// Last day of the current month
		end = time.Now().AddDate(0, 1, -time.Now().Day()).Format("2006-01-02")
	}

	endpoint := fmt.Sprintf(transactionsEndpoint, api.Config.ApiUrl, page, limit, start, end)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
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
	err = json.Unmarshal(body, &rawJson)
	if err != nil {
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

	transactions := make([]ApiTransaction, 0, len(data))

	for _, item := range data {
		itemMap, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid item format in data")
		}
		itemJson, err := json.Marshal(itemMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal item: %v", err)
		}
		var transaction ApiTransaction
		err = json.Unmarshal(itemJson, &transaction)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal item to Transaction: %v", err)
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (api *Api) SearchTransactions(query string) ([]Transaction, error) {
	transactions := []Transaction{}
	page := 1
	for {
		txs, err := api.searchTransactions(page, 50, query)
		if err != nil {
			return nil, fmt.Errorf("failed to search transactions: %v", err)
		}
		if len(txs) == 0 {
			break
		}
		for _, t := range txs {
			for _, subTx := range t.Attributes.Transactions {
				transaction := Transaction{
					TransactionID:   t.ID,
					Type:            subTx.Type,
					Date:            subTx.Date,
					Source:          subTx.SourceName,
					Destination:     subTx.DestinationName,
					Category:        subTx.CategoryName,
					Currency:        subTx.CurrencyCode,
					ForeignCurrency: subTx.ForeignCurrencyCode,
					Amount:          subTx.Amount,
					ForeignAmount:   subTx.ForeignAmount,
					Description:     subTx.Description,
				}
				transactions = append(transactions, transaction)
			}
		}
		page++
	}
	return transactions, nil
}

func (api *Api) searchTransactions(page, limit int, query string) ([]ApiTransaction, error) {

	endpoint := fmt.Sprintf(searchTransactionsEndpoint, api.Config.ApiUrl, page, limit, query)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
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

	err = json.Unmarshal(body, &rawJson)
	if err != nil {
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

	transactions := make([]ApiTransaction, 0, len(data))

	for _, item := range data {
		itemMap, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid item format in data")
		}
		itemJson, err := json.Marshal(itemMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal item: %v", err)
		}
		var transaction ApiTransaction
		err = json.Unmarshal(itemJson, &transaction)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal item to Transaction: %v", err)
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}
