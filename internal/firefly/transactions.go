/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import (
	"fmt"
	"slices"
)

type Transaction struct {
	ID            uint
	TransactionID string
	Type          string
	Date          string
	GroupTitle    string
	Splits        []Split
}

type Split struct {
	TransactionJournalID string
	Source               Account
	Destination          Account
	Category             Category
	Currency             string
	ForeignCurrency      string
	Amount               float64
	ForeignAmount        float64
	Description          string
}

type ResponseTransaction struct {
	Type       string                        `json:"type"`
	ID         string                        `json:"id"`
	Attributes ResponseTransactionAttributes `json:"attributes"`
}

type ResponseTransactionAttributes struct {
	CreatedAt    string                     `json:"created_at"`
	UpdatedAt    string                     `json:"updated_at"`
	User         string                     `json:"user"`
	GroupTitle   string                     `json:"group_title"`
	Transactions []ResponseTransactionSplit `json:"transactions"`
}

type ResponseTransactionSplit struct {
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
	Amount                       float64 `json:"amount,string"`
	PCAmount                     float64 `json:"pc_amount,string"`
	ForeignAmount                float64 `json:"foreign_amount,string"`
	PCForeignAmount              float64 `json:"pc_foreign_amount,string"`
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

func (api *Api) ListTransactions(query string) ([]Transaction, error) {
	var allData []any
	var err error
	if query != "" {
		allData, err = api.fetchPaginated("%s/search/transactions?&query=%s&page=%d",
			api.Config.ApiUrl,
			query)
	} else {
		allData, err = api.fetchPaginated("%s/transactions?start=%s&end=%s&page=%d",
			api.Config.ApiUrl,
			api.StartDate.Format("2006-01-02"),
			api.EndDate.Format("2006-01-02"))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch paginated transactions: %v", err)
	}

	txs, err := unmarshalItems[ResponseTransaction](allData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal transactions: %v", err)
	}

	transactions := []Transaction{}
	id := 0
	for _, t := range txs {
		var (
			splits []Split
			ttype  string
			tdate  string
		)
		for _, subTx := range t.Attributes.Transactions {
			if ttype == "" {
				ttype = subTx.Type
			}
			if tdate == "" {
				tdate = subTx.Date
			}
			source := api.GetAccountByID(subTx.SourceID)
			destination := api.GetAccountByID(subTx.DestinationID)
			category := api.GetCategoryByID(subTx.CategoryID)

			splits = append(splits, Split{
				Source:               source,
				Destination:          destination,
				Category:             category,
				Currency:             subTx.CurrencyCode,
				ForeignCurrency:      subTx.ForeignCurrencyCode,
				Amount:               subTx.Amount,
				ForeignAmount:        subTx.ForeignAmount,
				Description:          subTx.Description,
				TransactionJournalID: subTx.TransactionJournalID,
			},
			)
		}

		slices.Reverse(splits)
		transactions = append(transactions, Transaction{
			ID:            uint(id),
			TransactionID: t.ID,
			Type:          ttype,
			Date:          tdate,
			Splits:        splits,
			GroupTitle:    t.Attributes.GroupTitle,
		})
		id++
	}
	return transactions, nil
}

func (t *Transaction) Amount() float64 {
	total := 0.0
	for _, split := range t.Splits {
		total += split.Amount
	}
	return total
}

func (t *Transaction) ForeignAmount() float64 {
	total := 0.0
	if t.Type == "transfer" {
		for _, split := range t.Splits {
			total += split.ForeignAmount
		}
	}
	return total
}

func (t *Transaction) Description() string {
	l := len(t.Splits)
	if l > 1 {
		return t.GroupTitle
	}
	if l == 1 {
		return t.Splits[0].Description
	}
	return ""
}

func (t *Transaction) Source() Account {
	l := len(t.Splits)
	if t.Type == "withdrawal" || t.Type == "transfer" {
		if l > 0 {
			return t.Splits[0].Source
		} else {
			return Account{Name: "error"}
		}
	}
	if t.Type == "deposit" && l == 1 {
		return t.Splits[0].Source
	}
	return Account{Name: "multiple"}
}

func (t *Transaction) Destination() Account {
	l := len(t.Splits)
	if t.Type == "deposit" || t.Type == "transfer" {
		if l > 0 {
			return t.Splits[0].Destination
		} else {
			return Account{Name: "error"}
		}
	}
	if t.Type == "withdrawal" && l == 1 {
		return t.Splits[0].Destination
	}
	return Account{Name: "multiple"}
}

func (t *Transaction) Category() Category {
	if len(t.Splits) == 1 {
		return t.Splits[0].Category
	}
	return Category{Name: "multiple"}
}

func (t *Transaction) Currency() string {
	return t.Splits[0].Currency
}

func (t *Transaction) ForeignCurrency() string {
	if t.Type == "transfer" {
		return t.Splits[0].ForeignCurrency
	}
	if len(t.Splits) == 1 {
		return t.Splits[0].ForeignCurrency
	}
	return ""
}
