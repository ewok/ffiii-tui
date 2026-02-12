/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package period

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestNew(t *testing.T) {
	m := New()

	if m.Width != 80 {
		t.Errorf("Expected default width 80, got %d", m.Width)
	}
	if m.focus {
		t.Error("Expected new model to be unfocused")
	}
	if len(m.items) != 0 {
		t.Errorf("Expected empty items, got %d", len(m.items))
	}
}

func TestGenerateItems(t *testing.T) {
	items := generateItems(2026, time.February)

	expectedStart := monthEntry{year: 2021, month: time.January}
	expectedEnd := monthEntry{year: 2031, month: time.December}

	if len(items) == 0 {
		t.Fatal("Expected non-empty items")
	}
	if items[0] != expectedStart {
		t.Errorf("Expected first item %v, got %v", expectedStart, items[0])
	}
	if items[len(items)-1] != expectedEnd {
		t.Errorf("Expected last item %v, got %v", expectedEnd, items[len(items)-1])
	}

	expectedCount := (yearsRange*2 + 1) * 12
	if len(items) != expectedCount {
		t.Errorf("Expected %d items, got %d", expectedCount, len(items))
	}

	for i := 1; i < len(items); i++ {
		prev := time.Date(items[i-1].year, items[i-1].month, 1, 0, 0, 0, 0, time.UTC)
		curr := time.Date(items[i].year, items[i].month, 1, 0, 0, 0, 0, time.UTC)
		if !curr.After(prev) {
			t.Errorf("Items not in order at index %d: %v >= %v", i, prev, curr)
		}
	}
}

func TestFindIndex(t *testing.T) {
	items := generateItems(2026, time.February)

	idx := findIndex(items, 2026, time.February)
	if items[idx].year != 2026 || items[idx].month != time.February {
		t.Errorf("Expected 2026-Feb at index %d, got %v", idx, items[idx])
	}

	idx = findIndex(items, 2021, time.January)
	if idx != 0 {
		t.Errorf("Expected index 0 for first item, got %d", idx)
	}

	idx = findIndex(items, 1900, time.January)
	if idx != 0 {
		t.Errorf("Expected index 0 for non-existing item, got %d", idx)
	}
}

func TestMonthEntryLabel(t *testing.T) {
	e := monthEntry{year: 2026, month: time.March}
	label := e.label()

	if label != "March 2026" {
		t.Errorf("Expected 'March 2026', got '%s'", label)
	}
}

func TestInit(t *testing.T) {
	m := New()
	cmd := m.Init()
	if cmd != nil {
		t.Error("Expected Init to return nil")
	}
}

func TestUpdate_OpenMsg(t *testing.T) {
	m := New()

	updated, cmd := m.Update(OpenMsg{Year: 2026, Month: time.March})
	m = updated.(Model)

	if cmd != nil {
		t.Error("Expected nil command from OpenMsg")
	}
	if !m.focus {
		t.Error("Expected model to be focused after OpenMsg")
	}
	if len(m.items) == 0 {
		t.Error("Expected items to be generated")
	}
	if m.cursor != m.current {
		t.Error("Expected cursor and current to be the same")
	}
	if m.items[m.cursor].year != 2026 || m.items[m.cursor].month != time.March {
		t.Errorf("Expected cursor at 2026-March, got %v", m.items[m.cursor])
	}
}

func TestUpdate_CloseMsg(t *testing.T) {
	m := New()
	m.Focus()

	updated, cmd := m.Update(CloseMsg{})
	m = updated.(Model)

	if cmd != nil {
		t.Error("Expected nil command from CloseMsg")
	}
	if m.focus {
		t.Error("Expected model to be blurred after CloseMsg")
	}
}

