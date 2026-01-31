/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"strings"
	"sync"
	"testing"
	"time"

	"ffiii-tui/internal/firefly"
	"ffiii-tui/internal/ui/prompt"

	tea "github.com/charmbracelet/bubbletea"
)

// Mock UIAPI implementation for testing
type mockUIAPI struct {
	// PeriodAPI
	previousPeriodCalled int
	nextPeriodCalled     int

	// SummaryAPI
	updateSummaryCalled int
	getMaxWidthFunc     func() int
	summaryItemsFunc    func() map[string]firefly.SummaryItem

	// AccountsAPI
	updateAccountsCalled int
	accountsByTypeFunc   func(accountType string) []firefly.Account
	accountBalanceFunc   func(accountID string) float64

	// CategoriesAPI
	updateCategoriesCalled         int
	updateCategoriesInsightsCalled int
	categoriesListFunc             func() []firefly.Category
	getTotalSpentEarnedFunc        func() (float64, float64)
	categorySpentFunc              func(categoryID string) float64
	categoryEarnedFunc             func(categoryID string) float64

	// InsightsAPI
	updateExpenseInsightsCalled int
	getExpenseDiffFunc          func(accountID string) float64
	getTotalExpenseDiffFunc     func() float64
	updateRevenueInsightsCalled int
	getRevenueDiffFunc          func(accountID string) float64
	getTotalRevenueDiffFunc     func() float64

	// TransactionAPI
	listTransactionsFunc  func(query string) ([]firefly.Transaction, error)
	deleteTransactionFunc func(transactionID string) error

	// TransactionWriteAPI
	createTransactionFunc func(tx firefly.RequestTransaction) (string, error)
	updateTransactionFunc func(transactionID string, tx firefly.RequestTransaction) (string, error)

	// Account creation
	createAssetAccountFunc     func(name, currencyCode string) error
	createExpenseAccountFunc   func(name string) error
	createRevenueAccountFunc   func(name string) error
	createLiabilityAccountFunc func(nl firefly.NewLiability) error
	createCategoryFunc         func(name, notes string) error

	// Period and Currency
	timeoutSeconds  int
	periodStart     time.Time
	periodEnd       time.Time
	primaryCurrency firefly.Currency
}

func newTestUIAPI() *mockUIAPI {
	now := time.Now()
	return &mockUIAPI{
		timeoutSeconds: 10,
		periodStart:    time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		periodEnd:      time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, time.UTC),
		primaryCurrency: firefly.Currency{
			Code:   "USD",
			Symbol: "$",
		},
	}
}

// PeriodAPI methods
func (m *mockUIAPI) PreviousPeriod() {
	m.previousPeriodCalled++
	m.periodStart = m.periodStart.AddDate(0, -1, 0)
	m.periodEnd = m.periodEnd.AddDate(0, -1, 0)
}

func (m *mockUIAPI) NextPeriod() {
	m.nextPeriodCalled++
	m.periodStart = m.periodStart.AddDate(0, 1, 0)
	m.periodEnd = m.periodEnd.AddDate(0, 1, 0)
}

func (m *mockUIAPI) PeriodStart() time.Time { return m.periodStart }
func (m *mockUIAPI) PeriodEnd() time.Time   { return m.periodEnd }
func (m *mockUIAPI) TimeoutSeconds() int    { return m.timeoutSeconds }

// CurrencyAPI methods
func (m *mockUIAPI) PrimaryCurrency() firefly.Currency { return m.primaryCurrency }

// SummaryAPI methods
func (m *mockUIAPI) UpdateSummary() error {
	m.updateSummaryCalled++
	return nil
}

func (m *mockUIAPI) GetMaxWidth() int {
	if m.getMaxWidthFunc != nil {
		return m.getMaxWidthFunc()
	}
	return 40
}

func (m *mockUIAPI) SummaryItems() map[string]firefly.SummaryItem {
	if m.summaryItemsFunc != nil {
		return m.summaryItemsFunc()
	}
	return map[string]firefly.SummaryItem{}
}

// AccountsAPI methods
func (m *mockUIAPI) UpdateAccounts(accountType string) error {
	m.updateAccountsCalled++
	return nil
}

func (m *mockUIAPI) AccountsByType(accountType string) []firefly.Account {
	if m.accountsByTypeFunc != nil {
		return m.accountsByTypeFunc(accountType)
	}
	return []firefly.Account{}
}

func (m *mockUIAPI) AccountBalance(accountID string) float64 {
	if m.accountBalanceFunc != nil {
		return m.accountBalanceFunc(accountID)
	}
	return 0
}

