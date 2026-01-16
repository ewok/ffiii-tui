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

type mockCategoryAPI struct {
	updateCategoriesFunc           func() error
	updateCategoriesInsightsFunc   func() error
	categoriesListFunc             func() []firefly.Category
	getTotalSpentEarnedFunc        func() (float64, float64)
	categorySpentFunc              func(categoryID string) float64
	categoryEarnedFunc             func(categoryID string) float64
	createCategoryFunc             func(name, notes string) error
	primaryCurrencyFunc            func() firefly.Currency
	updateCategoriesCalled         bool
	updateCategoriesInsightsCalled bool
	createCategoryCalledWith       []struct{ name, notes string }
}

func (m *mockCategoryAPI) UpdateCategories() error {
	m.updateCategoriesCalled = true
	if m.updateCategoriesFunc != nil {
		return m.updateCategoriesFunc()
	}
	return nil
}

func (m *mockCategoryAPI) UpdateCategoriesInsights() error {
	m.updateCategoriesInsightsCalled = true
	if m.updateCategoriesInsightsFunc != nil {
		return m.updateCategoriesInsightsFunc()
	}
	return nil
}

func (m *mockCategoryAPI) CategoriesList() []firefly.Category {
	if m.categoriesListFunc != nil {
		return m.categoriesListFunc()
	}
	return nil
}

func (m *mockCategoryAPI) GetTotalSpentEarnedCategories() (spent, earned float64) {
	if m.getTotalSpentEarnedFunc != nil {
		return m.getTotalSpentEarnedFunc()
	}
	return 0, 0
}

func (m *mockCategoryAPI) CategorySpent(categoryID string) float64 {
	if m.categorySpentFunc != nil {
		return m.categorySpentFunc(categoryID)
	}
	return 0
}

func (m *mockCategoryAPI) CategoryEarned(categoryID string) float64 {
	if m.categoryEarnedFunc != nil {
		return m.categoryEarnedFunc(categoryID)
	}
	return 0
}

func (m *mockCategoryAPI) CreateCategory(name, notes string) error {
	m.createCategoryCalledWith = append(m.createCategoryCalledWith, struct{ name, notes string }{name, notes})
	if m.createCategoryFunc != nil {
		return m.createCategoryFunc(name, notes)
	}
	return nil
}

func (m *mockCategoryAPI) PrimaryCurrency() firefly.Currency {
	if m.primaryCurrencyFunc != nil {
		return m.primaryCurrencyFunc()
	}
	return firefly.Currency{Code: "USD", Symbol: "$"}
}

func newFocusedCategoriesModelWithCategory(t *testing.T, cat firefly.Category) modelCategories {
	t.Helper()

	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{cat}
		},
		categorySpentFunc:  func(categoryID string) float64 { return 0 },
		categoryEarnedFunc: func(categoryID string) float64 { return 0 },
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}

	m := newModelCategories(api)
	(&m).Focus()
	return m
}

// Basic functionality tests

func TestGetCategoriesItems_UsesCategorySpentAndEarnedAPI(t *testing.T) {
	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{
				{ID: "c1", Name: "Groceries", CurrencyCode: "USD"},
				{ID: "c2", Name: "Salary", CurrencyCode: "EUR"},
			}
		},
		categorySpentFunc: func(categoryID string) float64 {
			switch categoryID {
			case "c1":
				return 500.50
			case "c2":
				return 0
			default:
				return 0
			}
		},
		categoryEarnedFunc: func(categoryID string) float64 {
			switch categoryID {
			case "c1":
				return 0
			case "c2":
				return 3000.75
			default:
				return 0
			}
		},
	}

	items := getCategoriesItems(api, 0)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	first, ok := items[0].(categoryItem)
	if !ok {
		t.Fatalf("expected item type categoryItem, got %T", items[0])
	}
	if first.category.ID != "c1" {
		t.Errorf("expected first category ID 'c1', got %q", first.category.ID)
	}
	if first.spent != 500.50 {
		t.Errorf("expected first category spent 500.50, got %.2f", first.spent)
	}
	if first.earned != 0 {
		t.Errorf("expected first category earned 0, got %.2f", first.earned)
	}

	second, ok := items[1].(categoryItem)
	if !ok {
		t.Fatalf("expected item type categoryItem, got %T", items[1])
	}
	if second.category.ID != "c2" {
		t.Errorf("expected second category ID 'c2', got %q", second.category.ID)
	}
	if second.spent != 0 {
		t.Errorf("expected second category spent 0, got %.2f", second.spent)
	}
	if second.earned != 3000.75 {
		t.Errorf("expected second category earned 3000.75, got %.2f", second.earned)
	}
}

