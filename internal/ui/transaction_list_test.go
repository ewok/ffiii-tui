/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/

package ui

import (
	"errors"
	"fmt"
	"testing"

	"ffiii-tui/internal/firefly"
	"ffiii-tui/internal/ui/notify"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mockTransactionAPI struct {
	listTransactionsFunc        func(query string) ([]firefly.Transaction, error)
	deleteTransactionFunc       func(transactionID string) error
	listTransactionsCalledWith  []string
	deleteTransactionCalledWith []string
}

func (m *mockTransactionAPI) ListTransactions(query string) ([]firefly.Transaction, error) {
	m.listTransactionsCalledWith = append(m.listTransactionsCalledWith, query)
	if m.listTransactionsFunc != nil {
		return m.listTransactionsFunc(query)
	}
	return nil, nil
}

func (m *mockTransactionAPI) DeleteTransaction(transactionID string) error {
	m.deleteTransactionCalledWith = append(m.deleteTransactionCalledWith, transactionID)
	if m.deleteTransactionFunc != nil {
		return m.deleteTransactionFunc(transactionID)
	}
	return nil
}

func newTestTransaction(id uint, txID, txType, date, desc string) firefly.Transaction {
	return firefly.Transaction{
		ID:            id,
		TransactionID: txID,
		Type:          txType,
		Date:          date,
		GroupTitle:    fmt.Sprintf("Group %d", id),
		Splits: []firefly.Split{
			{
				TransactionJournalID: fmt.Sprintf("split-%d", id),
				Source:               firefly.Account{ID: "src1", Name: "Source Account"},
				Destination:          firefly.Account{ID: "dst1", Name: "Destination Account"},
				Category:             firefly.Category{ID: "cat1", Name: "Groceries"},
				Currency:             "USD",
				Amount:               100.00,
				Description:          desc,
			},
		},
	}
}

func newFocusedTransactionModel(t *testing.T, transactions []firefly.Transaction) modelTransactions {
	t.Helper()

	api := &mockTransactionAPI{
		listTransactionsFunc: func(query string) ([]firefly.Transaction, error) {
			return transactions, nil
		},
	}

	m := NewModelTransactions(api)
	m.transactions = transactions
	rows, columns := getRows(transactions)
	m.table.SetRows(rows)
	m.table.SetColumns(columns)
	(&m).Focus()
	return m
}

// Basic functionality tests

func TestNewModelTransactions_Initializes(t *testing.T) {
	api := &mockTransactionAPI{
		listTransactionsFunc: func(query string) ([]firefly.Transaction, error) {
			return []firefly.Transaction{}, nil
		},
	}

	m := NewModelTransactions(api)

	if m.api != api {
		t.Error("expected API to be set")
	}
	if m.focus {
		t.Error("expected model to be unfocused initially")
	}
	if len(m.transactions) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(m.transactions))
	}
}

func TestGetRows_EmptyTransactions(t *testing.T) {
	rows, columns := getRows([]firefly.Transaction{})

	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
	if len(columns) != 12 {
		t.Errorf("expected 12 columns, got %d", len(columns))
	}
}

func TestGetRows_SingleTransaction(t *testing.T) {
	tx := newTestTransaction(0, "tx1", "withdrawal", "2024-01-15T10:00:00Z", "Test transaction")

	rows, columns := getRows([]firefly.Transaction{tx})

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if len(columns) != 12 {
		t.Errorf("expected 12 columns, got %d", len(columns))
	}

	row := rows[0]
	if row[0] != "0" {
		t.Errorf("expected ID '0', got %q", row[0])
	}
	if row[1] != "←" {
		t.Errorf("expected withdrawal icon '←', got %q", row[1])
	}
	if row[2] != "2024-01-15" {
		t.Errorf("expected date '2024-01-15', got %q", row[2])
	}
	if row[11] != "tx1" {
		t.Errorf("expected transaction ID 'tx1', got %q", row[11])
	}
}