func TestUpdate_KeyNavigation(t *testing.T) {
	m := New()
	updated, _ := m.Update(OpenMsg{Year: 2026, Month: time.June})
	m = updated.(Model)

	startCursor := m.cursor

	tests := []struct {
		name     string
		key      string
		expected int
	}{
		{"right", "right", startCursor + 1},
		{"l", "l", startCursor + 2},
		{"down", "down", startCursor + 3},
		{"j", "j", startCursor + 4},
		{"left", "left", startCursor + 3},
		{"h", "h", startCursor + 2},
		{"up", "up", startCursor + 1},
		{"k", "k", startCursor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
			m = updated.(Model)

			if m.cursor != tt.expected {
				t.Errorf("After %s: expected cursor %d, got %d", tt.name, tt.expected, m.cursor)
			}
		})
	}
}

func TestUpdate_KeyNavigationBounds(t *testing.T) {
	m := New()
	updated, _ := m.Update(OpenMsg{Year: 2026, Month: time.June})
	m = updated.(Model)

	m.cursor = 0
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	m = updated.(Model)
	if m.cursor != 0 {
		t.Errorf("Expected cursor to stay at 0, got %d", m.cursor)
	}

	m.cursor = len(m.items) - 1
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m = updated.(Model)
	if m.cursor != len(m.items)-1 {
		t.Errorf("Expected cursor to stay at %d, got %d", len(m.items)-1, m.cursor)
	}
}

func TestUpdate_EnterKey(t *testing.T) {
	m := New()
	updated, _ := m.Update(OpenMsg{Year: 2026, Month: time.March})
	m = updated.(Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m = updated.(Model)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.focus {
		t.Error("Expected model to be blurred after Enter")
	}
	if cmd == nil {
		t.Fatal("Expected command from Enter key")
	}

	msg := cmd()
	selected, ok := msg.(SelectedMsg)
	if !ok {
		t.Fatalf("Expected SelectedMsg, got %T", msg)
	}
	if selected.Year != 2026 || selected.Month != time.April {
		t.Errorf("Expected 2026-April, got %d-%v", selected.Year, selected.Month)
	}
}

func TestUpdate_EscKey(t *testing.T) {
	m := New()
	updated, _ := m.Update(OpenMsg{Year: 2026, Month: time.March})
	m = updated.(Model)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)

	if m.focus {
		t.Error("Expected model to be blurred after Esc")
	}
	if cmd == nil {
		t.Fatal("Expected command from Esc key")
	}

	msg := cmd()
	if _, ok := msg.(CloseMsg); !ok {
		t.Errorf("Expected CloseMsg, got %T", msg)
	}
}

func TestUpdate_UnfocusedIgnoresKeys(t *testing.T) {
	m := New()
	updated, _ := m.Update(OpenMsg{Year: 2026, Month: time.March})
	m = updated.(Model)

	cursor := m.cursor
	m.Blur()

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	m = updated.(Model)

	if m.cursor != cursor {
		t.Error("Expected cursor not to change when unfocused")
	}
	if cmd != nil {
		t.Error("Expected nil command when unfocused")
	}
}

func TestView_Unfocused(t *testing.T) {
	m := New()
	view := m.View()

	if view != "" {
		t.Errorf("Expected empty view when unfocused, got '%s'", view)
	}
}

func TestView_EmptyItems(t *testing.T) {
	m := New()
	m.Focus()
	view := m.View()

	if view != "" {
		t.Errorf("Expected empty view with no items, got '%s'", view)
	}
}

func TestView_Focused(t *testing.T) {
	m := New()
	m.Width = 120
	updated, _ := m.Update(OpenMsg{Year: 2026, Month: time.March})
	m = updated.(Model)

	view := m.View()

	if view == "" {
		t.Error("Expected non-empty view when focused")
	}
	if !strings.Contains(view, "Period:") {
		t.Error("Expected view to contain 'Period:'")
	}
	if !strings.Contains(view, "March 2026") {
		t.Error("Expected view to contain 'March 2026'")
	}
}

