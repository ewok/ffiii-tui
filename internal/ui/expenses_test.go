/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/

package ui

import (
	"errors"
	"testing"

	"ffiii-tui/internal/firefly"
	"ffiii-tui/internal/ui/notify"
	"ffiii-tui/internal/ui/prompt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mockExpenseAPI struct {
	updateAccountsFunc          func(accountType string) error
	accountsByTypeFunc          func(accountType string) []firefly.Account
	accountBalanceFunc          func(accountID string) float64
	createExpenseAccountFunc    func(name string) error
	updateExpenseInsightsFunc   func() error
	getExpenseDiffFunc          func(accountID string) float64
	getTotalExpenseDiffFunc     func() float64
	primaryCurrencyFunc         func() firefly.Currency
	updateAccountsCalledWith    []string
	createExpenseCalledWith     []string
	updateExpenseInsightsCalled bool
}

func (m *mockExpenseAPI) UpdateAccounts(accountType string) error {
	m.updateAccountsCalledWith = append(m.updateAccountsCalledWith, accountType)
	if m.updateAccountsFunc != nil {
		return m.updateAccountsFunc(accountType)
	}
	return nil
}

func (m *mockExpenseAPI) AccountsByType(accountType string) []firefly.Account {
	if m.accountsByTypeFunc != nil {
		return m.accountsByTypeFunc(accountType)
	}
	return nil
}

func (m *mockExpenseAPI) AccountBalance(accountID string) float64 {
	if m.accountBalanceFunc != nil {
		return m.accountBalanceFunc(accountID)
	}
	return 0
}

func (m *mockExpenseAPI) CreateExpenseAccount(name string) error {
	m.createExpenseCalledWith = append(m.createExpenseCalledWith, name)
	if m.createExpenseAccountFunc != nil {
		return m.createExpenseAccountFunc(name)
	}
	return nil
}

func (m *mockExpenseAPI) UpdateExpenseInsights() error {
	m.updateExpenseInsightsCalled = true
	if m.updateExpenseInsightsFunc != nil {
		return m.updateExpenseInsightsFunc()
	}
	return nil
}

func (m *mockExpenseAPI) GetExpenseDiff(accountID string) float64 {
	if m.getExpenseDiffFunc != nil {
		return m.getExpenseDiffFunc(accountID)
	}
	return 0
}

func (m *mockExpenseAPI) GetTotalExpenseDiff() float64 {
	if m.getTotalExpenseDiffFunc != nil {
		return m.getTotalExpenseDiffFunc()
	}
	return 0
}

func (m *mockExpenseAPI) PrimaryCurrency() firefly.Currency {
	if m.primaryCurrencyFunc != nil {
		return m.primaryCurrencyFunc()
	}
	return firefly.Currency{Code: "USD", Symbol: "$"}
}

func newFocusedExpensesModelWithAccount(t *testing.T, acc firefly.Account) modelExpenses {
	t.Helper()

	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			if accountType != "expense" {
				t.Fatalf("expected accountType 'expense', got %q", accountType)
			}
			return []firefly.Account{acc}
		},
		getExpenseDiffFunc: func(accountID string) float64 { return 0 },
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}

	m := newModelExpenses(api)
	(&m).Focus()
	return m
}

// Basic functionality tests

func TestGetExpensesItems_UsesExpenseDiffAPI(t *testing.T) {
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			if accountType != "expense" {
				t.Fatalf("expected accountType 'expense', got %q", accountType)
			}
			return []firefly.Account{
				{ID: "e1", Name: "Groceries", CurrencyCode: "USD", Type: "expense"},
				{ID: "e2", Name: "Rent", CurrencyCode: "EUR", Type: "expense"},
			}
		},
		getExpenseDiffFunc: func(accountID string) float64 {
			switch accountID {
			case "e1":
				return 250.75
			case "e2":
				return 1200.00
			default:
				return 0
			}
		},
	}

	items := getExpensesItems(api, false)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	first, ok := items[0].(expenseItem)
	if !ok {
		t.Fatalf("expected item type expenseItem, got %T", items[0])
	}
	if first.Entity.ID != "e1" {
		t.Errorf("expected first account ID 'e1', got %q", first.Entity.ID)
	}
	if first.PrimaryVal != 250.75 {
		t.Errorf("expected first spent 250.75, got %v", first.PrimaryVal)
	}
	if first.Description() != "Spent: 250.75 USD" {
		t.Errorf("unexpected description: %q", first.Description())
	}
	if first.Title() != "Groceries" {
		t.Errorf("expected title 'Groceries', got %q", first.Title())
	}
}

