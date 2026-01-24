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

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	RefreshExpensesMsg        struct{}
	RefreshExpenseInsightsMsg struct{}
	ExpensesUpdatedMsg        struct{}
	NewExpenseMsg             struct {
		Account string
	}
)

type expenseItem = accountListItem[firefly.Account]

type modelExpenses struct {
	AccountListModel[firefly.Account]
}

func newModelExpenses(api ExpenseAPI) modelExpenses {
	config := &AccountListConfig[firefly.Account]{
		AccountType: "expense",
		Title:       "Expense accounts",
		GetItems: func(apiInterface any, sorted bool) []list.Item {
			return getExpensesItems(apiInterface.(ExpenseAPI), sorted)
		},
		RefreshItems: func(apiInterface any, accountType string) error {
			return apiInterface.(ExpenseAPI).UpdateAccounts(accountType)
		},
		RefreshMsgType: RefreshExpensesMsg{},
		UpdateMsgType:  ExpensesUpdatedMsg{},
		PromptNewFunc: func() tea.Cmd {
			return CmdPromptNewExpense(SetView(expensesView))
		},
		HasSort:     true,
		HasTotalRow: true,
		GetTotalFunc: func(api any) float64 {
			return api.(ExpenseAPI).GetTotalExpenseDiff()
		},
		FilterFunc: func(item list.Item) tea.Cmd {
			i, ok := item.(expenseItem)
			if ok {
				return Cmd(FilterMsg{Account: i.Entity})
			}
			return nil
		},
		SelectFunc: func(item list.Item) tea.Cmd {
			var cmds []tea.Cmd
			i, ok := item.(expenseItem)
			if ok {
				cmds = append(cmds, Cmd(FilterMsg{Account: i.Entity}))
			}
			cmds = append(cmds, SetView(transactionsView))
			return tea.Sequence(cmds...)
		},
	}
	return modelExpenses{
		AccountListModel: NewAccountListModel[firefly.Account](api, config),
	}
}

func (m modelExpenses) Init() tea.Cmd {
	return m.AccountListModel.Init()
}

func (m modelExpenses) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if newMsg, ok := msg.(NewExpenseMsg); ok {
		api := m.api.(ExpenseAPI)
		err := api.CreateExpenseAccount(newMsg.Account)
		if err != nil {
			return m, notify.NotifyWarn(err.Error())
		}
		return m, tea.Batch(
			Cmd(RefreshExpensesMsg{}),
			notify.NotifyLog(fmt.Sprintf("Expense account '%s' created", newMsg.Account)),
		)
	}

	switch msg.(type) {
	case RefreshExpenseInsightsMsg:
		return m, func() tea.Msg {
			err := m.api.(ExpenseAPI).UpdateExpenseInsights()
			if err != nil {
				return notify.NotifyWarn(err.Error())()
			}
			return ExpensesUpdatedMsg{}
		}
	}
	updated, cmd := m.AccountListModel.Update(msg)
	m.AccountListModel = updated.(AccountListModel[firefly.Account])
	return m, cmd
}

func getExpensesItems(api ExpenseAPI, sorted bool) []list.Item {
	items := []list.Item{}
	for _, account := range api.AccountsByType("expense") {
		spent := api.GetExpenseDiff(account.ID)
		if sorted && spent == 0 {
			continue
		}
		items = append(items, newAccountListItem(
			account,
			"Spent",
			spent,
		))
	}
	if sorted {
		slices.SortFunc(items, func(a, b list.Item) int {
			return int(b.(expenseItem).PrimaryVal) - int(a.(expenseItem).PrimaryVal)
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
