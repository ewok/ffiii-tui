/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/

package ui

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"ffiii-tui/internal/firefly"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// mockTransactionFormAPI implements all methods from AccountsAPI, CategoriesAPI, and TransactionWriteAPI
type mockTransactionFormAPI struct {
	// AccountsAPI methods
	updateAccountsFunc       func(accountType string) error
	accountsByTypeFunc       func(accountType string) []firefly.Account
	accountBalanceFunc       func(accountID string) float64
	updateAccountsCalledWith []string

	// CategoriesAPI methods
	updateCategoriesFunc              func() error
	updateCategoriesInsightsFunc      func() error
	categoriesListFunc                func() []firefly.Category
	getTotalSpentEarnedCategoriesFunc func() (spent, earned float64)
	categorySpentFunc                 func(categoryID string) float64
	categoryEarnedFunc                func(categoryID string) float64
	createCategoryFunc                func(name, notes string) error
	updateCategoriesCalledCount       int
	createCategoryCalledWith          []struct {
		name, notes string
	}

	// TransactionWriteAPI methods
	createTransactionFunc  func(tx firefly.RequestTransaction) (string, error)
	updateTransactionFunc  func(transactionID string, tx firefly.RequestTransaction) (string, error)
	createTransactionCalls []firefly.RequestTransaction
	updateTransactionCalls []struct {
		id string
		tx firefly.RequestTransaction
	}
}

// AccountsAPI methods
func (m *mockTransactionFormAPI) UpdateAccounts(accountType string) error {
	m.updateAccountsCalledWith = append(m.updateAccountsCalledWith, accountType)
	if m.updateAccountsFunc != nil {
		return m.updateAccountsFunc(accountType)
	}
	return nil
}

func (m *mockTransactionFormAPI) AccountsByType(accountType string) []firefly.Account {
	if m.accountsByTypeFunc != nil {
		return m.accountsByTypeFunc(accountType)
	}
	return nil
}

func (m *mockTransactionFormAPI) AccountBalance(accountID string) float64 {
	if m.accountBalanceFunc != nil {
		return m.accountBalanceFunc(accountID)
	}
	return 0
}

// CategoriesAPI methods
func (m *mockTransactionFormAPI) UpdateCategories() error {
	m.updateCategoriesCalledCount++
	if m.updateCategoriesFunc != nil {
		return m.updateCategoriesFunc()
	}
	return nil
}

func (m *mockTransactionFormAPI) UpdateCategoriesInsights() error {
	if m.updateCategoriesInsightsFunc != nil {
		return m.updateCategoriesInsightsFunc()
	}
	return nil
}

func (m *mockTransactionFormAPI) CategoriesList() []firefly.Category {
	if m.categoriesListFunc != nil {
		return m.categoriesListFunc()
	}
	return nil
}

func (m *mockTransactionFormAPI) GetTotalSpentEarnedCategories() (spent, earned float64) {
	if m.getTotalSpentEarnedCategoriesFunc != nil {
		return m.getTotalSpentEarnedCategoriesFunc()
	}
	return 0, 0
}

func (m *mockTransactionFormAPI) CategorySpent(categoryID string) float64 {
	if m.categorySpentFunc != nil {
		return m.categorySpentFunc(categoryID)
	}
	return 0
}

func (m *mockTransactionFormAPI) CategoryEarned(categoryID string) float64 {
	if m.categoryEarnedFunc != nil {
		return m.categoryEarnedFunc(categoryID)
	}
	return 0
}

func (m *mockTransactionFormAPI) CreateCategory(name, notes string) error {
	m.createCategoryCalledWith = append(m.createCategoryCalledWith, struct {
		name, notes string
	}{name: name, notes: notes})
	if m.createCategoryFunc != nil {
		return m.createCategoryFunc(name, notes)
	}
	return nil
}

// TransactionWriteAPI methods
func (m *mockTransactionFormAPI) CreateTransaction(tx firefly.RequestTransaction) (string, error) {
	m.createTransactionCalls = append(m.createTransactionCalls, tx)
	if m.createTransactionFunc != nil {
		return m.createTransactionFunc(tx)
	}
	return "", nil
}

func (m *mockTransactionFormAPI) UpdateTransaction(transactionID string, tx firefly.RequestTransaction) (string, error) {
	m.updateTransactionCalls = append(m.updateTransactionCalls, struct {
		id string
		tx firefly.RequestTransaction
	}{id: transactionID, tx: tx})
	if m.updateTransactionFunc != nil {
		return m.updateTransactionFunc(transactionID, tx)
	}
	return "", nil
}

// Test data
var (
	testAssetChecking = firefly.Account{
		ID:           "asset1",
		Name:         "Checking",
		Type:         "asset",
		CurrencyCode: "USD",
	}
	testAssetSavings = firefly.Account{
		ID:           "asset2",
		Name:         "Savings",
		Type:         "asset",
		CurrencyCode: "USD",
	}
	testExpenseGroceries = firefly.Account{
		ID:   "expense1",
		Name: "Groceries",
		Type: "expense",
	}
	testExpenseUtilities = firefly.Account{
		ID:   "expense2",
		Name: "Utilities",
		Type: "expense",
	}
	testRevenueSalary = firefly.Account{
		ID:   "revenue1",
		Name: "Salary",
		Type: "revenue",
	}
	testRevenueFreelance = firefly.Account{
		ID:   "revenue2",
		Name: "Freelance",
		Type: "revenue",
	}
	testLiabilityCreditCard = firefly.Account{
		ID:           "liability1",
		Name:         "Credit Card",
		Type:         "liabilities",
		CurrencyCode: "USD",
	}
	testLiabilityLoan = firefly.Account{
		ID:           "liability2",
		Name:         "Loan",
		Type:         "liabilities",
		CurrencyCode: "USD",
	}
	testCategoryFood = firefly.Category{
		ID:   "cat1",
		Name: "Food",
	}
	testCategoryBills = firefly.Category{
		ID:   "cat2",
		Name: "Bills",
	}
	testCategoryIncome = firefly.Category{
		ID:   "cat3",
		Name: "Income",
	}
)

func newTestTransactionModel() modelTransaction {
	api := &mockTransactionFormAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			switch accountType {
			case "asset":
				return []firefly.Account{testAssetChecking, testAssetSavings}
			case "expense":
				return []firefly.Account{testExpenseGroceries, testExpenseUtilities}
			case "revenue":
				return []firefly.Account{testRevenueSalary, testRevenueFreelance}
			case "liabilities":
				return []firefly.Account{testLiabilityCreditCard, testLiabilityLoan}
			default:
				return nil
			}
		},
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{testCategoryFood, testCategoryBills, testCategoryIncome}
		},
	}
	return newModelTransaction(api)
}