func TestGetExpensesItems_SortedFiltersZeroAndSorts(t *testing.T) {
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{
				{ID: "e1", Name: "Low", CurrencyCode: "USD", Type: "expense"},
				{ID: "e2", Name: "High", CurrencyCode: "USD", Type: "expense"},
				{ID: "e3", Name: "Zero", CurrencyCode: "USD", Type: "expense"},
			}
		},
		getExpenseDiffFunc: func(accountID string) float64 {
			switch accountID {
			case "e1":
				return 100
			case "e2":
				return 5000
			case "e3":
				return 0
			default:
				return 0
			}
		},
	}

	items := getExpensesItems(api, true)

	// Should filter out zero and sort by spent (descending)
	if len(items) != 2 {
		t.Fatalf("expected 2 items (zero filtered), got %d", len(items))
	}

	first := items[0].(expenseItem)
	if first.Entity.Name != "High" {
		t.Errorf("expected first item 'High', got %q", first.Entity.Name)
	}
	if first.PrimaryVal != 5000 {
		t.Errorf("expected first spent 5000, got %v", first.PrimaryVal)
	}

	second := items[1].(expenseItem)
	if second.Entity.Name != "Low" {
		t.Errorf("expected second item 'Low', got %q", second.Entity.Name)
	}
}

// Message handler tests

func TestModelExpenses_RefreshExpenses_Success(t *testing.T) {
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelExpenses(api)

	_, cmd := m.Update(RefreshExpensesMsg{})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	if _, ok := msg.(ExpensesUpdatedMsg); !ok {
		t.Fatalf("expected ExpensesUpdatedMsg, got %T", msg)
	}

	if len(api.updateAccountsCalledWith) != 1 || api.updateAccountsCalledWith[0] != "expense" {
		t.Fatalf("expected UpdateAccounts called with 'expense', got %v", api.updateAccountsCalledWith)
	}
}

func TestModelExpenses_RefreshExpenses_Error(t *testing.T) {
	expectedErr := errors.New("api failure")
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
		updateAccountsFunc: func(accountType string) error {
			return expectedErr
		},
	}
	m := newModelExpenses(api)

	_, cmd := m.Update(RefreshExpensesMsg{})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	notifyMsg, ok := msg.(notify.NotifyMsg)
	if !ok {
		t.Fatalf("expected notify.NotifyMsg, got %T", msg)
	}
	if notifyMsg.Level != notify.Warn {
		t.Fatalf("expected warn level, got %v", notifyMsg.Level)
	}
	if notifyMsg.Message != expectedErr.Error() {
		t.Fatalf("expected message %q, got %q", expectedErr.Error(), notifyMsg.Message)
	}
}

func TestModelExpenses_RefreshExpenseInsights_Success(t *testing.T) {
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelExpenses(api)

	_, cmd := m.Update(RefreshExpenseInsightsMsg{})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	if _, ok := msg.(ExpensesUpdatedMsg); !ok {
		t.Fatalf("expected ExpensesUpdatedMsg, got %T", msg)
	}

	if !api.updateExpenseInsightsCalled {
		t.Fatal("expected UpdateExpenseInsights to be called")
	}
}