func TestGetRows_TransactionTypes(t *testing.T) {
	tests := []struct {
		txType       string
		expectedIcon string
	}{
		{"withdrawal", "←"},
		{"deposit", "→"},
		{"transfer", "⇄"},
	}

	for _, tt := range tests {
		t.Run(tt.txType, func(t *testing.T) {
			tx := newTestTransaction(0, "tx1", tt.txType, "2024-01-15T10:00:00Z", "Test")
			rows, _ := getRows([]firefly.Transaction{tx})

			if len(rows) != 1 {
				t.Fatalf("expected 1 row, got %d", len(rows))
			}
			if rows[0][1] != tt.expectedIcon {
				t.Errorf("expected icon %q, got %q", tt.expectedIcon, rows[0][1])
			}
		})
	}
}

func TestGetRows_MultiSplitTransaction(t *testing.T) {
	tx := firefly.Transaction{
		ID:            0,
		TransactionID: "tx1",
		Type:          "withdrawal",
		Date:          "2024-01-15T10:00:00Z",
		Splits: []firefly.Split{
			{Description: "Split 1", Amount: 50.00, Currency: "USD"},
			{Description: "Split 2", Amount: 30.00, Currency: "USD"},
			{Description: "Split 3", Amount: 20.00, Currency: "USD"},
		},
	}

	rows, _ := getRows([]firefly.Transaction{tx})

	if len(rows) != 3 {
		t.Fatalf("expected 3 rows (one per split), got %d", len(rows))
	}

	if rows[0][1] != "←" {
		t.Errorf("expected first row icon '←', got %q", rows[0][1])
	}
	if rows[1][1] != " ↳" {
		t.Errorf("expected second row icon ' ↳', got %q", rows[1][1])
	}
	if rows[2][1] != " ↳" {
		t.Errorf("expected third row icon ' ↳', got %q", rows[2][1])
	}
}

// Message handler tests

func TestRefreshTransactionsMsg_Success(t *testing.T) {
	expectedTransactions := []firefly.Transaction{
		newTestTransaction(0, "tx1", "withdrawal", "2024-01-15T10:00:00Z", "Test 1"),
		newTestTransaction(1, "tx2", "deposit", "2024-01-16T10:00:00Z", "Test 2"),
	}

	api := &mockTransactionAPI{
		listTransactionsFunc: func(query string) ([]firefly.Transaction, error) {
			return expectedTransactions, nil
		},
	}

	m := NewModelTransactions(api)
	(&m).Focus()

	_, cmd := m.Update(RefreshTransactionsMsg{})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msg := cmd()
	txUpdateMsg, ok := msg.(TransactionsUpdateMsg)
	if !ok {
		t.Fatalf("expected TransactionsUpdateMsg, got %T", msg)
	}

	if len(txUpdateMsg.Transactions) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(txUpdateMsg.Transactions))
	}

	if len(api.listTransactionsCalledWith) != 1 {
		t.Fatalf("expected ListTransactions to be called once, got %d", len(api.listTransactionsCalledWith))
	}
	if api.listTransactionsCalledWith[0] != "" {
		t.Errorf("expected empty query, got %q", api.listTransactionsCalledWith[0])
	}
}

func TestRefreshTransactionsMsg_WithSearch(t *testing.T) {
	api := &mockTransactionAPI{
		listTransactionsFunc: func(query string) ([]firefly.Transaction, error) {
			return []firefly.Transaction{}, nil
		},
	}

	m := NewModelTransactions(api)
	m.currentSearch = "groceries"
	(&m).Focus()

	_, cmd := m.Update(RefreshTransactionsMsg{})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	_ = cmd()

	if len(api.listTransactionsCalledWith) != 1 {
		t.Fatalf("expected ListTransactions to be called once, got %d", len(api.listTransactionsCalledWith))
	}
	// Should be URL encoded
	if api.listTransactionsCalledWith[0] != "groceries" {
		t.Errorf("expected 'groceries', got %q", api.listTransactionsCalledWith[0])
	}
}

func TestRefreshTransactionsMsg_Error(t *testing.T) {
	expectedErr := errors.New("API error")
	api := &mockTransactionAPI{
		listTransactionsFunc: func(query string) ([]firefly.Transaction, error) {
			return nil, expectedErr
		},
	}

	m := NewModelTransactions(api)
	(&m).Focus()

	_, cmd := m.Update(RefreshTransactionsMsg{})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msg := cmd()
	notifyMsg, ok := msg.(notify.NotifyMsg)
	if !ok {
		t.Fatalf("expected notify.NotifyMsg, got %T", msg)
	}
	if notifyMsg.Level != notify.Warn {
		t.Errorf("expected notify level Warn, got %v", notifyMsg.Level)
	}
	if notifyMsg.Message != expectedErr.Error() {
		t.Errorf("expected message %q, got %q", expectedErr.Error(), notifyMsg.Message)
	}
}

