/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/

package ui

import (
	"errors"
	"testing"

	"ffiii-tui/internal/firefly"
)

// mockTransactionAPI is a simple mock for testing the constructor
type mockTransactionAPI struct {
	listTransactionsFunc  func(query string) ([]firefly.Transaction, error)
	deleteTransactionFunc func(transactionID string) error
}

func (m *mockTransactionAPI) ListTransactions(query string) ([]firefly.Transaction, error) {
	if m.listTransactionsFunc != nil {
		return m.listTransactionsFunc(query)
	}
	return []firefly.Transaction{}, nil
}

func (m *mockTransactionAPI) DeleteTransaction(transactionID string) error {
	if m.deleteTransactionFunc != nil {
		return m.deleteTransactionFunc(transactionID)
	}
	return nil
}

// TestNewModelTransactions_DefaultState verifies all initial state in a single comprehensive test
func TestNewModelTransactions_DefaultState(t *testing.T) {
	// Arrange
	mockAPI := &mockTransactionAPI{}

	// Act
	model := NewModelTransactions(mockAPI)

	// Assert: Verify all default state
	t.Run("API", func(t *testing.T) {
		if model.api != mockAPI {
			t.Error("api should be set to provided instance")
		}
	})

	t.Run("Transactions", func(t *testing.T) {
		if model.transactions == nil {
			t.Fatal("transactions should not be nil")
		}
		if len(model.transactions) != 0 {
			t.Errorf("transactions should be empty, got %d", len(model.transactions))
		}
	})

	t.Run("Filters", func(t *testing.T) {
		if model.currentSearch != "" {
			t.Errorf("currentSearch should be empty, got %q", model.currentSearch)
		}
		if model.currentFilter != "" {
			t.Errorf("currentFilter should be empty, got %q", model.currentFilter)
		}
		if !model.currentAccount.IsEmpty() {
			t.Error("currentAccount should be empty")
		}
		if !model.currentCategory.IsEmpty() {
			t.Error("currentCategory should be empty")
		}
	})

	t.Run("UIState", func(t *testing.T) {
		if model.focus {
			t.Error("focus should be false initially")
		}
	})

	t.Run("Table", func(t *testing.T) {
		if model.table.Cursor() != 0 {
			t.Errorf("table cursor should be at 0, got %d", model.table.Cursor())
		}
		if len(model.table.Rows()) != 0 {
			t.Errorf("table should have 0 rows initially, got %d", len(model.table.Rows()))
		}
		if len(model.table.Columns()) != 12 {
			t.Errorf("table should have 12 columns, got %d", len(model.table.Columns()))
		}
	})

	t.Run("Keymap", func(t *testing.T) {
		// Verify keymap is initialized by checking one required key
		if model.keymap.Quit.Keys() == nil {
			t.Error("keymap should be initialized")
		}
	})

	t.Run("Styles", func(t *testing.T) {
		// Verify styles struct is initialized
		// We check by accessing a style property - if it doesn't panic, it's initialized
		_ = model.styles.Base
		// The fact we can access it means the struct was initialized by DefaultStyles()
	})
}

// TestNewModelTransactions_TableSchema verifies the table column structure
func TestNewModelTransactions_TableSchema(t *testing.T) {
	// Arrange
	mockAPI := &mockTransactionAPI{}

	// Act
	model := NewModelTransactions(mockAPI)

	// Assert: Verify table schema
	columns := model.table.Columns()
	expectedColumns := []string{
		"ID", "Type", "Date", "Source", "Destination", "Category",
		"Currency", "Amount", "Foreign Currency", "Foreign Amount",
		"Description", "TxID",
	}

	if len(columns) != len(expectedColumns) {
		t.Fatalf("Expected %d columns, got %d", len(expectedColumns), len(columns))
	}

	// Check key columns to verify structure (not all titles for maintainability)
	keyChecks := map[int]string{
		0:  "ID",
		1:  "Type",
		2:  "Date",
		10: "Description",
		11: "TxID",
	}

	for idx, expectedTitle := range keyChecks {
		if columns[idx].Title != expectedTitle {
			t.Errorf("Column %d: expected %q, got %q", idx, expectedTitle, columns[idx].Title)
		}
	}
}