func TestGetCategoriesItems_SortsAndFiltersCorrectly(t *testing.T) {
	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{
				{ID: "c1", Name: "Groceries", CurrencyCode: "USD"},
				{ID: "c2", Name: "Transport", CurrencyCode: "USD"},
				{ID: "c3", Name: "Salary", CurrencyCode: "USD"},
				{ID: "c4", Name: "Investment", CurrencyCode: "USD"},
			}
		},
		categorySpentFunc: func(categoryID string) float64 {
			switch categoryID {
			case "c1":
				return 500.0
			case "c2":
				return 300.0
			case "c3":
				return 0
			case "c4":
				return 100.0
			default:
				return 0
			}
		},
		categoryEarnedFunc: func(categoryID string) float64 {
			switch categoryID {
			case "c1":
				return 0
			case "c2":
				return 50.0
			case "c3":
				return 3000.0
			case "c4":
				return 200.0
			default:
				return 0
			}
		},
	}

	tests := []struct {
		name      string
		sorted    int
		wantCount int
		firstID   string
		lastID    string
	}{
		{
			name:      "no sorting (sorted=0)",
			sorted:    0,
			wantCount: 4,
			firstID:   "c1",
			lastID:    "c4",
		},
		{
			name:      "sort by spent (sorted=-1), filters zero spent",
			sorted:    -1,
			wantCount: 3,
			firstID:   "c1",
			lastID:    "c4",
		},
		{
			name:      "sort by earned (sorted=1), filters zero earned",
			sorted:    1,
			wantCount: 3,
			firstID:   "c3",
			lastID:    "c2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := getCategoriesItems(api, tt.sorted)
			if len(items) != tt.wantCount {
				t.Errorf("expected %d items, got %d", tt.wantCount, len(items))
			}
			if len(items) > 0 {
				first := items[0].(categoryItem)
				if first.category.ID != tt.firstID {
					t.Errorf("expected first category ID %q, got %q", tt.firstID, first.category.ID)
				}
				last := items[len(items)-1].(categoryItem)
				if last.category.ID != tt.lastID {
					t.Errorf("expected last category ID %q, got %q", tt.lastID, last.category.ID)
				}
			}
		})
	}
}

func TestNewModelCategories_InitializesCurrencyCode(t *testing.T) {
	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}}
		},
		categorySpentFunc:  func(categoryID string) float64 { return 0 },
		categoryEarnedFunc: func(categoryID string) float64 { return 0 },
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "EUR", Symbol: "€"}
		},
	}

	_ = newModelCategories(api)

	if totalCategory.CurrencyCode != "EUR" {
		t.Errorf("expected totalCategory.CurrencyCode 'EUR', got %q", totalCategory.CurrencyCode)
	}
}

// Message handler tests

func TestRefreshCategoryInsightsMsg_Success(t *testing.T) {
	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}}
		},
		categorySpentFunc:  func(categoryID string) float64 { return 0 },
		categoryEarnedFunc: func(categoryID string) float64 { return 0 },
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}

	m := newModelCategories(api)
	_, cmd := m.Update(RefreshCategoryInsightsMsg{})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msg := cmd()
	if _, ok := msg.(CategoriesUpdateMsg); !ok {
		t.Errorf("expected CategoriesUpdateMsg, got %T", msg)
	}

	if !api.updateCategoriesInsightsCalled {
		t.Error("expected UpdateCategoriesInsights to be called")
	}
}

func TestRefreshCategoryInsightsMsg_Error(t *testing.T) {
	expectedErr := errors.New("insights API error")
	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}}
		},
		categorySpentFunc:  func(categoryID string) float64 { return 0 },
		categoryEarnedFunc: func(categoryID string) float64 { return 0 },
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
		updateCategoriesInsightsFunc: func() error {
			return expectedErr
		},
	}

	m := newModelCategories(api)
	_, cmd := m.Update(RefreshCategoryInsightsMsg{})

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

