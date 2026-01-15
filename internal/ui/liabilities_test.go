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

type mockLiabilityAPI struct {
	updateAccountsFunc         func(accountType string) error
	accountsByTypeFunc         func(accountType string) []firefly.Account
	accountBalanceFunc         func(accountID string) float64
	createLiabilityAccountFunc func(nl firefly.NewLiability) error
	updateAccountsCalledWith   []string
	createLiabilityCalledWith  []firefly.NewLiability
}

func (m *mockLiabilityAPI) UpdateAccounts(accountType string) error {
	m.updateAccountsCalledWith = append(m.updateAccountsCalledWith, accountType)
	if m.updateAccountsFunc != nil {
		return m.updateAccountsFunc(accountType)
	}
	return nil
}

func (m *mockLiabilityAPI) AccountsByType(accountType string) []firefly.Account {
	if m.accountsByTypeFunc != nil {
		return m.accountsByTypeFunc(accountType)
	}
	return nil
}

func (m *mockLiabilityAPI) AccountBalance(accountID string) float64 {
	if m.accountBalanceFunc != nil {
		return m.accountBalanceFunc(accountID)
	}
	return 0
}

func (m *mockLiabilityAPI) CreateLiabilityAccount(nl firefly.NewLiability) error {
	m.createLiabilityCalledWith = append(m.createLiabilityCalledWith, nl)
	if m.createLiabilityAccountFunc != nil {
		return m.createLiabilityAccountFunc(nl)
	}
	return nil
}

func newFocusedLiabilitiesModelWithAccount(t *testing.T, acc firefly.Account) modelLiabilities {
	t.Helper()

	api := &mockLiabilityAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			if accountType != "liabilities" {
				t.Fatalf("expected accountType 'liabilities', got %q", accountType)
			}
			return []firefly.Account{acc}
		},
		accountBalanceFunc: func(accountID string) float64 { return 0 },
	}

	m := newModelLiabilities(api)
	(&m).Focus()
	return m
}

// Basic functionality tests

func TestGetLiabilitiesItems_UsesAccountBalanceAPI(t *testing.T) {
	api := &mockLiabilityAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			if accountType != "liabilities" {
				t.Fatalf("expected accountType 'liabilities', got %q", accountType)
			}
			return []firefly.Account{
				{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"},
				{ID: "l2", Name: "Credit Card", CurrencyCode: "EUR", Type: "liabilities"},
			}
		},
		accountBalanceFunc: func(accountID string) float64 {
			switch accountID {
			case "l1":
				return -250000.00
			case "l2":
				return -5000.50
			default:
				return 0
			}
		},
	}

	items := getLiabilitiesItems(api)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	first, ok := items[0].(liabilityItem)
	if !ok {
		t.Fatalf("expected item type liabilityItem, got %T", items[0])
	}
	if first.account.ID != "l1" {
		t.Errorf("expected first account ID 'l1', got %q", first.account.ID)
	}
	if first.balance != -250000.00 {
		t.Errorf("expected first balance -250000.00, got %.2f", first.balance)
	}

	second, ok := items[1].(liabilityItem)
	if !ok {
		t.Fatalf("expected item type liabilityItem, got %T", items[1])
	}
	if second.account.ID != "l2" {
		t.Errorf("expected second account ID 'l2', got %q", second.account.ID)
	}
	if second.balance != -5000.50 {
		t.Errorf("expected second balance -5000.50, got %.2f", second.balance)
	}
}

func TestLiabilityItem_Methods(t *testing.T) {
	acc := firefly.Account{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}
	item := liabilityItem{
		account: acc,
		balance: -250000.00,
	}

	if item.Title() != acc.Name {
		t.Errorf("expected title %q, got %q", acc.Name, item.Title())
	}
	if item.FilterValue() != acc.Name {
		t.Errorf("expected filter value %q, got %q", acc.Name, item.FilterValue())
	}

	expectedDesc := "Balance: -250000.00 USD"
	if item.Description() != expectedDesc {
		t.Errorf("expected description %q, got %q", expectedDesc, item.Description())
	}
}