func TestModelExpenses_RefreshExpenseInsights_Error(t *testing.T) {
	expectedErr := errors.New("insights failure")
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
		updateExpenseInsightsFunc: func() error {
			return expectedErr
		},
	}
	m := newModelExpenses(api)

	_, cmd := m.Update(RefreshExpenseInsightsMsg{})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	notifyMsg, ok := msg.(notify.NotifyMsg)
	if !ok {
		t.Fatalf("expected notify.NotifyMsg, got %T", msg)
	}
	if notifyMsg.Message != expectedErr.Error() {
		t.Fatalf("expected message %q, got %q", expectedErr.Error(), notifyMsg.Message)
	}
}

func TestModelExpenses_NewExpense_Success(t *testing.T) {
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelExpenses(api)

	_, cmd := m.Update(NewExpenseMsg{Account: "New Expense"})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msgs := collectMsgsFromCmd(cmd)
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}

	if _, ok := msgs[0].(RefreshExpensesMsg); !ok {
		t.Fatalf("expected RefreshExpensesMsg, got %T", msgs[0])
	}

	n, ok := msgs[1].(notify.NotifyMsg)
	if !ok {
		t.Fatalf("expected notify.NotifyMsg, got %T", msgs[1])
	}
	if n.Level != notify.Log {
		t.Fatalf("expected log notify level, got %v", n.Level)
	}
	if n.Message != "Expense account 'New Expense' created" {
		t.Fatalf("unexpected notify message: %q", n.Message)
	}

	if len(api.createExpenseCalledWith) != 1 || api.createExpenseCalledWith[0] != "New Expense" {
		t.Fatalf("expected CreateExpenseAccount called with 'New Expense', got %v", api.createExpenseCalledWith)
	}
}

func TestModelExpenses_NewExpense_Error(t *testing.T) {
	expectedErr := errors.New("create failed")
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
		createExpenseAccountFunc: func(name string) error {
			return expectedErr
		},
	}
	m := newModelExpenses(api)

	_, cmd := m.Update(NewExpenseMsg{Account: "Bad Expense"})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	notifyMsg, ok := msg.(notify.NotifyMsg)
	if !ok {
		t.Fatalf("expected notify.NotifyMsg, got %T", msg)
	}
	if notifyMsg.Level != notify.Warn {
		t.Fatalf("expected warn level, got %v", notifyMsg.Level)
	}
	if notifyMsg.Message != expectedErr.Error() {
		t.Fatalf("expected message %q, got %q", expectedErr.Error(), notifyMsg.Message)
	}
}

func TestModelExpenses_ExpensesUpdated_EmitsDataLoadCompleted(t *testing.T) {
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{
				{ID: "e1", Name: "Groceries", CurrencyCode: "USD", Type: "expense"},
			}
		},
		getExpenseDiffFunc: func(accountID string) float64 {
			return 500
		},
		getTotalExpenseDiffFunc: func() float64 {
			return 1500
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelExpenses(api)

	updated, cmd := m.Update(ExpensesUpdatedMsg{})
	m2 := updated.(modelExpenses)
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msgs := collectMsgsFromMsg(cmd())
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	loader, ok := msgs[0].(DataLoadCompletedMsg)
	if !ok {
		t.Fatalf("expected DataLoadCompletedMsg, got %T", msgs[0])
	}
	if loader.DataType != "expense" {
		t.Fatalf("expected DataType 'expenses', got %q", loader.DataType)
	}

	// Verify list items include total + accounts
	listItems := m2.list.Items()
	if len(listItems) != 2 {
		t.Fatalf("expected 2 list items (total + 1 account), got %d", len(listItems))
	}

	// First item should be total
	totalItem, ok := listItems[0].(expenseItem)
	if !ok {
		t.Fatalf("expected first item to be expenseItem, got %T", listItems[0])
	}
	if totalItem.Entity.Name != "Total" {
		t.Errorf("expected first item name 'Total', got %q", totalItem.Entity.Name)
	}
	if totalItem.PrimaryVal != 1500 {
		t.Errorf("expected total spent 1500, got %v", totalItem.PrimaryVal)
	}
}

