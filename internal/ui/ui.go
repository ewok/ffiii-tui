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
	topSize             = 4
	leftSize            = 10
	summarySize         = 15
	fullTransactionView = false
	lazyLoadCounter     = 0

	globalWidth  = 0
	globalHeight = 0
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
	DataLoadCompletedMsg struct {
		DataType string
	}
	LazyLoadMsg          time.Time
	AllBaseDataLoadedMsg struct{}
	RefreshAllMsg        struct{}
	UpdatePositions      struct{}
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
	summary      modelSummary

	keymap UIKeyMap
	help   help.Model
	styles Styles

	Width int

	loadStatus map[string]bool
}

func Show(api *firefly.Api) {
	fullTransactionView = viper.GetBool("ui.full_view")

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
		notify:  newNotify(),
		summary: newModelSummary(api),
		keymap:  DefaultUIKeyMap(),
		help:    help.New(),
		styles:  DefaultStyles(),
		Width:   80,
		loadStatus: map[string]bool{
			"assets":      false,
			"expenses":    false,
			"revenues":    false,
			"liabilities": false,
			"categories":  false,
		},
	}

	m.help.Styles.FullKey = m.styles.HelpFullKey
	m.help.Styles.ShortKey = m.styles.HelpShortKey

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func (m modelUI) Init() tea.Cmd {
	return Cmd(RefreshAllMsg{})
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
	// zap.S().Debugf("UI Update: %+v", msg)

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
			m.transactions.currentSearch = ""
			m.api.PreviousPeriod()
			return m, tea.Batch(
				Cmd(RefreshTransactionsMsg{}),
				Cmd(RefreshSummaryMsg{}),
				Cmd(RefreshCategoryInsightsMsg{}),
				Cmd(RefreshRevenueInsightsMsg{}),
				Cmd(RefreshExpenseInsightsMsg{}),
			)
		case key.Matches(msg, m.keymap.NextPeriod):
			m.transactions.currentSearch = ""
			m.api.NextPeriod()
			return m, tea.Batch(
				Cmd(RefreshTransactionsMsg{}),
				Cmd(RefreshSummaryMsg{}),
				Cmd(RefreshCategoryInsightsMsg{}),
				Cmd(RefreshRevenueInsightsMsg{}),
				Cmd(RefreshExpenseInsightsMsg{}),
			)
		}
	case UpdatePositions:
		if m.help.ShowAll {
			topSize = 4 + lipgloss.Height(m.HelpView())
		} else {
			topSize = 4
		}

		h, _ := m.styles.Base.GetFrameSize()

		switch m.state {
		case transactionsView, assetsView:
			if fullTransactionView {
				leftSize = globalWidth
			} else {
				leftSize = max(
					lipgloss.Width(m.assets.View()),
					lipgloss.Width(m.summary.View()),
				) + h
			}
		case categoriesView:
			leftSize = lipgloss.Width(m.categories.View()) + h
		case expensesView:
			leftSize = lipgloss.Width(m.expenses.View()) + h
		case revenuesView:
			leftSize = lipgloss.Width(m.revenues.View()) + h
		case liabilitiesView:
			leftSize = lipgloss.Width(m.liabilities.View()) + h
		}
		m.Width = globalWidth - h
	case tea.WindowSizeMsg:
		globalWidth = msg.Width
		globalHeight = msg.Height
		zap.L().Debug(
			"Window size update: ",
			zap.Int("width", globalWidth),
			zap.Int("height", globalHeight))
		return m, Cmd(UpdatePositions{})

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
	case DataLoadCompletedMsg:
		m.loadStatus[msg.DataType] = true
	case LazyLoadMsg:
		lazyLoadCounter++
		for _, loaded := range m.loadStatus {
			if !loaded {
				if lazyLoadCounter > m.api.Config.TimeoutSeconds {
					lazyLoadCounter = 0
					return m, Notify("Could not load all resources", Warn)
				}
				return m, tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
					return LazyLoadMsg(t)
				})
			}
		}
		lazyLoadCounter = 0
		return m, tea.Batch(
			Cmd(RefreshTransactionsMsg{}),
			Cmd(RefreshSummaryMsg{}))
	case RefreshAllMsg:
		m.loadStatus = map[string]bool{
			"assets":      false,
			"expenses":    false,
			"revenues":    false,
			"liabilities": false,
			"categories":  false,
		}
		return m, tea.Batch(
			SetView(transactionsView),
			tea.WindowSize(),
			Cmd(RefreshAssetsMsg{}),
			Cmd(RefreshLiabilitiesMsg{}),
			Cmd(RefreshExpensesMsg{}),
			Cmd(RefreshRevenuesMsg{}),
			Cmd(RefreshCategoriesMsg{}),
			tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
				return LazyLoadMsg(t)
			}),
		)
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.prompt, cmd = updateModel(m.prompt, msg)
	cmds = append(cmds, cmd)

	m.notify, cmd = updateModel(m.notify, msg)
	cmds = append(cmds, cmd)

	m.summary, cmd = updateModel(m.summary, msg)
	cmds = append(cmds, cmd)

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

	return m, tea.Batch(cmds...)
}

func (m modelUI) View() string {
	// TODO: Refactor, too complicated
	s := ""

	// TODO: Move to model
	if m.prompt.focus {
		s = s + m.styles.PromptFocused.Render(m.prompt.View()) + "\n"
	} else if m.notify.text != "" {
		s = s + m.styles.PromptFocused.Render(m.notify.View()) + "\n"
	} else {
		header := " ffiii-tui"

		headerRenderer := m.styles.Prompt

		if m.state == newView {
			if m.new.new {
				header = header + " | New transaction"
				headerRenderer = m.styles.PromptNewTr
			} else {
				header = header + " | Editing transaction: " + m.new.attr.trxID
				headerRenderer = m.styles.PromptEditTr
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
		s = s + headerRenderer.Width(m.Width).Render(header) + "\n"
	}

	switch m.state {
	case transactionsView:
		if fullTransactionView {
			s = s + m.styles.BaseFocused.Render(m.transactions.View())
		} else {
			s = s + lipgloss.JoinHorizontal(lipgloss.Top,
				m.styles.Base.Render(
					lipgloss.JoinVertical(lipgloss.Left, m.summary.View(), m.assets.View())),
				m.styles.BaseFocused.Render(m.transactions.View()))
		}
	case assetsView:
		s = s + lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.styles.BaseFocused.Render(
				lipgloss.JoinVertical(lipgloss.Left, m.summary.View(), m.assets.View())),
			m.styles.Base.Render(m.transactions.View()))
	case categoriesView:
		s = s + lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.styles.BaseFocused.Render(m.categories.View()),
			m.styles.Base.Render(m.transactions.View()))
	case expensesView:
		s = s + lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.styles.BaseFocused.Render(m.expenses.View()),
			m.styles.Base.Render(m.transactions.View()))
	case revenuesView:
		s = s + lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.styles.BaseFocused.Render(m.revenues.View()),
			m.styles.Base.Render(m.transactions.View()))
	case liabilitiesView:
		s = s + lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.styles.BaseFocused.Render(m.liabilities.View()),
			m.styles.Base.Render(m.transactions.View()))
	case newView:
		s = s + lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.styles.Base.Render(
				lipgloss.JoinVertical(lipgloss.Left, m.summary.View(), m.assets.View())),
			m.styles.BaseFocused.Render(m.new.View()))
	}

	s = s + "\n" + m.help.Styles.ShortKey.Render(m.HelpView())

	return s
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
	return tea.Sequence(
		Cmd(SetFocusedViewMsg{state: state}),
		Cmd(UpdatePositions{}))
}