func TestTransactionsUpdateMsg_UpdatesTransactionsAndFilters(t *testing.T) {
	transactions := []firefly.Transaction{
		newTestTransaction(0, "tx1", "withdrawal", "2024-01-15T10:00:00Z", "Test 1"),
	}

	api := &mockTransactionAPI{}
	m := NewModelTransactions(api)
	(&m).Focus()

	updated, cmd := m.Update(TransactionsUpdateMsg{Transactions: transactions})
	m2 := updated.(modelTransactions)

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	if len(m2.transactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(m2.transactions))
	}

	msgs := collectMsgsFromCmd(cmd)
	foundFilter := false
	foundNotify := false
	for _, msg := range msgs {
		if _, ok := msg.(FilterMsg); ok {
			foundFilter = true
		}
		if notifyMsg, ok := msg.(notify.NotifyMsg); ok {
			if notifyMsg.Level == notify.Log {
				foundNotify = true
			}
		}
	}
	if !foundFilter {
		t.Error("expected FilterMsg in batch")
	}
	if !foundNotify {
		t.Error("expected notify.NotifyMsg in batch")
	}
}

func TestDeleteTransactionMsg_Success(t *testing.T) {
	tx := newTestTransaction(0, "tx-to-delete", "withdrawal", "2024-01-15T10:00:00Z", "Test")

	api := &mockTransactionAPI{}
	m := NewModelTransactions(api)
	(&m).Focus()

	_, cmd := m.Update(DeleteTransactionMsg{Transaction: tx})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	if len(api.deleteTransactionCalledWith) != 1 {
		t.Fatalf("expected DeleteTransaction to be called once, got %d", len(api.deleteTransactionCalledWith))
	}
	if api.deleteTransactionCalledWith[0] != "tx-to-delete" {
		t.Errorf("expected transaction ID 'tx-to-delete', got %q", api.deleteTransactionCalledWith[0])
	}

	msgs := collectMsgsFromCmd(cmd)
	foundRefreshTransactions := false
	foundRefreshSummary := false
	for _, msg := range msgs {
		if _, ok := msg.(RefreshTransactionsMsg); ok {
			foundRefreshTransactions = true
		}
		if _, ok := msg.(RefreshSummaryMsg); ok {
			foundRefreshSummary = true
		}
	}
	if !foundRefreshTransactions {
		t.Error("expected RefreshTransactionsMsg in batch")
	}
	if !foundRefreshSummary {
		t.Error("expected RefreshSummaryMsg in batch")
	}
}

func TestDeleteTransactionMsg_Error(t *testing.T) {
	expectedErr := errors.New("delete error")
	tx := newTestTransaction(0, "tx-to-delete", "withdrawal", "2024-01-15T10:00:00Z", "Test")

	api := &mockTransactionAPI{
		deleteTransactionFunc: func(transactionID string) error {
			return expectedErr
		},
	}
	m := NewModelTransactions(api)
	(&m).Focus()

	_, cmd := m.Update(DeleteTransactionMsg{Transaction: tx})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msgs := collectMsgsFromCmd(cmd)
	foundError := false
	for _, msg := range msgs {
		if notifyMsg, ok := msg.(notify.NotifyMsg); ok {
			if notifyMsg.Level == notify.Err {
				foundError = true
			}
		}
	}
	if !foundError {
		t.Error("expected error notification in batch")
	}
}

// FilterMsg tests