func TestRefreshCategoriesMsg_Success(t *testing.T) {
	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}}
		},
		categorySpentFunc:  func(categoryID string) float64 { return 0 },
		categoryEarnedFunc: func(categoryID string) float64 { return 0 },
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}

	m := newModelCategories(api)
	_, cmd := m.Update(RefreshCategoriesMsg{})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msg := cmd()
	if _, ok := msg.(CategoriesUpdateMsg); !ok {
		t.Errorf("expected CategoriesUpdateMsg, got %T", msg)
	}

	if !api.updateCategoriesCalled {
		t.Error("expected UpdateCategories to be called")
	}
}

func TestRefreshCategoriesMsg_Error(t *testing.T) {
	expectedErr := errors.New("categories API error")
	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}}
		},
		categorySpentFunc:  func(categoryID string) float64 { return 0 },
		categoryEarnedFunc: func(categoryID string) float64 { return 0 },
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
		updateCategoriesFunc: func() error {
			return expectedErr
		},
	}

	m := newModelCategories(api)
	_, cmd := m.Update(RefreshCategoriesMsg{})

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

func TestCategoriesUpdateMsg_SetItemsWithTotal(t *testing.T) {
	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}}
		},
		categorySpentFunc:  func(categoryID string) float64 { return 500.0 },
		categoryEarnedFunc: func(categoryID string) float64 { return 100.0 },
		getTotalSpentEarnedFunc: func() (float64, float64) {
			return 1500.0, 300.0
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}

	m := newModelCategories(api)
	updated, cmd := m.Update(CategoriesUpdateMsg{})
	m2 := updated.(modelCategories)

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
			if dlMsg.DataType != "categories" {
				t.Errorf("expected DataType 'categories', got %q", dlMsg.DataType)
			}
			foundDataLoadMsg = true
		}
	}
	if !foundDataLoadMsg {
		t.Error("expected DataLoadCompletedMsg in batch")
	}

	// Verify list items include total + categories
	// Note: The total is inserted via a separate cmd in the batch,
	// so we check the updated model's list after the Update() returns
	listItems := m2.list.Items()
	if len(listItems) != 2 {
		t.Fatalf("expected 2 list items (total + 1 category), got %d", len(listItems))
	}

	// First item should be total
	totalItem, ok := listItems[0].(categoryItem)
	if !ok {
		t.Fatalf("expected first item to be categoryItem, got %T", listItems[0])
	}
	if totalItem.category.Name != "Total" {
		t.Errorf("expected first item to be 'Total', got %q", totalItem.category.Name)
	}
	if totalItem.spent != 1500.0 {
		t.Errorf("expected total spent 1500.0, got %.2f", totalItem.spent)
	}
	if totalItem.earned != 300.0 {
		t.Errorf("expected total earned 300.0, got %.2f", totalItem.earned)
	}
}

func TestNewCategoryMsg_Success(t *testing.T) {
	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}}
		},
		categorySpentFunc:  func(categoryID string) float64 { return 0 },
		categoryEarnedFunc: func(categoryID string) float64 { return 0 },
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}

	m := newModelCategories(api)
	_, cmd := m.Update(NewCategoryMsg{Category: "NewCat"})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	if len(api.createCategoryCalledWith) != 1 {
		t.Fatalf("expected CreateCategory to be called once, got %d", len(api.createCategoryCalledWith))
	}
	if api.createCategoryCalledWith[0].name != "NewCat" {
		t.Errorf("expected category name 'NewCat', got %q", api.createCategoryCalledWith[0].name)
	}
	if api.createCategoryCalledWith[0].notes != "" {
		t.Errorf("expected empty notes, got %q", api.createCategoryCalledWith[0].notes)
	}

	msgs := collectMsgsFromCmd(cmd)
	foundRefresh := false
	foundNotify := false
	for _, msg := range msgs {
		if _, ok := msg.(RefreshCategoriesMsg); ok {
			foundRefresh = true
		}
		if notifyMsg, ok := msg.(notify.NotifyMsg); ok {
			if notifyMsg.Level == notify.Log {
				foundNotify = true
			}
		}
	}
	if !foundRefresh {
		t.Error("expected RefreshCategoriesMsg in batch")
	}
	if !foundNotify {
		t.Error("expected notify.NotifyMsg in batch")
	}
}