func TestNewModelLiabilities_InitializesCorrectly(t *testing.T) {
	api := &mockLiabilityAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}}
		},
		accountBalanceFunc: func(accountID string) float64 { return 0 },
	}

	m := newModelLiabilities(api)

	if m.api != api {
		t.Error("expected API to be set")
	}
	if m.list.Title != "Liabilities" {
		t.Errorf("expected list title 'Liabilities', got %q", m.list.Title)
	}
	if m.focus {
		t.Error("expected model to be unfocused initially")
	}
}

// Message handler tests

func TestRefreshLiabilitiesMsg_Success(t *testing.T) {
	api := &mockLiabilityAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}}
		},
		accountBalanceFunc: func(accountID string) float64 { return 0 },
	}

	m := newModelLiabilities(api)
	_, cmd := m.Update(RefreshLiabilitiesMsg{})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msg := cmd()
	if _, ok := msg.(LiabilitiesUpdateMsg); !ok {
		t.Errorf("expected LiabilitiesUpdateMsg, got %T", msg)
	}

	if len(api.updateAccountsCalledWith) != 1 {
		t.Fatalf("expected UpdateAccounts to be called once, got %d", len(api.updateAccountsCalledWith))
	}
	if api.updateAccountsCalledWith[0] != "liabilities" {
		t.Errorf("expected UpdateAccounts called with 'liabilities', got %q", api.updateAccountsCalledWith[0])
	}
}

func TestRefreshLiabilitiesMsg_Error(t *testing.T) {
	expectedErr := errors.New("liabilities API error")
	api := &mockLiabilityAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}}
		},
		accountBalanceFunc: func(accountID string) float64 { return 0 },
		updateAccountsFunc: func(accountType string) error {
			return expectedErr
		},
	}

	m := newModelLiabilities(api)
	_, cmd := m.Update(RefreshLiabilitiesMsg{})

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

func TestLiabilitiesUpdateMsg_SetsItems(t *testing.T) {
	api := &mockLiabilityAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}}
		},
		accountBalanceFunc: func(accountID string) float64 { return -250000.0 },
	}

	m := newModelLiabilities(api)
	updated, cmd := m.Update(LiabilitiesUpdateMsg{})
	m2 := updated.(modelLiabilities)

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msgs := collectMsgsFromCmd(cmd)
	if len(msgs) < 1 {
		t.Fatal("expected at least one message from batch command")
	}

	foundDataLoadMsg := false
	for _, msg := range msgs {
		if dlMsg, ok := msg.(DataLoadCompletedMsg); ok {
			if dlMsg.DataType != "liabilities" {
				t.Errorf("expected DataType 'liabilities', got %q", dlMsg.DataType)
			}
			foundDataLoadMsg = true
		}
	}
	if !foundDataLoadMsg {
		t.Error("expected DataLoadCompletedMsg in batch")
	}

	listItems := m2.list.Items()
	if len(listItems) != 1 {
		t.Fatalf("expected 1 list item, got %d", len(listItems))
	}

	item, ok := listItems[0].(liabilityItem)
	if !ok {
		t.Fatalf("expected first item to be liabilityItem, got %T", listItems[0])
	}
	if item.account.Name != "Mortgage" {
		t.Errorf("expected item name 'Mortgage', got %q", item.account.Name)
	}
	if item.balance != -250000.0 {
		t.Errorf("expected balance -250000.0, got %.2f", item.balance)
	}
}

