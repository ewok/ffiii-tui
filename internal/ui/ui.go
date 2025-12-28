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
	"github.com/spf13/viper"
)

var (
	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2)
	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(0).
				Foreground(lipgloss.Color("170"))
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
	baseStyleFocused = baseStyle.
				BorderForeground(lipgloss.Color("62")).
				BorderStyle(lipgloss.ThickBorder())
	promptStyle         = baseStyle
	promptStyleFocused  = baseStyleFocused
	topSize             = 3
	promptVisible       = false
	fullTransactionView = false
)

type state uint

const (
	transactionView state = iota
	periodView
	newView
	assetView
	categoryView
	expensesView
	revenuesView
)

type (
	ViewTransactionsMsg        struct{}
	ViewAssetsMsg              struct{}
	ViewNewMsg                 struct{}
	ViewCategoriesMsg          struct{}
	ViewExpensesMsg            struct{}
	ViewRevenuesMsg            struct{}
	ViewPromptMsg              struct{}
	ViewFullTransactionViewMsg struct{}
)

type modelUI struct {
	state        state
	transactions modelTransactions
	api          *firefly.Api
	new          modelNewTransaction
	assets       modelAssets
	categories   modelCategories
	expenses     modelExpenses
	revenues     modelRevenues
	prompt       modelPrompt
}

