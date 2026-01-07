/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"fmt"
	"strings"

	"ffiii-tui/internal/firefly"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	RefreshAssetsMsg struct{}
	AssetsUpdateMsg  struct{}
)

type NewAssetMsg struct {
	Account  string
	Currency string
}

type assetItem struct {
	account, currency string
	balance           float64
}

func (i assetItem) Title() string { return i.account }
func (i assetItem) Description() string {
	return fmt.Sprintf("Balance: %.2f %s", i.balance, i.currency)
}
func (i assetItem) FilterValue() string { return i.account }

type modelAssets struct {
	list   list.Model
	api    *firefly.Api
	focus  bool
	keymap AssetKeyMap
}

func newModelAssets(api *firefly.Api) modelAssets {
	items := getAssetsItems(api)

	m := modelAssets{
		list:   list.New(items, list.NewDefaultDelegate(), 0, 0),
		api:    api,
		keymap: DefaultAssetKeyMap(),
	}
	m.list.Title = "Asset accounts"
	m.list.SetShowStatusBar(false)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowHelp(false)
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
		return m, func() tea.Msg {
            err := m.api.UpdateAccounts("asset")
            if err != nil {
                return Notify(err.Error(), Warning)
            }
			return AssetsUpdateMsg{}
		}
	case AssetsUpdateMsg:
		return m, tea.Batch(
			m.list.SetItems(getAssetsItems(m.api)),
			Cmd(DataLoadCompletedMsg{DataType: "assets"}),
		)
	case NewAssetMsg:
		err := m.api.CreateAssetAccount(msg.Account, msg.Currency)
		if err != nil {
			return m, Notify(err.Error(), Warning)
		}
		return m, Cmd(RefreshAssetsMsg{})
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
			i, ok := m.list.SelectedItem().(assetItem)
			if ok {
				return m, Cmd(FilterMsg{Account: i.account})
			}
			return m, nil
		case key.Matches(msg, m.keymap.Select):
			i, ok := m.list.SelectedItem().(assetItem)
			if ok {
				cmds = append(cmds, Cmd(FilterMsg{Account: i.account}))
			}
			cmds = append(cmds, SetView(transactionsView))
			return m, tea.Sequence(cmds...)
		case key.Matches(msg, m.keymap.New):
			return m, CmdPromptNewAsset(SetView(assetsView))
		case key.Matches(msg, m.keymap.Refresh):
			return m, Cmd(RefreshAssetsMsg{})
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

func (m modelAssets) View() string {
	return lipgloss.NewStyle().PaddingRight(1).Render(m.list.View())
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
			currency: i.CurrencyCode,
		})
	}

	return items
}

func CmdPromptNewAsset(backCmd tea.Cmd) tea.Cmd {
	return Cmd(PromptMsg{
		Prompt: "New Asset(<name>,<currency>): ",
		Value:  "",
		Callback: func(value string) tea.Cmd {
			var cmds []tea.Cmd
			if value != "None" {
				split := strings.SplitN(value, ",", 2)
				if len(split) >= 2 {
					acc := strings.TrimSpace(split[0])
					cur := strings.TrimSpace(split[1])
					if acc != "" && cur != "" {
						cmds = append(cmds, Cmd(NewAssetMsg{Account: acc, Currency: cur}))
					} else {
						cmds = append(cmds, Notify("Invalid asset name or currency", Warning))
					}
				} else {
					cmds = append(cmds, Notify("Invalid asset name or currency", Warning))
				}
			}
			cmds = append(cmds, backCmd)
			return tea.Sequence(cmds...)
		},
	})
}
