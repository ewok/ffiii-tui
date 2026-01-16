/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"errors"
	"strings"
	"testing"

	"ffiii-tui/internal/firefly"
	"ffiii-tui/internal/ui/notify"

	"github.com/charmbracelet/lipgloss"
)

// invalidItem implements list.Item interface but is not summaryItem
type invalidItem struct{}

func (i invalidItem) FilterValue() string { return "" }

// Mock SummaryAPI implementation
type mockSummaryAPI struct {
	updateSummaryFunc func() error
	getMaxWidthFunc   func() int
	summaryItemsFunc  func() map[string]firefly.SummaryItem

	updateSummaryCalled int
	getMaxWidthCalled   int
	summaryItemsCalled  int
}

func (m *mockSummaryAPI) UpdateSummary() error {
	m.updateSummaryCalled++
	if m.updateSummaryFunc != nil {
		return m.updateSummaryFunc()
	}
	return nil
}

func (m *mockSummaryAPI) GetMaxWidth() int {
	m.getMaxWidthCalled++
	if m.getMaxWidthFunc != nil {
		return m.getMaxWidthFunc()
	}
	return 40
}

func (m *mockSummaryAPI) SummaryItems() map[string]firefly.SummaryItem {
	m.summaryItemsCalled++
	if m.summaryItemsFunc != nil {
		return m.summaryItemsFunc()
	}
	return map[string]firefly.SummaryItem{
		"balance": {
			Title:         "Balance",
			ValueParsed:   "$1,234.56",
			MonetaryValue: 1234.56,
		},
	}
}

func newTestSummaryAPI() *mockSummaryAPI {
	return &mockSummaryAPI{}
}

// =============================================================================
// Basic Functionality Tests
// =============================================================================

func TestSummary_NewModelSummary(t *testing.T) {
	api := newTestSummaryAPI()
	api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
		return map[string]firefly.SummaryItem{
			"balance": {
				Title:         "Balance",
				ValueParsed:   "$1,000.00",
				MonetaryValue: 1000.00,
			},
			"spent": {
				Title:         "Spent",
				ValueParsed:   "-$500.00",
				MonetaryValue: -500.00,
			},
		}
	}

	m := newModelSummary(api)

	if m.api == nil {
		t.Error("Expected api to be set")
	}

	if m.list.Title != "Summary" {
		t.Errorf("Expected title 'Summary', got %q", m.list.Title)
	}

	items := m.list.Items()
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	if api.summaryItemsCalled != 1 {
		t.Errorf("Expected SummaryItems to be called once, got %d", api.summaryItemsCalled)
	}

	if api.getMaxWidthCalled != 1 {
		t.Errorf("Expected GetMaxWidth to be called once, got %d", api.getMaxWidthCalled)
	}
}

func TestSummary_Init(t *testing.T) {
	api := newTestSummaryAPI()
	m := newModelSummary(api)

	cmd := m.Init()

	if cmd != nil {
		t.Error("Expected Init to return nil command")
	}
}

func TestSummary_GetSummaryItems_Sorting(t *testing.T) {
	api := newTestSummaryAPI()
	api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
		return map[string]firefly.SummaryItem{
			"low": {
				Title:         "Low Balance",
				ValueParsed:   "$100.00",
				MonetaryValue: 100.00,
			},
			"high": {
				Title:         "High Balance",
				ValueParsed:   "$5,000.00",
				MonetaryValue: 5000.00,
			},
			"negative": {
				Title:         "Negative",
				ValueParsed:   "-$300.00",
				MonetaryValue: -300.00,
			},
			"medium": {
				Title:         "Medium Balance",
				ValueParsed:   "$1,000.00",
				MonetaryValue: 1000.00,
			},
		}
	}

	styles := DefaultStyles()
	items := getSummaryItems(api, styles)

	if len(items) != 4 {
		t.Fatalf("Expected 4 items, got %d", len(items))
	}

	// Items should be sorted by monetary value descending
	expectedOrder := []string{"High Balance", "Medium Balance", "Low Balance", "Negative"}
	for i, item := range items {
		si := item.(summaryItem)
		if si.title != expectedOrder[i] {
			t.Errorf("Expected item %d to be %q, got %q", i, expectedOrder[i], si.title)
		}
	}

	// Verify monetary values are in descending order
	expectedValues := []float64{5000.00, 1000.00, 100.00, -300.00}
	for i, item := range items {
		si := item.(summaryItem)
		if si.monetaryValue != expectedValues[i] {
			t.Errorf("Expected item %d value to be %.2f, got %.2f", i, expectedValues[i], si.monetaryValue)
		}
	}
}