func TestNewCategoryMsg_Error(t *testing.T) {
	expectedErr := errors.New("create category error")
	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}}
		},
		categorySpentFunc:  func(categoryID string) float64 { return 0 },
		categoryEarnedFunc: func(categoryID string) float64 { return 0 },
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
		createCategoryFunc: func(name, notes string) error {
			return expectedErr
		},
	}

	m := newModelCategories(api)
	_, cmd := m.Update(NewCategoryMsg{Category: "BadCat"})

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

// Key handling tests

func TestFocusAndBlur(t *testing.T) {
	cat := firefly.Category{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}
	m := newFocusedCategoriesModelWithCategory(t, cat)

	if !m.focus {
		t.Error("expected model to be focused after Focus()")
	}

	(&m).Blur()
	if m.focus {
		t.Error("expected model to be blurred after Blur()")
	}
}

func TestKeyFilter_SendsFilterMsg(t *testing.T) {
	cat := firefly.Category{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}
	m := newFocusedCategoriesModelWithCategory(t, cat)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msg := cmd()
	filterMsg, ok := msg.(FilterMsg)
	if !ok {
		t.Fatalf("expected FilterMsg, got %T", msg)
	}
	if filterMsg.Category.Name != cat.Name {
		t.Errorf("expected category name %q, got %q", cat.Name, filterMsg.Category.Name)
	}
	if filterMsg.Reset {
		t.Error("expected Reset to be false")
	}
}

func TestKeyFilter_TotalCategory_NoAction(t *testing.T) {
	cat := firefly.Category{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}
	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{cat}
		},
		categorySpentFunc:  func(categoryID string) float64 { return 100.0 },
		categoryEarnedFunc: func(categoryID string) float64 { return 50.0 },
		getTotalSpentEarnedFunc: func() (float64, float64) {
			return 100.0, 50.0
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}

	m := newModelCategories(api)
	(&m).Focus()
	updated, _ := m.Update(CategoriesUpdateMsg{})
	m = updated.(modelCategories)

	m.list.Select(0)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})

	if cmd != nil {
		t.Error("expected no command when filtering on Total category")
	}
}

func TestKeyRefresh_SendsRefreshMsg(t *testing.T) {
	cat := firefly.Category{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}
	m := newFocusedCategoriesModelWithCategory(t, cat)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msg := cmd()
	if _, ok := msg.(RefreshCategoriesMsg); !ok {
		t.Errorf("expected RefreshCategoriesMsg, got %T", msg)
	}
}

func TestKeySort_CyclesThroughStates(t *testing.T) {
	cat := firefly.Category{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}
	m := newFocusedCategoriesModelWithCategory(t, cat)

	if m.sorted != 0 {
		t.Errorf("expected initial sorted state 0, got %d", m.sorted)
	}

	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = newM.(modelCategories)

	if m.sorted != -1 {
		t.Errorf("expected sorted state -1 after first press, got %d", m.sorted)
	}
	if cmd == nil {
		t.Fatal("expected command after sort key press")
	}
	msg := cmd()
	if _, ok := msg.(CategoriesUpdateMsg); !ok {
		t.Errorf("expected CategoriesUpdateMsg, got %T", msg)
	}

	newM, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = newM.(modelCategories)

	if m.sorted != 1 {
		t.Errorf("expected sorted state 1 after second press, got %d", m.sorted)
	}
	if cmd == nil {
		t.Fatal("expected command after sort key press")
	}
	msg = cmd()
	if _, ok := msg.(CategoriesUpdateMsg); !ok {
		t.Errorf("expected CategoriesUpdateMsg, got %T", msg)
	}

	newM, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = newM.(modelCategories)

	if m.sorted != 0 {
		t.Errorf("expected sorted state 0 after third press, got %d", m.sorted)
	}
	if cmd == nil {
		t.Fatal("expected command after sort key press")
	}
	msg = cmd()
	if _, ok := msg.(CategoriesUpdateMsg); !ok {
		t.Errorf("expected CategoriesUpdateMsg, got %T", msg)
	}
}

