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
	"net/http"
	"time"
)

const newTransactionEndpoint = "%s/transactions"

type NewTransaction struct {
	ErrorIfDuplicateHash bool                `json:"error_if_duplicate_hash,omitempty"`
	ApplyRules           bool                `json:"apply_rules,omitempty"`
	FireWebhooks         bool                `json:"fire_webhooks,omitempty"`
	GroupTitle           string              `json:"group_title,omitempty"`
	Transactions         []NewSubTransaction `json:"transactions,omitempty"`
}

type NewSubTransaction struct {
	Type                string   `json:"type,omitempty"`
	Date                string   `json:"date,omitempty"`
	Amount              string   `json:"amount,omitempty"`
	Description         string   `json:"description,omitempty"`
	Order               int      `json:"order,omitempty"`
	CurrencyID          string   `json:"currency_id,omitempty"`
	CurrencyCode        string   `json:"currency_code,omitempty"`
	ForeignAmount       string   `json:"foreign_amount,omitempty"`
	ForeignCurrencyID   string   `json:"foreign_currency_id,omitempty"`
	ForeignCurrencyCode string   `json:"foreign_currency_code,omitempty"`
	BudgetID            string   `json:"budget_id,omitempty"`
	BudgetName          string   `json:"budget_name,omitempty"`
	CategoryID          string   `json:"category_id,omitempty"`
	CategoryName        string   `json:"category_name,omitempty"`
	SourceID            string   `json:"source_id,omitempty"`
	SourceName          string   `json:"source_name,omitempty"`
	DestinationID       string   `json:"destination_id,omitempty"`
	DestinationName     string   `json:"destination_name,omitempty"`
	Reconciled          bool     `json:"reconciled,omitempty"`
	PiggyBankID         int      `json:"piggy_bank_id,omitempty"`
	PiggyBankName       string   `json:"piggy_bank_name,omitempty"`
	BillID              string   `json:"bill_id,omitempty"`
	BillName            string   `json:"bill_name,omitempty"`
	Tags                []string `json:"tags,omitempty"`
	Notes               string   `json:"notes,omitempty"`
	InternalReference   string   `json:"internal_reference,omitempty"`
	ExternalID          string   `json:"external_id,omitempty"`
	ExternalURL         string   `json:"external_url,omitempty"`
	SepaCC              string   `json:"sepa_cc,omitempty"`
	SepaCTOp            string   `json:"sepa_ct_op,omitempty"`
	SepaCTID            string   `json:"sepa_ct_id,omitempty"`
	SepaDB              string   `json:"sepa_db,omitempty"`
	SepaCountry         string   `json:"sepa_country,omitempty"`
	SepaEP              string   `json:"sepa_ep,omitempty"`
	SepaCI              string   `json:"sepa_ci,omitempty"`
	SepaBatchID         string   `json:"sepa_batch_id,omitempty"`
	InterestDate        string   `json:"interest_date,omitempty"`
	BookDate            string   `json:"book_date,omitempty"`
	ProcessDate         string   `json:"process_date,omitempty"`
	DueDate             string   `json:"due_date,omitempty"`
	PaymentDate         string   `json:"payment_date,omitempty"`
	InvoiceDate         string   `json:"invoice_date,omitempty"`
}

func (api *Api) CreateTransaction(newTransaction NewTransaction) error {
	endpoint := fmt.Sprintf(newTransactionEndpoint, api.Config.ApiUrl)

	payload, err := json.Marshal(newTransaction)
	if err != nil {
		return fmt.Errorf("failed to marshal new transaction: %v", err)
	}
	// fmt.Printf("CreateTransaction payload: %s\n", string(payload))
	// os.Exit(1)

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Config.ApiKey))

	client := &http.Client{Timeout: time.Duration(api.Config.TimeoutSeconds) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to send exchange rate: %d", resp.StatusCode)
	}

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
