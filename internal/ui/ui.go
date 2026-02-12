/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"ffiii-tui/internal/ui/notify"
	"ffiii-tui/internal/ui/period"
	"ffiii-tui/internal/ui/prompt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type state uint

var (
	loading        atomic.Int32
	loadingOps     sync.Map
	operationIDSeq atomic.Uint64 // for generating unique operation IDs
)

const (
	transactionsView state = iota
	periodView
	newView
	assetsView
	categoriesView
	expensesView
	revenuesView
	liabilitiesView
	// promptView
)

type (
	ViewFullTransactionViewMsg struct{}
	SetFocusedViewMsg          struct {
		state state
	}
	DataLoadCompletedMsg struct {
		DataType string
	}
	LazyLoadMsg struct {
		t time.Time
		c int
	}
	AllBaseDataLoadedMsg struct{}
	RefreshAllMsg        struct{}
	UpdatePositions      struct {
		layout *LayoutConfig
	}
)

type modelUI struct {
	state        state
	transactions modelTransactions
	api          UIAPI
	new          modelTransaction
	assets       modelAssets
	categories   modelCategories
	expenses     modelExpenses
	revenues     modelRevenues
	liabilities  modelLiabilities
	prompt       prompt.Model
	periodPicker period.Model
	notify       notify.Model
	summary      modelSummary
	spinner      spinner.Model

	keymap UIKeyMap
	help   help.Model
	styles Styles

	Width  int
	layout *LayoutConfig

	loadStatus map[string]bool
}