func TestNewLiabilityMsg_Success(t *testing.T) {
	api := &mockLiabilityAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}}
		},
		accountBalanceFunc: func(accountID string) float64 { return 0 },
	}

	m := newModelLiabilities(api)
	_, cmd := m.Update(NewLiabilityMsg{
		Account:   "Car Loan",
		Currency:  "USD",
		Type:      "loan",
		Direction: "debit",
	})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	if len(api.createLiabilityCalledWith) != 1 {
		t.Fatalf("expected CreateLiabilityAccount to be called once, got %d", len(api.createLiabilityCalledWith))
	}

	nl := api.createLiabilityCalledWith[0]
	if nl.Name != "Car Loan" {
		t.Errorf("expected name 'Car Loan', got %q", nl.Name)
	}
	if nl.CurrencyCode != "USD" {
		t.Errorf("expected currency 'USD', got %q", nl.CurrencyCode)
	}
	if nl.Type != "loan" {
		t.Errorf("expected type 'loan', got %q", nl.Type)
	}
	if nl.Direction != "debit" {
		t.Errorf("expected direction 'debit', got %q", nl.Direction)
	}

	msgs := collectMsgsFromCmd(cmd)
	foundRefresh := false
	foundNotify := false
	for _, msg := range msgs {
		if _, ok := msg.(RefreshLiabilitiesMsg); ok {
			foundRefresh = true
		}
		if notifyMsg, ok := msg.(notify.NotifyMsg); ok {
			if notifyMsg.Level == notify.Log {
				foundNotify = true
			}
		}
	}
	if !foundRefresh {
		t.Error("expected RefreshLiabilitiesMsg in batch")
	}
	if !foundNotify {
		t.Error("expected notify.NotifyMsg in batch")
	}
}

func TestNewLiabilityMsg_Error(t *testing.T) {
	expectedErr := errors.New("create liability error")
	api := &mockLiabilityAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}}
		},
		accountBalanceFunc: func(accountID string) float64 { return 0 },
		createLiabilityAccountFunc: func(nl firefly.NewLiability) error {
			return expectedErr
		},
	}

	m := newModelLiabilities(api)
	_, cmd := m.Update(NewLiabilityMsg{
		Account:   "Bad Loan",
		Currency:  "USD",
		Type:      "loan",
		Direction: "debit",
	})

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

func TestUpdatePositions_SetsListSize(t *testing.T) {
	globalWidth = 100
	globalHeight = 40
	topSize = 5

	api := &mockLiabilityAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
	}
	m := newModelLiabilities(api)

	updated, _ := m.Update(UpdatePositions{})
	m2 := updated.(modelLiabilities)

	h, v := m2.styles.Base.GetFrameSize()
	wantW := globalWidth - h
	wantH := globalHeight - v - topSize
	if m2.list.Width() != wantW {
		t.Errorf("expected width %d, got %d", wantW, m2.list.Width())
	}
	if m2.list.Height() != wantH {
		t.Errorf("expected height %d, got %d", wantH, m2.list.Height())
	}
}

// Key handling tests

func TestLiabilities_FocusAndBlur(t *testing.T) {
	acc := firefly.Account{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}
	m := newFocusedLiabilitiesModelWithAccount(t, acc)

	if !m.focus {
		t.Error("expected model to be focused after Focus()")
	}

	(&m).Blur()
	if m.focus {
		t.Error("expected model to be blurred after Blur()")
	}
}

func TestLiabilities_KeyFilter_SendsFilterMsg(t *testing.T) {
	acc := firefly.Account{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}
	m := newFocusedLiabilitiesModelWithAccount(t, acc)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msg := cmd()
	filterMsg, ok := msg.(FilterMsg)
	if !ok {
		t.Fatalf("expected FilterMsg, got %T", msg)
	}
	if filterMsg.Account.Name != acc.Name {
		t.Errorf("expected account name %q, got %q", acc.Name, filterMsg.Account.Name)
	}
	if filterMsg.Reset {
		t.Error("expected Reset to be false")
	}
}

func TestLiabilities_KeyRefresh_SendsRefreshMsg(t *testing.T) {
	acc := firefly.Account{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}
	m := newFocusedLiabilitiesModelWithAccount(t, acc)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msg := cmd()
	if _, ok := msg.(RefreshLiabilitiesMsg); !ok {
		t.Errorf("expected RefreshLiabilitiesMsg, got %T", msg)
	}
}

func TestLiabilities_KeyResetFilter_SendsResetFilterMsg(t *testing.T) {
	acc := firefly.Account{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}
	m := newFocusedLiabilitiesModelWithAccount(t, acc)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlA})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msg := cmd()
	filterMsg, ok := msg.(FilterMsg)
	if !ok {
		t.Fatalf("expected FilterMsg, got %T", msg)
	}
	if !filterMsg.Reset {
		t.Error("expected Reset to be true")
	}
}

