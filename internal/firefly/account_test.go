/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestAccountsServer(t *testing.T) *httptest.Server {
	t.Helper()
	writeBody := func(w http.ResponseWriter, body string) {
		if _, err := fmt.Fprint(w, body); err != nil {
			t.Errorf("failed to write response body: %v", err)
		}
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasPrefix(r.URL.Path, "/insight/"):
			writeBody(w, `[]`)
		case r.URL.Path == "/accounts":
			switch r.URL.Query().Get("type") {
			case "special":
				writeBody(w, `{"data":[{"id":"42","attributes":{"active":true,"name":"Cash account","currency_code":"EUR","current_balance":"0","type":"cash"}}],"meta":{"pagination":{"current_page":1,"total_pages":1,"total":1}}}`)
			case "expense":
				writeBody(w, `{"data":[{"id":"7","attributes":{"active":true,"name":"Groceries","currency_code":"EUR","current_balance":"0","type":"expense"}}],"meta":{"pagination":{"current_page":1,"total_pages":1,"total":1}}}`)
			default:
				writeBody(w, `{"data":[],"meta":{"pagination":{"current_page":1,"total_pages":1,"total":0}}}`)
			}
		default:
			http.NotFound(w, r)
		}
	}))
}

func newTestApi(t *testing.T, serverURL string) *Api {
	t.Helper()
	return &Api{
		Config: ApiConfig{
			ApiKey:         "test-key",
			ApiUrl:         serverURL,
			TimeoutSeconds: 5,
		},
		Accounts:        make(map[string][]Account),
		accountBalances: make(map[string]float64),
	}
}

func TestCashAccount_PopulatesCurrencyCode(t *testing.T) {
	server := newTestAccountsServer(t)
	defer server.Close()

	api := newTestApi(t, server.URL)

	cash := api.CashAccount()
	if cash.IsEmpty() {
		t.Fatal("expected cash account, got empty account")
	}
	if cash.ID != "42" {
		t.Errorf("expected cash account ID '42', got %q", cash.ID)
	}
	if cash.Type != "cash" {
		t.Errorf("expected cash account type 'cash', got %q", cash.Type)
	}
	if cash.CurrencyCode != "EUR" {
		t.Errorf("expected cash account currency code 'EUR', got %q", cash.CurrencyCode)
	}
}

// Regression test for issue #58: editing a transaction with the cash account
// reset the account choice because the cash account was cached twice with
// differing field values, breaking exact struct comparison in the form select.
func TestUpdateAccounts_CashAccountCacheConsistency(t *testing.T) {
	server := newTestAccountsServer(t)
	defer server.Close()

	api := newTestApi(t, server.URL)

	if err := api.UpdateAccounts("special"); err != nil {
		t.Fatalf("UpdateAccounts(special) failed: %v", err)
	}
	if err := api.UpdateAccounts("expense"); err != nil {
		t.Fatalf("UpdateAccounts(expense) failed: %v", err)
	}

	cashBucket := api.AccountsByType("cash")
	if len(cashBucket) != 1 {
		t.Fatalf("expected 1 account in 'cash' bucket, got %d", len(cashBucket))
	}

	var cashInExpenses Account
	for _, acc := range api.AccountsByType("expense") {
		if acc.Type == "cash" {
			cashInExpenses = acc
		}
	}
	if cashInExpenses.IsEmpty() {
		t.Fatal("expected cash account to be present in 'expense' bucket")
	}

	if cashBucket[0] != cashInExpenses {
		t.Errorf("cash account cached inconsistently: cash bucket %+v vs expense bucket %+v",
			cashBucket[0], cashInExpenses)
	}

	byID := api.GetAccountByID("42")
	if byID != cashInExpenses {
		t.Errorf("GetAccountByID returned %+v, expected it to equal expense bucket entry %+v",
			byID, cashInExpenses)
	}
	if byID.CurrencyCode != "EUR" {
		t.Errorf("expected currency code 'EUR' on cash account from GetAccountByID, got %q", byID.CurrencyCode)
	}
}

func TestUpdateAccounts_RemovesStaleAccounts(t *testing.T) {
	responses := map[string]string{
		"asset": `{"data":[
			{"id":"1","attributes":{"active":true,"name":"Checking","currency_code":"EUR","current_balance":"100","type":"asset"}},
			{"id":"2","attributes":{"active":true,"name":"Savings","currency_code":"EUR","current_balance":"200","type":"asset"}}
		],"meta":{"pagination":{"current_page":1,"total_pages":1,"total":2}}}`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body, ok := responses[r.URL.Query().Get("type")]
		if !ok {
			body = `{"data":[],"meta":{"pagination":{"current_page":1,"total_pages":1,"total":0}}}`
		}
		if _, err := fmt.Fprint(w, body); err != nil {
			t.Errorf("failed to write response body: %v", err)
		}
	}))
	defer server.Close()

	api := newTestApi(t, server.URL)

	if err := api.UpdateAccounts("asset"); err != nil {
		t.Fatalf("UpdateAccounts(asset) failed: %v", err)
	}
	if len(api.AccountsByType("asset")) != 2 {
		t.Fatalf("expected 2 asset accounts, got %d", len(api.AccountsByType("asset")))
	}

	responses["asset"] = `{"data":[
		{"id":"1","attributes":{"active":true,"name":"Checking","currency_code":"EUR","current_balance":"100","type":"asset"}}
	],"meta":{"pagination":{"current_page":1,"total_pages":1,"total":1}}}`

	if err := api.UpdateAccounts("asset"); err != nil {
		t.Fatalf("UpdateAccounts(asset) refresh failed: %v", err)
	}

	assets := api.AccountsByType("asset")
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset account after refresh, got %d", len(assets))
	}
	if assets[0].ID != "1" {
		t.Errorf("expected remaining account ID '1', got %q", assets[0].ID)
	}
	if balance := api.AccountBalance("2"); balance != 0 {
		t.Errorf("expected stale account balance to be removed, got %f", balance)
	}
}

func TestUpdateAccounts_RemovesEmptiedAccountType(t *testing.T) {
	responses := map[string]string{
		"asset": `{"data":[
			{"id":"1","attributes":{"active":true,"name":"Checking","currency_code":"EUR","current_balance":"100","type":"asset"}}
		],"meta":{"pagination":{"current_page":1,"total_pages":1,"total":1}}}`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body, ok := responses[r.URL.Query().Get("type")]
		if !ok {
			body = `{"data":[],"meta":{"pagination":{"current_page":1,"total_pages":1,"total":0}}}`
		}
		if _, err := fmt.Fprint(w, body); err != nil {
			t.Errorf("failed to write response body: %v", err)
		}
	}))
	defer server.Close()

	api := newTestApi(t, server.URL)

	if err := api.UpdateAccounts("asset"); err != nil {
		t.Fatalf("UpdateAccounts(asset) failed: %v", err)
	}
	if len(api.AccountsByType("asset")) != 1 {
		t.Fatalf("expected 1 asset account, got %d", len(api.AccountsByType("asset")))
	}

	responses["asset"] = `{"data":[],"meta":{"pagination":{"current_page":1,"total_pages":1,"total":0}}}`

	if err := api.UpdateAccounts("asset"); err != nil {
		t.Fatalf("UpdateAccounts(asset) refresh failed: %v", err)
	}

	if got := len(api.AccountsByType("asset")); got != 0 {
		t.Errorf("expected 0 asset accounts after all were deleted server-side, got %d", got)
	}
	if balance := api.AccountBalance("1"); balance != 0 {
		t.Errorf("expected removed account balance to be cleared, got %f", balance)
	}
}