// Account creation methods
func (m *mockUIAPI) CreateAssetAccount(name, currencyCode string) error {
	if m.createAssetAccountFunc != nil {
		return m.createAssetAccountFunc(name, currencyCode)
	}
	return nil
}

func (m *mockUIAPI) CreateExpenseAccount(name string) error {
	if m.createExpenseAccountFunc != nil {
		return m.createExpenseAccountFunc(name)
	}
	return nil
}

func (m *mockUIAPI) CreateRevenueAccount(name string) error {
	if m.createRevenueAccountFunc != nil {
		return m.createRevenueAccountFunc(name)
	}
	return nil
}

func (m *mockUIAPI) CreateLiabilityAccount(nl firefly.NewLiability) error {
	if m.createLiabilityAccountFunc != nil {
		return m.createLiabilityAccountFunc(nl)
	}
	return nil
}

// CategoriesAPI methods
func (m *mockUIAPI) UpdateCategories() error {
	m.updateCategoriesCalled++
	return nil
}

func (m *mockUIAPI) UpdateCategoriesInsights() error {
	m.updateCategoriesInsightsCalled++
	return nil
}

func (m *mockUIAPI) CategoriesList() []firefly.Category {
	if m.categoriesListFunc != nil {
		return m.categoriesListFunc()
	}
	return []firefly.Category{}
}

func (m *mockUIAPI) GetTotalSpentEarnedCategories() (float64, float64) {
	if m.getTotalSpentEarnedFunc != nil {
		return m.getTotalSpentEarnedFunc()
	}
	return 0, 0
}

func (m *mockUIAPI) CategorySpent(categoryID string) float64 {
	if m.categorySpentFunc != nil {
		return m.categorySpentFunc(categoryID)
	}
	return 0
}

func (m *mockUIAPI) CategoryEarned(categoryID string) float64 {
	if m.categoryEarnedFunc != nil {
		return m.categoryEarnedFunc(categoryID)
	}
	return 0
}

func (m *mockUIAPI) CreateCategory(name, notes string) error {
	if m.createCategoryFunc != nil {
		return m.createCategoryFunc(name, notes)
	}
	return nil
}

// InsightsAPI methods
func (m *mockUIAPI) UpdateExpenseInsights() error {
	m.updateExpenseInsightsCalled++
	return nil
}

func (m *mockUIAPI) GetExpenseDiff(accountID string) float64 {
	if m.getExpenseDiffFunc != nil {
		return m.getExpenseDiffFunc(accountID)
	}
	return 0
}

func (m *mockUIAPI) GetTotalExpenseDiff() float64 {
	if m.getTotalExpenseDiffFunc != nil {
		return m.getTotalExpenseDiffFunc()
	}
	return 0
}

func (m *mockUIAPI) UpdateRevenueInsights() error {
	m.updateRevenueInsightsCalled++
	return nil
}

func (m *mockUIAPI) GetRevenueDiff(accountID string) float64 {
	if m.getRevenueDiffFunc != nil {
		return m.getRevenueDiffFunc(accountID)
	}
	return 0
}

func (m *mockUIAPI) GetTotalRevenueDiff() float64 {
	if m.getTotalRevenueDiffFunc != nil {
		return m.getTotalRevenueDiffFunc()
	}
	return 0
}

// TransactionAPI methods
func (m *mockUIAPI) ListTransactions(query string) ([]firefly.Transaction, error) {
	if m.listTransactionsFunc != nil {
		return m.listTransactionsFunc(query)
	}
	return []firefly.Transaction{}, nil
}

func (m *mockUIAPI) DeleteTransaction(transactionID string) error {
	if m.deleteTransactionFunc != nil {
		return m.deleteTransactionFunc(transactionID)
	}
	return nil
}

// TransactionWriteAPI methods
func (m *mockUIAPI) CreateTransaction(tx firefly.RequestTransaction) (string, error) {
	if m.createTransactionFunc != nil {
		return m.createTransactionFunc(tx)
	}
	return "", nil
}

func (m *mockUIAPI) UpdateTransaction(transactionID string, tx firefly.RequestTransaction) (string, error) {
	if m.updateTransactionFunc != nil {
		return m.updateTransactionFunc(transactionID, tx)
	}
	return "", nil
}

// Helper function to create a test modelUI
func newTestModelUI() modelUI {
	api := newTestUIAPI()
	return NewModelUI(api)
}

// =============================================================================
// Basic Functionality Tests
// =============================================================================

func TestUI_Init(t *testing.T) {
	m := newTestModelUI()

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Expected Init to return a command")
	}

	// Init now returns tea.Batch with multiple commands (RefreshAllMsg and spinner.Tick)
	msg := cmd()
	if _, ok := msg.(tea.BatchMsg); !ok {
		t.Errorf("Expected tea.BatchMsg from Init, got %T", msg)
	}
}