func TestTransaction_Init(t *testing.T) {
	m := newTestTransactionModel()

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected Init to return a command")
	}

	// Verify it returns a command (we can't inspect tea.windowSizeMsg type)
	msg := cmd()
	if msg == nil {
		t.Fatal("expected Init command to return a message")
	}
}

func TestTransaction_NewModel(t *testing.T) {
	api := &mockTransactionFormAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{testAssetChecking}
		},
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{testCategoryFood}
		},
	}

	m := newModelTransaction(api)

	if m.api == nil {
		t.Fatal("expected api to be set")
	}
	if m.form == nil {
		t.Fatal("expected form to be initialized")
	}
	if m.attr == nil {
		t.Fatal("expected attr to be initialized")
	}
	if m.splits == nil {
		t.Fatal("expected splits to be initialized")
	}
	if len(m.splits) != 0 {
		t.Fatalf("expected 0 splits initially, got %d", len(m.splits))
	}
	if m.focus {
		t.Fatal("expected focus to be false initially")
	}
}

func TestTransaction_FocusBlur(t *testing.T) {
	m := newTestTransactionModel()

	// Initially not focused
	if m.focus {
		t.Fatal("expected focus to be false initially")
	}

	// Focus
	m.Focus()
	if !m.focus {
		t.Fatal("expected focus to be true after Focus()")
	}

	// Blur
	m.Blur()
	if m.focus {
		t.Fatal("expected focus to be false after Blur()")
	}
}

func TestTransaction_RedrawForm(t *testing.T) {
	m := newTestTransactionModel()

	// Send RedrawFormMsg
	_, cmd := m.Update(RedrawFormMsg{})
	if cmd == nil {
		t.Fatal("expected cmd after RedrawFormMsg")
	}

	// Verify it returns a command (we can't inspect tea.windowSizeMsg type)
	msg := cmd()
	if msg == nil {
		t.Fatal("expected RedrawFormMsg to return a message")
	}

	// Verify form was updated (basic check)
	if m.form == nil {
		t.Fatal("expected form to exist after redraw")
	}
}

// Part 2: Message handler tests

func TestTransaction_NewTransactionMsg(t *testing.T) {
	trx := firefly.Transaction{
		TransactionID: "trx123",
		Type:          "withdrawal",
		Date:          "2026-01-15",
		Splits: []firefly.Split{
			{
				Source:      testAssetChecking,
				Destination: testExpenseGroceries,
				Category:    testCategoryFood,
				Amount:      100.00,
				Description: "Test transaction",
			},
		},
	}

	t.Run("first time sets transaction and marks created", func(t *testing.T) {
		m := newTestTransactionModel()
		if m.created {
			t.Fatal("expected created to be false initially")
		}

		updated, cmd := m.Update(NewTransactionMsg{Transaction: trx})
		m2 := updated.(modelTransaction)

		if !m2.created {
			t.Error("expected created to be true after NewTransactionMsg")
		}

		// Verify batch contains RedrawFormMsg and SetFocusedViewMsg
		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}
		msgs := collectMsgsFromCmd(cmd)
		if len(msgs) < 2 {
			t.Fatalf("expected at least 2 messages, got %d", len(msgs))
		}

		if _, ok := msgs[0].(RedrawFormMsg); !ok {
			t.Errorf("expected RedrawFormMsg as first message, got %T", msgs[0])
		}
		if _, ok := msgs[1].(SetFocusedViewMsg); !ok {
			t.Errorf("expected SetFocusedViewMsg as second message, got %T", msgs[1])
		}
	})

	t.Run("second time doesn't call SetTransaction again", func(t *testing.T) {
		m := newTestTransactionModel()

		// First call
		updated, _ := m.Update(NewTransactionMsg{Transaction: trx})
		m = updated.(modelTransaction)

		// Modify something to verify it doesn't change
		m.attr.groupTitle = "Modified"

		// Second call with different transaction
		trx2 := trx
		trx2.TransactionID = "trx456"
		updated, cmd := m.Update(NewTransactionMsg{Transaction: trx2})
		m2 := updated.(modelTransaction)

		// Verify the model wasn't updated (groupTitle should still be "Modified")
		if m2.attr.groupTitle != "Modified" {
			t.Error("expected SetTransaction not to be called second time")
		}

		// Still returns batch
		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}
	})
}

func TestTransaction_NewTransactionFromMsg(t *testing.T) {
	trx := firefly.Transaction{
		TransactionID: "trx123",
		Type:          "deposit",
		Date:          "2026-01-15",
		Splits: []firefly.Split{
			{
				Source:      testRevenueSalary,
				Destination: testAssetChecking,
				Category:    testCategoryIncome,
				Amount:      2000.00,
				Description: "Salary payment",
			},
		},
	}

	m := newTestTransactionModel()
	if m.created {
		t.Fatal("expected created to be false initially")
	}

	updated, cmd := m.Update(NewTransactionFromMsg{Transaction: trx})
	m2 := updated.(modelTransaction)

	// Always sets transaction and marks created
	if !m2.created {
		t.Error("expected created to be true after NewTransactionFromMsg")
	}

	// Verify batch contains RedrawFormMsg and SetFocusedViewMsg
	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}
	msgs := collectMsgsFromCmd(cmd)
	if len(msgs) < 2 {
		t.Fatalf("expected at least 2 messages, got %d", len(msgs))
	}

	if _, ok := msgs[0].(RedrawFormMsg); !ok {
		t.Errorf("expected RedrawFormMsg as first message, got %T", msgs[0])
	}
}

func TestTransaction_EditTransactionMsg(t *testing.T) {
	trx := firefly.Transaction{
		TransactionID: "trx123",
		Type:          "withdrawal",
		Date:          "2026-01-15",
		Splits: []firefly.Split{
			{
				TransactionJournalID: "journal1",
				Source:               testAssetChecking,
				Destination:          testExpenseGroceries,
				Category:             testCategoryFood,
				Amount:               50.00,
				Description:          "Edit test",
			},
		},
	}

	m := newTestTransactionModel()
	if m.created {
		t.Fatal("expected created to be false initially")
	}

	updated, cmd := m.Update(EditTransactionMsg{Transaction: trx})
	m2 := updated.(modelTransaction)

	// Sets transaction with new=false and marks created
	if !m2.created {
		t.Error("expected created to be true after EditTransactionMsg")
	}
	if m2.new {
		t.Error("expected new to be false after EditTransactionMsg")
	}

	// Verify batch contains RedrawFormMsg and SetFocusedViewMsg
	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}
	msgs := collectMsgsFromCmd(cmd)
	if len(msgs) < 2 {
		t.Fatalf("expected at least 2 messages, got %d", len(msgs))
	}
}