func TestLiabilities_KeyNew_EmitsPrompt(t *testing.T) {
	acc := firefly.Account{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}
	m := newFocusedLiabilitiesModelWithAccount(t, acc)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msg := cmd()
	if _, ok := msg.(prompt.PromptMsg); !ok {
		t.Errorf("expected prompt.PromptMsg, got %T", msg)
	}
}

func TestLiabilities_View_RendersListView(t *testing.T) {
	lipgloss.SetColorProfile(0)
	acc := firefly.Account{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}
	m := newFocusedLiabilitiesModelWithAccount(t, acc)

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

func TestModelLiabilities_UnfocusedIgnoresKeys(t *testing.T) {
	acc := firefly.Account{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}
	m := newFocusedLiabilitiesModelWithAccount(t, acc)
	(&m).Blur()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if cmd != nil {
		t.Error("expected no command when unfocused")
	}
}

// View navigation tests

func TestLiabilities_KeyPresses_NavigateToCorrectViews(t *testing.T) {
	acc := firefly.Account{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}

	tests := []struct {
		name         string
		key          rune
		expectedView state
		disabled     bool
		expectedMsgs int
	}{
		{"assets", 'a', assetsView, false, 2},
		{"categories", 'c', categoriesView, false, 2},
		{"expenses", 'e', expensesView, false, 2},
		{"transactions", 't', transactionsView, false, 2},
		{"liabilities (self)", 'o', liabilitiesView, true, 0},
		{"revenues", 'i', revenuesView, false, 2},
		{"quit to transactions", 'q', transactionsView, false, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newFocusedLiabilitiesModelWithAccount(t, acc)
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

			if _, ok := msgs[1].(UpdatePositions); !ok {
				t.Fatalf("key %q: expected UpdatePositions as second message, got %T", tt.key, msgs[1])
			}
		})
	}
}

// Prompt callback tests

func TestCmdPromptNewLiability_EmitsPrompt(t *testing.T) {
	backCmd := Cmd(SetFocusedViewMsg{state: liabilitiesView})
	cmd := CmdPromptNewLiability(backCmd)

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msg := cmd()
	askMsg, ok := msg.(prompt.PromptMsg)
	if !ok {
		t.Fatalf("expected prompt.PromptMsg, got %T", msg)
	}
	expectedPrompt := "New Liabity(<name>,<currency>,<type:loan|debt|mortage>,<direction:credit|debit>): "
	if askMsg.Prompt != expectedPrompt {
		t.Errorf("expected prompt %q, got %q", expectedPrompt, askMsg.Prompt)
	}
}

func TestCmdPromptNewLiability_ValidInput(t *testing.T) {
	backCmdCalled := false
	backCmd := func() tea.Msg {
		backCmdCalled = true
		return nil
	}

	cmd := CmdPromptNewLiability(backCmd)
	askMsg := cmd().(prompt.PromptMsg)

	resultCmd := askMsg.Callback("Car Loan, USD, loan, debit")
	if resultCmd == nil {
		t.Fatal("expected a command from callback, got nil")
	}

	msgs := collectMsgsFromCmd(resultCmd)
	if len(msgs) < 1 {
		t.Fatal("expected at least one message from callback")
	}

	newLiabilityMsg, ok := msgs[0].(NewLiabilityMsg)
	if !ok {
		t.Fatalf("expected NewLiabilityMsg, got %T", msgs[0])
	}
	if newLiabilityMsg.Account != "Car Loan" {
		t.Errorf("expected account 'Car Loan', got %q", newLiabilityMsg.Account)
	}
	if newLiabilityMsg.Currency != "USD" {
		t.Errorf("expected currency 'USD', got %q", newLiabilityMsg.Currency)
	}
	if newLiabilityMsg.Type != "loan" {
		t.Errorf("expected type 'loan', got %q", newLiabilityMsg.Type)
	}
	if newLiabilityMsg.Direction != "debit" {
		t.Errorf("expected direction 'debit', got %q", newLiabilityMsg.Direction)
	}

	if !backCmdCalled {
		t.Error("expected back command to be called")
	}
}