func TestUI_UpdateModel_TypeAssertion(t *testing.T) {
	m := newTestModelUI()

	// Test successful update
	updated, cmd := updateModel(m.assets, tea.WindowSizeMsg{Width: 100, Height: 50})
	if updated.api == nil {
		t.Error("Expected updated model to have API set")
	}
	_ = cmd // May or may not be nil depending on the message
}

func TestUI_SetState(t *testing.T) {
	m := newTestModelUI()

	states := []state{
		transactionsView,
		assetsView,
		categoriesView,
		expensesView,
		revenuesView,
		liabilitiesView,
		newView,
	}

	for _, s := range states {
		m.SetState(s)
		if m.state != s {
			t.Errorf("Expected state %d, got %d", s, m.state)
		}
	}
}

func TestUI_SetView_Command(t *testing.T) {
	cmd := SetView(assetsView)
	if cmd == nil {
		t.Fatal("Expected SetView to return a command")
	}

	// SetView returns a tea.Sequence, which executes commands in order
	// We can verify it returns a command without trying to execute the full sequence
	// (tea.Sequence is not easily testable in unit tests as it requires the tea runtime)
}

// =============================================================================
// Key Binding Tests
// =============================================================================

func TestUI_KeyQuit(t *testing.T) {
	m := newTestModelUI()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Fatal("Expected Quit command")
	}

	// tea.Quit returns tea.QuitMsg when executed
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("Expected tea.QuitMsg, got %T", msg)
	}
}

func TestUI_KeyToggleHelp(t *testing.T) {
	m := newTestModelUI()

	initialShowAll := m.help.ShowAll

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if cmd == nil {
		t.Error("Expected command from help toggle")
	}

	m2 := updated.(modelUI)
	if m2.help.ShowAll == initialShowAll {
		t.Error("Expected ShowAll to toggle")
	}

	// Toggle again
	updated2, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m3 := updated2.(modelUI)
	if m3.help.ShowAll != initialShowAll {
		t.Error("Expected ShowAll to toggle back")
	}
}

func TestUI_KeyPreviousPeriod(t *testing.T) {
	api := newTestUIAPI()
	m := modelUI{
		api:          api,
		transactions: NewModelTransactions(api),
		new:          newModelTransaction(api),
		assets:       newModelAssets(api),
		categories:   newModelCategories(api),
		expenses:     newModelExpenses(api),
		revenues:     newModelRevenues(api),
		liabilities:  newModelLiabilities(api),
		summary:      newModelSummary(api),
		keymap:       DefaultUIKeyMap(),
		styles:       DefaultStyles(),
	}
	m.transactions.currentSearch = "test"

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})

	if cmd == nil {
		t.Fatal("Expected command from previous period")
	}

	if api.previousPeriodCalled != 1 {
		t.Errorf("Expected PreviousPeriod to be called once, got %d", api.previousPeriodCalled)
	}

	// Should clear search
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	m2 := updated.(modelUI)
	if m2.transactions.currentSearch != "" {
		t.Error("Expected search to be cleared")
	}

	// Check that refresh commands are sent
	msgs := collectMsgsFromCmd(cmd)
	foundRefreshTransactions := false
	foundRefreshSummary := false

	for _, msg := range msgs {
		switch msg.(type) {
		case RefreshTransactionsMsg:
			foundRefreshTransactions = true
		case RefreshSummaryMsg:
			foundRefreshSummary = true
		}
	}

	if !foundRefreshTransactions {
		t.Error("Expected RefreshTransactionsMsg in batch")
	}
	if !foundRefreshSummary {
		t.Error("Expected RefreshSummaryMsg in batch")
	}
}

func TestUI_KeyNextPeriod(t *testing.T) {
	api := newTestUIAPI()
	m := modelUI{
		api:          api,
		transactions: NewModelTransactions(api),
		new:          newModelTransaction(api),
		assets:       newModelAssets(api),
		categories:   newModelCategories(api),
		expenses:     newModelExpenses(api),
		revenues:     newModelRevenues(api),
		liabilities:  newModelLiabilities(api),
		summary:      newModelSummary(api),
		keymap:       DefaultUIKeyMap(),
		styles:       DefaultStyles(),
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})

	if cmd == nil {
		t.Fatal("Expected command from next period")
	}

	if api.nextPeriodCalled != 1 {
		t.Errorf("Expected NextPeriod to be called once, got %d", api.nextPeriodCalled)
	}
}

// =============================================================================
// Message Handler Tests
// =============================================================================

