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
	account, balance, currency string
}

func (i assetItem) Title() string       { return i.account }
func (i assetItem) Description() string { return fmt.Sprintf("%s %s", i.balance, i.currency) }
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
		m.api.UpdateAssets()
		cmds = append(cmds, m.list.SetItems(getAssetsItems(m.api)))
		return m, tea.Batch(cmds...)
	case NewAssetMsg:
		err := m.api.CreateAccount(msg.account, "asset", msg.currency)
		// TODO: Report error to user
		if err != nil {
			fmt.Println("Error creating asset:", err)
		} else {
			cmds = append(cmds, Cmd(RefreshAssetsMsg{}))
		}
		return m, tea.Batch(cmds...)
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
			case "esc", "q", "ctrl+c":
				cmds = append(cmds, Cmd(ViewTransactionsMsg{}))
			case "f":
				i, ok := m.list.SelectedItem().(assetItem)
				if ok {
					cmds = append(cmds, Cmd(FilterAssetMsg{account: i.account}))
				}
				return m, tea.Batch(cmds...)
			case "enter":
				i, ok := m.list.SelectedItem().(assetItem)
				if ok {
					cmds = append(cmds, Cmd(FilterAssetMsg{account: i.account}))
				}
				cmds = append(cmds, Cmd(ViewTransactionsMsg{}))
			case "n":
				cmds = append(cmds, Cmd(PromptMsg{
					Prompt: "New Asset(name,currency): ",
					Value:  "",
					Callback: func(value string) []tea.Cmd {
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
						return cmds
					}}))
				return m, tea.Batch(cmds...)
			case "c":
				cmds = append(cmds, Cmd(ViewCategoriesMsg{}))
			case "e":
				cmds = append(cmds, Cmd(ViewExpensesMsg{}))
			case "i":
				cmds = append(cmds, Cmd(ViewRevenuesMsg{}))
			case "r":
				cmds = append(cmds, Cmd(RefreshAssetsMsg{}))
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
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
	for _, i := range api.Assets {
		items = append(items, assetItem{account: i.Name, balance: fmt.Sprintf("%.2f", i.Balance), currency: i.CurrencyCode})
	}
	return items
}