func TestFilterMsg_ByAccount(t *testing.T) {
	srcAccount := firefly.Account{ID: "src1", Name: "Source Account"}
	dstAccount := firefly.Account{ID: "dst1", Name: "Destination Account"}
	otherAccount := firefly.Account{ID: "other", Name: "Other Account"}

	transactions := []firefly.Transaction{
		{
			ID:            0,
			TransactionID: "tx1",
			Type:          "withdrawal",
			Date:          "2024-01-15T10:00:00Z",
			Splits: []firefly.Split{
				{Source: srcAccount, Destination: dstAccount, Description: "Tx1"},
			},
		},
		{
			ID:            1,
			TransactionID: "tx2",
			Type:          "deposit",
			Date:          "2024-01-16T10:00:00Z",
			Splits: []firefly.Split{
				{Source: otherAccount, Destination: firefly.Account{Name: "Another"}, Description: "Tx2"},
			},
		},
	}

	m := newFocusedTransactionModel(t, transactions)

	updated, _ := m.Update(FilterMsg{Account: srcAccount})
	m2 := updated.(modelTransactions)

	if len(m2.table.Rows()) != 1 {
		t.Errorf("expected 1 filtered row, got %d", len(m2.table.Rows()))
	}
	if m2.currentAccount.ID != "src1" {
		t.Errorf("expected currentAccount ID 'src1', got %q", m2.currentAccount.ID)
	}
}

func TestFilterMsg_ByCategory(t *testing.T) {
	catGroceries := firefly.Category{ID: "cat1", Name: "Groceries"}
	catTransport := firefly.Category{ID: "cat2", Name: "Transport"}

	transactions := []firefly.Transaction{
		{
			ID:            0,
			TransactionID: "tx1",
			Type:          "withdrawal",
			Date:          "2024-01-15T10:00:00Z",
			Splits: []firefly.Split{
				{Category: catGroceries, Description: "Tx1", Amount: 50.00, Currency: "USD"},
			},
		},
		{
			ID:            1,
			TransactionID: "tx2",
			Type:          "withdrawal",
			Date:          "2024-01-16T10:00:00Z",
			Splits: []firefly.Split{
				{Category: catTransport, Description: "Tx2", Amount: 30.00, Currency: "USD"},
			},
		},
	}

	m := newFocusedTransactionModel(t, transactions)

	updated, _ := m.Update(FilterMsg{Category: catGroceries})
	m2 := updated.(modelTransactions)

	if len(m2.table.Rows()) != 1 {
		t.Errorf("expected 1 filtered row, got %d", len(m2.table.Rows()))
	}
	if m2.currentCategory.ID != "cat1" {
		t.Errorf("expected currentCategory ID 'cat1', got %q", m2.currentCategory.ID)
	}
}

func TestFilterMsg_ByQuery(t *testing.T) {
	transactions := []firefly.Transaction{
		{
			ID:            0,
			TransactionID: "tx1",
			Type:          "withdrawal",
			Date:          "2024-01-15T10:00:00Z",
			GroupTitle:    "Group Title",
			Splits: []firefly.Split{
				{Description: "Buy groceries", Amount: 50.00, Currency: "USD"},
			},
		},
		{
			ID:            1,
			TransactionID: "tx2",
			Type:          "deposit",
			Date:          "2024-01-16T10:00:00Z",
			GroupTitle:    "Other Group",
			Splits: []firefly.Split{
				{Description: "Salary payment", Amount: 1000.00, Currency: "USD"},
			},
		},
	}

	m := newFocusedTransactionModel(t, transactions)

	updated, _ := m.Update(FilterMsg{Query: "groceries"})
	m2 := updated.(modelTransactions)

	if len(m2.table.Rows()) != 1 {
		t.Errorf("expected 1 filtered row, got %d", len(m2.table.Rows()))
	}
	if m2.currentFilter != "groceries" {
		t.Errorf("expected currentFilter 'groceries', got %q", m2.currentFilter)
	}
}

func TestFilterMsg_Reset(t *testing.T) {
	srcAccount := firefly.Account{ID: "src1", Name: "Source"}
	catGroceries := firefly.Category{ID: "cat1", Name: "Groceries"}

	transactions := []firefly.Transaction{
		newTestTransaction(0, "tx1", "withdrawal", "2024-01-15T10:00:00Z", "Test"),
	}

	m := newFocusedTransactionModel(t, transactions)
	m.currentAccount = srcAccount
	m.currentCategory = catGroceries
	m.currentFilter = "test"

	updated, _ := m.Update(FilterMsg{Reset: true})
	m2 := updated.(modelTransactions)

	if !m2.currentAccount.IsEmpty() {
		t.Error("expected currentAccount to be empty after reset")
	}
	if !m2.currentCategory.IsEmpty() {
		t.Error("expected currentCategory to be empty after reset")
	}
	if m2.currentFilter != "" {
		t.Errorf("expected currentFilter to be empty after reset, got %q", m2.currentFilter)
	}
}