func TestCmdPromptNewLiability_InvalidFormat(t *testing.T) {
	backCmdCalled := false
	backCmd := func() tea.Msg {
		backCmdCalled = true
		return nil
	}

	cmd := CmdPromptNewLiability(backCmd)
	askMsg := cmd().(prompt.PromptMsg)

	resultCmd := askMsg.Callback("InvalidInput")
	if resultCmd == nil {
		t.Fatal("expected a command from callback, got nil")
	}

	msgs := collectMsgsFromCmd(resultCmd)

	// Should have warning message, no NewLiabilityMsg
	foundWarning := false
	for _, msg := range msgs {
		if _, ok := msg.(NewLiabilityMsg); ok {
			t.Error("expected no NewLiabilityMsg for invalid input")
		}
		if notifyMsg, ok := msg.(notify.NotifyMsg); ok {
			if notifyMsg.Level == notify.Warn {
				foundWarning = true
			}
		}
	}
	if !foundWarning {
		t.Error("expected warning message for invalid input")
	}

	if !backCmdCalled {
		t.Error("expected back command to be called even with invalid input")
	}
}

func TestCmdPromptNewLiability_EmptyNameOrCurrency(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty name", ", USD, loan, debit"},
		{"empty currency", "Car Loan, , loan, debit"},
		{"both empty", ", , loan, debit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backCmd := func() tea.Msg { return nil }
			cmd := CmdPromptNewLiability(backCmd)
			askMsg := cmd().(prompt.PromptMsg)

			resultCmd := askMsg.Callback(tt.input)
			msgs := collectMsgsFromCmd(resultCmd)

			// Should have warning message, no NewLiabilityMsg
			foundWarning := false
			for _, msg := range msgs {
				if _, ok := msg.(NewLiabilityMsg); ok {
					t.Error("expected no NewLiabilityMsg for empty name/currency")
				}
				if notifyMsg, ok := msg.(notify.NotifyMsg); ok {
					if notifyMsg.Level == notify.Warn {
						foundWarning = true
					}
				}
			}
			if !foundWarning {
				t.Error("expected warning message for empty name/currency")
			}
		})
	}
}

func TestCmdPromptNewLiability_CancelWithNone(t *testing.T) {
	backCmdCalled := false
	backCmd := func() tea.Msg {
		backCmdCalled = true
		return nil
	}

	cmd := CmdPromptNewLiability(backCmd)
	askMsg := cmd().(prompt.PromptMsg)

	resultCmd := askMsg.Callback("None")
	if resultCmd == nil {
		t.Fatal("expected a command from callback, got nil")
	}

	msgs := collectMsgsFromCmd(resultCmd)

	for _, msg := range msgs {
		if _, ok := msg.(NewLiabilityMsg); ok {
			t.Error("expected no NewLiabilityMsg when input is 'None'")
		}
	}

	if !backCmdCalled {
		t.Error("expected back command to be called even when canceled")
	}
}

// Edge case tests

func TestGetLiabilitiesItems_EmptyList(t *testing.T) {
	api := &mockLiabilityAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
	}

	items := getLiabilitiesItems(api)
	if len(items) != 0 {
		t.Errorf("expected 0 items for empty list, got %d", len(items))
	}
}

func TestGetLiabilitiesItems_NilAPI(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when calling getLiabilitiesItems with nil API")
		}
	}()

	getLiabilitiesItems(nil)
}

func TestModelLiabilities_LargeBalance(t *testing.T) {
	api := &mockLiabilityAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{{ID: "l1", Name: "BigLoan", CurrencyCode: "USD", Type: "liabilities"}}
		},
		accountBalanceFunc: func(accountID string) float64 { return -999999999.99 },
	}

	m := newModelLiabilities(api)
	items := m.list.Items()

	if len(items) != 1 {
		t.Fatal("expected 1 item")
	}

	item := items[0].(liabilityItem)
	if item.balance != -999999999.99 {
		t.Errorf("expected large balance -999999999.99, got %.2f", item.balance)
	}
}