func TestModelExpenses_UpdatePositions_SetsListSize(t *testing.T) {
	globalWidth := 100
	globalHeight := 40
	topSize := 5

	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelExpenses(api)

	updated, _ := m.Update(UpdatePositions{
		layout: &LayoutConfig{
			Width:       globalWidth,
			Height:      globalHeight,
			TopSize:     topSize,
			SummarySize: 10,
		},
	})
	m2 := updated.(modelExpenses)

	h, v := m2.styles.Base.GetFrameSize()
	wantW := globalWidth - h
	wantH := globalHeight - v - topSize
	if m2.list.Width() != wantW {
		t.Fatalf("expected width %d, got %d", wantW, m2.list.Width())
	}
	if m2.list.Height() != wantH {
		t.Fatalf("expected height %d, got %d", wantH, m2.list.Height())
	}
}

// Key handling tests

func TestModelExpenses_IgnoresKeysWhenNotFocused(t *testing.T) {
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{{ID: "e1", Name: "Groceries", CurrencyCode: "USD", Type: "expense"}}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelExpenses(api) // focus is false by default

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if cmd != nil {
		t.Fatalf("expected nil cmd when not focused, got %T", cmd)
	}
}

func TestModelExpenses_KeyFilter_EmitsFilterMsgWithSelectedAccount(t *testing.T) {
	acc := firefly.Account{ID: "e1", Name: "Groceries", CurrencyCode: "USD", Type: "expense"}
	m := newFocusedExpensesModelWithAccount(t, acc)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	msgs := collectMsgsFromCmd(cmd)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	filter, ok := msgs[0].(FilterMsg)
	if !ok {
		t.Fatalf("expected FilterMsg, got %T", msgs[0])
	}
	if filter.Account.ID != "e1" {
		t.Fatalf("expected account ID 'e1', got %q", filter.Account.ID)
	}
}

func TestModelExpenses_KeyFilter_IgnoresTotalAccount(t *testing.T) {
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{
				{ID: "e1", Name: "Groceries", CurrencyCode: "USD", Type: "expense"},
			}
		},
		getExpenseDiffFunc:      func(accountID string) float64 { return 500 },
		getTotalExpenseDiffFunc: func() float64 { return 500 },
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelExpenses(api)
	(&m).Focus()

	// Trigger ExpensesUpdatedMsg to add total account
	updated, cmd := m.Update(ExpensesUpdatedMsg{})
	m = updated.(modelExpenses)
	_ = cmd

	// Select first item (which should be total)
	m.list.Select(0)

	// Try to filter - should return nil
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	if cmd != nil {
		t.Fatalf("expected nil cmd when filtering total account, got %T", cmd)
	}
}

func TestModelExpenses_KeyRefresh_EmitsRefreshExpensesMsg(t *testing.T) {
	m := newFocusedExpensesModelWithAccount(t, firefly.Account{
		ID: "e1", Name: "Groceries", CurrencyCode: "USD", Type: "expense",
	})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	if _, ok := msg.(RefreshExpensesMsg); !ok {
		t.Fatalf("expected RefreshExpensesMsg, got %T", msg)
	}
}

func TestModelExpenses_KeySort_TogglesSort(t *testing.T) {
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{
				{ID: "e1", Name: "Low", CurrencyCode: "USD", Type: "expense"},
				{ID: "e2", Name: "High", CurrencyCode: "USD", Type: "expense"},
			}
		},
		getExpenseDiffFunc: func(accountID string) float64 {
			if accountID == "e1" {
				return 100
			}
			return 5000
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelExpenses(api)
	(&m).Focus()

	if m.sorted {
		t.Fatal("expected sorted to be false initially")
	}

	// Press 's' to toggle sort
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m2 := updated.(modelExpenses)

	if !m2.sorted {
		t.Fatal("expected sorted to be true after toggle")
	}

	msg := cmd()
	if _, ok := msg.(ExpensesUpdatedMsg); !ok {
		t.Fatalf("expected ExpensesUpdatedMsg, got %T", msg)
	}

	// Press 's' again to toggle back
	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m3 := updated.(modelExpenses)

	if m3.sorted {
		t.Fatal("expected sorted to be false after second toggle")
	}
}