func TestKeyResetFilter_SendsResetFilterMsg(t *testing.T) {
	cat := firefly.Category{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}
	m := newFocusedCategoriesModelWithCategory(t, cat)

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

func TestKeyNew_EmitsPrompt(t *testing.T) {
	cat := firefly.Category{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}
	m := newFocusedCategoriesModelWithCategory(t, cat)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msg := cmd()
	if _, ok := msg.(prompt.PromptMsg); !ok {
		t.Errorf("expected prompt.PromptMsg, got %T", msg)
	}
}

func TestView_RendersListView(t *testing.T) {
	lipgloss.SetColorProfile(0)
	cat := firefly.Category{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}
	m := newFocusedCategoriesModelWithCategory(t, cat)

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
}

// View navigation tests

func TestKeyPresses_NavigateToCorrectViews(t *testing.T) {
	cat := firefly.Category{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}

	tests := []struct {
		name         string
		key          rune
		expectedView state
		disabled     bool
		expectedMsgs int
	}{
		{"assets", 'a', assetsView, false, 1},
		{"categories (self)", 'c', categoriesView, true, 0},
		{"expenses", 'e', expensesView, false, 1},
		{"transactions", 't', transactionsView, false, 1},
		{"liabilities", 'o', liabilitiesView, false, 1},
		{"revenues", 'i', revenuesView, false, 1},
		{"quit to transactions", 'q', transactionsView, false, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newFocusedCategoriesModelWithCategory(t, cat)
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

func TestCmdPromptNewCategory_EmitsPrompt(t *testing.T) {
	backCmd := Cmd(SetFocusedViewMsg{state: categoriesView})
	cmd := CmdPromptNewCategory(backCmd)

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	msg := cmd()
	askMsg, ok := msg.(prompt.PromptMsg)
	if !ok {
		t.Fatalf("expected prompt.PromptMsg, got %T", msg)
	}
	if askMsg.Prompt != "New Category(<name>): " {
		t.Errorf("expected prompt 'New Category(<name>): ', got %q", askMsg.Prompt)
	}
}

func TestCmdPromptNewCategory_ValidInput(t *testing.T) {
	backCmdCalled := false
	backCmd := func() tea.Msg {
		backCmdCalled = true
		return nil
	}

	cmd := CmdPromptNewCategory(backCmd)
	askMsg := cmd().(prompt.PromptMsg)

	resultCmd := askMsg.Callback("NewCategory")
	if resultCmd == nil {
		t.Fatal("expected a command from callback, got nil")
	}

	msgs := collectMsgsFromCmd(resultCmd)
	if len(msgs) < 1 {
		t.Fatal("expected at least one message from callback")
	}

	newCatMsg, ok := msgs[0].(NewCategoryMsg)
	if !ok {
		t.Fatalf("expected NewCategoryMsg, got %T", msgs[0])
	}
	if newCatMsg.Category != "NewCategory" {
		t.Errorf("expected category 'NewCategory', got %q", newCatMsg.Category)
	}

	if !backCmdCalled {
		t.Error("expected back command to be called")
	}
}

func TestCmdPromptNewCategory_CancelWithNone(t *testing.T) {
	backCmdCalled := false
	backCmd := func() tea.Msg {
		backCmdCalled = true
		return nil
	}

	cmd := CmdPromptNewCategory(backCmd)
	askMsg := cmd().(prompt.PromptMsg)

	resultCmd := askMsg.Callback("None")
	if resultCmd == nil {
		t.Fatal("expected a command from callback, got nil")
	}

	msgs := collectMsgsFromCmd(resultCmd)

	for _, msg := range msgs {
		if _, ok := msg.(NewCategoryMsg); ok {
			t.Error("expected no NewCategoryMsg when input is 'None'")
		}
	}

	if !backCmdCalled {
		t.Error("expected back command to be called even when canceled")
	}
}

// Edge case tests

func TestCategoryItem_Description(t *testing.T) {
	tests := []struct {
		name     string
		item     categoryItem
		wantDesc string
	}{
		{
			name: "both spent and earned",
			item: categoryItem{
				category: firefly.Category{Name: "Mixed", CurrencyCode: "USD"},
				spent:    100.50,
				earned:   200.75,
			},
			wantDesc: "Spent: 100.50 USD | Earned: 200.75 USD",
		},
		{
			name: "only spent",
			item: categoryItem{
				category: firefly.Category{Name: "Expense", CurrencyCode: "EUR"},
				spent:    50.00,
				earned:   0,
			},
			wantDesc: "Spent: 50.00 EUR",
		},
		{
			name: "only earned",
			item: categoryItem{
				category: firefly.Category{Name: "Income", CurrencyCode: "GBP"},
				spent:    0,
				earned:   1000.00,
			},
			wantDesc: "Earned: 1000.00 GBP",
		},
		{
			name: "no transactions",
			item: categoryItem{
				category: firefly.Category{Name: "Empty", CurrencyCode: "USD"},
				spent:    0,
				earned:   0,
			},
			wantDesc: "No transactions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := tt.item.Description()
			if desc != tt.wantDesc {
				t.Errorf("expected description %q, got %q", tt.wantDesc, desc)
			}
		})
	}
}

func TestCategoryItem_TitleAndFilterValue(t *testing.T) {
	cat := firefly.Category{ID: "c1", Name: "TestCategory", CurrencyCode: "USD"}
	item := categoryItem{
		category: cat,
		spent:    100.0,
		earned:   50.0,
	}

	if item.Title() != cat.Name {
		t.Errorf("expected title %q, got %q", cat.Name, item.Title())
	}
	if item.FilterValue() != cat.Name {
		t.Errorf("expected filter value %q, got %q", cat.Name, item.FilterValue())
	}
}

func TestModelCategories_UnfocusedIgnoresKeys(t *testing.T) {
	cat := firefly.Category{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}
	m := newFocusedCategoriesModelWithCategory(t, cat)
	(&m).Blur()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if cmd != nil {
		t.Error("expected no command when unfocused")
	}
}

func TestGetCategoriesItems_EmptyList(t *testing.T) {
	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{}
		},
		categorySpentFunc:  func(categoryID string) float64 { return 0 },
		categoryEarnedFunc: func(categoryID string) float64 { return 0 },
	}

	items := getCategoriesItems(api, 0)
	if len(items) != 0 {
		t.Errorf("expected 0 items for empty list, got %d", len(items))
	}
}

func TestGetCategoriesItems_NilAPI(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when calling getCategoriesItems with nil API")
		}
	}()

	getCategoriesItems(nil, 0)
}