func TestUI_WindowSizeMsg(t *testing.T) {
	m := newTestModelUI()

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})

	if cmd == nil {
		t.Fatal("Expected command from WindowSizeMsg")
	}

	if m.layout.GetWidth() != 200 {
		t.Errorf("Expected m.layout.GetWidth() 200, got %d", m.layout.GetWidth())
	}
	if m.layout.GetHeight() != 60 {
		t.Errorf("Expected m.layout.GetHeight() 60, got %d", m.layout.GetHeight())
	}

	// Should return UpdatePositions command
	msg := cmd()
	if _, ok := msg.(UpdatePositions); !ok {
		t.Errorf("Expected UpdatePositions, got %T", msg)
	}

	_ = updated
}

func TestUI_UpdatePositions_TransactionsView(t *testing.T) {
	m := newTestModelUI()
	m.state = transactionsView
	globalWidth := 200
	globalHeight := 60

	updated, _ := m.Update(UpdatePositions{
		layout: &LayoutConfig{
			Width:  globalWidth,
			Height: globalHeight,
		},
	})
	m2 := updated.(modelUI)

	if m2.Width == 0 {
		t.Error("Expected Width to be set")
	}

	if m2.layout.TopSize == 0 {
		t.Error("Expected topSize to be set")
	}

	m2.layout = m2.layout.WithFullTransactionView(true)
	updated2, _ := m2.Update(UpdatePositions{
		layout: m2.layout, // Use the modified layout
	})

	m3 := updated2.(modelUI)
	if m3.layout.LeftSize != 0 {
		t.Errorf("Expected leftSize to be 0 in full view, got %d", m3.layout.LeftSize)
	}

	_ = m3
}

func TestUI_UpdatePositions_DifferentViews(t *testing.T) {
	tests := []struct {
		name  string
		state state
	}{
		{"assets view", assetsView},
		{"categories view", categoriesView},
		{"expenses view", expensesView},
		{"revenues view", revenuesView},
		{"liabilities view", liabilitiesView},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModelUI()
			m.state = tt.state

			updated, _ := m.Update(UpdatePositions{
				layout: &LayoutConfig{
					Width:  200,
					Height: 60,
				},
			})
			m2 := updated.(modelUI)

			if m2.Width == 0 {
				t.Error("Expected Width to be set")
			}
		})
	}
}

func TestUI_SetFocusedViewMsg_TransactionsView(t *testing.T) {
	m := newTestModelUI()

	updated, _ := m.Update(SetFocusedViewMsg{state: transactionsView})
	m2 := updated.(modelUI)

	if m2.state != transactionsView {
		t.Errorf("Expected state transactionsView, got %d", m2.state)
	}

	// Transactions should be focused
	if !m2.transactions.focus {
		t.Error("Expected transactions to be focused")
	}

	// Others should be blurred
	if m2.assets.focus {
		t.Error("Expected assets to be blurred")
	}
	if m2.categories.focus {
		t.Error("Expected categories to be blurred")
	}
}

func TestUI_SetFocusedViewMsg_AllViews(t *testing.T) {
	tests := []struct {
		name  string
		state state
	}{
		{"transactions", transactionsView},
		{"assets", assetsView},
		{"categories", categoriesView},
		{"expenses", expensesView},
		{"revenues", revenuesView},
		{"liabilities", liabilitiesView},
		{"new transaction", newView},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModelUI()

			updated, cmd := m.Update(SetFocusedViewMsg{state: tt.state})
			m2 := updated.(modelUI)

			if m2.state != tt.state {
				t.Errorf("Expected state %d, got %d", tt.state, m2.state)
			}

			msg := cmd()
			if _, ok := msg.(UpdatePositions); !ok {
				t.Errorf("Expected UpdatePositions, got %T", msg)
			}
		})
	}
}

func TestUI_ViewFullTransactionViewMsg(t *testing.T) {
	m := newTestModelUI()

	m.layout.FullTransactionView = false

	updated, _ := m.Update(ViewFullTransactionViewMsg{})
	_ = updated
	m2 := updated.(modelUI)

	if !m2.layout.FullTransactionView {
		t.Error("Expected fullTransactionView to toggle to true")
	}

	updated2, _ := m.Update(ViewFullTransactionViewMsg{})
	_ = updated2
	m3 := updated2.(modelUI)

	if m3.layout.FullTransactionView {
		t.Error("Expected fullTransactionView to toggle to false")
	}
}

func TestUI_DataLoadCompletedMsg(t *testing.T) {
	m := newTestModelUI()
	m.loadStatus["assets"] = false

	updated, _ := m.Update(DataLoadCompletedMsg{DataType: "assets"})
	m2 := updated.(modelUI)

	if !m2.loadStatus["assets"] {
		t.Error("Expected assets to be marked as loaded")
	}
}

