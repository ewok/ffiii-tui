/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"ffiii-tui/internal/firefly"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/charmbracelet/bubbles/textinput"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2).PaddingRight(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("170"))
)

type state uint

const (
	transactionView state = iota
	filterView
	periodView
	newView
	accountView
	categoryView
	expensesView
)

type (
	viewTransactionsMsg struct{}
	viewAccountsMsg     struct{}
	viewFilterMsg       struct{}
	viewNewMsg          struct{}
	viewCategoriesMsg   struct{}
	viewExpensesMsg     struct{}
)

type modelUI struct {
	state        state
	transactions modelTransactions
	filter       textinput.Model
	fireflyApi   *firefly.Api
	new          modelNewTransaction
	accounts     modelAccounts
	categories   modelCategories
	expenses     modelExpenses
}

func Show(api *firefly.Api) {

	ti := textinput.New()
	ti.Placeholder = "Filter"
	ti.CharLimit = 156
	ti.Width = 20

	n := newModelNewTransaction(api)
	a := newModelAccounts(api)
	t := newModelTransactions(api)
	c := newModelCategories(api)
	e := newModelExpenses(api)

	m := modelUI{filter: ti, fireflyApi: api, transactions: t, new: n, accounts: a, categories: c, expenses: e}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func (m modelUI) Init() tea.Cmd {
	return func() tea.Msg { return viewTransactionsMsg{} }
}

func (m modelUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case viewTransactionsMsg:
		m.state = transactionView
		m.filter.Blur()
		m.transactions.Focus()
		m.accounts.Blur()
		m.categories.Blur()
		m.expenses.Blur()
	case viewAccountsMsg:
		m.state = accountView
		m.filter.Blur()
		m.transactions.Blur()
		m.accounts.Focus()
		m.categories.Blur()
		m.expenses.Blur()
	case viewFilterMsg:
		m.state = filterView
		m.filter.Focus()
		m.transactions.Blur()
		m.accounts.Blur()
		m.categories.Blur()
		m.expenses.Blur()
	case viewNewMsg:
		m.state = newView
		m.filter.Blur()
		m.transactions.Blur()
		m.accounts.Blur()
		m.categories.Blur()
		m.expenses.Blur()
	case viewCategoriesMsg:
		m.state = categoryView
		m.filter.Blur()
		m.transactions.Blur()
		m.accounts.Blur()
		m.categories.Focus()
		m.expenses.Blur()
	case viewExpensesMsg:
		m.state = expensesView
		m.filter.Blur()
		m.transactions.Blur()
		m.accounts.Blur()
		m.categories.Blur()
		m.expenses.Focus()
	}

	switch m.state {
	case accountView, transactionView, categoryView, expensesView:
		nModel, nCmd := m.transactions.Update(msg)
		listModel, ok := nModel.(modelTransactions)
		if !ok {
			panic("Somthing bad happened")
		}
		m.transactions = listModel
		cmds = append(cmds, nCmd)

		nModel, nCmd = m.accounts.Update(msg)
		accountsModel, ok := nModel.(modelAccounts)
		if !ok {
			panic("Somthing bad happened")
		}
		m.accounts = accountsModel
		cmds = append(cmds, nCmd)

		nModel, nCmd = m.categories.Update(msg)
		categoryModel, ok := nModel.(modelCategories)
		if !ok {
			panic("Somthing bad happened")
		}
		m.categories = categoryModel
		cmds = append(cmds, nCmd)

		nModel, nCmd = m.expenses.Update(msg)
		expensesModel, ok := nModel.(modelExpenses)
		if !ok {
			panic("Somthing bad happened")
		}
		m.expenses = expensesModel
		cmds = append(cmds, nCmd)
	case filterView:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				cmds = append(cmds, Cmd(FilterMsg{query: ""}))
				cmds = append(cmds, Cmd(viewTransactionsMsg{}))
			case "enter":
				value := m.filter.Value()
				cmds = append(cmds, Cmd(FilterMsg{query: value}))
				cmds = append(cmds, Cmd(viewTransactionsMsg{}))
			}
		}
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		cmds = append(cmds, cmd)
	case newView:
		nModel, nCmd := m.new.Update(msg)
		newModel, ok := nModel.(modelNewTransaction)
		if !ok {
			panic("Somthing bad happened")
		}
		m.new = newModel
		cmds = append(cmds, nCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m modelUI) View() string {
	switch m.state {
	case transactionView:
		// return baseStyle.Render(m.transactions.View())
		return lipgloss.JoinHorizontal(lipgloss.Top,
		    baseStyle.Render(m.accounts.View()),
		    baseStyle.Render(m.transactions.View()))
	case accountView:
		// return baseStyle.Render(m.accounts.View())
		return lipgloss.JoinHorizontal(
		    lipgloss.Top,
		    baseStyle.Render(m.accounts.View()),
		    baseStyle.Render(m.transactions.View()))
    // TODO: Make prompt view
	case filterView:
		return baseStyle.Render(fmt.Sprintf("filter: %s", m.filter.View()) + "\n" + m.transactions.View())
	case newView:
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			baseStyle.Render(m.accounts.View()),
			baseStyle.Render(m.new.View()))
	case categoryView:
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			baseStyle.Render(m.categories.View()),
			baseStyle.Render(m.transactions.View()))
	case expensesView:
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			baseStyle.Render(m.expenses.View()),
			baseStyle.Render(m.transactions.View()))
	}
	return baseStyle.Render(m.transactions.View()) + "\n"
}
