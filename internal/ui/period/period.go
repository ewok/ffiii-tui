/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package period

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	yearsRange = 5
)

type OpenMsg struct {
	Year  int
	Month time.Month
}

type SelectedMsg struct {
	Year  int
	Month time.Month
}

type CloseMsg struct{}

type monthEntry struct {
	year  int
	month time.Month
}

func (e monthEntry) label() string {
	return fmt.Sprintf("%s %d", e.month.String(), e.year)
}

type Model struct {
	items   []monthEntry
	cursor  int
	current int
	focus   bool
	styles  Styles
	Width   int
}

func New() Model {
	return Model{
		styles: DefaultStyles(),
		Width:  80,
	}
}

func generateItems(year int, month time.Month) []monthEntry {
	start := time.Date(year-yearsRange, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+yearsRange, 12, 1, 0, 0, 0, 0, time.UTC)

	var items []monthEntry
	for d := start; !d.After(end); d = d.AddDate(0, 1, 0) {
		items = append(items, monthEntry{
			year:  d.Year(),
			month: d.Month(),
		})
	}
	return items
}

func findIndex(items []monthEntry, year int, month time.Month) int {
	for i, e := range items {
		if e.year == year && e.month == month {
			return i
		}
	}
	return 0
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case OpenMsg:
		m.items = generateItems(msg.Year, msg.Month)
		idx := findIndex(m.items, msg.Year, msg.Month)
		m.cursor = idx
		m.current = idx
		m.Focus()
		return m, nil
	case CloseMsg:
		m.Blur()
		return m, nil
	}

	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h", "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right", "l", "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			selected := m.items[m.cursor]
			m.Blur()
			return m, func() tea.Msg {
				return SelectedMsg{
					Year:  selected.year,
					Month: selected.month,
				}
			}
		case "esc":
			m.Blur()
			return m, func() tea.Msg {
				return CloseMsg{}
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.focus || len(m.items) == 0 {
		return ""
	}

	prefix := " Period: "
	separator := " | "
	arrowLeft := "<< "
	arrowRight := " >>"
	borderOverhead := 4

	available := m.Width - borderOverhead - lipgloss.Width(prefix) -
		lipgloss.Width(arrowLeft) - lipgloss.Width(arrowRight)

	visibleLabels := m.buildVisibleLabels(available, separator)

	var line strings.Builder
	line.WriteString(prefix)

	if m.cursor > 0 {
		line.WriteString(arrowLeft)
	} else {
		line.WriteString("   ")
	}

	line.WriteString(visibleLabels)

	if m.cursor < len(m.items)-1 {
		line.WriteString(arrowRight)
	} else {
		line.WriteString("   ")
	}

	return m.styles.Border.Width(m.Width).Render(line.String())
}

func (m Model) buildVisibleLabels(maxWidth int, separator string) string {
	sepWidth := lipgloss.Width(separator)

	cursorLabel := m.renderLabel(m.cursor)
	cursorWidth := lipgloss.Width(cursorLabel)
	remaining := maxWidth - cursorWidth

	var leftLabels []string
	var rightLabels []string
	li := m.cursor - 1
	ri := m.cursor + 1

	for remaining > 0 {
		addedAny := false

		if li >= 0 {
			label := m.renderLabel(li)
			w := lipgloss.Width(label) + sepWidth
			if w <= remaining {
				leftLabels = append(leftLabels, label)
				remaining -= w
				li--
				addedAny = true
			} else {
				li = -1
			}
		}

		if ri < len(m.items) {
			label := m.renderLabel(ri)
			w := lipgloss.Width(label) + sepWidth
			if w <= remaining {
				rightLabels = append(rightLabels, label)
				remaining -= w
				ri++
				addedAny = true
			} else {
				ri = len(m.items)
			}
		}

		if !addedAny {
			break
		}
	}

	var parts []string
	for i := len(leftLabels) - 1; i >= 0; i-- {
		parts = append(parts, leftLabels[i])
	}
	parts = append(parts, cursorLabel)
	parts = append(parts, rightLabels...)

	return strings.Join(parts, separator)
}

func (m Model) renderLabel(i int) string {
	entry := m.items[i]
	label := entry.label()

	switch i {
	case m.cursor:
		return m.styles.Selected.Render("> " + label + " <")
	case m.current:
		return m.styles.Current.Render(label)
	default:
		return m.styles.Item.Render(label)
	}
}

func (m *Model) Focus() {
	m.focus = true
}

func (m *Model) Blur() {
	m.focus = false
}

func (m *Model) Focused() bool {
	return m.focus
}

func (m *Model) WithWidth(width int) *Model {
	m.Width = width
	return m
}

func (m *Model) WithStyles(styles Styles) *Model {
	m.styles = styles
	return m
}

func Open(year int, month time.Month) tea.Cmd {
	return func() tea.Msg {
		return OpenMsg{
			Year:  year,
			Month: month,
		}
	}
}
