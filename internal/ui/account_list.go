/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"reflect"

	"ffiii-tui/internal/firefly"
	"ffiii-tui/internal/ui/notify"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// AccountListModel is a generic model for account/category list views
type AccountListModel[T ListEntity] struct {
	list   list.Model
	api    any // Specific API interface
	focus  bool
	sorted bool
	config *AccountListConfig[T]
	styles Styles
	keymap AccountKeyMap
}

// NewAccountListModel creates a new generic account list model
func NewAccountListModel[T ListEntity](api any, config *AccountListConfig[T]) AccountListModel[T] {
	items := config.GetItems(api, false)

	m := AccountListModel[T]{
		list:   list.New(items, list.NewDefaultDelegate(), 0, 0),
		api:    api,
		config: config,
		styles: DefaultStyles(),
		keymap: DefaultAccountKeyMap(),
	}
	m.list.Title = config.Title
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowHelp(false)
	m.list.DisableQuitKeybindings()

	return m
}

func (m AccountListModel[T]) Init() tea.Cmd {
	return nil
}

func (m AccountListModel[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if matchMsgType(msg, m.config.RefreshMsgType) {
		return m, func() tea.Msg {
			startLoading("Loading accounts...")
			defer stopLoading()
			err := m.config.RefreshItems(m.api, m.config.AccountType)
			if err != nil {
				return notify.NotifyWarn(err.Error())()
			}
			return m.config.UpdateMsgType
		}
	}

	if matchMsgType(msg, m.config.UpdateMsgType) {
		return m, m.updateItemsCmd()
	}

	if msg, ok := msg.(UpdatePositions); ok {
		if msg.layout != nil {
			h, v := m.styles.Base.GetFrameSize()
			var height int
			if m.config.HasSummary {
				height = msg.layout.Height - v - msg.layout.TopSize - msg.layout.SummarySize
			} else {
				height = msg.layout.Height - v - msg.layout.TopSize
			}
			m.list.SetSize(msg.layout.Width-h, height)
		}
		return m, nil
	}

	if !m.focus {
		return m, nil
	}

	// Common keys
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Quit):
			return m, SetView(transactionsView)
		case key.Matches(msg, m.keymap.ViewTransactions):
			return m, SetView(transactionsView)
		case key.Matches(msg, m.keymap.ViewAssets):
			return m, SetView(assetsView)
		case key.Matches(msg, m.keymap.ViewCategories):
			return m, SetView(categoriesView)
		case key.Matches(msg, m.keymap.ViewExpenses):
			return m, SetView(expensesView)
		case key.Matches(msg, m.keymap.ViewRevenues):
			return m, SetView(revenuesView)
		case key.Matches(msg, m.keymap.ViewLiabilities):
			return m, SetView(liabilitiesView)
		case key.Matches(msg, m.keymap.Refresh):
			return m, Cmd(m.config.RefreshMsgType)
		case key.Matches(msg, m.keymap.ResetFilter):
			return m, Cmd(FilterMsg{Reset: true})
		case key.Matches(msg, m.keymap.Sort):
			if m.config.HasSort {
				m.sorted = !m.sorted
				return m, Cmd(m.config.UpdateMsgType)
			}
		case key.Matches(msg, m.keymap.New):
			return m, m.config.PromptNewFunc()
		case key.Matches(msg, m.keymap.Filter):
			i, ok := m.list.SelectedItem().(accountListItem[T])
			if ok {
				if m.config.HasTotalRow && i.Entity.GetName() == "Total" {
					return m, nil
				}
				return m, m.config.FilterFunc(i)
			}
			return m, nil
		case key.Matches(msg, m.keymap.Select):
			i, ok := m.list.SelectedItem().(accountListItem[T])
			if ok {
				if m.config.HasTotalRow && i.Entity.GetName() == "Total" {
					return m, nil
				}
				return m, m.config.SelectFunc(i)
			}
			return m, nil
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m AccountListModel[T]) View() string {
	return m.styles.LeftPanel.Render(m.list.View())
}

func (m *AccountListModel[T]) Focus() {
	m.list.FilterInput.Focus()
	m.focus = true
}

func (m *AccountListModel[T]) Blur() {
	m.list.FilterInput.Blur()
	m.focus = false
}

func (m AccountListModel[T]) createTotalEntity(primary float64) list.Item {
	var entity T

	acc := firefly.Account{Name: "Total", CurrencyCode: ""}
	if api, ok := m.api.(interface{ PrimaryCurrency() firefly.Currency }); ok {
		acc.CurrencyCode = api.PrimaryCurrency().Code
	}
	entity = any(acc).(T)
	return newAccountListItem(entity, "Total", primary)
}

func (m *AccountListModel[T]) updateItemsCmd() tea.Cmd {
	items := m.config.GetItems(m.api, m.sorted)

	if m.config.HasTotalRow && m.config.GetTotalFunc != nil {
		primary := m.config.GetTotalFunc(m.api)
		totalEntity := m.createTotalEntity(primary)

		cmds := []tea.Cmd{
			m.list.SetItems(items),
			m.list.InsertItem(0, totalEntity),
		}

		cmds = append(cmds, Cmd(DataLoadCompletedMsg{DataType: m.config.AccountType}))

		return tea.Sequence(cmds...)
	}

	return tea.Batch(
		m.list.SetItems(items),
		Cmd(DataLoadCompletedMsg{DataType: m.config.AccountType}),
	)
}

func matchMsgType(msg, ty tea.Msg) bool {
	return reflect.TypeOf(msg) == reflect.TypeOf(ty)
}
