/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"ffiii-tui/internal/firefly"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type RefreshAssetsMsg struct{}

type NewAssetMsg struct {
	account  string
	currency string
}

type assetItem struct {
	account, currency string
	balance           float64
}

func (i assetItem) Title() string       { return i.account }
func (i assetItem) Description() string { return fmt.Sprintf("%.2f %s", i.balance, i.currency) }
func (i assetItem) FilterValue() string { return i.account }

type modelAssets struct {
	list  list.Model
	api   *firefly.Api
	focus bool
}

func newModelAssets(api *firefly.Api) modelAssets {
	items := getAssetsItems(api)

	m := modelAssets{list: list.New(items, list.NewDefaultDelegate(), 0, 0), api: api}
	m.list.Title = "Assets"
	m.list.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(false)
	m.list.DisableQuitKeybindings()

	return m
}

func (m modelAssets) Init() tea.Cmd {
	return nil
}

func (m modelAssets) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case RefreshAssetsMsg:
		return m, tea.Sequence(
			Cmd(m.api.UpdateAccounts("asset")),
			m.list.SetItems(getAssetsItems(m.api)))
	case NewAssetMsg:
		err := m.api.CreateAccount(msg.account, "asset", msg.currency)
		// TODO: Report error to user
		if err != nil {
			cmd = tea.Println("Error creating asset:", err)
		} else {
			cmd = Cmd(RefreshAssetsMsg{})
		}
		return m, cmd
	case tea.WindowSizeMsg:
		h, v := baseStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-topSize)
	}

	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.focus {
			switch msg.String() {
			case "esc", "q":
				return m, Cmd(ViewTransactionsMsg{})
			case "f":
				i, ok := m.list.SelectedItem().(assetItem)
				if ok {
					return m, Cmd(FilterItemMsg{account: i.account})
				}
				return m, nil
			case "enter":
				i, ok := m.list.SelectedItem().(assetItem)
				if ok {
					cmds = append(cmds, Cmd(FilterItemMsg{account: i.account}))
				}
				cmds = append(cmds, Cmd(ViewTransactionsMsg{}))
				return m, tea.Sequence(cmds...)
			case "n":
				cmd = Cmd(PromptMsg{
					Prompt: "New Asset(name,currency): ",
					Value:  "",
					Callback: func(value string) tea.Cmd {
						var cmds []tea.Cmd
						if value != "" {
							split := strings.SplitN(value, ",", 2)
							if len(split) >= 2 {
								acc := strings.TrimSpace(split[0])
								cur := strings.TrimSpace(split[1])
								if acc != "" && cur != "" {
									cmds = append(cmds, Cmd(NewAssetMsg{account: acc, currency: cur}))
								} else {
									// TODO: Report error to user
								}
							}
						}
						cmds = append(cmds, Cmd(ViewAssetsMsg{}))
						return tea.Sequence(cmds...)
					}})
				return m, cmd
			case "c":
				return m, Cmd(ViewCategoriesMsg{})
			case "e":
				return m, Cmd(ViewExpensesMsg{})
			case "i":
				return m, Cmd(ViewRevenuesMsg{})
			case "r":
				return m, Cmd(RefreshAssetsMsg{})
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m modelAssets) View() string {
	return m.list.View()
}

func (m *modelAssets) Focus() {
	m.list.FilterInput.Focus()
	m.focus = true
}

func (m *modelAssets) Blur() {
	m.list.FilterInput.Blur()
	m.focus = false
}

func getAssetsItems(api *firefly.Api) []list.Item {
	items := []list.Item{}
	for _, i := range api.Accounts["asset"] {
		items = append(items, assetItem{
			account:  i.Name,
			balance:  i.Balance,
			currency: i.CurrencyCode})
	}

	return items
}