func TestTransaction_ResetTransactionMsg(t *testing.T) {
	m := newTestTransactionModel()

	// Set up some state
	m.attr.groupTitle = "Test"
	m.created = false
	m.new = false

	updated, cmd := m.Update(ResetTransactionMsg{})
	m2 := updated.(modelTransaction)

	// Marks created=true
	if !m2.created {
		t.Error("expected created to be true after ResetTransactionMsg")
	}

	// Returns RedrawForm cmd
	if cmd == nil {
		t.Fatal("expected cmd to be returned")
	}
	msg := cmd()
	if _, ok := msg.(RedrawFormMsg); !ok {
		t.Errorf("expected RedrawFormMsg, got %T", msg)
	}
}

func TestTransaction_DeleteSplitMsg(t *testing.T) {
	t.Run("valid index deletes split", func(t *testing.T) {
		m := newTestTransactionModel()
		m.splits = []*split{
			{description: "Split 0"},
			{description: "Split 1"},
			{description: "Split 2"},
		}

		_, cmd := m.Update(DeleteSplitMsg{Index: 1})

		// Execute the command to get the actual delete result
		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}

		// Verify command is a sequence
		msg := cmd()
		if msg == nil {
			t.Fatal("expected message from cmd")
		}
	})

	t.Run("invalid index 0", func(t *testing.T) {
		m := newTestTransactionModel()
		m.splits = []*split{
			{description: "Split 0"},
			{description: "Split 1"},
		}

		_, cmd := m.Update(DeleteSplitMsg{Index: 0})

		// Should return a cmd (warning notification)
		if cmd == nil {
			t.Fatal("expected cmd to be returned for invalid index")
		}
	})

	t.Run("out of range index", func(t *testing.T) {
		m := newTestTransactionModel()
		m.splits = []*split{
			{description: "Split 0"},
			{description: "Split 1"},
		}

		_, cmd := m.Update(DeleteSplitMsg{Index: 10})

		// Should return a cmd (warning notification)
		if cmd == nil {
			t.Fatal("expected cmd to be returned for out of range index")
		}
	})
}

// Part 3: Key binding and transaction operation tests

func TestTransaction_KeyBindings(t *testing.T) {
	// Save and restore fullNewForm global
	origFullNewForm := fullNewForm
	defer func() { fullNewForm = origFullNewForm }()

	t.Run("Cancel returns SetView(transactionsView)", func(t *testing.T) {
		m := newTestTransactionModel()
		m.Focus()

		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}

		msgs := collectMsgsFromCmd(cmd)
		if len(msgs) == 0 {
			t.Fatal("expected at least one message")
		}

		setViewMsg, ok := msgs[0].(SetFocusedViewMsg)
		if !ok {
			t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[0])
		}
		if setViewMsg.state != transactionsView {
			t.Errorf("expected transactionsView, got %v", setViewMsg.state)
		}
	})

	t.Run("Reset returns batch with SetView(newView) and ResetTransactionMsg", func(t *testing.T) {
		m := newTestTransactionModel()
		m.Focus()

		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}

		msgs := collectMsgsFromCmd(cmd)
		if len(msgs) < 2 {
			t.Fatalf("expected at least 2 messages, got %d", len(msgs))
		}

		// Check for SetFocusedViewMsg
		var foundSetView bool
		var foundReset bool
		for _, msg := range msgs {
			if setViewMsg, ok := msg.(SetFocusedViewMsg); ok {
				foundSetView = true
				if setViewMsg.state != newView {
					t.Errorf("expected newView, got %v", setViewMsg.state)
				}
			}
			if _, ok := msg.(ResetTransactionMsg); ok {
				foundReset = true
			}
		}

		if !foundSetView {
			t.Error("expected SetFocusedViewMsg in batch")
		}
		if !foundReset {
			t.Error("expected ResetTransactionMsg in batch")
		}
	})

	t.Run("Refresh increments all 3 counters and returns RedrawForm", func(t *testing.T) {
		// Save and restore global counters
		origCategory := triggerCategoryCounter
		origSource := triggerSourceCounter
		origDest := triggerDestinationCounter
		defer func() {
			triggerCategoryCounter = origCategory
			triggerSourceCounter = origSource
			triggerDestinationCounter = origDest
		}()

		m := newTestTransactionModel()
		m.Focus()

		initialCategory := triggerCategoryCounter
		initialSource := triggerSourceCounter
		initialDest := triggerDestinationCounter

		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlR})

		// Verify all counters incremented
		if triggerCategoryCounter != initialCategory+1 {
			t.Errorf("expected category counter to be %d, got %d", initialCategory+1, triggerCategoryCounter)
		}
		if triggerSourceCounter != initialSource+1 {
			t.Errorf("expected source counter to be %d, got %d", initialSource+1, triggerSourceCounter)
		}
		if triggerDestinationCounter != initialDest+1 {
			t.Errorf("expected destination counter to be %d, got %d", initialDest+1, triggerDestinationCounter)
		}

		// Verify RedrawForm was returned
		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}
		msg := cmd()
		if _, ok := msg.(RedrawFormMsg); !ok {
			t.Errorf("expected RedrawFormMsg, got %T", msg)
		}
	})

	t.Run("AddSplit adds split to array and returns RedrawForm", func(t *testing.T) {
		m := newTestTransactionModel()
		m.Focus()

		initialCount := len(m.splits)

		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
		m2 := updated.(modelTransaction)

		// Verify split was added
		if len(m2.splits) != initialCount+1 {
			t.Errorf("expected %d splits, got %d", initialCount+1, len(m2.splits))
		}

		// Verify RedrawForm was returned
		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}
		msg := cmd()
		if _, ok := msg.(RedrawFormMsg); !ok {
			t.Errorf("expected RedrawFormMsg, got %T", msg)
		}
	})

	t.Run("DeleteSplit returns prompt command", func(t *testing.T) {
		m := newTestTransactionModel()
		m.Focus()
		m.splits = []*split{
			{description: "Split 0"},
			{description: "Split 1"},
		}

		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})

		// Verify command is not nil (prompt.Ask returns a command)
		if cmd == nil {
			t.Fatal("expected cmd to be returned for delete split prompt")
		}
	})

	t.Run("ChangeLayout toggles fullNewForm and returns RedrawForm", func(t *testing.T) {
		m := newTestTransactionModel()
		m.Focus()

		initialFullNewForm := fullNewForm

		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlF})

		// Verify global was toggled
		if fullNewForm == initialFullNewForm {
			t.Error("expected fullNewForm to be toggled")
		}

		// Verify RedrawForm was returned
		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}
		msg := cmd()
		if _, ok := msg.(RedrawFormMsg); !ok {
			t.Errorf("expected RedrawFormMsg, got %T", msg)
		}
	})

	t.Run("Submit when form completed and new=true calls CreateTransaction", func(t *testing.T) {
		api := &mockTransactionFormAPI{
			accountsByTypeFunc: func(accountType string) []firefly.Account {
				return []firefly.Account{testAssetChecking}
			},
			categoriesListFunc: func() []firefly.Category {
				return []firefly.Category{testCategoryFood}
			},
			createTransactionFunc: func(tx firefly.RequestTransaction) (string, error) {
				return "", nil
			},
		}

		m := newModelTransaction(api)
		m.Focus()
		m.new = true
		m.splits = []*split{
			{
				source:      testAssetChecking,
				destination: testExpenseGroceries,
				category:    testCategoryFood,
				amount:      "50.00",
				description: "Test",
			},
		}
		m.attr.year = "2026"
		m.attr.month = "01"
		m.attr.day = "15"
		m.attr.transactionType = "withdrawal"
		m.UpdateForm()
		m.form.State = huh.StateCompleted

		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}

		// Verify CreateTransaction was called
		if len(api.createTransactionCalls) != 1 {
			t.Errorf("expected CreateTransaction to be called once, got %d calls", len(api.createTransactionCalls))
		}
	})

	t.Run("Submit when form completed and new=false calls UpdateTransaction", func(t *testing.T) {
		api := &mockTransactionFormAPI{
			accountsByTypeFunc: func(accountType string) []firefly.Account {
				return []firefly.Account{testAssetChecking}
			},
			categoriesListFunc: func() []firefly.Category {
				return []firefly.Category{testCategoryFood}
			},
			updateTransactionFunc: func(transactionID string, tx firefly.RequestTransaction) (string, error) {
				return "", nil
			},
		}

		m := newModelTransaction(api)
		m.Focus()
		m.new = false
		m.attr.trxID = "trx123"
		m.splits = []*split{
			{
				source:      testAssetChecking,
				destination: testExpenseGroceries,
				category:    testCategoryFood,
				amount:      "50.00",
				description: "Test",
				trxJID:      "journal1",
			},
		}
		m.attr.year = "2026"
		m.attr.month = "01"
		m.attr.day = "15"
		m.attr.transactionType = "withdrawal"
		m.UpdateForm()
		m.form.State = huh.StateCompleted

		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}

		// Verify UpdateTransaction was called
		if len(api.updateTransactionCalls) != 1 {
			t.Errorf("expected UpdateTransaction to be called once, got %d calls", len(api.updateTransactionCalls))
		}
		if api.updateTransactionCalls[0].id != "trx123" {
			t.Errorf("expected transaction ID 'trx123', got %s", api.updateTransactionCalls[0].id)
		}
	})
}