func TestUI_LazyLoadMsg_AllResourcesLoaded(t *testing.T) {
	m := newTestModelUI()
	m.loadStatus = map[string]bool{
		"assets":      true,
		"expenses":    true,
		"revenues":    true,
		"liabilities": true,
		"categories":  true,
	}

	updated, cmd := m.Update(LazyLoadMsg{c: 5})

	if cmd == nil {
		t.Fatal("Expected command from LazyLoadMsg")
	}

	msgs := collectMsgsFromCmd(cmd)
	foundRefreshTransactions := false
	foundRefreshSummary := false

	for _, msg := range msgs {
		switch msg.(type) {
		case RefreshTransactionsMsg:
			foundRefreshTransactions = true
		case RefreshSummaryMsg:
			foundRefreshSummary = true
		}
	}

	if !foundRefreshTransactions {
		t.Error("Expected RefreshTransactionsMsg in batch")
	}
	if !foundRefreshSummary {
		t.Error("Expected RefreshSummaryMsg in batch")
	}

	_ = updated
}

func TestUI_LazyLoadMsg_ResourcesStillLoading_Retry(t *testing.T) {
	m := newTestModelUI()
	m.loadStatus = map[string]bool{
		"assets":      true,
		"expenses":    false, // Still loading
		"revenues":    true,
		"liabilities": true,
		"categories":  true,
	}

	updated, cmd := m.Update(LazyLoadMsg{c: 3})

	if cmd == nil {
		t.Fatal("Expected command from LazyLoadMsg")
	}

	// Execute the command to get the message
	msg := cmd()
	if msg == nil {
		t.Fatal("Expected message from command")
	}

	// The command should produce another LazyLoadMsg after the tick
	// We can't easily test the timing, but we verify a command was returned
	// which would be the tea.Tick that schedules the next check

	_ = updated
}

func TestUI_LazyLoadMsg_Timeout_ShowsWarning(t *testing.T) {
	m := newTestModelUI()
	m.loadStatus = map[string]bool{
		"assets":      true,
		"expenses":    false, // Still loading
		"revenues":    true,
		"liabilities": true,
		"categories":  true,
	}

	updated, cmd := m.Update(LazyLoadMsg{c: 1})

	if cmd == nil {
		t.Fatal("Expected warning command from LazyLoadMsg")
	}

	// The command should be a notify.NotifyWarn
	// We can't easily inspect the notification content in unit tests,
	// but we verify that a command was returned (the warning)

	_ = updated
}

func TestUI_LazyLoadMsg_FullLifecycle(t *testing.T) {
	m := newTestModelUI()
	m.loadStatus = map[string]bool{
		"assets":      false,
		"expenses":    false,
		"revenues":    false,
		"liabilities": false,
		"categories":  false,
	}

	// Step 1: First LazyLoadMsg with resources unloaded - should retry
	updated, cmd := m.Update(LazyLoadMsg{c: 3})
	m = updated.(modelUI)

	if cmd == nil {
		t.Fatal("Expected retry command from LazyLoadMsg")
	}

	// Verify it's a retry (not a refresh batch)
	msgs := collectMsgsFromCmd(cmd)
	hasRefreshTransactions := false
	for _, msg := range msgs {
		if _, ok := msg.(RefreshTransactionsMsg); ok {
			hasRefreshTransactions = true
		}
	}
	if hasRefreshTransactions {
		t.Error("Expected retry, not refresh commands on first attempt")
	}

	// Step 2: Second LazyLoadMsg, still loading - should retry
	updated, cmd = m.Update(LazyLoadMsg{c: 2})
	m = updated.(modelUI)

	if cmd == nil {
		t.Fatal("Expected retry command from second LazyLoadMsg")
	}

	// Step 3: Mark all resources as loaded
	m.loadStatus = map[string]bool{
		"assets":      true,
		"expenses":    true,
		"revenues":    true,
		"liabilities": true,
		"categories":  true,
	}

	// Step 4: Third LazyLoadMsg, now all loaded - should trigger refresh
	updated, cmd = m.Update(LazyLoadMsg{c: 1})
	m = updated.(modelUI)

	if cmd == nil {
		t.Fatal("Expected refresh command from LazyLoadMsg after loading complete")
	}

	msgs = collectMsgsFromCmd(cmd)
	foundRefreshTransactions := false
	foundRefreshSummary := false

	for _, msg := range msgs {
		switch msg.(type) {
		case RefreshTransactionsMsg:
			foundRefreshTransactions = true
		case RefreshSummaryMsg:
			foundRefreshSummary = true
		}
	}

	if !foundRefreshTransactions {
		t.Error("Expected RefreshTransactionsMsg after all resources loaded")
	}
	if !foundRefreshSummary {
		t.Error("Expected RefreshSummaryMsg after all resources loaded")
	}
}