func TestFilterMsg_ComplexFiltering(t *testing.T) {
	transactions := []firefly.Transaction{
		{
			ID:            0,
			TransactionID: "tx1",
			Type:          "withdrawal",
			Date:          "2024-01-15T10:00:00Z",
			Splits: []firefly.Split{
				{
					Source:      firefly.Account{Name: "Checking"},
					Destination: firefly.Account{Name: "Walmart"},
					Category:    firefly.Category{Name: "Groceries"},
					Description: "Weekly shopping",
					Amount:      150.50,
					Currency:    "USD",
				},
			},
		},
	}

	m := newFocusedTransactionModel(t, transactions)

	tests := []struct {
		name         string
		query        string
		expectedRows int
	}{
		{"filter by source", "Checking", 1},
		{"filter by destination", "Walmart", 1},
		{"filter by category", "Groceries", 1},
		{"filter by description", "shopping", 1},
		{"filter by amount", "150.50", 1},
		{"filter by currency", "USD", 1},
		{"filter no match", "nomatch", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, _ := m.Update(FilterMsg{Query: tt.query})
			m2 := updated.(modelTransactions)

			if len(m2.table.Rows()) != tt.expectedRows {
				t.Errorf("expected %d rows, got %d", tt.expectedRows, len(m2.table.Rows()))
			}
		})
	}
}

// SearchMsg tests

func TestSearchMsg_SetSearch(t *testing.T) {
	api := &mockTransactionAPI{
		listTransactionsFunc: func(query string) ([]firefly.Transaction, error) {
			return []firefly.Transaction{}, nil
		},
	}

	m := NewModelTransactions(api)
	(&m).Focus()

	updated, cmd := m.Update(SearchMsg{Query: "test search"})
	m2 := updated.(modelTransactions)

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	if m2.currentSearch != "test search" {
		t.Errorf("expected currentSearch 'test search', got %q", m2.currentSearch)
	}

	msg := cmd()
	if _, ok := msg.(RefreshTransactionsMsg); !ok {
		t.Errorf("expected RefreshTransactionsMsg, got %T", msg)
	}
}

func TestSearchMsg_ClearSearch(t *testing.T) {
	api := &mockTransactionAPI{
		listTransactionsFunc: func(query string) ([]firefly.Transaction, error) {
			return []firefly.Transaction{}, nil
		},
	}

	m := NewModelTransactions(api)
	m.currentSearch = "existing search"
	(&m).Focus()

	updated, cmd := m.Update(SearchMsg{Query: "None"})
	m2 := updated.(modelTransactions)

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	if m2.currentSearch != "" {
		t.Errorf("expected empty currentSearch, got %q", m2.currentSearch)
	}
}

func TestSearchMsg_ClearSearchNoOp(t *testing.T) {
	api := &mockTransactionAPI{}
	m := NewModelTransactions(api)
	(&m).Focus()

	_, cmd := m.Update(SearchMsg{Query: "None"})

	if cmd != nil {
		t.Error("expected no command when clearing already empty search")
	}
}

// Key handling tests

func TestTransactionList_FocusAndBlur(t *testing.T) {
	m := newFocusedTransactionModel(t, []firefly.Transaction{})

	if !m.focus {
		t.Error("expected model to be focused after Focus()")
	}

	(&m).Blur()
	if m.focus {
		t.Error("expected model to be blurred after Blur()")
	}
}

