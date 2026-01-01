/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"ffiii-tui/internal/firefly"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type RefreshExpensesMsg struct{}
type RefreshExpenseInsightsMsg struct{}
type NewExpenseMsg struct {
	Account string
}

type expenseItem struct {
	account, currency string
	spent             float64
}

func (i expenseItem) Title() string { return i.account }
func (i expenseItem) Description() string {
	return fmt.Sprintf("Spent: %.2f %s", i.spent, i.currency)
}
func (i expenseItem) FilterValue() string { return i.account }

type modelExpenses struct {
	list   list.Model
	api    *firefly.Api
	focus  bool
	sorted bool
	keymap ExpenseKeyMap
}

func newModelExpenses(api *firefly.Api) modelExpenses {
	items := getExpensesItems(api, false)

	m := modelExpenses{
		list:   list.New(items, list.NewDefaultDelegate(), 0, 0),
		api:    api,
		keymap: DefaultExpenseKeyMap(),
	}
	m.list.Title = "Expenses"
	m.list.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowStatusBar(false)
	m.list.SetShowHelp(false)
	m.list.DisableQuitKeybindings()

	return m
}

func (m modelExpenses) Init() tea.Cmd {
	return nil
}

func (m modelExpenses) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case RefreshExpenseInsightsMsg:
		return m, tea.Sequence(
			Cmd(m.api.UpdateExpenseInsights()),
			m.list.SetItems(getExpensesItems(m.api, m.sorted)))
	case RefreshExpensesMsg:
		return m, tea.Sequence(
			Cmd(m.api.UpdateAccounts("expense")),
			m.list.SetItems(getExpensesItems(m.api, m.sorted)))
	case NewExpenseMsg:
		err := m.api.CreateAccount(msg.Account, "expense", "")
		if err != nil {
			return m, Notify(err.Error(), Warning)
		}
		return m, Cmd(RefreshExpensesMsg{})
	case tea.WindowSizeMsg:
		h, v := baseStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-topSize)
	}

	if !m.focus {
		return m, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Quit):
			return m, Cmd(ViewTransactionsMsg{})
		case key.Matches(msg, m.keymap.New):
			return m, Cmd(PromptMsg{
				Prompt: "New Expense: ",
				Value:  "",
				Callback: func(value string) tea.Cmd {
					var cmds []tea.Cmd
					if value != "None" {
						cmds = append(cmds, Cmd(NewExpenseMsg{Account: value}))
					}
					cmds = append(cmds, Cmd(ViewExpensesMsg{}))
					return tea.Sequence(cmds...)
				}})
		case key.Matches(msg, m.keymap.Filter):
			i, ok := m.list.SelectedItem().(expenseItem)
			if ok {
				return m, Cmd(FilterMsg{Account: i.account})
			}
			return m, nil
		case key.Matches(msg, m.keymap.Refresh):
			return m, Cmd(RefreshExpenseInsightsMsg{})
		case key.Matches(msg, m.keymap.Sort):
			m.sorted = !m.sorted
			return m, m.list.SetItems(getExpensesItems(m.api, m.sorted))
		case key.Matches(msg, m.keymap.ResetFilter):
			return m, Cmd(FilterMsg{Reset: true})
		case key.Matches(msg, m.keymap.ViewAssets):
			return m, Cmd(ViewAssetsMsg{})
		case key.Matches(msg, m.keymap.ViewExpenses):
			return m, Cmd(ViewTransactionsMsg{})
		case key.Matches(msg, m.keymap.ViewRevenues):
			return m, Cmd(ViewRevenuesMsg{})
		case key.Matches(msg, m.keymap.ViewCategories):
			return m, Cmd(ViewCategoriesMsg{})
		case key.Matches(msg, m.keymap.ViewTransactions):
			return m, Cmd(ViewTransactionsMsg{})
			// case "R":
			// 	return m, Cmd(RefreshExpensesMsg{})
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m modelExpenses) View() string {
	return lipgloss.NewStyle().PaddingRight(1).Render(m.list.View())
}

func (m *modelExpenses) Focus() {
	m.list.FilterInput.Focus()
	m.focus = true
}

func (m *modelExpenses) Blur() {
	m.list.FilterInput.Blur()
	m.focus = false
}

func getExpensesItems(api *firefly.Api, sorted bool) []list.Item {
	items := []list.Item{}
	for _, i := range api.Accounts["expense"] {
		spent := api.GetExpenseDiff(i.ID)

		if sorted && spent == 0 {
			continue
		}
		items = append(items, expenseItem{
			account:  i.Name,
			currency: i.CurrencyCode,
			spent:    spent,
		})
	}
	if sorted {
		slices.SortFunc(items, func(a, b list.Item) int {
			return int(b.(expenseItem).spent) - int(a.(expenseItem).spent)
		})
	}

	return items
}