func TestView_ArrowsDisplay(t *testing.T) {
	m := New()
	m.Width = 120
	updated, _ := m.Update(OpenMsg{Year: 2026, Month: time.June})
	m = updated.(Model)

	view := m.View()
	if !strings.Contains(view, "<<") {
		t.Error("Expected left arrow when cursor > 0")
	}
	if !strings.Contains(view, ">>") {
		t.Error("Expected right arrow when cursor < last")
	}

	m.cursor = 0
	view = m.View()
	if strings.Contains(view, "<<") {
		t.Error("Expected no left arrow when cursor is 0")
	}

	m.cursor = len(m.items) - 1
	view = m.View()
	if strings.Contains(view, ">>") {
		t.Error("Expected no right arrow when cursor is at last item")
	}
}

func TestView_CursorMarkers(t *testing.T) {
	m := New()
	m.Width = 120
	updated, _ := m.Update(OpenMsg{Year: 2026, Month: time.March})
	m = updated.(Model)

	view := m.View()
	if !strings.Contains(view, "> March 2026 <") {
		t.Error("Expected cursor markers around selected month")
	}
}

func TestFocusBlurFocused(t *testing.T) {
	m := New()

	if m.Focused() {
		t.Error("Expected unfocused initially")
	}

	m.Focus()
	if !m.Focused() {
		t.Error("Expected focused after Focus()")
	}

	m.Blur()
	if m.Focused() {
		t.Error("Expected unfocused after Blur()")
	}
}

func TestWithWidth(t *testing.T) {
	m := New()
	m.WithWidth(200)

	if m.Width != 200 {
		t.Errorf("Expected width 200, got %d", m.Width)
	}
}

func TestWithStyles(t *testing.T) {
	m := New()
	custom := Styles{
		Border: lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()),
	}
	m.WithStyles(custom)

	if m.styles.Border.GetBorderStyle() != lipgloss.NormalBorder() {
		t.Error("Expected custom border style")
	}
}

func TestOpen(t *testing.T) {
	cmd := Open(2026, time.March)
	if cmd == nil {
		t.Fatal("Expected command from Open")
	}

	msg := cmd()
	openMsg, ok := msg.(OpenMsg)
	if !ok {
		t.Fatalf("Expected OpenMsg, got %T", msg)
	}
	if openMsg.Year != 2026 || openMsg.Month != time.March {
		t.Errorf("Expected 2026-March, got %d-%v", openMsg.Year, openMsg.Month)
	}
}

func TestBuildVisibleLabels_NarrowWidth(t *testing.T) {
	m := New()
	m.Width = 40
	updated, _ := m.Update(OpenMsg{Year: 2026, Month: time.June})
	m = updated.(Model)

	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view even with narrow width")
	}
	if !strings.Contains(view, "June 2026") {
		t.Error("Expected at least the cursor month to be visible")
	}
}

func TestRenderLabel_Styles(t *testing.T) {
	m := New()
	updated, _ := m.Update(OpenMsg{Year: 2026, Month: time.June})
	m = updated.(Model)

	cursorLabel := m.renderLabel(m.cursor)
	if !strings.Contains(cursorLabel, "> ") || !strings.Contains(cursorLabel, " <") {
		t.Error("Expected cursor label to have > < markers")
	}

	if m.cursor > 0 {
		otherLabel := m.renderLabel(m.cursor - 1)
		if strings.Contains(otherLabel, "> ") {
			t.Error("Expected non-cursor label without markers")
		}
	}
}

func TestUpdate_OpenMsg_DifferentMonths(t *testing.T) {
	tests := []struct {
		year  int
		month time.Month
	}{
		{2021, time.January},
		{2026, time.December},
		{2030, time.July},
	}

	for _, tt := range tests {
		t.Run(tt.month.String(), func(t *testing.T) {
			m := New()
			updated, _ := m.Update(OpenMsg{Year: tt.year, Month: tt.month})
			m = updated.(Model)

			if m.items[m.cursor].year != tt.year || m.items[m.cursor].month != tt.month {
				t.Errorf("Expected cursor at %d-%v, got %v", tt.year, tt.month, m.items[m.cursor])
			}
		})
	}
}
