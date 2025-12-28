/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"ffiii-tui/internal/firefly"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type RefreshRevenuesMsg struct{}
type NewRevenueMsg struct {
	account string
}

type revenueItem struct {
	id, name, balance, currency string
}

func (i revenueItem) Title() string       { return i.name }
func (i revenueItem) Description() string {
	if i.balance != "" && i.currency != "" {
		return fmt.Sprintf("%s %s", i.balance, i.currency)
	}
	return ""
}
func (i revenueItem) FilterValue() string { return i.name }

type modelRevenues struct {
	list  list.Model
	api   *firefly.Api
	focus bool
}

func newModelRevenues(api *firefly.Api) modelRevenues {
	items := getRevenuesItems(api)

	m := modelRevenues{list: list.New(items, list.NewDefaultDelegate(), 0, 0), api: api}
	m.list.Title = "Revenues"
	m.list.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowStatusBar(false)
	m.list.DisableQuitKeybindings()

	return m
}

func (m modelRevenues) Init() tea.Cmd {
	return nil
}

func (m modelRevenues) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case RefreshRevenuesMsg:
		m.api.UpdateRevenues()
		return m, m.list.SetItems(getRevenuesItems(m.api))
	case NewRevenueMsg:
		err := m.api.CreateAccount(msg.account, "revenue", "")
		// TODO: Report error to user
		if err == nil {
			return m, Cmd(RefreshRevenuesMsg{})
		}
		return m, nil
	case tea.WindowSizeMsg:
		h, v := baseStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-topSize)
	}

	if !m.focus {
		return m, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.focus {
			switch msg.String() {
			case "esc", "q":
				return m, Cmd(ViewTransactionsMsg{})
			case "n":
				return m, Cmd(PromptMsg{
					Prompt: "New Revenue: ",
					Value:  "",
					Callback: func(value string) tea.Cmd {
						var cmds []tea.Cmd
						if value != "" {
							cmds = append(cmds, Cmd(NewRevenueMsg{account: value}))
						}
						cmds = append(cmds, Cmd(ViewRevenuesMsg{}))
						return tea.Sequence(cmds...)
					}})
			case "f":
				i, ok := m.list.SelectedItem().(revenueItem)
				if ok {
					return m, Cmd(FilterItemMsg{account: i.name})
				}
				return m, nil
			case "a":
				return m, Cmd(ViewAssetsMsg{})
			case "c":
				return m, Cmd(ViewCategoriesMsg{})
			case "e":
				return m, Cmd(ViewExpensesMsg{})
			case "r":
				return m, Cmd(RefreshRevenuesMsg{})
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m modelRevenues) View() string {
	return m.list.View()
}

func (m *modelRevenues) Focus() {
	m.list.FilterInput.Focus()
	m.focus = true
}

func (m *modelRevenues) Blur() {
	m.list.FilterInput.Blur()
	m.focus = false
}

func getRevenuesItems(api *firefly.Api) []list.Item {
	items := []list.Item{}
	for _, i := range api.Revenues {
		items = append(items, revenueItem{
			id:       i.ID, 
			name:     i.Name,
			balance:  fmt.Sprintf("%.2f", i.Balance),
			currency: i.CurrencyCode,
		})
	}
	return items
}