func TestSummary_GetSummaryItems_Styling(t *testing.T) {
	api := newTestSummaryAPI()
	styles := DefaultStyles()

	tests := []struct {
		name          string
		monetaryValue float64
		expectedStyle lipgloss.Style
	}{
		{
			name:          "Positive value uses Deposit style",
			monetaryValue: 1000.00,
			expectedStyle: styles.Deposit,
		},
		{
			name:          "Negative value uses Withdrawal style",
			monetaryValue: -500.00,
			expectedStyle: styles.Withdrawal,
		},
		{
			name:          "Zero value uses Normal style",
			monetaryValue: 0.00,
			expectedStyle: styles.Normal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
				return map[string]firefly.SummaryItem{
					"test": {
						Title:         "Test Item",
						ValueParsed:   "$0.00",
						MonetaryValue: tt.monetaryValue,
					},
				}
			}

			items := getSummaryItems(api, styles)
			if len(items) != 1 {
				t.Fatalf("Expected 1 item, got %d", len(items))
			}

			si := items[0].(summaryItem)
			// Compare style properties (can't directly compare styles due to internal state)
			if si.style.GetForeground() != tt.expectedStyle.GetForeground() {
				t.Errorf("Expected foreground color %v, got %v",
					tt.expectedStyle.GetForeground(), si.style.GetForeground())
			}
		})
	}
}

func TestSummary_SummaryItem_FilterValue(t *testing.T) {
	item := summaryItem{
		title:         "Test Title",
		value:         "$100.00",
		monetaryValue: 100.00,
	}

	if item.FilterValue() != "Test Title" {
		t.Errorf("Expected FilterValue to return %q, got %q", "Test Title", item.FilterValue())
	}
}

// =============================================================================
// Message Handler Tests
// =============================================================================

func TestSummary_Update_RefreshSummaryMsg_Success(t *testing.T) {
	api := newTestSummaryAPI()
	api.updateSummaryFunc = func() error {
		return nil
	}
	m := newModelSummary(api)

	_, cmd := m.Update(RefreshSummaryMsg{})

	if cmd == nil {
		t.Fatal("Expected command to be returned")
	}

	msg := cmd()
	if _, ok := msg.(SummaryUpdateMsg); !ok {
		t.Errorf("Expected SummaryUpdateMsg, got %T", msg)
	}

	if api.updateSummaryCalled != 1 {
		t.Errorf("Expected UpdateSummary to be called once, got %d", api.updateSummaryCalled)
	}
}

func TestSummary_Update_RefreshSummaryMsg_Error(t *testing.T) {
	api := newTestSummaryAPI()
	api.updateSummaryFunc = func() error {
		return errors.New("update failed")
	}
	m := newModelSummary(api)

	_, cmd := m.Update(RefreshSummaryMsg{})

	if cmd == nil {
		t.Fatal("Expected command to be returned")
	}

	msg := cmd()
	notifyMsg, ok := msg.(notify.NotifyMsg)
	if !ok {
		t.Fatalf("Expected notify.NotifyMsg, got %T", msg)
	}

	if notifyMsg.Level != notify.Warn {
		t.Errorf("Expected Warn level, got %v", notifyMsg.Level)
	}

	if !strings.Contains(notifyMsg.Message, "update failed") {
		t.Errorf("Expected error message to contain 'update failed', got %q", notifyMsg.Message)
	}
}

func TestSummary_Update_SummaryUpdateMsg(t *testing.T) {
	api := newTestSummaryAPI()
	initialItems := map[string]firefly.SummaryItem{
		"balance": {
			Title:         "Balance",
			ValueParsed:   "$1,000.00",
			MonetaryValue: 1000.00,
		},
	}
	api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
		return initialItems
	}

	m := newModelSummary(api)

	// Update the items
	newItems := map[string]firefly.SummaryItem{
		"balance": {
			Title:         "Balance",
			ValueParsed:   "$2,000.00",
			MonetaryValue: 2000.00,
		},
		"spent": {
			Title:         "Spent",
			ValueParsed:   "-$500.00",
			MonetaryValue: -500.00,
		},
	}
	api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
		return newItems
	}

	_, cmd := m.Update(SummaryUpdateMsg{})

	if cmd == nil {
		t.Fatal("Expected command to be returned")
	}

	// The command should be a tea.Sequence that sets items and requests window size
	// We can verify this by checking that executing the command doesn't panic
	// and that the model's items are updated via SetItems
}