func Show(api UIAPI) {
	m := NewModelUI(api)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func NewModelUI(api UIAPI) modelUI {
	lc := NewDefaultLayout()
	lc = lc.WithFullTransactionView(viper.GetBool("ui.full_view"))

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	m := modelUI{
		api:          api,
		transactions: NewModelTransactions(api),
		new:          newModelTransaction(api),
		assets:       newModelAssets(api),
		categories:   newModelCategories(api),
		expenses:     newModelExpenses(api),
		revenues:     newModelRevenues(api),
		liabilities:  newModelLiabilities(api),
		prompt:       prompt.New(),
		periodPicker: period.New(),
		notify:       notify.New(),
		summary:      newModelSummary(api),
		spinner:      sp,
		keymap:       DefaultUIKeyMap(),
		help:         help.New(),
		styles:       DefaultStyles(),
		Width:        80,
		layout:       lc,
		loadStatus: map[string]bool{
			"asset":      false,
			"expense":    false,
			"revenue":    false,
			"liability":  false,
			"categories": false,
		},
	}

	m.help.Styles.FullKey = m.styles.HelpFullKey
	m.help.Styles.ShortKey = m.styles.HelpShortKey

	return m
}

func (m modelUI) Init() tea.Cmd {
	return tea.Batch(
		Cmd(RefreshAllMsg{}),
		m.spinner.Tick)
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
			if !m.isAnyInputFocused() {
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
			}
		case key.Matches(msg, m.keymap.PeriodPicker):
			if !m.isAnyInputFocused() {
				return m, period.Open(
					m.api.PeriodStart().Year(),
					m.api.PeriodStart().Month(),
				)
			}
		}
	case period.SelectedMsg:
		m.transactions.currentSearch = ""
		m.api.SetPeriod(msg.Year, msg.Month)
		return m, tea.Batch(
			Cmd(RefreshTransactionsMsg{}),
			Cmd(RefreshSummaryMsg{}),
			Cmd(RefreshCategoryInsightsMsg{}),
			Cmd(RefreshRevenueInsightsMsg{}),
			Cmd(RefreshExpenseInsightsMsg{}),
		)
	case period.CloseMsg:
	case UpdatePositions:
		// TODO: Refactor, bad design
		// Use current layout
		globalWidth := m.layout.Width
		fullView := msg.layout.GetFullTransactionView()
		if msg.layout.Width != 0 {
			globalWidth = msg.layout.Width
		}

		h, _ := m.styles.Base.GetFrameSize()
		m.Width = globalWidth - h

		topSize := 5
		if m.help.ShowAll {
			topSize += lipgloss.Height(m.HelpView())
		}

		leftSize := 0
		tabBarSize := 2
		tabBarWidth := lipgloss.Width(m.tabBar())
		switch m.state {
		case transactionsView, assetsView:
			if !fullView {
				leftSize = max(
					lipgloss.Width(m.assets.View()),
					lipgloss.Width(m.summary.View()),
					tabBarWidth,
				) + h
			} else {
				tabBarSize = 0
			}
		case categoriesView:
			leftSize = max(lipgloss.Width(m.categories.View()), tabBarWidth) + h
		case expensesView:
			leftSize = max(lipgloss.Width(m.expenses.View()), tabBarWidth) + h
		case revenuesView:
			leftSize = max(lipgloss.Width(m.revenues.View()), tabBarWidth) + h
		case liabilitiesView:
			leftSize = max(lipgloss.Width(m.liabilities.View()), tabBarWidth) + h
		}
		m.layout = m.layout.
			WithTopSize(topSize).
			WithLeftSize(leftSize)
		m.layout.TabBarSize = tabBarSize

	case tea.WindowSizeMsg:
		return m, Cmd(UpdatePositions{
			layout: m.layout.WithSize(msg.Width, msg.Height),
		})

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

		m.SetState(msg.state)
		return m, Cmd(UpdatePositions{layout: m.layout})

	case ViewFullTransactionViewMsg:
		viper.Set("ui.full_view", m.layout.ToggleFullTransactionView())
		return m, Cmd(UpdatePositions{layout: m.layout})
	case DataLoadCompletedMsg:
		m.loadStatus[msg.DataType] = true
	case LazyLoadMsg:
		c := msg.c - 1
		for _, loaded := range m.loadStatus {
			if !loaded {
				if c <= 0 {
					return m, notify.NotifyWarn("Could not load all resources")
				}
				return m, tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
					return LazyLoadMsg{t: t, c: c}
				})
			}
		}
		return m, tea.Batch(
			Cmd(RefreshTransactionsMsg{}),
			Cmd(RefreshSummaryMsg{}))
	case RefreshAllMsg:
		m.loadStatus = map[string]bool{
			"asset":      false,
			"expense":    false,
			"revenue":    false,
			"liability":  false,
			"categories": false,
		}
		return m, tea.Batch(
			SetView(transactionsView),
			tea.WindowSize(),
			Cmd(RefreshAssetsMsg{}),
			Cmd(RefreshLiabilitiesMsg{}),
			Cmd(RefreshExpensesMsg{}),
			Cmd(RefreshRevenuesMsg{}),
			Cmd(RefreshCategoriesMsg{}),
			tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
				return LazyLoadMsg{
					t: t,
					c: m.api.TimeoutSeconds(),
				}
			}),
		)
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.prompt, cmd = updateModel(m.prompt, msg)
	cmds = append(cmds, cmd)
	if m.prompt.Focused() {
		return m, tea.Batch(cmds...)
	}

	periodPickerWasFocused := m.periodPicker.Focused()
	m.periodPicker, cmd = updateModel(m.periodPicker, msg)
	cmds = append(cmds, cmd)
	if periodPickerWasFocused {
		return m, tea.Batch(cmds...)
	}

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

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m modelUI) View() string {
	// TODO: Refactor, too complicated
	var s strings.Builder

	// TODO: Move to model
	if m.prompt.Focused() {
		s.WriteString(m.prompt.WithWidth(m.layout.GetWidth()).View() + "\n")
	} else if m.periodPicker.Focused() {
		s.WriteString(m.periodPicker.WithWidth(m.layout.GetWidth()).View() + "\n")
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
				header = header + fmt.Sprintf(" | p %s %d",
					m.api.PeriodStart().Month(),
					m.api.PeriodStart().Year())
			}
			if !m.transactions.currentAccount.IsEmpty() {
				header = header + " | Account: " + m.transactions.currentAccount.Name
			}
			if !m.transactions.currentCategory.IsEmpty() {
				header = header + " | Category: " + m.transactions.currentCategory.Name
			}
			if m.transactions.currentFilter != "" {
				header = header + " | Filter: " + m.transactions.currentFilter
			}
		}

		if loading.Load() > 0 {
			msg := buildLoadingMessage()
			header += " | " + m.spinner.View() + msg
		}
		s.WriteString(headerRenderer.Width(m.Width).Render(header) + "\n")
	}

	switch m.state {
	case transactionsView:
		if m.layout.GetFullTransactionView() {
			s.WriteString(m.styles.BaseFocused.Render(m.transactions.View()))
		} else {
			s.WriteString(lipgloss.JoinHorizontal(lipgloss.Top,
				m.styles.Base.Render(
					lipgloss.JoinVertical(lipgloss.Left, m.tabBar(), m.summary.View(), m.assets.View())),
				m.styles.BaseFocused.Render(m.transactions.View())))
		}
	case assetsView:
		s.WriteString(lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.styles.BaseFocused.Render(
				lipgloss.JoinVertical(lipgloss.Left, m.tabBar(), m.summary.View(), m.assets.View())),
			m.styles.Base.Render(m.transactions.View())))
	case categoriesView:
		s.WriteString(lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.styles.BaseFocused.Render(
				lipgloss.JoinVertical(lipgloss.Left, m.tabBar(), m.categories.View())),
			m.styles.Base.Render(m.transactions.View())))
	case expensesView:
		s.WriteString(lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.styles.BaseFocused.Render(
				lipgloss.JoinVertical(lipgloss.Left, m.tabBar(), m.expenses.View())),
			m.styles.Base.Render(m.transactions.View())))
	case revenuesView:
		s.WriteString(lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.styles.BaseFocused.Render(
				lipgloss.JoinVertical(lipgloss.Left, m.tabBar(), m.revenues.View())),
			m.styles.Base.Render(m.transactions.View())))
	case liabilitiesView:
		s.WriteString(lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.styles.BaseFocused.Render(
				lipgloss.JoinVertical(lipgloss.Left, m.tabBar(), m.liabilities.View())),
			m.styles.Base.Render(m.transactions.View())))
	case newView:
		s.WriteString(lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.styles.Base.Render(
				lipgloss.JoinVertical(lipgloss.Left, m.summary.View(), m.assets.View())),
			m.styles.BaseFocused.Render(m.new.View())))
	}
	s.WriteString("\n")

	s.WriteString(m.notify.WithWidth(m.layout.GetWidth()).View() + "\n")
	s.WriteString(m.help.Styles.ShortKey.Render(m.HelpView()))

	return s.String()
}

