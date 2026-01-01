/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"ffiii-tui/internal/firefly"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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
	topSize             = 4
	promptVisible       = false
	fullTransactionView = false
)

type state uint

const (
	transactionsView state = iota
	periodView
	newView
	assetsView
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
	notify       modelNotify

	keymap UIKeyMap
	help   help.Model
}

func Show(api *firefly.Api) {

	fullTransactionView = viper.GetBool("ui.full_view")
	h := help.New()
	h.Styles.FullKey = lipgloss.NewStyle().PaddingLeft(1)
	h.Styles.ShortKey = lipgloss.NewStyle().PaddingLeft(1)

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
		notify: newNotify(NotifyMsg{Message: ""}),
		keymap: DefaultUIKeyMap(),
		help:   h,
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
		switch {
		case key.Matches(msg, m.keymap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymap.ShowShortHelp):
			m.help.ShowAll = !m.help.ShowAll
			m.assets.list.Help.ShowAll = m.help.ShowAll
			m.categories.list.Help.ShowAll = m.help.ShowAll
			m.expenses.list.Help.ShowAll = m.help.ShowAll
			m.revenues.list.Help.ShowAll = m.help.ShowAll

			m.assets.list.SetShowHelp(m.help.ShowAll)
			m.categories.list.SetShowHelp(m.help.ShowAll)
			m.expenses.list.SetShowHelp(m.help.ShowAll)
			m.revenues.list.SetShowHelp(m.help.ShowAll)
			return m, tea.WindowSize()
		case key.Matches(msg, m.keymap.PreviousPeriod):
			m.api.PreviousPeriod()
			return m, tea.Batch(
				Cmd(RefreshTransactionsMsg{}),
				Cmd(RefreshCategoryInsightsMsg{}),
				Cmd(RefreshRevenueInsightsMsg{}),
				Cmd(RefreshExpenseInsightsMsg{}),
			)
		case key.Matches(msg, m.keymap.NextPeriod):
			m.api.NextPeriod()
			return m, tea.Batch(
				Cmd(RefreshTransactionsMsg{}),
				Cmd(RefreshCategoryInsightsMsg{}),
				Cmd(RefreshRevenueInsightsMsg{}),
				Cmd(RefreshExpenseInsightsMsg{}),
			)
		}
	case tea.WindowSizeMsg:
		promptStyle = baseStyle.
			Width(msg.Width - 2)
		promptStyleFocused = baseStyleFocused.
			BorderForeground(lipgloss.Color("#FF5555")).
			Width(msg.Width - 2)
		if m.help.ShowAll {
			topSize = 4 + lipgloss.Height(m.HelpView())
		} else {
			topSize = 4
		}

	case ViewTransactionsMsg:
		m.SetState(transactionsView)
		promptVisible = false
		m.transactions.Focus()
		m.assets.Blur()
		m.categories.Blur()
		m.expenses.Blur()
		m.revenues.Blur()
		m.prompt.Blur()
		m.new.Blur()
	case ViewAssetsMsg:
		m.SetState(assetsView)
		promptVisible = false
		m.transactions.Blur()
		m.assets.Focus()
		m.categories.Blur()
		m.expenses.Blur()
		m.revenues.Blur()
		m.prompt.Blur()
		m.new.Blur()
	case ViewNewMsg:
		m.SetState(newView)
		promptVisible = false
		m.transactions.Blur()
		m.assets.Blur()
		m.categories.Blur()
		m.expenses.Blur()
		m.revenues.Blur()
		m.prompt.Blur()
		m.new.Focus()
	case ViewCategoriesMsg:
		m.SetState(categoryView)
		promptVisible = false
		m.transactions.Blur()
		m.assets.Blur()
		m.categories.Focus()
		m.expenses.Blur()
		m.revenues.Blur()
		m.prompt.Blur()
	case ViewExpensesMsg:
		m.SetState(expensesView)
		promptVisible = false
		m.transactions.Blur()
		m.assets.Blur()
		m.categories.Blur()
		m.expenses.Focus()
		m.revenues.Blur()
		m.prompt.Blur()
		m.new.Blur()
	case ViewRevenuesMsg:
		m.SetState(revenuesView)
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

	nModel, cmd = m.notify.Update(msg)
	notifyModel, ok := nModel.(modelNotify)
	if !ok {
		panic("Somthing bad happened")
	}
	m.notify = notifyModel
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m modelUI) View() string {
	s := ""

	if promptVisible {
		s = s + promptStyleFocused.Render(" "+m.prompt.View()) + "\n"
	} else if m.notify.text != "" {
		s = s + promptStyleFocused.Render(" Notification: "+m.notify.View()) + "\n"
	} else {
		header := " ffiii-tui"

		if m.transactions.currentSearch != "" {
			header = header + " | Search: " + m.transactions.currentSearch
		} else {
			header = header + fmt.Sprintf(" | Period: %s - %s",
				m.api.StartDate.Format(time.DateOnly),
				m.api.EndDate.Format(time.DateOnly))

		}
		if m.transactions.currentAccount != "" {
			header = header + " | Account: " + m.transactions.currentAccount
		}
		if m.transactions.currentCategory != "" {
			header = header + " | Category: " + m.transactions.currentCategory
		}
		if m.transactions.currentFilter != "" {
			header = header + " | Filter: " + m.transactions.currentFilter
		}
		if m.notify.text != "" {
			header = header + "\n" + " Notification: " + m.notify.View()
		}
		s = s + promptStyle.Render(header) + "\n"
	}

	switch m.state {
	case transactionsView:
		if fullTransactionView {
			s = s + baseStyleFocused.Render(m.transactions.View())
		} else {
			s = s + lipgloss.JoinHorizontal(lipgloss.Top,
				baseStyle.Render(m.assets.View()),
				baseStyleFocused.Render(m.transactions.View()))
		}
	case assetsView:
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
	return s + "\n" + m.help.Styles.ShortKey.Render(m.HelpView())

}

func (m *modelUI) SetState(s state) {
	m.state = s
}

func (m *modelUI) HelpView() string {
	help := ""
	switch m.state {
	case transactionsView:
		help += m.help.View(m.transactions.keymap)
	case assetsView:
		help += m.help.View(m.assets.keymap)
	case categoryView:
		help += m.help.View(m.categories.keymap)
	case expensesView:
		help += m.help.View(m.expenses.keymap)
	case revenuesView:
		help += m.help.View(m.revenues.keymap)
	case newView:
		help += m.help.View(m.new.keymap)
	}
	if m.help.ShowAll {
		help = lipgloss.JoinHorizontal(lipgloss.Left, help, m.help.View(m.keymap))
	}
	return help
}