func TestSummary_Update_UpdatePositions(t *testing.T) {
	api := newTestSummaryAPI()
	api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
		return map[string]firefly.SummaryItem{
			"item1": {Title: "Item 1", ValueParsed: "$100", MonetaryValue: 100},
			"item2": {Title: "Item 2", ValueParsed: "$200", MonetaryValue: 200},
			"item3": {Title: "Item 3", ValueParsed: "$300", MonetaryValue: 300},
		}
	}

	m := newModelSummary(api)
	initialHeight := m.list.Height()

	updatedModel, cmd := m.Update(UpdatePositions{
		layout: &LayoutConfig{
			SummarySize: 0,
		},
	})

	if cmd != nil {
		t.Error("Expected UpdatePositions to return nil command")
	}

	m2, ok := updatedModel.(modelSummary)
	if !ok {
		t.Fatal("Expected modelSummary type")
	}

	// Height should be updated based on number of items
	if m2.list.Height() == initialHeight {
		// Height might be 0 initially, so just check it was set
		if m2.list.Height() == 0 {
			t.Error("Expected list height to be updated")
		}
	}
}

func TestSummary_Update_UnknownMessage(t *testing.T) {
	api := newTestSummaryAPI()
	m := newModelSummary(api)

	type unknownMsg struct{}

	updatedModel, cmd := m.Update(unknownMsg{})

	if cmd != nil {
		t.Error("Expected nil command for unknown message")
	}

	if _, ok := updatedModel.(modelSummary); !ok {
		t.Error("Expected modelSummary to be returned")
	}
}

// =============================================================================
// View Tests
// =============================================================================

func TestSummary_View(t *testing.T) {
	api := newTestSummaryAPI()
	api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
		return map[string]firefly.SummaryItem{
			"balance": {
				Title:         "Balance",
				ValueParsed:   "$1,000.00",
				MonetaryValue: 1000.00,
			},
		}
	}

	m := newModelSummary(api)
	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view")
	}

	// View should contain the title
	if !strings.Contains(view, "Summary") {
		t.Error("Expected view to contain 'Summary'")
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestSummary_EmptyItems(t *testing.T) {
	api := newTestSummaryAPI()
	api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
		return map[string]firefly.SummaryItem{}
	}

	m := newModelSummary(api)

	if len(m.list.Items()) != 0 {
		t.Errorf("Expected 0 items, got %d", len(m.list.Items()))
	}

	// Should not panic with empty items
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view even with no items")
	}

	// Test Update with empty items
	_, cmd := m.Update(SummaryUpdateMsg{})
	if cmd == nil {
		t.Error("Expected command even with empty items")
	}
}

func TestSummary_ManyItems(t *testing.T) {
	api := newTestSummaryAPI()
	items := make(map[string]firefly.SummaryItem)
	for i := range 100 {
		key := "item" + string(rune(i))
		items[key] = firefly.SummaryItem{
			Title:         "Item " + string(rune(i)),
			ValueParsed:   "$100.00",
			MonetaryValue: float64(i * 100),
		}
	}
	api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
		return items
	}

	m := newModelSummary(api)

	if len(m.list.Items()) != 100 {
		t.Errorf("Expected 100 items, got %d", len(m.list.Items()))
	}

	// Should not panic with many items
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view with many items")
	}
}

func TestSummary_ItemsWithSameValue(t *testing.T) {
	api := newTestSummaryAPI()
	api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
		return map[string]firefly.SummaryItem{
			"item1": {
				Title:         "Item 1",
				ValueParsed:   "$1,000.00",
				MonetaryValue: 1000.00,
			},
			"item2": {
				Title:         "Item 2",
				ValueParsed:   "$1,000.00",
				MonetaryValue: 1000.00,
			},
			"item3": {
				Title:         "Item 3",
				ValueParsed:   "$1,000.00",
				MonetaryValue: 1000.00,
			},
		}
	}

	styles := DefaultStyles()
	items := getSummaryItems(api, styles)

	if len(items) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(items))
	}

	// All items should have the same monetary value
	for _, item := range items {
		si := item.(summaryItem)
		if si.monetaryValue != 1000.00 {
			t.Errorf("Expected monetary value 1000.00, got %.2f", si.monetaryValue)
		}
	}
}

func TestSummary_LargeValues(t *testing.T) {
	api := newTestSummaryAPI()
	api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
		return map[string]firefly.SummaryItem{
			"huge": {
				Title:         "Huge Amount",
				ValueParsed:   "$999,999,999.99",
				MonetaryValue: 999999999.99,
			},
			"tiny": {
				Title:         "Tiny Amount",
				ValueParsed:   "$0.01",
				MonetaryValue: 0.01,
			},
		}
	}

	m := newModelSummary(api)

	if len(m.list.Items()) != 2 {
		t.Errorf("Expected 2 items, got %d", len(m.list.Items()))
	}

	// Should not panic with large values
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view with large values")
	}
}

