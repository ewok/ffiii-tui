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
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2).PaddingRight(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("170"))
)

type state uint

const (
	transactionView state = iota
	periodView
	newView
	accountView
	categoryView
	expensesView
	promptView
)

type (
	ViewTransactionsMsg struct{}
	ViewAccountsMsg     struct{}
	ViewNewMsg          struct{}
	ViewCategoriesMsg   struct{}
	ViewExpensesMsg     struct{}
	ViewPromptMsg       struct{}
)

type modelUI struct {
	state        state
	transactions modelTransactions
	api          *firefly.Api
	new          modelNewTransaction
	accounts     modelAccounts
	categories   modelCategories
	expenses     modelExpenses
	prompt       modelPrompt
}

func Show(api *firefly.Api) {

	m := modelUI{
		api:          api,
		transactions: newModelTransactions(api),
		new:          newModelNewTransaction(api),
		accounts:     newModelAccounts(api),
		categories:   newModelCategories(api),
		expenses:     newModelExpenses(api),
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func (m modelUI) Init() tea.Cmd {
	return func() tea.Msg { return ViewTransactionsMsg{} }
}

func (m modelUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

    // FIXME: Remake routing

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
		// TODO: Make it prettier
	case ViewTransactionsMsg:
		m.state = transactionView
		m.transactions.Focus()
		m.accounts.Blur()
		m.categories.Blur()
		m.expenses.Blur()
		m.prompt.Blur()
	case ViewAccountsMsg:
		m.state = accountView
		m.transactions.Blur()
		m.accounts.Focus()
		m.categories.Blur()
		m.expenses.Blur()
		m.prompt.Blur()
	case ViewNewMsg:
		m.state = newView
		m.transactions.Blur()
		m.accounts.Blur()
		m.categories.Blur()
		m.expenses.Blur()
		m.prompt.Blur()
	case ViewCategoriesMsg:
		m.state = categoryView
		m.transactions.Blur()
		m.accounts.Blur()
		m.categories.Focus()
		m.expenses.Blur()
		m.prompt.Blur()
	case ViewExpensesMsg:
		m.state = expensesView
		m.transactions.Blur()
		m.accounts.Blur()
		m.categories.Blur()
		m.expenses.Focus()
		m.prompt.Blur()
	case ViewPromptMsg:
		m.state = promptView
		m.transactions.Blur()
		m.accounts.Blur()
		m.categories.Blur()
		m.expenses.Blur()
		m.prompt.Focus()
	case PromptMsg:
		m.prompt = newPrompt(msg)
		cmds = append(cmds, cmd, Cmd(ViewPromptMsg{}))
		return m, tea.Batch(cmds...)

	case RefreshCategoriesMsg:
		m.api.UpdateCategories()
		cmds = append(cmds, m.categories.list.SetItems(getCategoriesItems(m.api)))
		return m, tea.Batch(cmds...)
	case NewCategoryMsg:
		err := m.api.CreateCategory(msg.category, "")
		if err != nil {
			cmds = append(cmds, Cmd(ViewCategoriesMsg{}))
			return m, tea.Batch(cmds...)
		}
		cmds = append(cmds, Cmd(RefreshCategoriesMsg{}))
		return m, tea.Batch(cmds...)

	case RefreshBalanceMsg:
		m.api.UpdateAssets()
		cmds = append(cmds, m.accounts.list.SetItems(getAssetsItems(m.api)))
		return m, tea.Batch(cmds...)
	case RefreshExpensesMsg:
		m.api.UpdateExpenses()
		cmds = append(cmds, m.expenses.list.SetItems(getExpensesItems(m.api)))
		return m, tea.Batch(cmds...)

	case FilterMsg:
		value := msg.query
		if value == "" {
			rows, columns := getRows(m.transactions.transactions)
			m.transactions.table.SetRows(rows)
			m.transactions.table.SetColumns(columns)
			return m, nil
		}
		transactions, err := m.api.SearchTransactions(value)
		if err != nil {
			return m, nil
		}
		rows, columns := getRows(transactions)
		m.transactions.table.SetRows(rows)
		m.transactions.table.SetColumns(columns)
		cmds = append(cmds, tea.Printf("Filtered"))
	case FilterAccountMsg:
		value := msg.account
		m.transactions.currentAccount = value
		if value != "" {
			transactions := []firefly.Transaction{}
			for _, tx := range m.transactions.transactions {
				if tx.Source == value || tx.Destination == value {
					transactions = append(transactions, tx)
				}
			}
			rows, columns := getRows(transactions)
			m.transactions.table.SetRows(rows)
			m.transactions.table.SetColumns(columns)
		} else {
			rows, columns := getRows(m.transactions.transactions)
			m.transactions.table.SetRows(rows)
			m.transactions.table.SetColumns(columns)
		}
	case RefreshTransactionsMsg:
		transactions, err := m.api.ListTransactions("", "", "")
		if err != nil {
			return m, nil
		}
		m.transactions.transactions = transactions
		cmds = append(cmds, Cmd(FilterAccountMsg{account: m.transactions.currentAccount}))

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
	case newView:
		nModel, nCmd := m.new.Update(msg)
		newModel, ok := nModel.(modelNewTransaction)
		if !ok {
			panic("Somthing bad happened")
		}
		m.new = newModel
		cmds = append(cmds, nCmd)
	case promptView:
		nModel, nCmd := m.prompt.Update(msg)
		promptModel, ok := nModel.(modelPrompt)
		if !ok {
			panic("Somthing bad happened")
		}
		cmds = append(cmds, nCmd)
		m.prompt = promptModel
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
	case promptView:
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			baseStyle.Render(m.accounts.View()),
			baseStyle.Render(m.prompt.View()))
	}
	return baseStyle.Render(m.transactions.View()) + "\n"
}