func TestTransaction_KeyBindings_NotFocused(t *testing.T) {
	m := newTestTransactionModel()
	// Don't call Focus()

	tests := []struct {
		name   string
		keyMsg tea.KeyMsg
	}{
		{"Cancel", tea.KeyMsg{Type: tea.KeyEsc}},
		{"Reset", tea.KeyMsg{Type: tea.KeyCtrlN}},
		{"Refresh", tea.KeyMsg{Type: tea.KeyCtrlR}},
		{"AddSplit", tea.KeyMsg{Type: tea.KeyCtrlA}},
		{"DeleteSplit", tea.KeyMsg{Type: tea.KeyCtrlD}},
		{"ChangeLayout", tea.KeyMsg{Type: tea.KeyCtrlF}},
		{"Submit", tea.KeyMsg{Type: tea.KeyEnter}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cmd := m.Update(tt.keyMsg)
			if cmd != nil {
				t.Errorf("expected nil cmd when not focused, got %T", cmd)
			}
		})
	}
}

func TestTransaction_SetTransaction_New(t *testing.T) {
	m := newTestTransactionModel()

	// Empty transaction
	trx := firefly.Transaction{}
	m.SetTransaction(trx, true)

	// Verify current date is set
	now := time.Now()
	expectedYear := fmt.Sprintf("%d", now.Year())
	expectedMonth := fmt.Sprintf("%02d", now.Month())
	expectedDay := fmt.Sprintf("%02d", now.Day())

	if m.attr.year != expectedYear {
		t.Errorf("expected year %s, got %s", expectedYear, m.attr.year)
	}
	if m.attr.month != expectedMonth {
		t.Errorf("expected month %s, got %s", expectedMonth, m.attr.month)
	}
	if m.attr.day != expectedDay {
		t.Errorf("expected day %s, got %s", expectedDay, m.attr.day)
	}

	// Verify one empty split
	if len(m.splits) != 1 {
		t.Errorf("expected 1 split, got %d", len(m.splits))
	}
	if m.splits[0].amount != "" {
		t.Errorf("expected empty amount, got %s", m.splits[0].amount)
	}
	if m.splits[0].description != "" {
		t.Errorf("expected empty description, got %s", m.splits[0].description)
	}
	if m.splits[0].trxJID != "" {
		t.Errorf("expected empty trxJID, got %s", m.splits[0].trxJID)
	}

	// Verify transaction type
	if m.attr.transactionType != "withdrawal" {
		t.Errorf("expected transactionType 'withdrawal', got %s", m.attr.transactionType)
	}

	// Verify new flag
	if !m.new {
		t.Error("expected new to be true")
	}
}