func TestTransactionList_View_RendersTable(t *testing.T) {
	lipgloss.SetColorProfile(0)
	tx := newTestTransaction(0, "tx1", "withdrawal", "2024-01-15T10:00:00Z", "Test")
	m := newFocusedTransactionModel(t, []firefly.Transaction{tx})

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestTransactionList_UnfocusedIgnoresKeys(t *testing.T) {
	m := newFocusedTransactionModel(t, []firefly.Transaction{})
	(&m).Blur()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if cmd != nil {
		t.Error("expected no command when unfocused")
	}
}

func TestGetCurrentTransaction_Success(t *testing.T) {
	tx := newTestTransaction(0, "tx1", "withdrawal", "2024-01-15T10:00:00Z", "Test")
	m := newFocusedTransactionModel(t, []firefly.Transaction{tx})

	currentTx, err := m.GetCurrentTransaction()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if currentTx.TransactionID != "tx1" {
		t.Errorf("expected transaction ID 'tx1', got %q", currentTx.TransactionID)
	}
}

func TestGetCurrentTransaction_NoTransactions(t *testing.T) {
	m := newFocusedTransactionModel(t, []firefly.Transaction{})

	_, err := m.GetCurrentTransaction()
	if err == nil {
		t.Error("expected error when no transactions")
	}
}

func TestGetCurrentTransaction_NoSelection(t *testing.T) {
	api := &mockTransactionAPI{}
	m := NewModelTransactions(api)
	m.transactions = []firefly.Transaction{
		newTestTransaction(0, "tx1", "withdrawal", "2024-01-15T10:00:00Z", "Test"),
	}
	(&m).Focus()

	_, err := m.GetCurrentTransaction()
	if err == nil {
		t.Error("expected error when no selection")
	}
}

func TestUpdatePositions_SetsTableSize(t *testing.T) {
	globalWidth = 200
	globalHeight = 60
	topSize = 5
	leftSize = 50
	summarySize = 0 // Reset summary size for this test

	m := newFocusedTransactionModel(t, []firefly.Transaction{})

	updated, _ := m.Update(UpdatePositions{})
	m2 := updated.(modelTransactions)

	h, v := m2.styles.Base.GetFrameSize()
	wantW := globalWidth - leftSize - h
	// Note: The table.SetHeight() call sets the desired height, but the table component
	// may apply its own internal padding/margins, so we check that Update was called correctly
	expectedSetHeight := globalHeight - topSize - v

	if m2.table.Width() != wantW {
		t.Errorf("expected width %d, got %d", wantW, m2.table.Width())
	}

	// The actual table height may differ from what was set due to internal table styling
	// We verify that the Update method calculated and set the height based on global dimensions
	if m2.table.Height() < expectedSetHeight-5 || m2.table.Height() > expectedSetHeight {
		t.Errorf("expected height around %d (globalHeight=%d - topSize=%d - frameV=%d), got %d",
			expectedSetHeight, globalHeight, topSize, v, m2.table.Height())
	}
}

// Edge case tests

func TestGetRows_CalculatesColumnWidths(t *testing.T) {
	tx := firefly.Transaction{
		ID:            0,
		TransactionID: "tx1",
		Type:          "withdrawal",
		Date:          "2024-01-15T10:00:00Z",
		Splits: []firefly.Split{
			{
				Source:          firefly.Account{Name: "Very Long Source Account Name Here"},
				Destination:     firefly.Account{Name: "Very Long Destination Account Name Here"},
				Category:        firefly.Category{Name: "Very Long Category Name Here"},
				Description:     "Very long description that should affect column width",
				Amount:          99999.99,
				Currency:        "USDT",
				ForeignAmount:   88888.88,
				ForeignCurrency: "EURO",
			},
		},
	}

	_, columns := getRows([]firefly.Transaction{tx})

	sourceCol := columns[3]
	if sourceCol.Width < len("Very Long Source Account Name Here") {
		t.Errorf("expected source width >= %d, got %d", len("Very Long Source Account Name Here"), sourceCol.Width)
	}

	destCol := columns[4]
	if destCol.Width < len("Very Long Destination Account Name Here") {
		t.Errorf("expected destination width >= %d, got %d", len("Very Long Destination Account Name Here"), destCol.Width)
	}

	catCol := columns[5]
	if catCol.Width < len("Very Long Category Name Here") {
		t.Errorf("expected category width >= %d, got %d", len("Very Long Category Name Here"), catCol.Width)
	}
}

func TestGetRows_HandlesMultipleTransactions(t *testing.T) {
	transactions := []firefly.Transaction{
		newTestTransaction(0, "tx1", "withdrawal", "2024-01-15T10:00:00Z", "Test 1"),
		newTestTransaction(1, "tx2", "deposit", "2024-01-16T10:00:00Z", "Test 2"),
		newTestTransaction(2, "tx3", "transfer", "2024-01-17T10:00:00Z", "Test 3"),
	}

	rows, _ := getRows(transactions)

	if len(rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(rows))
	}

	if rows[0][0] != "0" {
		t.Errorf("expected first row ID '0', got %q", rows[0][0])
	}
	if rows[1][0] != "1" {
		t.Errorf("expected second row ID '1', got %q", rows[1][0])
	}
	if rows[2][0] != "2" {
		t.Errorf("expected third row ID '2', got %q", rows[2][0])
	}
}