func TestModelExpenses_KeyResetFilter_EmitsResetFilterMsg(t *testing.T) {
	m := newFocusedExpensesModelWithAccount(t, firefly.Account{
		ID: "e1", Name: "Groceries", CurrencyCode: "USD", Type: "expense",
	})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	filter, ok := msg.(FilterMsg)
	if !ok {
		t.Fatalf("expected FilterMsg, got %T", msg)
	}
	if !filter.Reset {
		t.Fatalf("expected Reset=true, got %+v", filter)
	}
}

func TestModelExpenses_KeyNew_ReturnsPromptMsg(t *testing.T) {
	m := newFocusedExpensesModelWithAccount(t, firefly.Account{
		ID: "e1", Name: "Groceries", CurrencyCode: "USD", Type: "expense",
	})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	p, ok := msg.(prompt.PromptMsg)
	if !ok {
		t.Fatalf("expected prompt.PromptMsg, got %T", msg)
	}
	if p.Prompt != "New Expense(<name>): " {
		t.Fatalf("unexpected prompt: %q", p.Prompt)
	}
	if p.Callback == nil {
		t.Fatal("expected callback")
	}
}

func TestModelExpenses_View_UsesLeftPanelStyle(t *testing.T) {
	m := newFocusedExpensesModelWithAccount(t, firefly.Account{
		ID: "e1", Name: "Groceries", CurrencyCode: "USD", Type: "expense",
	})

	m.styles.LeftPanel = lipgloss.NewStyle().PaddingLeft(2).PaddingRight(3)

	got := m.View()
	want := m.styles.LeftPanel.Render(m.list.View())

	if got != want {
		t.Fatal("unexpected view output")
	}
	if got == m.list.View() {
		t.Fatal("expected left panel styling to change output")
	}
}

// Table-driven view navigation tests

func TestModelExpenses_KeyViewNavigation(t *testing.T) {
	tests := []struct {
		name         string
		key          rune
		expectedView state
		disabled     bool
		expectedMsgs int
	}{
		{"assets", 'a', assetsView, false, 1},
		{"categories", 'c', categoriesView, false, 1},
		{"revenues", 'i', revenuesView, false, 1},
		{"transactions", 't', transactionsView, false, 1},
		{"liabilities", 'o', liabilitiesView, false, 1},
		{"expenses (self)", 'e', expensesView, false, 1},
		{"quit to transactions", 'q', transactionsView, false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newFocusedExpensesModelWithAccount(t, firefly.Account{
				ID:           "e1",
				Name:         "Groceries",
				CurrencyCode: "USD",
				Type:         "expense",
			})

			_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})

			if tt.disabled {
				// Key is disabled, should not emit any command
				if cmd != nil {
					msgs := collectMsgsFromCmd(cmd)
					for _, msg := range msgs {
						if _, ok := msg.(SetFocusedViewMsg); ok {
							t.Fatalf("key %q is disabled but emitted SetFocusedViewMsg", tt.key)
						}
					}
				}
				return
			}

			if cmd == nil {
				t.Fatalf("expected cmd for key %q", tt.key)
			}

			msgs := collectMsgsFromCmd(cmd)
			if len(msgs) != tt.expectedMsgs {
				t.Fatalf("key %q: expected %d messages, got %d (%T)", tt.key, tt.expectedMsgs, len(msgs), msgs)
			}

			focused, ok := msgs[0].(SetFocusedViewMsg)
			if !ok {
				t.Fatalf("key %q: expected SetFocusedViewMsg, got %T", tt.key, msgs[0])
			}
			if focused.state != tt.expectedView {
				t.Fatalf("key %q: expected view %v, got %v", tt.key, tt.expectedView, focused.state)
			}
		})
	}
}

// Prompt callback tests

func TestCmdPromptNewExpense_EmitsPromptMsgWithCallback(t *testing.T) {
	backCmd := Cmd(SetFocusedViewMsg{state: expensesView})
	cmd := CmdPromptNewExpense(backCmd)

	msg := cmd()
	p, ok := msg.(prompt.PromptMsg)
	if !ok {
		t.Fatalf("expected prompt.PromptMsg, got %T", msg)
	}
	if p.Prompt != "New Expense(<name>): " {
		t.Fatalf("unexpected prompt: %q", p.Prompt)
	}
	if p.Callback == nil {
		t.Fatal("expected callback")
	}
}