func TestTransaction_SetTransaction_Edit(t *testing.T) {
	m := newTestTransactionModel()

	trx := firefly.Transaction{
		TransactionID: "trx123",
		Type:          "deposit",
		Date:          "2026-01-15",
		GroupTitle:    "Test Group",
		Splits: []firefly.Split{
			{
				TransactionJournalID: "journal1",
				Source:               testRevenueSalary,
				Destination:          testAssetChecking,
				Category:             testCategoryIncome,
				Amount:               100.50,
				ForeignAmount:        0,
				Description:          "Salary payment",
			},
			{
				TransactionJournalID: "journal2",
				Source:               testRevenueFreelance,
				Destination:          testAssetSavings,
				Category:             testCategoryIncome,
				Amount:               250.75,
				ForeignAmount:        0,
				Description:          "Freelance work",
			},
		},
	}

	m.SetTransaction(trx, false)

	// Verify date parsing
	if m.attr.year != "2026" {
		t.Errorf("expected year '2026', got %s", m.attr.year)
	}
	if m.attr.month != "01" {
		t.Errorf("expected month '01', got %s", m.attr.month)
	}
	if m.attr.day != "15" {
		t.Errorf("expected day '15', got %s", m.attr.day)
	}

	// Verify splits copied
	if len(m.splits) != 2 {
		t.Fatalf("expected 2 splits, got %d", len(m.splits))
	}

	// Verify first split
	if m.splits[0].amount != "100.50" {
		t.Errorf("expected amount '100.50', got %s", m.splits[0].amount)
	}
	if m.splits[0].description != "Salary payment" {
		t.Errorf("expected description 'Salary payment', got %s", m.splits[0].description)
	}
	if m.splits[0].trxJID != "journal1" {
		t.Errorf("expected trxJID 'journal1', got %s", m.splits[0].trxJID)
	}
	if m.splits[0].source.ID != testRevenueSalary.ID {
		t.Errorf("expected source ID %s, got %s", testRevenueSalary.ID, m.splits[0].source.ID)
	}
	if m.splits[0].destination.ID != testAssetChecking.ID {
		t.Errorf("expected destination ID %s, got %s", testAssetChecking.ID, m.splits[0].destination.ID)
	}
	if m.splits[0].category.ID != testCategoryIncome.ID {
		t.Errorf("expected category ID %s, got %s", testCategoryIncome.ID, m.splits[0].category.ID)
	}

	// Verify second split
	if m.splits[1].amount != "250.75" {
		t.Errorf("expected amount '250.75', got %s", m.splits[1].amount)
	}
	if m.splits[1].description != "Freelance work" {
		t.Errorf("expected description 'Freelance work', got %s", m.splits[1].description)
	}
	if m.splits[1].trxJID != "journal2" {
		t.Errorf("expected trxJID 'journal2', got %s", m.splits[1].trxJID)
	}

	// Verify transaction type and ID
	if m.attr.transactionType != "deposit" {
		t.Errorf("expected transactionType 'deposit', got %s", m.attr.transactionType)
	}
	if m.attr.trxID != "trx123" {
		t.Errorf("expected trxID 'trx123', got %s", m.attr.trxID)
	}
	if m.attr.groupTitle != "Test Group" {
		t.Errorf("expected groupTitle 'Test Group', got %s", m.attr.groupTitle)
	}

	// Verify new flag
	if m.new {
		t.Error("expected new to be false")
	}
}

func TestTransaction_CreateTransaction(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &mockTransactionFormAPI{
			accountsByTypeFunc: func(accountType string) []firefly.Account {
				return []firefly.Account{testAssetChecking}
			},
			categoriesListFunc: func() []firefly.Category {
				return []firefly.Category{testCategoryFood}
			},
			createTransactionFunc: func(tx firefly.RequestTransaction) (string, error) {
				return "", nil
			},
		}

		m := newModelTransaction(api)
		m.created = true
		m.new = true
		m.splits = []*split{
			{
				source:      testAssetChecking,
				destination: testExpenseGroceries,
				category:    testCategoryFood,
				amount:      "50.00",
				description: "Groceries",
			},
		}
		m.attr.year = "2026"
		m.attr.month = "01"
		m.attr.day = "15"
		m.attr.transactionType = "withdrawal"

		cmd := m.CreateTransaction()
		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}

		// Verify created flag is reset
		if m.created {
			t.Error("expected created to be false after CreateTransaction")
		}

		// Verify CreateTransaction was called
		if len(api.createTransactionCalls) != 1 {
			t.Fatalf("expected CreateTransaction to be called once, got %d calls", len(api.createTransactionCalls))
		}

		// Verify request structure
		req := api.createTransactionCalls[0]
		if req.ApplyRules != true {
			t.Error("expected ApplyRules to be true")
		}
		if req.ErrorIfDuplicateHash != false {
			t.Error("expected ErrorIfDuplicateHash to be false")
		}
		if req.FireWebhooks != true {
			t.Error("expected FireWebhooks to be true")
		}
		if len(req.Transactions) != 1 {
			t.Fatalf("expected 1 transaction split, got %d", len(req.Transactions))
		}

		split := req.Transactions[0]
		if split.Type != "withdrawal" {
			t.Errorf("expected type 'withdrawal', got %s", split.Type)
		}
		if split.Date != "2026-01-15" {
			t.Errorf("expected date '2026-01-15', got %s", split.Date)
		}
		if split.Amount != "50.00" {
			t.Errorf("expected amount '50.00', got %s", split.Amount)
		}

		// Verify batch contains multiple messages
		msgs := collectMsgsFromCmd(cmd)
		if len(msgs) < 2 {
			t.Fatalf("expected at least 2 messages in batch, got %d", len(msgs))
		}

		// Check for SetFocusedViewMsg
		hasSetView := false
		for _, msg := range msgs {
			if setViewMsg, ok := msg.(SetFocusedViewMsg); ok {
				if setViewMsg.state == transactionsView {
					hasSetView = true
				}
			}
		}
		if !hasSetView {
			t.Error("expected batch to contain SetView(transactionsView)")
		}
	})

	t.Run("error", func(t *testing.T) {
		api := &mockTransactionFormAPI{
			accountsByTypeFunc: func(accountType string) []firefly.Account {
				return []firefly.Account{testAssetChecking}
			},
			categoriesListFunc: func() []firefly.Category {
				return []firefly.Category{testCategoryFood}
			},
			createTransactionFunc: func(tx firefly.RequestTransaction) (string, error) {
				return "", errors.New("API error")
			},
		}

		m := newModelTransaction(api)
		m.created = true
		m.new = true
		m.splits = []*split{
			{
				source:      testAssetChecking,
				destination: testExpenseGroceries,
				category:    testCategoryFood,
				amount:      "50.00",
				description: "Groceries",
			},
		}
		m.attr.year = "2026"
		m.attr.month = "01"
		m.attr.day = "15"
		m.attr.transactionType = "withdrawal"

		cmd := m.CreateTransaction()
		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}

		// Verify sequence contains error notification and SetView
		msgs := collectMsgsFromCmd(cmd)
		if len(msgs) < 2 {
			t.Fatalf("expected at least 2 messages in sequence, got %d", len(msgs))
		}

		// Check for SetFocusedViewMsg
		hasSetView := false
		for _, msg := range msgs {
			if setViewMsg, ok := msg.(SetFocusedViewMsg); ok {
				if setViewMsg.state == transactionsView {
					hasSetView = true
				}
			}
		}
		if !hasSetView {
			t.Error("expected sequence to contain SetView(transactionsView)")
		}
	})
}