func TestUI_RefreshAllMsg(t *testing.T) {
	m := newTestModelUI()
	m.loadStatus = map[string]bool{
		"assets":      true,
		"expenses":    true,
		"revenues":    true,
		"liabilities": true,
		"categories":  true,
	}

	updated, cmd := m.Update(RefreshAllMsg{})

	if cmd == nil {
		t.Fatal("Expected command from RefreshAllMsg")
	}

	m2 := updated.(modelUI)

	// All load statuses should be reset to false
	for key, loaded := range m2.loadStatus {
		if loaded {
			t.Errorf("Expected %s to be marked as not loaded", key)
		}
	}

	// Command should be a batch - we can verify it's not nil
	// The actual execution would require the tea runtime
	// Just verify that important refresh messages would be sent
}

// =============================================================================
// Message Routing Tests
// =============================================================================

func TestUI_MessageRoutingToSubModels(t *testing.T) {
	m := newTestModelUI()

	// Send a WindowSizeMsg which should be routed to all sub-models
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 150, Height: 80})
	m2 := updated.(modelUI)

	// Verify that the message was processed (globalWidth should be updated)
	if m2.layout.Width != 150 {
		t.Errorf("Expected m2.layout.Width 150, got %d", m2.layout.Width)
	}

	_ = m2
}

func TestUI_PromptFocusedBlocksOtherUpdates(t *testing.T) {
	m := newTestModelUI()
	m.prompt = prompt.New() // Properly initialize prompt

	// Focus the prompt
	m.prompt.Focus()

	// Send a message that would normally update other models
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	// Command should only contain prompt updates
	if cmd == nil {
		// Prompt might return nil if it doesn't handle the key
		return
	}

	m2 := updated.(modelUI)
	if !m2.prompt.Focused() {
		t.Error("Expected prompt to remain focused")
	}
}

// =============================================================================
// View Tests
// =============================================================================

func TestUI_View_TransactionsView(t *testing.T) {
	m := newTestModelUI()
	m.state = transactionsView

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}

	if !strings.Contains(view, "ffiii-tui") {
		t.Error("Expected view to contain 'ffiii-tui' header")
	}
}

func TestUI_View_AllStates(t *testing.T) {
	tests := []struct {
		name  string
		state state
	}{
		{"transactions", transactionsView},
		{"assets", assetsView},
		{"categories", categoriesView},
		{"expenses", expensesView},
		{"revenues", revenuesView},
		{"liabilities", liabilitiesView},
		{"new", newView},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModelUI()
			m.state = tt.state

			view := m.View()

			if view == "" {
				t.Error("Expected non-empty view")
			}

			// All views should contain the header
			if !strings.Contains(view, "ffiii-tui") {
				t.Error("Expected view to contain 'ffiii-tui' header")
			}
		})
	}
}

func TestUI_View_FullTransactionView(t *testing.T) {
	m := newTestModelUI()
	m.state = transactionsView

	// Test normal view
	m.layout = m.layout.WithFullTransactionView(false)
	view1 := m.View()

	// Test full view
	m.layout = m.layout.WithFullTransactionView(true)
	view2 := m.View()

	// Views should be different
	if view1 == view2 {
		t.Error("Expected different views for full vs normal transaction view")
	}
}

func TestUI_View_WithSearch(t *testing.T) {
	m := newTestModelUI()
	m.state = transactionsView
	m.transactions.currentSearch = "test search"

	view := m.View()

	if !strings.Contains(view, "Search: test search") {
		t.Error("Expected view to contain search term")
	}
}

func TestUI_View_WithAccountFilter(t *testing.T) {
	m := newTestModelUI()
	m.state = transactionsView
	m.transactions.currentAccount = firefly.Account{
		ID:   "1",
		Name: "Test Account",
		Type: "asset",
	}

	view := m.View()

	if !strings.Contains(view, "Account: Test Account") {
		t.Error("Expected view to contain account filter")
	}
}

func TestUI_View_WithCategoryFilter(t *testing.T) {
	m := newTestModelUI()
	m.state = transactionsView
	m.transactions.currentCategory = firefly.Category{
		ID:   "1",
		Name: "Test Category",
	}

	view := m.View()

	if !strings.Contains(view, "Category: Test Category") {
		t.Error("Expected view to contain category filter")
	}
}