func TestCmdPromptNewExpense_CallbackValid_EmitsNewExpenseAndBackCmd(t *testing.T) {
	backCmd := Cmd(SetFocusedViewMsg{state: expensesView})
	cmd := CmdPromptNewExpense(backCmd)

	p := cmd().(prompt.PromptMsg)
	msgs := collectMsgsFromCmd(p.Callback("Utilities"))

	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}

	newExpense, ok := msgs[0].(NewExpenseMsg)
	if !ok {
		t.Fatalf("expected NewExpenseMsg, got %T", msgs[0])
	}
	if newExpense.Account != "Utilities" {
		t.Fatalf("expected account 'Utilities', got %q", newExpense.Account)
	}

	focused, ok := msgs[1].(SetFocusedViewMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[1])
	}
	if focused.state != expensesView {
		t.Fatalf("expected expensesView, got %v", focused.state)
	}
}

func TestCmdPromptNewExpense_CallbackNone_EmitsOnlyBackCmd(t *testing.T) {
	backCmd := Cmd(SetFocusedViewMsg{state: expensesView})
	cmd := CmdPromptNewExpense(backCmd)

	p := cmd().(prompt.PromptMsg)
	msgs := collectMsgsFromCmd(p.Callback("None"))

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	focused, ok := msgs[0].(SetFocusedViewMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[0])
	}
	if focused.state != expensesView {
		t.Fatalf("expected expensesView, got %v", focused.state)
	}
}

// Edge case tests

func TestModelExpenses_EmptyAccountList(t *testing.T) {
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelExpenses(api)
	(&m).Focus()

	// Verify no panics when updating with empty list
	_, cmd := m.Update(ExpensesUpdatedMsg{})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	// Verify view rendering with empty list doesn't panic
	view := m.View()
	if view == "" {
		t.Fatal("expected non-empty view")
	}
}

func TestModelExpenses_NilAPI(t *testing.T) {
	// Document that nil API is not handled gracefully
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when constructing with nil API, but no panic occurred")
		}
	}()

	_ = newModelExpenses(nil)
}

func TestModelExpenses_SpentBoundaryValues(t *testing.T) {
	tests := []struct {
		name            string
		spent           float64
		expectedDisplay string
	}{
		{"zero spent", 0.0, "Spent: 0.00 USD"},
		{"negative spent", -150.50, "Spent: -150.50 USD"},
		{"very large spent", 999999999.99, "Spent: 999999999.99 USD"},
		{"small positive", 0.01, "Spent: 0.01 USD"},
		{"small negative", -0.01, "Spent: -0.01 USD"},
		{"many decimals", 123.456789, "Spent: 123.46 USD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &mockExpenseAPI{
				accountsByTypeFunc: func(accountType string) []firefly.Account {
					return []firefly.Account{
						{ID: "e1", Name: "Test Expense", CurrencyCode: "USD", Type: "expense"},
					}
				},
				getExpenseDiffFunc: func(accountID string) float64 {
					return tt.spent
				},
			}

			items := getExpensesItems(api, false)
			if len(items) != 1 {
				t.Fatalf("expected 1 item, got %d", len(items))
			}

			item, ok := items[0].(expenseItem)
			if !ok {
				t.Fatalf("expected expenseItem, got %T", items[0])
			}

			if item.PrimaryVal != tt.spent {
				t.Errorf("expected spent %v, got %v", tt.spent, item.PrimaryVal)
			}

			if item.Description() != tt.expectedDisplay {
				t.Errorf("expected description %q, got %q", tt.expectedDisplay, item.Description())
			}
		})
	}
}