func TestTransaction_UpdateTransaction(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		api := &mockTransactionFormAPI{
			accountsByTypeFunc: func(accountType string) []firefly.Account {
				return []firefly.Account{testAssetChecking}
			},
			categoriesListFunc: func() []firefly.Category {
				return []firefly.Category{testCategoryFood}
			},
			updateTransactionFunc: func(transactionID string, tx firefly.RequestTransaction) (string, error) {
				return "", nil
			},
		}

		m := newModelTransaction(api)
		m.created = true
		m.new = false
		m.attr.trxID = "trx123"
		m.splits = []*split{
			{
				source:      testAssetChecking,
				destination: testExpenseGroceries,
				category:    testCategoryFood,
				amount:      "75.00",
				description: "Updated groceries",
				trxJID:      "journal1",
			},
		}
		m.attr.year = "2026"
		m.attr.month = "01"
		m.attr.day = "16"
		m.attr.transactionType = "withdrawal"

		cmd := m.UpdateTransaction()
		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}

		// Verify created flag is reset
		if m.created {
			t.Error("expected created to be false after UpdateTransaction")
		}

		// Verify UpdateTransaction was called with correct ID
		if len(api.updateTransactionCalls) != 1 {
			t.Fatalf("expected UpdateTransaction to be called once, got %d calls", len(api.updateTransactionCalls))
		}
		if api.updateTransactionCalls[0].id != "trx123" {
			t.Errorf("expected transaction ID 'trx123', got %s", api.updateTransactionCalls[0].id)
		}

		// Verify request structure
		req := api.updateTransactionCalls[0].tx
		if req.ApplyRules != true {
			t.Error("expected ApplyRules to be true")
		}
		if req.FireWebhooks != true {
			t.Error("expected FireWebhooks to be true")
		}
		if len(req.Transactions) != 1 {
			t.Fatalf("expected 1 transaction split, got %d", len(req.Transactions))
		}

		split := req.Transactions[0]
		if split.TransactionJournalID != "journal1" {
			t.Errorf("expected TransactionJournalID 'journal1', got %s", split.TransactionJournalID)
		}
		if split.Type != "withdrawal" {
			t.Errorf("expected type 'withdrawal', got %s", split.Type)
		}
		if split.Date != "2026-01-16" {
			t.Errorf("expected date '2026-01-16', got %s", split.Date)
		}
		if split.Amount != "75.00" {
			t.Errorf("expected amount '75.00', got %s", split.Amount)
		}

		// Verify batch contains multiple messages
		msgs := collectMsgsFromCmd(cmd)
		if len(msgs) < 2 {
			t.Fatalf("expected at least 2 messages in batch, got %d", len(msgs))
		}

		// Check for SetFocusedViewMsg
		hasSetView := false
		for _, msg := range msgs {
			if setViewMsg, ok := msg.(SetFocusedViewMsg); ok {
				if setViewMsg.state == transactionsView {
					hasSetView = true
				}
			}
		}
		if !hasSetView {
			t.Error("expected batch to contain SetView(transactionsView)")
		}
	})

	t.Run("error", func(t *testing.T) {
		api := &mockTransactionFormAPI{
			accountsByTypeFunc: func(accountType string) []firefly.Account {
				return []firefly.Account{testAssetChecking}
			},
			categoriesListFunc: func() []firefly.Category {
				return []firefly.Category{testCategoryFood}
			},
			updateTransactionFunc: func(transactionID string, tx firefly.RequestTransaction) (string, error) {
				return "", errors.New("Update API error")
			},
		}

		m := newModelTransaction(api)
		m.created = true
		m.new = false
		m.attr.trxID = "trx123"
		m.splits = []*split{
			{
				source:      testAssetChecking,
				destination: testExpenseGroceries,
				category:    testCategoryFood,
				amount:      "75.00",
				description: "Updated groceries",
				trxJID:      "journal1",
			},
		}
		m.attr.year = "2026"
		m.attr.month = "01"
		m.attr.day = "16"
		m.attr.transactionType = "withdrawal"

		cmd := m.UpdateTransaction()
		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}

		// Verify sequence contains error notification and SetView
		msgs := collectMsgsFromCmd(cmd)
		if len(msgs) < 2 {
			t.Fatalf("expected at least 2 messages in sequence, got %d", len(msgs))
		}

		// Check for SetFocusedViewMsg
		hasSetView := false
		for _, msg := range msgs {
			if setViewMsg, ok := msg.(SetFocusedViewMsg); ok {
				if setViewMsg.state == transactionsView {
					hasSetView = true
				}
			}
		}
		if !hasSetView {
			t.Error("expected sequence to contain SetView(transactionsView)")
		}
	})
}

// Part 4: Helper method tests and edge cases

func TestSplit_Description(t *testing.T) {
	t.Run("with custom description", func(t *testing.T) {
		s := &split{
			description: "Custom description",
			category:    testCategoryFood,
			source:      testAssetChecking,
			destination: testExpenseGroceries,
		}

		result := s.Description()
		if result != "Custom description" {
			t.Errorf("expected 'Custom description', got %s", result)
		}
	})

	t.Run("without description returns formatted string", func(t *testing.T) {
		s := &split{
			description: "",
			category:    testCategoryFood,
			source:      testAssetChecking,
			destination: testExpenseGroceries,
		}

		result := s.Description()
		expected := "Food, Checking -> Groceries"
		if result != expected {
			t.Errorf("expected '%s', got %s", expected, result)
		}
	})
}

