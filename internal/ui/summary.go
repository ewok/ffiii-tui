/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"unicode/utf8"

	"ffiii-tui/internal/ui/notify"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap"
)

type (
	RefreshSummaryMsg struct{}
	SummaryUpdateMsg  struct{}
)

type summaryItem struct {
	title, value  string
	monetaryValue float64
	style         lipgloss.Style
}

func (i summaryItem) FilterValue() string { return i.title }

type summaryDelegate struct{}

func (d summaryDelegate) Height() int                             { return 1 }
func (d summaryDelegate) Spacing() int                            { return 0 }
func (d summaryDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d summaryDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(summaryItem)
	if !ok {
		return
	}

	styledTitle := i.title
	styledValue := i.style.Render(i.value)

	availableWidth := m.Width() + 4

	valueLen := utf8.RuneCountInString(i.value)
	titleLen := utf8.RuneCountInString(i.title)

	spacingNeeded := max(availableWidth-titleLen-valueLen, 1)
	if spacingNeeded == 1 {
		m.SetWidth(titleLen + valueLen + 5)
	}

	str := fmt.Sprintf(" %s%s%s", styledTitle,
		strings.Repeat(" ", spacingNeeded),
		styledValue)

	_, err := fmt.Fprint(w, str)
	if err != nil {
		zap.L().Debug("failed to render summary item", zap.Error(err))
	}
}

type modelSummary struct {
	list   list.Model
	api    SummaryAPI
	styles Styles
}

func newModelSummary(api SummaryAPI) modelSummary {
	styles := DefaultStyles()
	items := getSummaryItems(api, styles)
	m := modelSummary{
		list:   list.New(items, summaryDelegate{}, 0, 0),
		api:    api,
		styles: styles,
	}
	m.list.Title = "Summary"
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowHelp(false)
	m.list.DisableQuitKeybindings()
	m.list.SetShowPagination(false)
	m.list.SetWidth(api.GetMaxWidth())
	return m
}

func (m modelSummary) Init() tea.Cmd {
	return nil
}

func (m modelSummary) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case RefreshSummaryMsg:
		return m, func() tea.Msg {
			err := m.api.UpdateSummary()
			if err != nil {
				return notify.NotifyWarn(err.Error())()
			}
			return SummaryUpdateMsg{}
		}
	case SummaryUpdateMsg:
		return m, tea.Sequence(
			m.list.SetItems(getSummaryItems(m.api, m.styles)),
			tea.WindowSize())
	case UpdatePositions:
		_, v := m.styles.Base.GetFrameSize()
		l := len(m.list.Items())
		m.list.SetHeight(l + v + 1)
		summarySize = m.list.Height()
	}
	return m, nil
}

func (m modelSummary) View() string {
	return m.styles.LeftPanel.Render(m.list.View())
}

func getSummaryItems(api SummaryAPI, styles Styles) []list.Item {
	var style lipgloss.Style
	items := []list.Item{}
	for _, si := range api.SummaryItems() {
		switch {
		case si.MonetaryValue < 0:
			style = styles.Withdrawal
		case si.MonetaryValue > 0:
			style = styles.Deposit
		default:
			style = styles.Normal
		}
		item := summaryItem{
			title:         si.Title,
			value:         si.ValueParsed,
			monetaryValue: si.MonetaryValue,
			style:         style,
		}
		items = append(items, item)
	}

	slices.SortFunc(items, func(a, b list.Item) int {
		sa := a.(summaryItem)
		sb := b.(summaryItem)
		if sa.monetaryValue != sb.monetaryValue {
			if sa.monetaryValue < sb.monetaryValue {
				return 1
			}
			return -1
		}
		return 0
	})
	return items
}
