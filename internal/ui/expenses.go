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

var totalExpenseAccount = firefly.Account{Name: "Total", CurrencyCode: ""}

type (
	RefreshExpensesMsg        struct{}
	RefreshExpenseInsightsMsg struct{}
	ExpensesUpdatedMsg        struct{}
	NewExpenseMsg             struct {
		Account string
	}
)

type expenseItem struct {
	account firefly.Account
	spent   float64
}

func (i expenseItem) Title() string { return i.account.Name }
func (i expenseItem) Description() string {
	return fmt.Sprintf("Spent: %.2f %s", i.spent, i.account.CurrencyCode)
}
func (i expenseItem) FilterValue() string { return i.account.Name }

type modelExpenses struct {
	list   list.Model
	api    ExpenseAPI
	focus  bool
	sorted bool
	keymap ExpenseKeyMap
	styles Styles
}

func newModelExpenses(api ExpenseAPI) modelExpenses {
	// Set total expense account currency
	totalExpenseAccount.CurrencyCode = api.PrimaryCurrency().Code

	items := getExpensesItems(api, false)

	m := modelExpenses{
		list:   list.New(items, list.NewDefaultDelegate(), 0, 0),
		api:    api,
		keymap: DefaultExpenseKeyMap(),
		styles: DefaultStyles(),
	}
	m.list.Title = "Expense accounts"
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
		return m, func() tea.Msg {
			err := m.api.UpdateExpenseInsights()
			if err != nil {
				return notify.NotifyWarn(err.Error())()
			}
			return ExpensesUpdatedMsg{}
		}
	case RefreshExpensesMsg:
		return m, func() tea.Msg {
			err := m.api.UpdateAccounts("expense")
			if err != nil {
				return notify.NotifyWarn(err.Error())()
			}
			return ExpensesUpdatedMsg{}
		}
	case ExpensesUpdatedMsg:
		return m, tea.Sequence(
			m.list.SetItems(getExpensesItems(m.api, m.sorted)),
			m.list.InsertItem(0, expenseItem{
				account: totalExpenseAccount,
				spent:   m.api.GetTotalExpenseDiff(),
			}),
			Cmd(DataLoadCompletedMsg{DataType: "expenses"}),
		)
	case NewExpenseMsg:
		err := m.api.CreateExpenseAccount(msg.Account)
		if err != nil {
			return m, notify.NotifyWarn(err.Error())
		}
		return m, tea.Batch(
			Cmd(RefreshExpensesMsg{}),
			notify.NotifyLog(fmt.Sprintf("Expense account '%s' created", msg.Account)),
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
			return m, CmdPromptNewExpense(SetView(expensesView))
		case key.Matches(msg, m.keymap.Filter):
			i, ok := m.list.SelectedItem().(expenseItem)
			if ok {
				if i.account == totalExpenseAccount {
					return m, nil
				}
				return m, Cmd(FilterMsg{Account: i.account})
			}
			return m, nil
		case key.Matches(msg, m.keymap.Refresh):
			return m, Cmd(RefreshExpensesMsg{})
		case key.Matches(msg, m.keymap.Sort):
			m.sorted = !m.sorted
			return m, Cmd(ExpensesUpdatedMsg{})
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
			// 	return m, Cmd(RefreshExpensesMsg{})
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m modelExpenses) View() string {
	return m.styles.LeftPanel.Render(m.list.View())
}

func (m *modelExpenses) Focus() {
	m.list.FilterInput.Focus()
	m.focus = true
}

func (m *modelExpenses) Blur() {
	m.list.FilterInput.Blur()
	m.focus = false
}

func getExpensesItems(api ExpenseAPI, sorted bool) []list.Item {
	items := []list.Item{}
	for _, account := range api.AccountsByType("expense") {
		spent := api.GetExpenseDiff(account.ID)

		if sorted && spent == 0 {
			continue
		}
		items = append(items, expenseItem{
			account: account,
			spent:   spent,
		})
	}
	if sorted {
		slices.SortFunc(items, func(a, b list.Item) int {
			return int(b.(expenseItem).spent) - int(a.(expenseItem).spent)
		})
	}

	return items
}

func CmdPromptNewExpense(backCmd tea.Cmd) tea.Cmd {
	return prompt.Ask(
		"New Expense(<name>): ",
		"",
		func(value string) tea.Cmd {
			var cmds []tea.Cmd
			if value != "None" {
				cmds = append(cmds, Cmd(NewExpenseMsg{Account: value}))
			}
			cmds = append(cmds, backCmd)
			return tea.Sequence(cmds...)
		},
	)
}