func TestUI_View_NewTransaction(t *testing.T) {
	m := newTestModelUI()
	m.state = newView
	m.new.new = true

	view := m.View()

	if !strings.Contains(view, "New transaction") {
		t.Error("Expected view to contain 'New transaction'")
	}
}

func TestUI_View_EditTransaction(t *testing.T) {
	m := newTestModelUI()
	m.state = newView
	m.new.new = false
	m.new.attr = &transactionAttr{
		trxID: "123",
	}

	view := m.View()

	if !strings.Contains(view, "Editing transaction: 123") {
		t.Error("Expected view to contain 'Editing transaction: 123'")
	}
}

func TestUI_View_WithPromptFocused(t *testing.T) {
	m := newTestModelUI()
	m.prompt.Focus()

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view with prompt focused")
	}
}

// =============================================================================
// Help View Tests
// =============================================================================

func TestUI_HelpView_AllStates(t *testing.T) {
	tests := []struct {
		name  string
		state state
	}{
		{"transactions", transactionsView},
		{"assets", assetsView},
		{"categories", categoriesView},
		{"expenses", expensesView},
		{"revenues", revenuesView},
		{"liabilities", liabilitiesView},
		{"new", newView},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModelUI()
			m.state = tt.state

			helpView := m.HelpView()

			// Help view should not be empty
			if helpView == "" {
				t.Error("Expected non-empty help view")
			}
		})
	}
}

func TestUI_HelpView_ShowAll(t *testing.T) {
	m := newTestModelUI()
	m.state = transactionsView

	// Test with ShowAll = false
	m.help.ShowAll = false
	helpView1 := m.HelpView()

	// Test with ShowAll = true
	m.help.ShowAll = true
	helpView2 := m.HelpView()

	// ShowAll view should be longer
	if len(helpView2) <= len(helpView1) {
		t.Error("Expected ShowAll help view to be longer")
	}
}

// =============================================================================
// Loading Indicator Tests
// =============================================================================

func TestLoading_BasicOperations(t *testing.T) {
	// Reset
	loading.Store(0)
	loadingOps = sync.Map{}

	// Test start increments counter and stores message
	opID := startLoading("Loading test...")
	if opID == "" {
		t.Fatal("Expected non-empty operation ID")
	}
	if loading.Load() != 1 {
		t.Errorf("Expected counter to be 1, got %d", loading.Load())
	}

	// Verify message is stored
	msg := buildLoadingMessage()
	if !strings.Contains(msg, "test...") {
		t.Errorf("Expected message to contain 'test...', got '%s'", msg)
	}

	// Test stop decrements counter and removes operation
	stopLoading(opID)
	if loading.Load() != 0 {
		t.Errorf("Expected counter to be 0, got %d", loading.Load())
	}

	// Verify message is cleared
	msg = buildLoadingMessage()
	if msg != "..." {
		t.Errorf("Expected fallback '...', got '%s'", msg)
	}
}

func TestLoading_OverflowUnderflowProtection(t *testing.T) {
	// Test overflow protection
	loading.Store(99)
	loadingOps = sync.Map{}

	op1 := startLoading("Test")
	if loading.Load() != 100 {
		t.Errorf("Expected counter to be 100, got %d", loading.Load())
	}

	op2 := startLoading("Test")
	if op2 != "" {
		t.Error("Expected empty operation ID when max reached")
	}
	if loading.Load() != 100 {
		t.Errorf("Expected counter to stay at 100, got %d", loading.Load())
	}

	// Cleanup
	stopLoading(op1)

	// Test underflow protection
	loading.Store(0)
	loadingOps = sync.Map{}
	stopLoading("")
	stopLoading("invalid_id")
	if loading.Load() != 0 {
		t.Errorf("Expected counter to stay at 0, got %d", loading.Load())
	}
}

func TestLoading_NestedOperations(t *testing.T) {
	// Reset
	loading.Store(0)
	loadingOps = sync.Map{}

	// Start multiple operations
	op1 := startLoading("Operation 1")
	op2 := startLoading("Operation 2")
	op3 := startLoading("Operation 3")

	if loading.Load() != 3 {
		t.Errorf("Expected counter to be 3, got %d", loading.Load())
	}

	// Message should show all operations
	msg := buildLoadingMessage()
	if !strings.Contains(msg, "(3)") {
		t.Errorf("Expected message to show count (3), got '%s'", msg)
	}

	// Stop one operation - should show 2 operations
	stopLoading(op2)
	if loading.Load() != 2 {
		t.Errorf("Expected counter to be 2, got %d", loading.Load())
	}

	msg = buildLoadingMessage()
	if !strings.Contains(msg, "(2)") {
		t.Errorf("Expected message to show count (2), got '%s'", msg)
	}

	// Stop remaining operations
	stopLoading(op1)
	stopLoading(op3)

	if loading.Load() != 0 {
		t.Errorf("Expected counter to be 0, got %d", loading.Load())
	}

	msg = buildLoadingMessage()
	if msg != "..." {
		t.Errorf("Expected fallback '...', got '%s'", msg)
	}
}

