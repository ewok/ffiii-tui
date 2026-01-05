/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"fmt"
	"os"
	"time"

	"ffiii-tui/internal/firefly"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
	promptStyleFocused  = baseStyleFocused.BorderForeground(lipgloss.Color("#FF5555"))
	promptStyleNewTr    = baseStyle.BorderForeground(lipgloss.Color("34"))
	promptStyleEditTr   = baseStyle.BorderForeground(lipgloss.Color("214"))
	topSize             = 4
	fullTransactionView = false
)

type state uint

const (
	transactionsView state = iota
	periodView
	newView
	assetsView
	categoriesView
	expensesView
	revenuesView
	liabilitiesView
	promptView
)

type (
	ViewFullTransactionViewMsg struct{}
	SetFocusedViewMsg          struct {
		state state
	}
)

type modelUI struct {
	state        state
	transactions modelTransactions
	api          *firefly.Api
	new          modelTransaction
	assets       modelAssets
	categories   modelCategories
	expenses     modelExpenses
	revenues     modelRevenues
	liabilities  modelLiabilities
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
		new:          newModelTransaction(api, firefly.Transaction{}, true),
		assets:       newModelAssets(api),
		categories:   newModelCategories(api),
		expenses:     newModelExpenses(api),
		revenues:     newModelRevenues(api),
		liabilities:  newModelLiabilities(api),
		prompt: newPrompt(PromptMsg{
			Prompt: "",
			Value:  "",
			Callback: func(value string) tea.Cmd {
				return Cmd(SetFocusedViewMsg{state: transactionsView})
			},
		}),
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
	return SetView(transactionsView)
}

func updateModel[T tea.Model](current T, msg tea.Msg) (T, tea.Cmd) {
	model, cmd := current.Update(msg)
	if converted, ok := model.(T); ok {
		return converted, cmd
	}
	zap.S().Errorf("Failed to update model: type assertion failed for %T", current)
	return current, cmd
}

func (m modelUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	zap.S().Debugf("UI Update: %+v", msg)

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
		promptStyle = promptStyle.
			Width(msg.Width - 2)
		promptStyleFocused = promptStyleFocused.
			Width(msg.Width - 2)
		promptStyleNewTr = promptStyleNewTr.
			Width(msg.Width - 2)
		promptStyleEditTr = promptStyleEditTr.
			Width(msg.Width - 2)
		if m.help.ShowAll {
			topSize = 4 + lipgloss.Height(m.HelpView())
		} else {
			topSize = 4
		}

	case SetFocusedViewMsg:
		if msg.state == transactionsView {
			m.transactions.Focus()
		} else {
			m.transactions.Blur()
		}
		if msg.state == assetsView {
			m.assets.Focus()
		} else {
			m.assets.Blur()
		}
		if msg.state == categoriesView {
			m.categories.Focus()
		} else {
			m.categories.Blur()
		}
		if msg.state == expensesView {
			m.expenses.Focus()
		} else {
			m.expenses.Blur()
		}
		if msg.state == revenuesView {
			m.revenues.Focus()
		} else {
			m.revenues.Blur()
		}
		if msg.state == liabilitiesView {
			m.liabilities.Focus()
		} else {
			m.liabilities.Blur()
		}
		if msg.state == newView {
			m.new.Focus()
		} else {
			m.new.Blur()
		}
		if msg.state == promptView {
			m.prompt.Focus()
		} else {
			m.prompt.Blur()
			m.SetState(msg.state)
		}

	case ViewFullTransactionViewMsg:
		fullTransactionView = !fullTransactionView
		viper.Set("ui.full_view", fullTransactionView)
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.transactions, cmd = updateModel(m.transactions, msg)
	cmds = append(cmds, cmd)

	m.assets, cmd = updateModel(m.assets, msg)
	cmds = append(cmds, cmd)

	m.categories, cmd = updateModel(m.categories, msg)
	cmds = append(cmds, cmd)

	m.expenses, cmd = updateModel(m.expenses, msg)
	cmds = append(cmds, cmd)

	m.revenues, cmd = updateModel(m.revenues, msg)
	cmds = append(cmds, cmd)

	m.liabilities, cmd = updateModel(m.liabilities, msg)
	cmds = append(cmds, cmd)

	m.new, cmd = updateModel(m.new, msg)
	cmds = append(cmds, cmd)

	m.prompt, cmd = updateModel(m.prompt, msg)
	cmds = append(cmds, cmd)

	m.notify, cmd = updateModel(m.notify, msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m modelUI) View() string {
	// TODO: Refactor, too complicated
	s := ""

	if m.prompt.focus {
		s = s + promptStyleFocused.Render(" "+m.prompt.View()) + "\n"
	} else if m.notify.text != "" {
		s = s + promptStyleFocused.Render(" Notification: "+m.notify.View()) + "\n"
	} else {
		header := " ffiii-tui"

		headerRenderer := promptStyle

		if m.state == newView {
			if m.new.new {
				header = header + " | New transaction"
				headerRenderer = promptStyleNewTr
			} else {
				header = header + " | Editing transaction: " + m.new.attr.trxID
				headerRenderer = promptStyleEditTr
			}
		} else {
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
		}
		s = s + headerRenderer.Render(header) + "\n"
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
	case categoriesView:
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
	case liabilitiesView:
		s = s + lipgloss.JoinHorizontal(
			lipgloss.Top,
			baseStyleFocused.Render(m.liabilities.View()),
			baseStyle.Render(m.transactions.View()))
	case newView:
		s = s + lipgloss.JoinHorizontal(
			lipgloss.Top,
			baseStyle.Render(m.assets.View()),
			baseStyleFocused.Render(m.new.View()))
	}
	return s + "\n" + m.help.Styles.ShortKey.Render(m.HelpView())
}

func (m *modelUI) HelpView() string {
	help := ""
	switch m.state {
	case transactionsView:
		help += m.help.View(m.transactions.keymap)
	case assetsView:
		help += m.help.View(m.assets.keymap)
	case categoriesView:
		help += m.help.View(m.categories.keymap)
	case expensesView:
		help += m.help.View(m.expenses.keymap)
	case revenuesView:
		help += m.help.View(m.revenues.keymap)
	case liabilitiesView:
		help += m.help.View(m.liabilities.keymap)
	case newView:
		help += m.help.View(m.new.keymap)
	}
	if m.help.ShowAll {
		help = lipgloss.JoinHorizontal(lipgloss.Left, help, m.help.View(m.keymap))
	}
	return help
}

func (m *modelUI) SetState(s state) {
	m.state = s
}

func SetView(state state) tea.Cmd {
	return Cmd(SetFocusedViewMsg{state: state})
}
