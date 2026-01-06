/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import (
	"fmt"
)

type RequestTransaction struct {
	ErrorIfDuplicateHash bool                      `json:"error_if_duplicate_hash,omitempty"`
	ApplyRules           bool                      `json:"apply_rules,omitempty"`
	FireWebhooks         bool                      `json:"fire_webhooks,omitempty"`
	GroupTitle           string                    `json:"group_title,omitempty"`
	Transactions         []RequestTransactionSplit `json:"transactions,omitempty"`
}

type RequestTransactionSplit struct {
	TransactionJournalID string   `json:"transaction_journal_id,omitempty"`
	Type                 string   `json:"type,omitempty"`
	Date                 string   `json:"date,omitempty"`
	Amount               string   `json:"amount,omitempty"`
	Description          string   `json:"description,omitempty"`
	Order                int      `json:"order,omitempty"`
	CurrencyID           string   `json:"currency_id,omitempty"`
	CurrencyCode         string   `json:"currency_code,omitempty"`
	ForeignAmount        string   `json:"foreign_amount,omitempty"`
	ForeignCurrencyID    string   `json:"foreign_currency_id,omitempty"`
	ForeignCurrencyCode  string   `json:"foreign_currency_code,omitempty"`
	BudgetID             string   `json:"budget_id,omitempty"`
	CategoryID           string   `json:"category_id,omitempty"`
	CategoryName         string   `json:"category_name,omitempty"`
	SourceID             string   `json:"source_id,omitempty"`
	SourceName           string   `json:"source_name,omitempty"`
	SourceIBAN           string   `json:"source_iban,omitempty"`
	DestinationID        string   `json:"destination_id,omitempty"`
	DestinationName      string   `json:"destination_name,omitempty"`
	DestinationIBAN      string   `json:"destination_iban,omitempty"`
	Reconciled           bool     `json:"reconciled,omitempty"`
	BillID               string   `json:"bill_id,omitempty"`
	BillName             string   `json:"bill_name,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
	Notes                string   `json:"notes,omitempty"`
	InternalReference    string   `json:"internal_reference,omitempty"`
	ExternalID           string   `json:"external_id,omitempty"`
	ExternalURL          string   `json:"external_url,omitempty"`
	SepaCC               string   `json:"sepa_cc,omitempty"`
	SepaCTOp             string   `json:"sepa_ct_op,omitempty"`
	SepaCTID             string   `json:"sepa_ct_id,omitempty"`
	SepaDB               string   `json:"sepa_db,omitempty"`
	SepaCountry          string   `json:"sepa_country,omitempty"`
	SepaEP               string   `json:"sepa_ep,omitempty"`
	SepaCI               string   `json:"sepa_ci,omitempty"`
	SepaBatchID          string   `json:"sepa_batch_id,omitempty"`
	InterestDate         string   `json:"interest_date,omitempty"`
	BookDate             string   `json:"book_date,omitempty"`
	ProcessDate          string   `json:"process_date,omitempty"`
	DueDate              string   `json:"due_date,omitempty"`
	PaymentDate          string   `json:"payment_date,omitempty"`
	InvoiceDate          string   `json:"invoice_date,omitempty"`
}

func (api *Api) CreateTransaction(newTransaction RequestTransaction) error {
	endpoint := fmt.Sprintf("%s/transactions", api.Config.ApiUrl)

	response, err := api.postRequest(endpoint, newTransaction)
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

func (api *Api) UpdateTransaction(transactionId string, transaction RequestTransaction) error {
	endpoint := fmt.Sprintf("%s/transactions/%s", api.Config.ApiUrl, transactionId)

	response, err := api.putRequest(endpoint, transaction)
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

func (api *Api) DeleteTransaction(transactionId string) error {
	endpoint := fmt.Sprintf("%s/transactions/%s", api.Config.ApiUrl, transactionId)

	_, err := api.deleteRequest(endpoint)
	if err != nil {
		return err
	}

	return nil
}