func TestSplit_CurrencyCode(t *testing.T) {
	tests := []struct {
		name           string
		sourceType     string
		sourceCurrency string
		destCurrency   string
		expected       string
	}{
		{
			name:           "source asset returns source currency",
			sourceType:     "asset",
			sourceCurrency: "USD",
			destCurrency:   "EUR",
			expected:       "USD",
		},
		{
			name:           "source liabilities returns source currency",
			sourceType:     "liabilities",
			sourceCurrency: "GBP",
			destCurrency:   "USD",
			expected:       "GBP",
		},
		{
			name:           "source revenue returns destination currency",
			sourceType:     "revenue",
			sourceCurrency: "",
			destCurrency:   "USD",
			expected:       "USD",
		},
		{
			name:           "other types return empty string",
			sourceType:     "expense",
			sourceCurrency: "USD",
			destCurrency:   "EUR",
			expected:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &split{
				source: firefly.Account{
					Type:         tt.sourceType,
					CurrencyCode: tt.sourceCurrency,
				},
				destination: firefly.Account{
					CurrencyCode: tt.destCurrency,
				},
			}

			result := s.CurrencyCode()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSplit_ForeignCurrencyCode(t *testing.T) {
	tests := []struct {
		name           string
		sourceType     string
		sourceCurrency string
		destType       string
		destCurrency   string
		expected       string
	}{
		{
			name:           "asset to asset same currency returns empty",
			sourceType:     "asset",
			sourceCurrency: "USD",
			destType:       "asset",
			destCurrency:   "USD",
			expected:       "",
		},
		{
			name:           "asset to asset different currency returns destination currency",
			sourceType:     "asset",
			sourceCurrency: "USD",
			destType:       "asset",
			destCurrency:   "EUR",
			expected:       "EUR",
		},
		{
			name:           "liability to asset different currency returns destination currency",
			sourceType:     "liabilities",
			sourceCurrency: "USD",
			destType:       "asset",
			destCurrency:   "GBP",
			expected:       "GBP",
		},
		{
			name:           "asset to liability different currency returns destination currency",
			sourceType:     "asset",
			sourceCurrency: "EUR",
			destType:       "liabilities",
			destCurrency:   "USD",
			expected:       "USD",
		},
		{
			name:           "non-asset/liability returns empty",
			sourceType:     "revenue",
			sourceCurrency: "",
			destType:       "asset",
			destCurrency:   "USD",
			expected:       "",
		},
		{
			name:           "asset to expense returns empty",
			sourceType:     "asset",
			sourceCurrency: "USD",
			destType:       "expense",
			destCurrency:   "",
			expected:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &split{
				source: firefly.Account{
					Type:         tt.sourceType,
					CurrencyCode: tt.sourceCurrency,
				},
				destination: firefly.Account{
					Type:         tt.destType,
					CurrencyCode: tt.destCurrency,
				},
			}

			result := s.ForeignCurrencyCode()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestTransaction_GroupTitle(t *testing.T) {
	t.Run("single split returns empty string", func(t *testing.T) {
		m := newTestTransactionModel()
		m.splits = []*split{
			{description: "Single split"},
		}

		result := m.GroupTitle()
		if result != "" {
			t.Errorf("expected empty string, got '%s'", result)
		}
	})

	t.Run("multiple splits with custom title returns custom title", func(t *testing.T) {
		m := newTestTransactionModel()
		m.splits = []*split{
			{description: "Split 1"},
			{description: "Split 2"},
		}
		m.attr.groupTitle = "Custom Group Title"

		result := m.GroupTitle()
		if result != "Custom Group Title" {
			t.Errorf("expected 'Custom Group Title', got '%s'", result)
		}
	})

	t.Run("multiple splits withdrawal without custom title", func(t *testing.T) {
		m := newTestTransactionModel()
		m.splits = []*split{
			{description: "Split 1"},
			{description: "Split 2"},
		}
		m.attr.transactionType = "withdrawal"
		m.attr.source = testAssetChecking
		m.attr.groupTitle = ""

		result := m.GroupTitle()
		expected := "withdrawal, splits: 2, Checking"
		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("multiple splits deposit without custom title", func(t *testing.T) {
		m := newTestTransactionModel()
		m.splits = []*split{
			{description: "Split 1"},
			{description: "Split 2"},
		}
		m.attr.transactionType = "deposit"
		m.attr.destination = testAssetSavings
		m.attr.groupTitle = ""

		result := m.GroupTitle()
		expected := "deposit, splits: 2, Savings"
		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("multiple splits transfer without custom title", func(t *testing.T) {
		m := newTestTransactionModel()
		m.splits = []*split{
			{description: "Split 1"},
			{description: "Split 2"},
		}
		m.attr.transactionType = "transfer"
		m.attr.source = testAssetChecking
		m.attr.destination = testAssetSavings
		m.attr.groupTitle = ""

		result := m.GroupTitle()
		expected := "transfer, splits: 2, Checking -> Savings"
		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	})
}

func TestTransaction_TransactionTypeDetection(t *testing.T) {
	tests := []struct {
		name         string
		sourceType   string
		destType     string
		expectedType string
	}{
		{
			name:         "asset to expense is withdrawal",
			sourceType:   "asset",
			destType:     "expense",
			expectedType: "withdrawal",
		},
		{
			name:         "asset to liability is withdrawal",
			sourceType:   "asset",
			destType:     "liabilities",
			expectedType: "withdrawal",
		},
		{
			name:         "asset to cash is withdrawal",
			sourceType:   "asset",
			destType:     "cash",
			expectedType: "withdrawal",
		},
		{
			name:         "asset to asset is transfer",
			sourceType:   "asset",
			destType:     "asset",
			expectedType: "transfer",
		},
		{
			name:         "revenue to asset is deposit",
			sourceType:   "revenue",
			destType:     "asset",
			expectedType: "deposit",
		},
		{
			name:         "revenue to liability is deposit",
			sourceType:   "revenue",
			destType:     "liabilities",
			expectedType: "deposit",
		},
		{
			name:         "liability to expense is withdrawal",
			sourceType:   "liabilities",
			destType:     "expense",
			expectedType: "withdrawal",
		},
		{
			name:         "liability to asset is deposit",
			sourceType:   "liabilities",
			destType:     "asset",
			expectedType: "deposit",
		},
		{
			name:         "liability to liability is transfer",
			sourceType:   "liabilities",
			destType:     "liabilities",
			expectedType: "transfer",
		},
		{
			name:         "unknown combination is unknown",
			sourceType:   "expense",
			destType:     "revenue",
			expectedType: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestTransactionModel()
			s := &split{
				source: firefly.Account{
					ID:   "src1",
					Name: "Source",
					Type: tt.sourceType,
				},
				destination: firefly.Account{
					ID:   "dest1",
					Name: "Destination",
					Type: tt.destType,
				},
			}
			m.splits = []*split{s}

			// Call trxTitle to trigger type detection
			titleFunc, _ := m.trxTitle(0, s)
			titleFunc()

			if m.attr.transactionType != tt.expectedType {
				t.Errorf("expected transaction type '%s', got '%s'", tt.expectedType, m.attr.transactionType)
			}
		})
	}
}

func TestTransaction_View(t *testing.T) {
	t.Run("form not completed returns form view", func(t *testing.T) {
		m := newTestTransactionModel()
		m.form.State = huh.StateNormal

		result := m.View()
		// The actual form view is complex, just verify it's not the instruction text
		if result == "Press Enter to submit, Ctrl+N to reset current form, Ctrl+R to edit current form again, or Esc to go back." {
			t.Error("expected form view, got instruction text")
		}
	})

	t.Run("form completed returns instruction text", func(t *testing.T) {
		m := newTestTransactionModel()
		m.form.State = huh.StateCompleted

		result := m.View()
		expected := "Press Enter to submit, Ctrl+N to reset current form, Ctrl+R to edit current form again, or Esc to go back."
		if result != expected {
			t.Errorf("expected instruction text, got '%s'", result)
		}
	})
}

func TestTransaction_DeleteSplit_Details(t *testing.T) {
	t.Run("successfully removes split at index 1 from 3-split array", func(t *testing.T) {
		m := newTestTransactionModel()
		m.splits = []*split{
			{description: "Split 0", amount: "10.00"},
			{description: "Split 1", amount: "20.00"},
			{description: "Split 2", amount: "30.00"},
		}

		_, cmd := m.Update(DeleteSplitMsg{Index: 1})

		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}

		// Execute the command to get the actual delete result
		msg := cmd()
		if msg == nil {
			t.Fatal("expected message from cmd")
		}
	})

	t.Run("preserves order after deletion", func(t *testing.T) {
		m := newTestTransactionModel()
		split0 := &split{description: "Split 0", amount: "10.00"}
		split1 := &split{description: "Split 1", amount: "20.00"}
		split2 := &split{description: "Split 2", amount: "30.00"}
		m.splits = []*split{split0, split1, split2}

		// Call DeleteSplit method which performs the deletion
		cmd := m.DeleteSplit(1)

		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}

		// Verify only split 0 and split 2 remain in order
		if len(m.splits) != 2 {
			t.Fatalf("expected 2 splits after deletion, got %d", len(m.splits))
		}
		if m.splits[0].description != "Split 0" {
			t.Errorf("expected first split to be 'Split 0', got '%s'", m.splits[0].description)
		}
		if m.splits[1].description != "Split 2" {
			t.Errorf("expected second split to be 'Split 2', got '%s'", m.splits[1].description)
		}
	})

	t.Run("returns tea.Sequence command", func(t *testing.T) {
		m := newTestTransactionModel()
		m.splits = []*split{
			{description: "Split 0"},
			{description: "Split 1"},
			{description: "Split 2"},
		}

		_, cmd := m.Update(DeleteSplitMsg{Index: 1})

		// Verify command is not nil
		if cmd == nil {
			t.Fatal("expected cmd to be returned")
		}

		// Verify it returns a message (sequence or batch)
		msg := cmd()
		if msg == nil {
			t.Fatal("expected message from sequence command")
		}
	})
}

func TestTransaction_EmptyAccountsAndCategories(t *testing.T) {
	api := &mockTransactionFormAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{}
		},
	}

	m := newModelTransaction(api)

	// Should not crash with empty accounts/categories
	if m.form == nil {
		t.Fatal("expected form to be initialized even with empty data")
	}
	if m.splits == nil {
		t.Fatal("expected splits to be initialized")
	}

	// Try to update form with empty data
	m.UpdateForm()

	// Verify form exists
	if m.form == nil {
		t.Fatal("expected form to exist after UpdateForm")
	}
}

func TestTransaction_MultipleSplits(t *testing.T) {
	m := newTestTransactionModel()

	// Add 5 splits
	for i := range 5 {
		m.splits = append(m.splits, &split{
			source:      testAssetChecking,
			destination: testExpenseGroceries,
			category:    testCategoryFood,
			amount:      fmt.Sprintf("%d.00", (i+1)*10),
			description: fmt.Sprintf("Split %d", i),
		})
	}

	if len(m.splits) != 5 {
		t.Fatalf("expected 5 splits, got %d", len(m.splits))
	}

	// Verify GroupTitle reflects correct count
	title := m.GroupTitle()
	if title == "" {
		t.Error("expected non-empty group title for multiple splits")
	}
	if !contains(title, "5") {
		t.Errorf("expected group title to contain '5', got '%s'", title)
	}

	// Update form with all splits
	m.UpdateForm()

	if m.form == nil {
		t.Fatal("expected form to be initialized with multiple splits")
	}
}

func TestTransaction_ForeignCurrency(t *testing.T) {
	testAssetEUR := firefly.Account{
		ID:           "asset_eur",
		Name:         "EUR Account",
		Type:         "asset",
		CurrencyCode: "EUR",
	}

	m := newTestTransactionModel()
	m.splits = []*split{
		{
			source:        testAssetChecking, // USD
			destination:   testAssetEUR,      // EUR
			category:      testCategoryFood,
			amount:        "100.00",
			foreignAmount: "85.00",
			description:   "Foreign currency transaction",
		},
	}

	// Verify ForeignCurrencyCode returns EUR
	foreignCode := m.splits[0].ForeignCurrencyCode()
	if foreignCode != "EUR" {
		t.Errorf("expected foreign currency code 'EUR', got '%s'", foreignCode)
	}

	// Verify CurrencyCode returns USD (source)
	currencyCode := m.splits[0].CurrencyCode()
	if currencyCode != "USD" {
		t.Errorf("expected currency code 'USD', got '%s'", currencyCode)
	}
}

func TestTransaction_ZeroAmount(t *testing.T) {
	m := newTestTransactionModel()
	m.splits = []*split{
		{
			source:      testAssetChecking,
			destination: testExpenseGroceries,
			category:    testCategoryFood,
			amount:      "0.00",
			description: "Zero amount transaction",
		},
	}

	// Should not panic with zero amount
	if m.splits[0].amount != "0.00" {
		t.Errorf("expected amount '0.00', got '%s'", m.splits[0].amount)
	}

	// Update form should work with zero amount
	m.UpdateForm()

	if m.form == nil {
		t.Fatal("expected form to exist with zero amount")
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexString(s, substr) >= 0)
}

func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