func TestModelCategories_LargeValues(t *testing.T) {
	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{{ID: "c1", Name: "BigCategory", CurrencyCode: "USD"}}
		},
		categorySpentFunc:  func(categoryID string) float64 { return 999999999.99 },
		categoryEarnedFunc: func(categoryID string) float64 { return 888888888.88 },
		getTotalSpentEarnedFunc: func() (float64, float64) {
			return 999999999.99, 888888888.88
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}

	m := newModelCategories(api)
	_, cmd := m.Update(CategoriesUpdateMsg{})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	items := m.list.Items()
	if len(items) < 1 {
		t.Fatal("expected at least one item")
	}

	totalItem := items[0].(categoryItem)
	if totalItem.spent != 999999999.99 {
		t.Errorf("expected large spent value 999999999.99, got %.2f", totalItem.spent)
	}
	if totalItem.earned != 888888888.88 {
		t.Errorf("expected large earned value 888888888.88, got %.2f", totalItem.earned)
	}
}

func TestModelCategories_NegativeValues(t *testing.T) {
	api := &mockCategoryAPI{
		categoriesListFunc: func() []firefly.Category {
			return []firefly.Category{{ID: "c1", Name: "NegCategory", CurrencyCode: "USD"}}
		},
		categorySpentFunc:  func(categoryID string) float64 { return -100.0 },
		categoryEarnedFunc: func(categoryID string) float64 { return -50.0 },
		getTotalSpentEarnedFunc: func() (float64, float64) {
			return -100.0, -50.0
		},
		primaryCurrencyFunc: func() firefly.Currency {
			return firefly.Currency{Code: "USD", Symbol: "$"}
		},
	}

	m := newModelCategories(api)
	_, cmd := m.Update(CategoriesUpdateMsg{})

	if cmd == nil {
		t.Fatal("expected a command, got nil")
	}

	items := m.list.Items()
	if len(items) < 1 {
		t.Fatal("expected at least one item")
	}

	totalItem := items[0].(categoryItem)
	if totalItem.spent != -100.0 {
		t.Errorf("expected negative spent value -100.0, got %.2f", totalItem.spent)
	}
	if totalItem.earned != -50.0 {
		t.Errorf("expected negative earned value -50.0, got %.2f", totalItem.earned)
	}
}