func (m *modelUI) HelpView() string {
	help := ""
	switch m.state {
	case transactionsView:
		help += m.help.View(m.transactions.keymap)
	case assetsView:
		help += m.help.View(m.assets.keymap)
	case expensesView:
		help += m.help.View(m.expenses.keymap)
	case revenuesView:
		help += m.help.View(m.revenues.keymap)
	case liabilitiesView:
		help += m.help.View(m.liabilities.keymap)
	case categoriesView:
		help += m.help.View(m.categories.keymap)
	case newView:
		help += m.help.View(m.new.keymap)
	}
	if m.help.ShowAll {
		help = lipgloss.JoinHorizontal(lipgloss.Left, help, m.help.View(m.keymap))
	}
	return help
}

func (m *modelUI) tabBar() string {
	type tab struct {
		key   string
		label string
		state state
	}

	tabs := []tab{
		{"a", "Assets", assetsView},
		{"c", "Categ.", categoriesView},
		{"e", "Expns.", expensesView},
		{"i", "Revnu.", revenuesView},
		{"o", "Liab.", liabilitiesView},
	}

	var parts []string
	for _, t := range tabs {
		active := m.state == t.state ||
			(m.state == transactionsView && t.state == assetsView)
		label := t.key + " " + t.label
		if active {
			parts = append(parts, m.styles.TabActive.Render(label))
		} else {
			parts = append(parts, m.styles.TabInactive.Render(label))
		}
	}

	return strings.Join(parts, m.styles.TabInactive.Render(" ")) + "\n"
}

func (m *modelUI) SetState(s state) {
	m.state = s
}

func (m *modelUI) isAnyInputFocused() bool {
	return m.prompt.Focused() ||
		m.new.Focused() ||
		m.assets.list.FilterInput.Focused() ||
		m.expenses.list.FilterInput.Focused() ||
		m.revenues.list.FilterInput.Focused() ||
		m.categories.list.FilterInput.Focused() ||
		m.liabilities.list.FilterInput.Focused()
}

func SetView(state state) tea.Cmd {
	return Cmd(SetFocusedViewMsg{state: state})
}

func startLoading(message string) string {
	for {
		current := loading.Load()
		if current >= 100 {
			return "" // Max operations reached
		}
		if loading.CompareAndSwap(current, current+1) {
			break
		}
	}

	// Generate unique operation ID
	opID := fmt.Sprintf("op_%d", operationIDSeq.Add(1))

	loadingOps.Store(opID, message)

	return opID
}

func stopLoading(opID string) {
	if opID == "" {
		return // Invalid operation ID
	}

	loadingOps.Delete(opID)

	for {
		current := loading.Load()
		if current <= 0 {
			return
		}
		if loading.CompareAndSwap(current, current-1) {
			return
		}
	}
}

func buildLoadingMessage() string {
	var messages []string

	loadingOps.Range(func(key, value any) bool {
		if msg, ok := value.(string); ok {
			abbrev := msg

			if len(abbrev) > 25 {
				abbrev = abbrev[:22] + "..."
			}

			messages = append(messages, abbrev)
		}
		return true
	})

	if len(messages) == 0 {
		return "..."
	}

	const maxDisplay = 5
	count := len(messages)

	if count == 1 {
		return messages[0]
	}

	if count <= maxDisplay {
		return fmt.Sprintf("(%d) %s", count, strings.Join(messages, " "))
	}

	// Show first messages + remaining count
	shown := messages[:maxDisplay]
	remaining := count - maxDisplay
	return fmt.Sprintf("(%d) %s | +%d more", count, strings.Join(shown, " "), remaining)
}