func Show(api *firefly.Api) {

	fullTransactionView = viper.GetBool("ui.full_view")

	m := modelUI{
		api:          api,
		transactions: newModelTransactions(api),
		new:          newModelNewTransaction(api, firefly.Transaction{}),
		assets:       newModelAssets(api),
		categories:   newModelCategories(api),
		expenses:     newModelExpenses(api),
		revenues:     newModelRevenues(api),
		prompt: newPrompt(PromptMsg{
			Prompt: "",
			Value:  "",
			Callback: func(value string) tea.Cmd {
				return Cmd(ViewTransactionsMsg{})
			}}),
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

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		promptStyle = baseStyle.
			Width(msg.Width - 2)
		promptStyleFocused = baseStyleFocused.
			BorderForeground(lipgloss.Color("#FF5555")).
			Width(msg.Width - 2)
	case ViewTransactionsMsg:
		m.state = transactionView
		promptVisible = false
		m.transactions.Focus()
		m.assets.Blur()
		m.categories.Blur()
		m.expenses.Blur()
		m.revenues.Blur()
		m.prompt.Blur()
		m.new.Blur()
	case ViewAssetsMsg:
		m.state = assetView
		promptVisible = false
		m.transactions.Blur()
		m.assets.Focus()
		m.categories.Blur()
		m.expenses.Blur()
		m.revenues.Blur()
		m.prompt.Blur()
		m.new.Blur()
	case ViewNewMsg:
		m.state = newView
		promptVisible = false
		m.transactions.Blur()
		m.assets.Blur()
		m.categories.Blur()
		m.expenses.Blur()
		m.revenues.Blur()
		m.prompt.Blur()
		m.new.Focus()
	case ViewCategoriesMsg:
		m.state = categoryView
		promptVisible = false
		m.transactions.Blur()
		m.assets.Blur()
		m.categories.Focus()
		m.expenses.Blur()
		m.revenues.Blur()
		m.prompt.Blur()
	case ViewExpensesMsg:
		m.state = expensesView
		promptVisible = false
		m.transactions.Blur()
		m.assets.Blur()
		m.categories.Blur()
		m.expenses.Focus()
		m.revenues.Blur()
		m.prompt.Blur()
		m.new.Blur()
	case ViewRevenuesMsg:
		m.state = revenuesView
		promptVisible = false
		m.transactions.Blur()
		m.assets.Blur()
		m.categories.Blur()
		m.expenses.Blur()
		m.revenues.Focus()
		m.prompt.Blur()
		m.new.Blur()
	case ViewPromptMsg:
		promptVisible = true
		m.transactions.Blur()
		m.assets.Blur()
		m.categories.Blur()
		m.expenses.Blur()
		m.revenues.Blur()
		m.prompt.Focus()
		m.new.Blur()
	case ViewFullTransactionViewMsg:
		fullTransactionView = !fullTransactionView
		viper.Set("ui.full_view", fullTransactionView)
	}

	var cmds []tea.Cmd

	nModel, cmd := m.transactions.Update(msg)
	listModel, ok := nModel.(modelTransactions)
	if !ok {
		panic("Somthing bad happened")
	}
	m.transactions = listModel
	cmds = append(cmds, cmd)

	nModel, cmd = m.assets.Update(msg)
	assetsModel, ok := nModel.(modelAssets)
	if !ok {
		panic("Somthing bad happened")
	}
	m.assets = assetsModel
	cmds = append(cmds, cmd)

	nModel, cmd = m.categories.Update(msg)
	categoryModel, ok := nModel.(modelCategories)
	if !ok {
		panic("Somthing bad happened")
	}
	m.categories = categoryModel
	cmds = append(cmds, cmd)

	nModel, cmd = m.expenses.Update(msg)
	expensesModel, ok := nModel.(modelExpenses)
	if !ok {
		panic("Somthing bad happened")
	}
	m.expenses = expensesModel
	cmds = append(cmds, cmd)

	nModel, cmd = m.revenues.Update(msg)
	revenuesModel, ok := nModel.(modelRevenues)
	if !ok {
		panic("Somthing bad happened")
	}
	m.revenues = revenuesModel
	cmds = append(cmds, cmd)

	nModel, cmd = m.new.Update(msg)
	newModel, ok := nModel.(modelNewTransaction)
	if !ok {
		panic("Somthing bad happened")
	}
	m.new = newModel
	cmds = append(cmds, cmd)

	nModel, cmd = m.prompt.Update(msg)
	promptModel, ok := nModel.(modelPrompt)
	if !ok {
		panic("Somthing bad happened")
	}
	m.prompt = promptModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m modelUI) View() string {
	s := ""

	if promptVisible {
		s = s + promptStyleFocused.Render(" "+m.prompt.View()) + "\n"
	} else {
		header := " ffiii-tui"
		if m.transactions.currentItem != "" {
			header = header + " | Item: " + m.transactions.currentItem
		}
		if m.transactions.currentFilter != "" {
			header = header + " | Filter: " + m.transactions.currentFilter
		}
		if m.transactions.currentSearch != "" {
			header = header + " | Search: " + m.transactions.currentSearch
		}
		s = s + promptStyle.Render(header) + "\n"
	}

	switch m.state {
	case transactionView:
		if fullTransactionView {
			s = s + baseStyleFocused.Render(m.transactions.View())
		} else {
			s = s + lipgloss.JoinHorizontal(lipgloss.Top,
				baseStyle.Render(m.assets.View()),
				baseStyleFocused.Render(m.transactions.View()))
		}
	case assetView:
		s = s + lipgloss.JoinHorizontal(
			lipgloss.Top,
			baseStyleFocused.Render(m.assets.View()),
			baseStyle.Render(m.transactions.View()))
	case newView:
		if fullTransactionView {
			s = s + baseStyleFocused.Render(m.new.View())
		} else {
			s = s + lipgloss.JoinHorizontal(
				lipgloss.Top,
				baseStyle.Render(m.assets.View()),
				baseStyleFocused.Render(m.new.View()))
		}
	case categoryView:
		s = s + lipgloss.JoinHorizontal(
			lipgloss.Top,
			baseStyleFocused.Render(m.categories.View()),
			baseStyle.Render(m.transactions.View()))
	case expensesView:
		s = s + lipgloss.JoinHorizontal(
			lipgloss.Top,
			baseStyleFocused.Render(m.expenses.View()),
			baseStyle.Render(m.transactions.View()))
	case revenuesView:
		s = s + lipgloss.JoinHorizontal(
			lipgloss.Top,
			baseStyleFocused.Render(m.revenues.View()),
			baseStyle.Render(m.transactions.View()))
	}
	return s
}