func TestModelCategories_SpecialCharactersInCategoryName(t *testing.T) {
	specialNames := []string{
		"Café & Restaurant",
		"日本語",
		"Category with emoji 🍕",
		"<script>alert('xss')</script>",
		"Category\nWith\nNewlines",
		"",
	}

	for _, name := range specialNames {
		t.Run(name, func(t *testing.T) {
			api := &mockCategoryAPI{
				categoriesListFunc: func() []firefly.Category {
					return []firefly.Category{{ID: "c1", Name: name, CurrencyCode: "USD"}}
				},
				categorySpentFunc:  func(categoryID string) float64 { return 0 },
				categoryEarnedFunc: func(categoryID string) float64 { return 0 },
				primaryCurrencyFunc: func() firefly.Currency {
					return firefly.Currency{Code: "USD", Symbol: "$"}
				},
			}

			m := newModelCategories(api)
			items := m.list.Items()

			if len(items) != 1 {
				t.Errorf("expected 1 item, got %d", len(items))
			}

			item := items[0].(categoryItem)
			if item.Title() != name {
				t.Errorf("expected title %q, got %q", name, item.Title())
			}
		})
	}
}

func TestModelCategories_UpdatePositions(t *testing.T) {
	cat := firefly.Category{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}
	m := newFocusedCategoriesModelWithCategory(t, cat)

	globalWidth := 100
	globalHeight := 50
	topSize := 5

	updated, _ := m.Update(UpdatePositions{
		layout: &LayoutConfig{
			Width:   globalWidth,
			Height:  globalHeight,
			TopSize: topSize,
		},
	})
	m2 := updated.(modelCategories)

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

func TestModelCategories_FocusToggle(t *testing.T) {
	cat := firefly.Category{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}
	m := newFocusedCategoriesModelWithCategory(t, cat)

	if !m.focus {
		t.Error("expected model to be focused initially")
	}

	(&m).Blur()
	if m.focus {
		t.Error("expected model to be blurred after Blur()")
	}

	(&m).Focus()
	if !m.focus {
		t.Error("expected model to be focused after Focus()")
	}
}

func TestModelCategories_SmallDimensions(t *testing.T) {
	cat := firefly.Category{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}
	m := newFocusedCategoriesModelWithCategory(t, cat)

	globalWidth := 10
	globalHeight := 5
	topSize := 2

	updated, _ := m.Update(UpdatePositions{
		layout: &LayoutConfig{
			Width:   globalWidth,
			Height:  globalHeight,
			TopSize: topSize,
		},
	})
	m2 := updated.(modelCategories)

	w, h := m2.list.Width(), m2.list.Height()
	if w < 0 || h < 0 {
		t.Error("expected non-negative list dimensions even with small screen")
	}
}

func TestModelCategories_LargeDimensions(t *testing.T) {
	cat := firefly.Category{ID: "c1", Name: "Groceries", CurrencyCode: "USD"}
	m := newFocusedCategoriesModelWithCategory(t, cat)

	globalWidth := 1000
	globalHeight := 1000
	topSize := 10

	updated, _ := m.Update(UpdatePositions{
		layout: &LayoutConfig{
			Width:   globalWidth,
			Height:  globalHeight,
			TopSize: topSize,
		},
	})
	m2 := updated.(modelCategories)

	w, h := m2.list.Width(), m2.list.Height()
	if w <= 0 || h <= 0 {
		t.Error("expected positive list dimensions with large screen")
	}
}
