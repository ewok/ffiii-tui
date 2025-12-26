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

type RefreshBalanceMsg struct{}

type accountItem struct {
	account, balance, currency string
}

func (i accountItem) Title() string       { return i.account }
func (i accountItem) Description() string { return fmt.Sprintf("%s %s", i.balance, i.currency) }
func (i accountItem) FilterValue() string { return i.account }

type modelAccounts struct {
	list  list.Model
	api   *firefly.Api
	focus bool
}

func newModelAccounts(api *firefly.Api) modelAccounts {
	items := getAssetsItems(api)

	m := modelAccounts{list: list.New(items, list.NewDefaultDelegate(), 0, 0), api: api}
	m.list.Title = "Accounts"
	m.list.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(false)
	m.list.DisableQuitKeybindings()

	return m
}

func (m modelAccounts) Init() tea.Cmd {
	return nil
}

func (m modelAccounts) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.focus {
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				cmds = append(cmds, Cmd(ViewTransactionsMsg{}))
			case "f":
				i, ok := m.list.SelectedItem().(accountItem)
				if ok {
					cmds = append(cmds, Cmd(FilterAccountMsg{account: i.account}))
				}
				return m, tea.Batch(cmds...)
			case "enter":
				i, ok := m.list.SelectedItem().(accountItem)
				if ok {
					cmds = append(cmds, Cmd(FilterAccountMsg{account: i.account}))
				}
				cmds = append(cmds, Cmd(ViewTransactionsMsg{}))
			// case "n":
			// 	cmds = append(cmds, func() tea.Msg { return viewNewMsg{} })
			case "c":
				cmds = append(cmds, Cmd(ViewCategoriesMsg{}))
			case "e":
				cmds = append(cmds, Cmd(ViewExpensesMsg{}))
			}
		}
	case tea.WindowSizeMsg:
		h, v := baseStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	if m.focus {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m modelAccounts) View() string {
	return m.list.View()
}

func (m *modelAccounts) Focus() {
	m.list.FilterInput.Focus()
	m.focus = true
}

func (m *modelAccounts) Blur() {
	m.list.FilterInput.Blur()
	m.focus = false
}

func getAssetsItems(api *firefly.Api) []list.Item {
	items := []list.Item{}
	for _, i := range api.Assets {
		items = append(items, accountItem{account: i.Name, balance: fmt.Sprintf("%.2f", i.Balance), currency: i.CurrencyCode})
	}
	return items
}