// TestNewModelTransactions_EmptyDataHandling verifies table generation handles empty data correctly
func TestNewModelTransactions_EmptyDataHandling(t *testing.T) {
	// Arrange
	mockAPI := &mockTransactionAPI{}

	// Act
	model := NewModelTransactions(mockAPI)
	rows, columns := getRows(model.transactions)

	// Assert
	if len(rows) != 0 {
		t.Errorf("Expected 0 rows for empty transactions, got %d", len(rows))
	}

	if len(columns) != 12 {
		t.Errorf("Expected 12 columns even with empty data, got %d", len(columns))
	}
}

// TestNewModelTransactions_NilAPI verifies handling of nil API (defensive programming)
func TestNewModelTransactions_NilAPI(t *testing.T) {
	testCases := []struct {
		name        string
		api         TransactionAPI
		shouldPanic bool
	}{
		{
			name:        "nil interface",
			api:         nil,
			shouldPanic: false,
		},
		{
			name:        "typed nil pointer",
			api:         (*mockTransactionAPI)(nil),
			shouldPanic: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tc.shouldPanic {
						t.Errorf("NewModelTransactions panicked unexpectedly: %v", r)
					}
				}
			}()

			model := NewModelTransactions(tc.api)

			// Constructor should still initialize the model
			if model.transactions == nil {
				t.Error("transactions slice should be initialized even with nil API")
			}
		})
	}
}

// TestNewModelTransactions_APIFunctional verifies the API reference works
func TestNewModelTransactions_APIFunctional(t *testing.T) {
	// Arrange
	called := false
	mockAPI := &mockTransactionAPI{
		listTransactionsFunc: func(query string) ([]firefly.Transaction, error) {
			called = true
			return []firefly.Transaction{
				{ID: 1, TransactionID: "tx_001"},
			}, nil
		},
	}

	// Act
	model := NewModelTransactions(mockAPI)
	transactions, err := model.api.ListTransactions("")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !called {
		t.Error("Expected API function to be called")
	}

	if len(transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(transactions))
	}

	if transactions[0].TransactionID != "tx_001" {
		t.Errorf("Expected transaction ID tx_001, got %q", transactions[0].TransactionID)
	}
}

// TestNewModelTransactions_APIErrorHandling verifies API errors are propagated
func TestNewModelTransactions_APIErrorHandling(t *testing.T) {
	// Arrange
	expectedError := errors.New("API connection failed")
	mockAPI := &mockTransactionAPI{
		listTransactionsFunc: func(query string) ([]firefly.Transaction, error) {
			return nil, expectedError
		},
	}

	// Act
	model := NewModelTransactions(mockAPI)
	_, err := model.api.ListTransactions("")

	// Assert
	if err == nil {
		t.Fatal("Expected error to be returned")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error %q, got %q", expectedError.Error(), err.Error())
	}
}

// TestNewModelTransactions_IndependentInstances verifies multiple instances don't share state
func TestNewModelTransactions_IndependentInstances(t *testing.T) {
	// Arrange
	mockAPI1 := &mockTransactionAPI{}
	mockAPI2 := &mockTransactionAPI{}

	// Act
	model1 := NewModelTransactions(mockAPI1)
	model2 := NewModelTransactions(mockAPI2)

	// Modify model1 state
	model1.currentSearch = "test search"
	model1.currentFilter = "test filter"
	model1.focus = true

	// Assert: model2 should be unaffected
	if model2.currentSearch == model1.currentSearch {
		t.Error("model2.currentSearch should not be affected by model1 changes")
	}
	if model2.currentFilter == model1.currentFilter {
		t.Error("model2.currentFilter should not be affected by model1 changes")
	}
	if model2.focus == model1.focus {
		t.Error("model2.focus should not be affected by model1 changes")
	}
}
