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

type mockRevenueAPI struct {
	updateAccountsFunc          func(accountType string) error
	accountsByTypeFunc          func(accountType string) []firefly.Account
	accountBalanceFunc          func(accountID string) float64
	createRevenueAccountFunc    func(name string) error
	updateRevenueInsightsFunc   func() error
	getRevenueDiffFunc          func(accountID string) float64
	getTotalRevenueDiffFunc     func() float64
	primaryCurrencyFunc         func() firefly.Currency
	updateAccountsCalledWith    []string
	createRevenueCalledWith     []string
	updateRevenueInsightsCalled bool
}

func (m *mockRevenueAPI) UpdateAccounts(accountType string) error {
	m.updateAccountsCalledWith = append(m.updateAccountsCalledWith, accountType)
	if m.updateAccountsFunc != nil {
		return m.updateAccountsFunc(accountType)
	}
	return nil
}

func (m *mockRevenueAPI) AccountsByType(accountType string) []firefly.Account {
	if m.accountsByTypeFunc != nil {
		return m.accountsByTypeFunc(accountType)
	}
	return nil
}

func (m *mockRevenueAPI) AccountBalance(accountID string) float64 {
	if m.accountBalanceFunc != nil {
		return m.accountBalanceFunc(accountID)
	}
	return 0
}

func (m *mockRevenueAPI) CreateRevenueAccount(name string) error {
	m.createRevenueCalledWith = append(m.createRevenueCalledWith, name)
	if m.createRevenueAccountFunc != nil {
		return m.createRevenueAccountFunc(name)
	}
	return nil
}

func (m *mockRevenueAPI) UpdateRevenueInsights() error {
	m.updateRevenueInsightsCalled = true
	if m.updateRevenueInsightsFunc != nil {
		return m.updateRevenueInsightsFunc()
	}
	return nil
}

func (m *mockRevenueAPI) GetRevenueDiff(accountID string) float64 {
	if m.getRevenueDiffFunc != nil {
		return m.getRevenueDiffFunc(accountID)
	}
	return 0
}

func (m *mockRevenueAPI) GetTotalRevenueDiff() float64 {
	if m.getTotalRevenueDiffFunc != nil {
		return m.getTotalRevenueDiffFunc()
	}
	return 0
}

func (m *mockRevenueAPI) PrimaryCurrency() firefly.Currency {
	if m.primaryCurrencyFunc != nil {
		return m.primaryCurrencyFunc()
	}
	return firefly.Currency{Code: "USD", Symbol: "$"}
}

func newFocusedRevenuesModelWithAccount(t *testing.T, acc firefly.Account) modelRevenues {
	t.Helper()

	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			if accountType != "revenue" {
				t.Fatalf("expected accountType 'revenue', got %q", accountType)
			}
			return []firefly.Account{acc}
		},
		getRevenueDiffFunc: func(accountID string) float64 { return 0 },
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}

	m := newModelRevenues(api)
	(&m).Focus()
	return m
}

// Basic functionality tests

func TestGetRevenuesItems_UsesRevenueDiffAPI(t *testing.T) {
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			if accountType != "revenue" {
				t.Fatalf("expected accountType 'revenue', got %q", accountType)
			}
			return []firefly.Account{
				{ID: "r1", Name: "Salary", CurrencyCode: "USD", Type: "revenue"},
				{ID: "r2", Name: "Freelance", CurrencyCode: "EUR", Type: "revenue"},
			}
		},
		getRevenueDiffFunc: func(accountID string) float64 {
			switch accountID {
			case "r1":
				return 5000.50
			case "r2":
				return 1200.75
			default:
				return 0
			}
		},
	}

	items := getRevenuesItems(api, false)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	first, ok := items[0].(revenueItem)
	if !ok {
		t.Fatalf("expected item type revenueItem, got %T", items[0])
	}
	if first.account.ID != "r1" {
		t.Errorf("expected first account ID 'r1', got %q", first.account.ID)
	}
	if first.earned != 5000.50 {
		t.Errorf("expected first earned 5000.50, got %v", first.earned)
	}
	if first.Description() != "Earned: 5000.50 USD" {
		t.Errorf("unexpected description: %q", first.Description())
	}
	if first.Title() != "Salary" {
		t.Errorf("expected title 'Salary', got %q", first.Title())
	}
}

