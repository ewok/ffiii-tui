/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"fmt"
	"regexp"
	"strings"

	"ffiii-tui/internal/firefly"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var promptValue string

type (
	RefreshLiabilitiesMsg struct{}
	LiabilitiesUpdateMsg  struct{}
)

type NewLiabilityMsg struct {
	Account   string
	Currency  string
	Type      string
	Direction string
}

type liabilityItem struct {
	account, currency string
	balance           float64
}

func (i liabilityItem) Title() string { return i.account }
func (i liabilityItem) Description() string {
	return fmt.Sprintf("Balance: %.2f %s", i.balance, i.currency)
}
func (i liabilityItem) FilterValue() string { return i.account }

type modelLiabilities struct {
	list   list.Model
	api    *firefly.Api
	focus  bool
	keymap LiabilityKeyMap
}

func newModelLiabilities(api *firefly.Api) modelLiabilities {
	items := getLiabilitiesItems(api)

	m := modelLiabilities{
		list:   list.New(items, list.NewDefaultDelegate(), 0, 0),
		api:    api,
		keymap: DefaultLiabilityKeyMap(),
	}
	m.list.Title = "Liabilities"
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowHelp(false)
	m.list.DisableQuitKeybindings()

	return m
}

func (m modelLiabilities) Init() tea.Cmd {
	return nil
}

func (m modelLiabilities) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case RefreshLiabilitiesMsg:
		return m, func() tea.Msg {
            err := m.api.UpdateAccounts("liabilities")
            if err != nil {
                return Notify(err.Error(), Warning)
            }
			return LiabilitiesUpdateMsg{}
		}
	case LiabilitiesUpdateMsg:
		return m, tea.Batch(
			m.list.SetItems(getLiabilitiesItems(m.api)),
			Cmd(DataLoadCompletedMsg{DataType: "liabilities"}),
		)
	case NewLiabilityMsg:
		err := m.api.CreateLiabilityAccount(firefly.NewLiability{
			Name:         msg.Account,
			CurrencyCode: msg.Currency,
			Type:         msg.Type,
			Direction:    msg.Direction,
		})
		if err != nil {
			return m, Notify(err.Error(), Warning)
		}
		promptValue = ""
		return m, Cmd(RefreshLiabilitiesMsg{})
	case tea.WindowSizeMsg:
		h, v := baseStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-topSize)
	}

	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Quit):
			return m, SetView(transactionsView)
		case key.Matches(msg, m.keymap.Filter):
			i, ok := m.list.SelectedItem().(liabilityItem)
			if ok {
				return m, Cmd(FilterMsg{Account: i.account})
			}
			return m, nil
		case key.Matches(msg, m.keymap.New):
			return m, CmdPromptNewLiability(SetView(liabilitiesView))
		case key.Matches(msg, m.keymap.Refresh):
			return m, Cmd(RefreshLiabilitiesMsg{})
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

		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m modelLiabilities) View() string {
	return lipgloss.NewStyle().PaddingRight(1).Render(m.list.View())
}

func (m *modelLiabilities) Focus() {
	m.list.FilterInput.Focus()
	m.focus = true
}

func (m *modelLiabilities) Blur() {
	m.list.FilterInput.Blur()
	m.focus = false
}

func getLiabilitiesItems(api *firefly.Api) []list.Item {
	items := []list.Item{}
	for _, i := range api.Accounts["liabilities"] {
		items = append(items, liabilityItem{
			account:  i.Name,
			balance:  i.Balance,
			currency: i.CurrencyCode,
		})
	}

	return items
}

func CmdPromptNewLiability(backCmd tea.Cmd) tea.Cmd {
	return Cmd(PromptMsg{
		Prompt: "New Liabity(<name>,<currency>,<type:loan|debt|mortage>,<direction:credit|debit>): ",
		Value:  promptValue,
		Callback: func(value string) tea.Cmd {
			var cmds []tea.Cmd
			if value != "None" {
				promptValue = value
				// String: <name>,<currency>,<type>,<direction>
				re := regexp.MustCompile(`^\s*([^,]+)\s*,\s*([^,]+)\s*,\s*([^,]+)\s*,\s*([^,]+)\s*$`)
				matches := re.FindStringSubmatch(value)
				if len(matches) == 5 {
					acc := strings.TrimSpace(matches[1])
					cur := strings.TrimSpace(matches[2])
					typ := strings.TrimSpace(matches[3])
					dir := strings.TrimSpace(matches[4])
					if acc != "" && cur != "" {
						cmds = append(cmds, Cmd(NewLiabilityMsg{Account: acc, Currency: cur, Type: typ, Direction: dir}))
					} else {
						cmds = append(cmds, Notify("Invalid liability name or currency", Warning))
					}
				} else {
					cmds = append(cmds, Notify("Invalid liability request", Warning))
				}

			}
			cmds = append(cmds, backCmd)
			return tea.Sequence(cmds...)
		},
	})
}