func TestModelLiabilities_PositiveBalance(t *testing.T) {
	api := &mockLiabilityAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{{ID: "l1", Name: "Loan", CurrencyCode: "USD", Type: "liabilities"}}
		},
		accountBalanceFunc: func(accountID string) float64 { return 100.0 },
	}

	m := newModelLiabilities(api)
	items := m.list.Items()

	if len(items) != 1 {
		t.Fatal("expected 1 item")
	}

	item := items[0].(liabilityItem)
	if item.balance != 100.0 {
		t.Errorf("expected positive balance 100.0, got %.2f", item.balance)
	}
}

func TestModelLiabilities_SpecialCharactersInName(t *testing.T) {
	specialNames := []string{
		"Credit Card Café",
		"日本語",
		"Mortgage 🏠",
		"<script>alert('xss')</script>",
		"Loan\nWith\nNewlines",
		"",
	}

	for _, name := range specialNames {
		t.Run(name, func(t *testing.T) {
			api := &mockLiabilityAPI{
				accountsByTypeFunc: func(accountType string) []firefly.Account {
					return []firefly.Account{{ID: "l1", Name: name, CurrencyCode: "USD", Type: "liabilities"}}
				},
				accountBalanceFunc: func(accountID string) float64 { return 0 },
			}

			m := newModelLiabilities(api)
			items := m.list.Items()

			if len(items) != 1 {
				t.Errorf("expected 1 item, got %d", len(items))
			}

			item := items[0].(liabilityItem)
			if item.Title() != name {
				t.Errorf("expected title %q, got %q", name, item.Title())
			}
		})
	}
}

func TestModelLiabilities_SmallDimensions(t *testing.T) {
	acc := firefly.Account{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}
	m := newFocusedLiabilitiesModelWithAccount(t, acc)

	globalWidth = 10
	globalHeight = 5
	topSize = 2

	updated, _ := m.Update(UpdatePositions{})
	m2 := updated.(modelLiabilities)

	w, h := m2.list.Width(), m2.list.Height()
	if w < 0 || h < 0 {
		t.Error("expected non-negative list dimensions even with small screen")
	}
}

func TestModelLiabilities_LargeDimensions(t *testing.T) {
	acc := firefly.Account{ID: "l1", Name: "Mortgage", CurrencyCode: "USD", Type: "liabilities"}
	m := newFocusedLiabilitiesModelWithAccount(t, acc)

	globalWidth = 1000
	globalHeight = 1000
	topSize = 10

	updated, _ := m.Update(UpdatePositions{})
	m2 := updated.(modelLiabilities)

	w, h := m2.list.Width(), m2.list.Height()
	if w <= 0 || h <= 0 {
		t.Error("expected positive list dimensions with large screen")
	}
}

func TestCmdPromptNewLiability_WithSpaces(t *testing.T) {
	backCmd := func() tea.Msg { return nil }
	cmd := CmdPromptNewLiability(backCmd)
	askMsg := cmd().(prompt.PromptMsg)

	// Test with extra spaces
	resultCmd := askMsg.Callback("  Car Loan  ,  USD  ,  loan  ,  debit  ")
	msgs := collectMsgsFromCmd(resultCmd)

	var newLiabilityMsg *NewLiabilityMsg
	for _, msg := range msgs {
		if nlMsg, ok := msg.(NewLiabilityMsg); ok {
			newLiabilityMsg = &nlMsg
			break
		}
	}

	if newLiabilityMsg == nil {
		t.Fatal("expected NewLiabilityMsg")
	}

	if newLiabilityMsg.Account != "Car Loan" {
		t.Errorf("expected trimmed account 'Car Loan', got %q", newLiabilityMsg.Account)
	}
	if newLiabilityMsg.Currency != "USD" {
		t.Errorf("expected trimmed currency 'USD', got %q", newLiabilityMsg.Currency)
	}
	if newLiabilityMsg.Type != "loan" {
		t.Errorf("expected trimmed type 'loan', got %q", newLiabilityMsg.Type)
	}
	if newLiabilityMsg.Direction != "debit" {
		t.Errorf("expected trimmed direction 'debit', got %q", newLiabilityMsg.Direction)
	}
}