func TestGetRevenuesItems_SortedFiltersZeroAndSorts(t *testing.T) {
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{
				{ID: "r1", Name: "Low", CurrencyCode: "USD", Type: "revenue"},
				{ID: "r2", Name: "High", CurrencyCode: "USD", Type: "revenue"},
				{ID: "r3", Name: "Zero", CurrencyCode: "USD", Type: "revenue"},
			}
		},
		getRevenueDiffFunc: func(accountID string) float64 {
			switch accountID {
			case "r1":
				return 100
			case "r2":
				return 5000
			case "r3":
				return 0
			default:
				return 0
			}
		},
	}

	items := getRevenuesItems(api, true)

	// Should filter out zero and sort by earned (descending)
	if len(items) != 2 {
		t.Fatalf("expected 2 items (zero filtered), got %d", len(items))
	}

	first := items[0].(revenueItem)
	if first.account.Name != "High" {
		t.Errorf("expected first item 'High', got %q", first.account.Name)
	}
	if first.earned != 5000 {
		t.Errorf("expected first earned 5000, got %v", first.earned)
	}

	second := items[1].(revenueItem)
	if second.account.Name != "Low" {
		t.Errorf("expected second item 'Low', got %q", second.account.Name)
	}
}

func TestNewModelRevenues_SetsPrimaryCurrency(t *testing.T) {
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "EUR", Symbol: "€"}
		},
	}

	_ = newModelRevenues(api)

	// Verify totalRevenueAccount got the primary currency
	if totalRevenueAccount.CurrencyCode != "EUR" {
		t.Errorf("expected totalRevenueAccount currency 'EUR', got %q", totalRevenueAccount.CurrencyCode)
	}
}

// Message handler tests

func TestModelRevenues_RefreshRevenues_Success(t *testing.T) {
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelRevenues(api)

	_, cmd := m.Update(RefreshRevenuesMsg{})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	if _, ok := msg.(RevenuesUpdateMsg); !ok {
		t.Fatalf("expected RevenuesUpdateMsg, got %T", msg)
	}

	if len(api.updateAccountsCalledWith) != 1 || api.updateAccountsCalledWith[0] != "revenue" {
		t.Fatalf("expected UpdateAccounts called with 'revenue', got %v", api.updateAccountsCalledWith)
	}
}

