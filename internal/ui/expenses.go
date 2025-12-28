/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"ffiii-tui/internal/firefly"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type RefreshEspensesMsg struct{}
type NewExpenseMsg struct {
	account string
}

type expenseItem struct {
	id, name, balance, currency string
}

func (i expenseItem) Title() string { return i.name }
func (i expenseItem) Description() string {
	if i.balance != "" && i.currency != "" {
		return fmt.Sprintf("%s %s", i.balance, i.currency)
	}
	return ""
}
func (i expenseItem) FilterValue() string { return i.name }

type modelExpenses struct {
	list  list.Model
	api   *firefly.Api
	focus bool
}

func newModelExpenses(api *firefly.Api) modelExpenses {
	items := getExpensesItems(api)

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
	case RefreshExpensesMsg:
		m.api.UpdateExpenses()
		return m, m.list.SetItems(getExpensesItems(m.api))
	case NewExpenseMsg:
		err := m.api.CreateAccount(msg.account, "expense", "")
		// TODO: Report error to user
		if err == nil {
			return m, Cmd(RefreshExpensesMsg{})
		}
		return m, nil
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
						if value != "" {
							cmds = append(cmds, Cmd(NewExpenseMsg{account: value}))
						}
						cmds = append(cmds, Cmd(ViewExpensesMsg{}))
						return tea.Sequence(cmds...)
					}})
			case "f":
				i, ok := m.list.SelectedItem().(expenseItem)
				if ok {
					return m, Cmd(FilterItemMsg{account: i.name})
				}
				return m, nil
			case "a":
				return m, Cmd(ViewAssetsMsg{})
			case "c":
				return m, Cmd(ViewCategoriesMsg{})
			case "r":
				return m, Cmd(RefreshExpensesMsg{})
			case "i":
				return m, Cmd(ViewRevenuesMsg{})
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

func getExpensesItems(api *firefly.Api) []list.Item {
	items := []list.Item{}
	for _, i := range api.Expenses {
		items = append(items, expenseItem{
			id:       i.ID,
			name:     i.Name,
			balance:  fmt.Sprintf("%.2f", i.Balance),
			currency: i.CurrencyCode,
		})
	}
	return items
}