func TestFilterMsg_CaseInsensitiveSearch(t *testing.T) {
	transactions := []firefly.Transaction{
		{
			ID:            0,
			TransactionID: "tx1",
			Type:          "withdrawal",
			Date:          "2024-01-15T10:00:00Z",
			Splits: []firefly.Split{
				{Description: "Buy GROCERIES", Amount: 50.00, Currency: "USD"},
			},
		},
	}

	m := newFocusedTransactionModel(t, transactions)

	tests := []struct {
		query string
	}{
		{"groceries"},
		{"GROCERIES"},
		{"GrOcErIeS"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			updated, _ := m.Update(FilterMsg{Query: tt.query})
			m2 := updated.(modelTransactions)

			if len(m2.table.Rows()) != 1 {
				t.Errorf("expected 1 row for query %q, got %d", tt.query, len(m2.table.Rows()))
			}
		})
	}
}

func TestFilterMsg_MatchesGroupTitle(t *testing.T) {
	transactions := []firefly.Transaction{
		{
			ID:            0,
			TransactionID: "tx1",
			Type:          "withdrawal",
			Date:          "2024-01-15T10:00:00Z",
			GroupTitle:    "Shopping trip to the mall",
			Splits: []firefly.Split{
				{Description: "Various items", Amount: 50.00, Currency: "USD"},
			},
		},
	}

	m := newFocusedTransactionModel(t, transactions)

	updated, _ := m.Update(FilterMsg{Query: "mall"})
	m2 := updated.(modelTransactions)

	if len(m2.table.Rows()) != 1 {
		t.Errorf("expected 1 row matching group title, got %d", len(m2.table.Rows()))
	}
}

func TestFilterMsg_ForeignCurrencyAndAmount(t *testing.T) {
	transactions := []firefly.Transaction{
		{
			ID:            0,
			TransactionID: "tx1",
			Type:          "withdrawal",
			Date:          "2024-01-15T10:00:00Z",
			Splits: []firefly.Split{
				{
					Description:     "Foreign transaction",
					Amount:          50.00,
					Currency:        "USD",
					ForeignAmount:   45.50,
					ForeignCurrency: "EUR",
				},
			},
		},
	}

	m := newFocusedTransactionModel(t, transactions)

	tests := []struct {
		name  string
		query string
		match bool
	}{
		{"foreign currency", "EUR", true},
		{"foreign amount", "45.50", true},
		{"no match", "GBP", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, _ := m.Update(FilterMsg{Query: tt.query})
			m2 := updated.(modelTransactions)

			expectedRows := 0
			if tt.match {
				expectedRows = 1
			}

			if len(m2.table.Rows()) != expectedRows {
				t.Errorf("expected %d rows, got %d", expectedRows, len(m2.table.Rows()))
			}
		})
	}
}

func TestDeleteTransactionMsg_EmptyTransactionID(t *testing.T) {
	tx := firefly.Transaction{
		ID:            0,
		TransactionID: "",
		Type:          "withdrawal",
		Date:          "2024-01-15T10:00:00Z",
	}

	api := &mockTransactionAPI{}
	m := NewModelTransactions(api)
	(&m).Focus()

	_, cmd := m.Update(DeleteTransactionMsg{Transaction: tx})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	// Should not call DeleteTransaction with empty ID
	if len(api.deleteTransactionCalledWith) != 0 {
		t.Error("expected DeleteTransaction not to be called with empty ID")
	}
}