func TestModelRevenues_RefreshRevenues_Error(t *testing.T) {
	expectedErr := errors.New("api failure")
	api := &mockRevenueAPI{
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
	m := newModelRevenues(api)

	_, cmd := m.Update(RefreshRevenuesMsg{})
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

func TestModelRevenues_RefreshRevenueInsights_Success(t *testing.T) {
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelRevenues(api)

	_, cmd := m.Update(RefreshRevenueInsightsMsg{})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	if _, ok := msg.(RevenuesUpdateMsg); !ok {
		t.Fatalf("expected RevenuesUpdateMsg, got %T", msg)
	}

	if !api.updateRevenueInsightsCalled {
		t.Fatal("expected UpdateRevenueInsights to be called")
	}
}

func TestModelRevenues_RefreshRevenueInsights_Error(t *testing.T) {
	expectedErr := errors.New("insights failure")
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
		updateRevenueInsightsFunc: func() error {
			return expectedErr
		},
	}
	m := newModelRevenues(api)

	_, cmd := m.Update(RefreshRevenueInsightsMsg{})
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

func TestModelRevenues_NewRevenue_Success(t *testing.T) {
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelRevenues(api)

	_, cmd := m.Update(NewRevenueMsg{Account: "New Revenue"})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msgs := collectMsgsFromCmd(cmd)
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}

	if _, ok := msgs[0].(RefreshRevenuesMsg); !ok {
		t.Fatalf("expected RefreshRevenuesMsg, got %T", msgs[0])
	}

	n, ok := msgs[1].(notify.NotifyMsg)
	if !ok {
		t.Fatalf("expected notify.NotifyMsg, got %T", msgs[1])
	}
	if n.Level != notify.Log {
		t.Fatalf("expected log notify level, got %v", n.Level)
	}
	if n.Message != "Revenue account 'New Revenue' created" {
		t.Fatalf("unexpected notify message: %q", n.Message)
	}

	if len(api.createRevenueCalledWith) != 1 || api.createRevenueCalledWith[0] != "New Revenue" {
		t.Fatalf("expected CreateRevenueAccount called with 'New Revenue', got %v", api.createRevenueCalledWith)
	}
}

func TestModelRevenues_NewRevenue_Error(t *testing.T) {
	expectedErr := errors.New("create failed")
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
		createRevenueAccountFunc: func(name string) error {
			return expectedErr
		},
	}
	m := newModelRevenues(api)

	_, cmd := m.Update(NewRevenueMsg{Account: "Bad Revenue"})
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

func TestModelRevenues_RevenuesUpdate_EmitsDataLoadCompleted(t *testing.T) {
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{
				{ID: "r1", Name: "Salary", CurrencyCode: "USD", Type: "revenue"},
			}
		},
		getRevenueDiffFunc: func(accountID string) float64 {
			return 1000
		},
		getTotalRevenueDiffFunc: func() float64 {
			return 2500
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelRevenues(api)

	updated, cmd := m.Update(RevenuesUpdateMsg{})
	m2 := updated.(modelRevenues)
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
	if loader.DataType != "revenues" {
		t.Fatalf("expected DataType 'revenues', got %q", loader.DataType)
	}

	// Verify list items include total + accounts
	// Note: The total is inserted via a separate cmd in the sequence,
	// so we check the updated model's list after the Update() returns
	listItems := m2.list.Items()
	if len(listItems) != 2 {
		t.Fatalf("expected 2 list items (total + 1 account), got %d", len(listItems))
	}

	// First item should be total
	totalItem, ok := listItems[0].(revenueItem)
	if !ok {
		t.Fatalf("expected first item to be revenueItem, got %T", listItems[0])
	}
	if totalItem.account.Name != "Total" {
		t.Errorf("expected first item name 'Total', got %q", totalItem.account.Name)
	}
	if totalItem.earned != 2500 {
		t.Errorf("expected total earned 2500, got %v", totalItem.earned)
	}
}

func TestModelRevenues_UpdatePositions_SetsListSize(t *testing.T) {
	globalWidth := 100
	globalHeight := 40
	topSize := 5

	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelRevenues(api)

	updated, _ := m.Update(UpdatePositions{
		layout: &LayoutConfig{
			Width:   globalWidth,
			Height:  globalHeight,
			TopSize: topSize,
		},
	})
	m2 := updated.(modelRevenues)

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

func TestModelRevenues_IgnoresKeysWhenNotFocused(t *testing.T) {
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{{ID: "r1", Name: "Salary", CurrencyCode: "USD", Type: "revenue"}}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelRevenues(api) // focus is false by default

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if cmd != nil {
		t.Fatalf("expected nil cmd when not focused, got %T", cmd)
	}
}

func TestModelRevenues_KeyFilter_EmitsFilterMsgWithSelectedAccount(t *testing.T) {
	acc := firefly.Account{ID: "r1", Name: "Salary", CurrencyCode: "USD", Type: "revenue"}
	m := newFocusedRevenuesModelWithAccount(t, acc)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	msgs := collectMsgsFromCmd(cmd)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	filter, ok := msgs[0].(FilterMsg)
	if !ok {
		t.Fatalf("expected FilterMsg, got %T", msgs[0])
	}
	if filter.Account.ID != "r1" {
		t.Fatalf("expected account ID 'r1', got %q", filter.Account.ID)
	}
}

func TestModelRevenues_KeyFilter_IgnoresTotalAccount(t *testing.T) {
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{
				{ID: "r1", Name: "Salary", CurrencyCode: "USD", Type: "revenue"},
			}
		},
		getRevenueDiffFunc:      func(accountID string) float64 { return 1000 },
		getTotalRevenueDiffFunc: func() float64 { return 1000 },
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelRevenues(api)
	(&m).Focus()

	// Trigger RevenuesUpdateMsg to add total account
	updated, cmd := m.Update(RevenuesUpdateMsg{})
	m = updated.(modelRevenues)
	_ = cmd

	// Select first item (which should be total)
	m.list.Select(0)

	// Try to filter - should return nil
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	if cmd != nil {
		t.Fatalf("expected nil cmd when filtering total account, got %T", cmd)
	}
}

func TestModelRevenues_KeyRefresh_EmitsRefreshRevenuesMsg(t *testing.T) {
	m := newFocusedRevenuesModelWithAccount(t, firefly.Account{
		ID: "r1", Name: "Salary", CurrencyCode: "USD", Type: "revenue",
	})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	msg := cmd()
	if _, ok := msg.(RefreshRevenuesMsg); !ok {
		t.Fatalf("expected RefreshRevenuesMsg, got %T", msg)
	}
}

func TestModelRevenues_KeySort_TogglesSort(t *testing.T) {
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{
				{ID: "r1", Name: "Low", CurrencyCode: "USD", Type: "revenue"},
				{ID: "r2", Name: "High", CurrencyCode: "USD", Type: "revenue"},
			}
		},
		getRevenueDiffFunc: func(accountID string) float64 {
			if accountID == "r1" {
				return 100
			}
			return 5000
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelRevenues(api)
	(&m).Focus()

	if m.sorted {
		t.Fatal("expected sorted to be false initially")
	}

	// Press 's' to toggle sort
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m2 := updated.(modelRevenues)

	if !m2.sorted {
		t.Fatal("expected sorted to be true after toggle")
	}

	msg := cmd()
	if _, ok := msg.(RevenuesUpdateMsg); !ok {
		t.Fatalf("expected RevenuesUpdateMsg, got %T", msg)
	}

	// Press 's' again to toggle back
	updated, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m3 := updated.(modelRevenues)

	if m3.sorted {
		t.Fatal("expected sorted to be false after second toggle")
	}
}

func TestModelRevenues_KeyResetFilter_EmitsResetFilterMsg(t *testing.T) {
	m := newFocusedRevenuesModelWithAccount(t, firefly.Account{
		ID: "r1", Name: "Salary", CurrencyCode: "USD", Type: "revenue",
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

func TestModelRevenues_KeyNew_ReturnsPromptMsg(t *testing.T) {
	m := newFocusedRevenuesModelWithAccount(t, firefly.Account{
		ID: "r1", Name: "Salary", CurrencyCode: "USD", Type: "revenue",
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
	if p.Prompt != "New Revenue(<name>): " {
		t.Fatalf("unexpected prompt: %q", p.Prompt)
	}
	if p.Callback == nil {
		t.Fatal("expected callback")
	}
}

func TestModelRevenues_View_UsesLeftPanelStyle(t *testing.T) {
	m := newFocusedRevenuesModelWithAccount(t, firefly.Account{
		ID: "r1", Name: "Salary", CurrencyCode: "USD", Type: "revenue",
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

func TestModelRevenues_KeyViewNavigation(t *testing.T) {
	tests := []struct {
		name         string
		key          rune
		expectedView state
		disabled     bool
		expectedMsgs int
	}{
		{"assets", 'a', assetsView, false, 1},
		{"categories", 'c', categoriesView, false, 1},
		{"expenses", 'e', expensesView, false, 1},
		{"transactions", 't', transactionsView, false, 1},
		{"liabilities", 'o', liabilitiesView, false, 1},
		{"revenues (self)", 'i', revenuesView, true, 0},
		{"quit to transactions", 'q', transactionsView, false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newFocusedRevenuesModelWithAccount(t, firefly.Account{
				ID:           "r1",
				Name:         "Salary",
				CurrencyCode: "USD",
				Type:         "revenue",
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

func TestCmdPromptNewRevenue_EmitsPromptMsgWithCallback(t *testing.T) {
	backCmd := Cmd(SetFocusedViewMsg{state: revenuesView})
	cmd := CmdPromptNewRevenue(backCmd)

	msg := cmd()
	p, ok := msg.(prompt.PromptMsg)
	if !ok {
		t.Fatalf("expected prompt.PromptMsg, got %T", msg)
	}
	if p.Prompt != "New Revenue(<name>): " {
		t.Fatalf("unexpected prompt: %q", p.Prompt)
	}
	if p.Callback == nil {
		t.Fatal("expected callback")
	}
}

func TestCmdPromptNewRevenue_CallbackValid_EmitsNewRevenueAndBackCmd(t *testing.T) {
	backCmd := Cmd(SetFocusedViewMsg{state: revenuesView})
	cmd := CmdPromptNewRevenue(backCmd)

	p := cmd().(prompt.PromptMsg)
	msgs := collectMsgsFromCmd(p.Callback("Freelance Income"))

	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}

	newRevenue, ok := msgs[0].(NewRevenueMsg)
	if !ok {
		t.Fatalf("expected NewRevenueMsg, got %T", msgs[0])
	}
	if newRevenue.Account != "Freelance Income" {
		t.Fatalf("expected account 'Freelance Income', got %q", newRevenue.Account)
	}

	focused, ok := msgs[1].(SetFocusedViewMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[1])
	}
	if focused.state != revenuesView {
		t.Fatalf("expected revenuesView, got %v", focused.state)
	}
}

func TestCmdPromptNewRevenue_CallbackNone_EmitsOnlyBackCmd(t *testing.T) {
	backCmd := Cmd(SetFocusedViewMsg{state: revenuesView})
	cmd := CmdPromptNewRevenue(backCmd)

	p := cmd().(prompt.PromptMsg)
	msgs := collectMsgsFromCmd(p.Callback("None"))

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	focused, ok := msgs[0].(SetFocusedViewMsg)
	if !ok {
		t.Fatalf("expected SetFocusedViewMsg, got %T", msgs[0])
	}
	if focused.state != revenuesView {
		t.Fatalf("expected revenuesView, got %v", focused.state)
	}
}

// Edge case tests

func TestModelRevenues_EmptyAccountList(t *testing.T) {
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{}
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}
	m := newModelRevenues(api)
	(&m).Focus()

	// Verify no panics when updating with empty list
	_, cmd := m.Update(RevenuesUpdateMsg{})
	if cmd == nil {
		t.Fatal("expected cmd")
	}

	// Verify view rendering with empty list doesn't panic
	view := m.View()
	if view == "" {
		t.Fatal("expected non-empty view")
	}
}

func TestModelRevenues_NilAPI(t *testing.T) {
	// Document that nil API is not handled gracefully
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when constructing with nil API, but no panic occurred")
		}
	}()

	_ = newModelRevenues(nil)
}

func TestModelRevenues_EarnedBoundaryValues(t *testing.T) {
	tests := []struct {
		name            string
		earned          float64
		expectedDisplay string
	}{
		{"zero earned", 0.0, "Earned: 0.00 USD"},
		{"negative earned", -150.50, "Earned: -150.50 USD"},
		{"very large earned", 999999999.99, "Earned: 999999999.99 USD"},
		{"small positive", 0.01, "Earned: 0.01 USD"},
		{"small negative", -0.01, "Earned: -0.01 USD"},
		{"many decimals", 123.456789, "Earned: 123.46 USD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &mockRevenueAPI{
				accountsByTypeFunc: func(accountType string) []firefly.Account {
					return []firefly.Account{
						{ID: "r1", Name: "Test Revenue", CurrencyCode: "USD", Type: "revenue"},
					}
				},
				getRevenueDiffFunc: func(accountID string) float64 {
					return tt.earned
				},
			}

			items := getRevenuesItems(api, false)
			if len(items) != 1 {
				t.Fatalf("expected 1 item, got %d", len(items))
			}

			item, ok := items[0].(revenueItem)
			if !ok {
				t.Fatalf("expected revenueItem, got %T", items[0])
			}

			if item.earned != tt.earned {
				t.Errorf("expected earned %v, got %v", tt.earned, item.earned)
			}

			if item.Description() != tt.expectedDisplay {
				t.Errorf("expected description %q, got %q", tt.expectedDisplay, item.Description())
			}
		})
	}
}

func TestModelRevenues_AccountNameEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		accountName string
	}{
		{"empty name", ""},
		{"very long name", "This is an extremely long revenue account name that might cause display issues"},
		{"special characters", "Salary™ with 💰 émojis & spëcial çhars"},
		{"whitespace only", "   "},
		{"newlines", "Revenue\nWith\nNewlines"},
		{"tabs", "Revenue\tWith\tTabs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &mockRevenueAPI{
				accountsByTypeFunc: func(accountType string) []firefly.Account {
					return []firefly.Account{
						{ID: "r1", Name: tt.accountName, CurrencyCode: "USD", Type: "revenue"},
					}
				},
				getRevenueDiffFunc: func(accountID string) float64 {
					return 1000.0
				},
			}

			items := getRevenuesItems(api, false)
			if len(items) != 1 {
				t.Fatalf("expected 1 item, got %d", len(items))
			}

			item, ok := items[0].(revenueItem)
			if !ok {
				t.Fatalf("expected revenueItem, got %T", items[0])
			}

			if item.account.Name != tt.accountName {
				t.Errorf("expected name %q, got %q", tt.accountName, item.account.Name)
			}

			title := item.Title()
			if title != tt.accountName {
				t.Errorf("expected title %q, got %q", tt.accountName, title)
			}
		})
	}
}

func TestModelRevenues_MissingCurrencyCode(t *testing.T) {
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{
				{ID: "r1", Name: "No Currency", CurrencyCode: "", Type: "revenue"},
			}
		},
		getRevenueDiffFunc: func(accountID string) float64 {
			return 500.0
		},
	}

	items := getRevenuesItems(api, false)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item, ok := items[0].(revenueItem)
	if !ok {
		t.Fatalf("expected revenueItem, got %T", items[0])
	}

	desc := item.Description()
	if desc != "Earned: 500.00 " {
		t.Errorf("expected description %q, got %q", "Earned: 500.00 ", desc)
	}
}

func TestModelRevenues_FocusToggle(t *testing.T) {
	api := &mockRevenueAPI{
		accountsByTypeFunc: func(accountType string) []firefly.Account {
			return []firefly.Account{
				{ID: "r1", Name: "Test", CurrencyCode: "USD", Type: "revenue"},
			}
		},
		getRevenueDiffFunc: func(accountID string) float64 { return 0 },
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}

	m := newModelRevenues(api)

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

func TestModelRevenues_UpdatePositions_WithVariousDimensions(t *testing.T) {
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
			api := &mockRevenueAPI{
				accountsByTypeFunc: func(accountType string) []firefly.Account {
					return []firefly.Account{}
				},
				primaryCurrencyFunc: func() firefly.Currency {
					return firefly.Currency{Code: "USD", Symbol: "$"}
				},
			}
			m := newModelRevenues(api)

			updated, _ := m.Update(UpdatePositions{
				layout: &LayoutConfig{
					Width:   tt.globalWidth,
					Height:  tt.globalHeight,
					TopSize: tt.topSize,
				},
			})
			m2 := updated.(modelRevenues)

			if m2.list.Width() < tt.expectedMinWidth {
				t.Errorf("expected width >= %d, got %d", tt.expectedMinWidth, m2.list.Width())
			}
			if m2.list.Height() < tt.expectedMinHeight {
				t.Errorf("expected height >= %d, got %d", tt.expectedMinHeight, m2.list.Height())
			}
		})
	}
}
