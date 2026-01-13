/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"fmt"
	"strings"

	"ffiii-tui/internal/firefly"
	"ffiii-tui/internal/ui/notify"
	"ffiii-tui/internal/ui/prompt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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
	account firefly.Account
	balance float64
}

func (i assetItem) Title() string { return i.account.Name }
func (i assetItem) Description() string {
	return fmt.Sprintf("Balance: %.2f %s", i.balance, i.account.CurrencyCode)
}
func (i assetItem) FilterValue() string { return i.account.Name }

type modelAssets struct {
	list   list.Model
	api    *firefly.Api
	focus  bool
	keymap AssetKeyMap
	styles Styles
}

func newModelAssets(api *firefly.Api) modelAssets {
	items := getAssetsItems(api)

	m := modelAssets{
		list:   list.New(items, list.NewDefaultDelegate(), 0, 0),
		api:    api,
		keymap: DefaultAssetKeyMap(),
		styles: DefaultStyles(),
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
				return notify.NotifyWarn(err.Error())
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
			return m, notify.NotifyWarn(err.Error())
		}
		return m, tea.Batch(
			Cmd(RefreshAssetsMsg{}),
			notify.NotifyLog(fmt.Sprintf("Asset account '%s' created", msg.Account)),
		)
	case UpdatePositions:
		h, v := m.styles.Base.GetFrameSize()
		m.list.SetSize(globalWidth-h, globalHeight-v-topSize-summarySize)
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
			return m, tea.Batch(
				Cmd(RefreshAssetsMsg{}),
				Cmd(RefreshSummaryMsg{}))
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
	return m.styles.LeftPanel.Render(m.list.View())
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
			account: i,
			balance: i.GetBalance(api),
		})
	}

	return items
}

func CmdPromptNewAsset(backCmd tea.Cmd) tea.Cmd {
	return prompt.Ask(
		"New Asset(<name>,<currency>): ",
		"",
		func(value string) tea.Cmd {
			var cmds []tea.Cmd
			if value != "None" {
				split := strings.SplitN(value, ",", 2)
				if len(split) >= 2 {
					acc := strings.TrimSpace(split[0])
					cur := strings.TrimSpace(split[1])
					if acc != "" && cur != "" {
						cmds = append(cmds, Cmd(NewAssetMsg{Account: acc, Currency: cur}))
					} else {
						cmds = append(cmds, notify.NotifyWarn("Invalid asset name or currency"))
					}
				} else {
					cmds = append(cmds, notify.NotifyWarn("Invalid asset name or currency"))
				}
			}
			cmds = append(cmds, backCmd)
			return tea.Sequence(cmds...)
		},
	)
}
