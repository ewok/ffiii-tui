/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"fmt"
	"slices"

	"ffiii-tui/internal/firefly"
	"ffiii-tui/internal/ui/notify"
	"ffiii-tui/internal/ui/prompt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

var totalRevenueAccount = firefly.Account{Name: "Total", CurrencyCode: ""}

type (
	RefreshRevenuesMsg        struct{}
	RefreshRevenueInsightsMsg struct{}
	RevenuesUpdateMsg         struct{}
	NewRevenueMsg             struct {
		Account string
	}
)

type revenueItem struct {
	account firefly.Account
	earned  float64
}

func (i revenueItem) Title() string { return i.account.Name }
func (i revenueItem) Description() string {
	return fmt.Sprintf("Earned: %.2f %s", i.earned, i.account.CurrencyCode)
}
func (i revenueItem) FilterValue() string { return i.account.Name }

type modelRevenues struct {
	list   list.Model
	api    RevenueAPI
	focus  bool
	sorted bool
	keymap RevenueKeyMap
	styles Styles
}

func newModelRevenues(api RevenueAPI) modelRevenues {
	// Set total revenue account currency
	totalRevenueAccount.CurrencyCode = api.PrimaryCurrency().Code

	items := getRevenuesItems(api, false)

	m := modelRevenues{
		list:   list.New(items, list.NewDefaultDelegate(), 0, 0),
		api:    api,
		keymap: DefaultRevenueKeyMap(),
		styles: DefaultStyles(),
	}
	m.list.Title = "Revenue accounts"
	m.list.SetFilteringEnabled(false)
	m.list.SetShowStatusBar(false)
	m.list.SetShowHelp(false)
	m.list.DisableQuitKeybindings()

	return m
}

func (m modelRevenues) Init() tea.Cmd {
	return nil
}

func (m modelRevenues) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case RefreshRevenueInsightsMsg:
		return m, func() tea.Msg {
			err := m.api.UpdateRevenueInsights()
			if err != nil {
				return notify.NotifyWarn(err.Error())()
			}
			return RevenuesUpdateMsg{}
		}
	case RefreshRevenuesMsg:
		return m, func() tea.Msg {
			err := m.api.UpdateAccounts("revenue")
			if err != nil {
				return notify.NotifyWarn(err.Error())()
			}
			return RevenuesUpdateMsg{}
		}
	case RevenuesUpdateMsg:
		return m, tea.Sequence(
			m.list.SetItems(getRevenuesItems(m.api, m.sorted)),
			m.list.InsertItem(0, revenueItem{
				account: totalRevenueAccount,
				earned:  m.api.GetTotalRevenueDiff(),
			}),
			Cmd(DataLoadCompletedMsg{DataType: "revenues"}),
		)
	case NewRevenueMsg:
		err := m.api.CreateRevenueAccount(msg.Account)
		if err != nil {
			return m, notify.NotifyWarn(err.Error())
		}
		return m, tea.Batch(
			Cmd(RefreshRevenuesMsg{}),
			notify.NotifyLog(fmt.Sprintf("Revenue account '%s' created", msg.Account)),
		)
	case UpdatePositions:
		h, v := m.styles.Base.GetFrameSize()
		m.list.SetSize(globalWidth-h, globalHeight-v-topSize)
	}

	if !m.focus {
		return m, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Quit):
			return m, SetView(transactionsView)
		case key.Matches(msg, m.keymap.New):
			return m, CmdPromptNewRevenue(SetView(revenuesView))
		case key.Matches(msg, m.keymap.Filter):
			i, ok := m.list.SelectedItem().(revenueItem)
			if ok {
				if i.account == totalRevenueAccount {
					return m, nil
				}
				return m, Cmd(FilterMsg{Account: i.account})
			}
			return m, nil
		case key.Matches(msg, m.keymap.Refresh):
			return m, Cmd(RefreshRevenuesMsg{})
		case key.Matches(msg, m.keymap.Sort):
			m.sorted = !m.sorted
			return m, Cmd(RevenuesUpdateMsg{})
		case key.Matches(msg, m.keymap.ResetFilter):
			return m, Cmd(FilterMsg{Reset: true})
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
			// case "R":
			// 	return m, Cmd(RefreshRevenuesMsg{})
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m modelRevenues) View() string {
	return m.styles.LeftPanel.Render(m.list.View())
}

func (m *modelRevenues) Focus() {
	m.list.FilterInput.Focus()
	m.focus = true
}

func (m *modelRevenues) Blur() {
	m.list.FilterInput.Blur()
	m.focus = false
}

func getRevenuesItems(api RevenueAPI, sorted bool) []list.Item {
	items := []list.Item{}
	for _, account := range api.AccountsByType("revenue") {
		earned := api.GetRevenueDiff(account.ID)
		if sorted && earned == 0 {
			continue
		}
		items = append(items, revenueItem{
			account: account,
			earned:  earned,
		})
	}
	if sorted {
		slices.SortFunc(items, func(a, b list.Item) int {
			return int(b.(revenueItem).earned) - int(a.(revenueItem).earned)
		})
	}
	return items
}

func CmdPromptNewRevenue(backCmd tea.Cmd) tea.Cmd {
	return prompt.Ask(
		"New Revenue(<name>): ",
		"",
		func(value string) tea.Cmd {
			var cmds []tea.Cmd
			if value != "None" {
				cmds = append(cmds, Cmd(NewRevenueMsg{Account: value}))
			}
			cmds = append(cmds, backCmd)
			return tea.Sequence(cmds...)
		},
	)
}