func TestModelExpenses_AccountNameEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		accountName string
	}{
		{"empty name", ""},
		{"very long name", "This is an extremely long expense account name that might cause display issues"},
		{"special characters", "Groceries™ with 💰 émojis & spëcial çhars"},
		{"whitespace only", "   "},
		{"newlines", "Expense\nWith\nNewlines"},
		{"tabs", "Expense\tWith\tTabs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &mockExpenseAPI{
				accountsByTypeFunc: func(accountType string) []firefly.Account {
					return []firefly.Account{
						{ID: "e1", Name: tt.accountName, CurrencyCode: "USD", Type: "expense"},
					}
				},
				getExpenseDiffFunc: func(accountID string) float64 {
					return 1000.0
				},
			}

			items := getExpensesItems(api, false)
			if len(items) != 1 {
				t.Fatalf("expected 1 item, got %d", len(items))
			}

			item, ok := items[0].(expenseItem)
			if !ok {
				t.Fatalf("expected expenseItem, got %T", items[0])
			}

			if item.Entity.Name != tt.accountName {
				t.Errorf("expected name %q, got %q", tt.accountName, item.Entity.Name)
			}

			title := item.Title()
			if title != tt.accountName {
				t.Errorf("expected title %q, got %q", tt.accountName, title)
			}
		})
	}
}

func TestModelExpenses_MissingCurrencyCode(t *testing.T) {
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{
				{ID: "e1", Name: "No Currency", CurrencyCode: "", Type: "expense"},
			}
		},
		getExpenseDiffFunc: func(accountID string) float64 {
			return 500.0
		},
	}

	items := getExpensesItems(api, false)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item, ok := items[0].(expenseItem)
	if !ok {
		t.Fatalf("expected expenseItem, got %T", items[0])
	}

	desc := item.Description()
	if desc != "Spent: 500.00 " {
		t.Errorf("expected description %q, got %q", "Spent: 500.00 ", desc)
	}
}

func TestModelExpenses_FocusToggle(t *testing.T) {
	api := &mockExpenseAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{
				{ID: "e1", Name: "Test", CurrencyCode: "USD", Type: "expense"},
			}
		},
		getExpenseDiffFunc: func(accountID string) float64 { return 0 },
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}

	m := newModelExpenses(api)

	if m.focus {
		t.Fatal("expected focus to be false initially")
	}

	(&m).Focus()
	if !m.focus {
		t.Fatal("expected focus to be true after Focus()")
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if cmd == nil {
		t.Fatal("expected cmd when focused")
	}

	(&m).Blur()
	if m.focus {
		t.Fatal("expected focus to be false after Blur()")
	}

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if cmd != nil {
		t.Fatal("expected nil cmd when blurred")
	}
}

func TestModelExpenses_UpdatePositions_WithVariousDimensions(t *testing.T) {
	tests := []struct {
		name              string
		globalWidth       int
		globalHeight      int
		topSize           int
		expectedMinWidth  int
		expectedMinHeight int
	}{
		{"normal dimensions", 100, 40, 5, 1, 1},
		{"minimum dimensions", 20, 20, 5, 1, 1},
		{"very large dimensions", 1000, 500, 5, 1, 1},
		{"zero top", 100, 40, 0, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &mockExpenseAPI{
				accountsByTypeFunc: func(accountType string) []firefly.Account {
					return []firefly.Account{}
				},
				primaryCurrencyFunc: func() firefly.Currency {
					return firefly.Currency{Code: "USD", Symbol: "$"}
				},
			}
			m := newModelExpenses(api)

			updated, _ := m.Update(UpdatePositions{
				layout: &LayoutConfig{
					Width:   tt.globalWidth,
					Height:  tt.globalHeight,
					TopSize: tt.topSize,
				},
			})
			m2 := updated.(modelExpenses)

			if m2.list.Width() < tt.expectedMinWidth {
				t.Errorf("expected width >= %d, got %d", tt.expectedMinWidth, m2.list.Width())
			}
			if m2.list.Height() < tt.expectedMinHeight {
				t.Errorf("expected height >= %d, got %d", tt.expectedMinHeight, m2.list.Height())
			}
		})
	}
}