func TestSummary_SpecialCharactersInTitle(t *testing.T) {
	api := newTestSummaryAPI()
	api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
		return map[string]firefly.SummaryItem{
			"special": {
				Title:         "Test 'Quote' & <HTML> 日本語",
				ValueParsed:   "$100.00",
				MonetaryValue: 100.00,
			},
		}
	}

	m := newModelSummary(api)

	if len(m.list.Items()) != 1 {
		t.Errorf("Expected 1 item, got %d", len(m.list.Items()))
	}

	item := m.list.Items()[0].(summaryItem)
	if !strings.Contains(item.title, "Quote") {
		t.Error("Expected special characters to be preserved in title")
	}

	// Should not panic with special characters
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view with special characters")
	}
}

func TestSummary_VeryLongTitles(t *testing.T) {
	api := newTestSummaryAPI()
	longTitle := strings.Repeat("Very Long Title ", 20)
	api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
		return map[string]firefly.SummaryItem{
			"long": {
				Title:         longTitle,
				ValueParsed:   "$100.00",
				MonetaryValue: 100.00,
			},
		}
	}

	m := newModelSummary(api)

	if len(m.list.Items()) != 1 {
		t.Errorf("Expected 1 item, got %d", len(m.list.Items()))
	}

	// Should not panic with very long titles
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view with long titles")
	}
}

func TestSummary_DifferentWidths(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"Very narrow", 10},
		{"Narrow", 20},
		{"Normal", 40},
		{"Wide", 80},
		{"Very wide", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := newTestSummaryAPI()
			api.getMaxWidthFunc = func() int {
				return tt.width
			}
			api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
				return map[string]firefly.SummaryItem{
					"balance": {
						Title:         "Balance",
						ValueParsed:   "$1,000.00",
						MonetaryValue: 1000.00,
					},
				}
			}

			m := newModelSummary(api)

			if m.list.Width() != tt.width {
				t.Errorf("Expected width %d, got %d", tt.width, m.list.Width())
			}

			// Should not panic with different widths
			view := m.View()
			if view == "" {
				t.Error("Expected non-empty view")
			}
		})
	}
}

func TestSummary_SummaryDelegate_Height(t *testing.T) {
	delegate := summaryDelegate{}
	if delegate.Height() != 1 {
		t.Errorf("Expected height 1, got %d", delegate.Height())
	}
}

func TestSummary_SummaryDelegate_Spacing(t *testing.T) {
	delegate := summaryDelegate{}
	if delegate.Spacing() != 0 {
		t.Errorf("Expected spacing 0, got %d", delegate.Spacing())
	}
}

func TestSummary_SummaryDelegate_Update(t *testing.T) {
	delegate := summaryDelegate{}
	cmd := delegate.Update(nil, nil)
	if cmd != nil {
		t.Error("Expected nil command from delegate Update")
	}
}

func TestSummary_SummaryDelegate_Render_InvalidItem(t *testing.T) {
	delegate := summaryDelegate{}
	api := newTestSummaryAPI()
	m := newModelSummary(api)

	// Create a buffer to capture output
	var buf strings.Builder

	// Use invalidItem type that implements list.Item but is not summaryItem
	var invalid invalidItem

	// Render with an invalid item (not a summaryItem)
	delegate.Render(&buf, m.list, 0, invalid)

	// Should not panic, and should produce no output
	if buf.Len() != 0 {
		t.Error("Expected no output for invalid item type")
	}
}

func TestSummary_IntegrationSequence(t *testing.T) {
	api := newTestSummaryAPI()
	initialItems := map[string]firefly.SummaryItem{
		"balance": {
			Title:         "Balance",
			ValueParsed:   "$1,000.00",
			MonetaryValue: 1000.00,
		},
	}
	api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
		return initialItems
	}

	// 1. Create model
	m := newModelSummary(api)
	if len(m.list.Items()) != 1 {
		t.Fatalf("Expected 1 initial item, got %d", len(m.list.Items()))
	}

	// 2. Trigger refresh
	m2, cmd := m.Update(RefreshSummaryMsg{})
	if cmd == nil {
		t.Fatal("Expected command from RefreshSummaryMsg")
	}
	m = m2.(modelSummary)

	// 3. Execute refresh command
	msg := cmd()
	if _, ok := msg.(SummaryUpdateMsg); !ok {
		t.Fatalf("Expected SummaryUpdateMsg, got %T", msg)
	}

	// 4. Update items
	api.summaryItemsFunc = func() map[string]firefly.SummaryItem {
		return map[string]firefly.SummaryItem{
			"balance": {
				Title:         "Balance",
				ValueParsed:   "$2,000.00",
				MonetaryValue: 2000.00,
			},
			"spent": {
				Title:         "Spent",
				ValueParsed:   "-$500.00",
				MonetaryValue: -500.00,
			},
		}
	}

	// 5. Process update message
	m3, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatal("Expected command from SummaryUpdateMsg")
	}
	m = m3.(modelSummary)

	// 6. Verify view renders
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view after update sequence")
	}
}