func TestLoading_ViewIntegration(t *testing.T) {
	// Reset
	loading.Store(0)
	loadingOps = sync.Map{}

	m := newTestModelUI()
	m.Width = 100

	// Initially no spinner
	view := m.View()
	if strings.Contains(view, "Loading") {
		t.Error("Expected no loading indicator when counter is 0")
	}

	// Start loading
	opID := startLoading("Loading transactions...")
	view = m.View()
	if !strings.Contains(view, "transactions...") {
		t.Error("Expected to see loading message in view")
	}

	// Stop loading
	stopLoading(opID)
	view = m.View()
	if strings.Contains(view, "transactions...") {
		t.Error("Expected loading indicator to be gone when counter is 0")
	}
}

func TestLoading_FallbackMessage(t *testing.T) {
	// Test with no operations
	loading.Store(0)
	loadingOps = sync.Map{}

	msg := buildLoadingMessage()
	if msg != "..." {
		t.Errorf("Expected fallback '...', got '%s'", msg)
	}

	// Cleanup
	loading.Store(0)
	loadingOps = sync.Map{}
}

func TestLoading_MultipleOperationsDisplay(t *testing.T) {
	// Reset
	loading.Store(0)
	loadingOps = sync.Map{}

	// Start multiple operations
	op1 := startLoading("Loading transactions...")
	op2 := startLoading("Loading categories...")
	op3 := startLoading("Creating category...")

	// Verify all tracked
	if loading.Load() != 3 {
		t.Errorf("Expected 3 operations, got %d", loading.Load())
	}

	// Check display message contains count
	msg := buildLoadingMessage()
	if !strings.Contains(msg, "(3)") {
		t.Errorf("Expected message to contain '(3)', got '%s'", msg)
	}

	// Should show abbreviated messages
	if !strings.Contains(msg, " ") {
		t.Error("Expected separator ' ' between messages")
	}

	// Stop one operation
	stopLoading(op2)
	if loading.Load() != 2 {
		t.Errorf("Expected 2 operations, got %d", loading.Load())
	}

	// Check message updated
	msg = buildLoadingMessage()
	if !strings.Contains(msg, "(2)") {
		t.Errorf("Expected message to contain '(2)', got '%s'", msg)
	}

	// Cleanup
	stopLoading(op1)
	stopLoading(op3)
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestUI_UnknownMessage(t *testing.T) {
	m := newTestModelUI()

	type unknownMsg struct{}

	updated, cmd := m.Update(unknownMsg{})

	// Should not panic and should return the model
	if _, ok := updated.(modelUI); !ok {
		t.Error("Expected modelUI to be returned")
	}

	// Command may or may not be nil depending on sub-models
	// The important thing is that the update doesn't panic
	_ = cmd
}

func TestUI_MultipleStateTransitions(t *testing.T) {
	m := newTestModelUI()

	states := []state{
		transactionsView,
		assetsView,
		categoriesView,
		expensesView,
		revenuesView,
		liabilitiesView,
		newView,
		transactionsView,
	}

	for _, s := range states {
		updated, _ := m.Update(SetFocusedViewMsg{state: s})
		m = updated.(modelUI)

		if m.state != s {
			t.Errorf("Expected state %d, got %d", s, m.state)
		}
	}
}

func TestUI_IntegrationSequence(t *testing.T) {
	m := newTestModelUI()

	// 1. Initialize
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Expected Init command")
	}

	// 2. Process RefreshAllMsg
	msg := cmd()
	updated, _ := m.Update(msg)
	m = updated.(modelUI)

	// 3. Process WindowSize
	updated, _ = m.Update(tea.WindowSizeMsg{Width: 150, Height: 80})
	m = updated.(modelUI)

	// 4. Switch to assets view
	updated, _ = m.Update(SetFocusedViewMsg{state: assetsView})
	m = updated.(modelUI)

	if m.state != assetsView {
		t.Error("Expected assets view")
	}

	// 5. Navigate to categories
	updated, _ = m.Update(SetFocusedViewMsg{state: categoriesView})
	m = updated.(modelUI)

	if m.state != categoriesView {
		t.Error("Expected categories view")
	}

	// 6. Toggle help
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(modelUI)

	if !m.help.ShowAll {
		t.Error("Expected help to be shown")
	}

	// 7. Render view
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
}
