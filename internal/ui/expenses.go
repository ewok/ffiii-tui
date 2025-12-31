/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"ffiii-tui/internal/firefly"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type RefreshExpensesMsg struct{}
type RefreshExpenseInsightsMsg struct{}
type NewExpenseMsg struct {
	account string
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
}

func newModelExpenses(api *firefly.Api) modelExpenses {
	items := getExpensesItems(api, false)

	m := modelExpenses{list: list.New(items, list.NewDefaultDelegate(), 0, 0), api: api}
	m.list.Title = "Expenses"
	m.list.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowStatusBar(false)
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
		err := m.api.CreateAccount(msg.account, "expense", "")
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
		if m.focus {
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				return m, Cmd(ViewTransactionsMsg{})
			case "n":
				return m, Cmd(PromptMsg{
					Prompt: "New Expense: ",
					Value:  "",
					Callback: func(value string) tea.Cmd {
						var cmds []tea.Cmd
						if value != "None" {
							cmds = append(cmds, Cmd(NewExpenseMsg{account: value}))
						}
						cmds = append(cmds, Cmd(ViewExpensesMsg{}))
						return tea.Sequence(cmds...)
					}})
			case "f":
				i, ok := m.list.SelectedItem().(expenseItem)
				if ok {
					return m, Cmd(FilterMsg{account: i.account})
				}
				return m, nil
			case "a":
				return m, Cmd(ViewAssetsMsg{})
			case "c":
				return m, Cmd(ViewCategoriesMsg{})
			case "r":
				return m, Cmd(RefreshExpenseInsightsMsg{})
			case "R":
				return m, Cmd(RefreshExpensesMsg{})
			case "i":
				return m, Cmd(ViewRevenuesMsg{})
			case "s":
				m.sorted = !m.sorted
				return m, m.list.SetItems(getExpensesItems(m.api, m.sorted))
			case "t":
				return m, Cmd(ViewTransactionsMsg{})
			case "ctrl+a":
				return m, Cmd(FilterMsg{reset: true})
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m modelExpenses) View() string {
	return m.list.View()
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
