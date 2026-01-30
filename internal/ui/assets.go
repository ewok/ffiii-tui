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

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	RefreshAssetsMsg struct{}
	AssetsUpdateMsg  struct{}
	NewAssetMsg      struct {
		Account  string
		Currency string
	}
)

type assetItem = accountListItem[firefly.Account]

type modelAssets struct {
	AccountListModel[firefly.Account]
}

func newModelAssets(api AssetAPI) modelAssets {
	config := &AccountListConfig[firefly.Account]{
		AccountType: "asset",
		Title:       "Asset accounts",
		GetItems: func(apiInterface any, sorted bool) []list.Item {
			return getAssetsItems(apiInterface.(AssetAPI))
		},
		RefreshItems: func(apiInterface any, accountType string) error {
			return apiInterface.(AssetAPI).UpdateAccounts(accountType)
		},
		RefreshMsgType: RefreshAssetsMsg{},
		UpdateMsgType:  AssetsUpdateMsg{},
		PromptNewFunc: func() tea.Cmd {
			return CmdPromptNewAsset(SetView(assetsView))
		},
		HasSort:     false,
		HasTotalRow: false,
		HasSummary:  true,
		FilterFunc: func(item list.Item) tea.Cmd {
			i, ok := item.(assetItem)
			if ok {
				return Cmd(FilterMsg{Account: i.Entity})
			}
			return nil
		},
		SelectFunc: func(item list.Item) tea.Cmd {
			var cmds []tea.Cmd
			i, ok := item.(assetItem)
			if ok {
				cmds = append(cmds, Cmd(FilterMsg{Account: i.Entity}))
			}
			cmds = append(cmds, SetView(transactionsView))
			return tea.Sequence(cmds...)
		},
	}
	return modelAssets{
		AccountListModel: NewAccountListModel(api, config),
	}
}

func (m modelAssets) Init() tea.Cmd {
	return m.AccountListModel.Init()
}

func (m modelAssets) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if newMsg, ok := msg.(NewAssetMsg); ok {
		api := m.api.(AssetAPI)
		err := api.CreateAssetAccount(newMsg.Account, newMsg.Currency)
		if err != nil {
			return m, notify.NotifyWarn(err.Error())
		}
		return m, tea.Batch(
			Cmd(RefreshAssetsMsg{}),
			notify.NotifyLog(fmt.Sprintf("Asset account '%s' created", newMsg.Account)),
		)
	}

	if _, ok := msg.(RefreshAssetsMsg); ok {
		updated, cmd := m.AccountListModel.Update(msg)
		m.AccountListModel = updated.(AccountListModel[firefly.Account])
		if cmd != nil {
			return m, tea.Batch(
				cmd,
				Cmd(RefreshSummaryMsg{}),
			)
		}
	}
	updated, cmd := m.AccountListModel.Update(msg)
	m.AccountListModel = updated.(AccountListModel[firefly.Account])
	return m, cmd
}

func getAssetsItems(api AccountsAPI) []list.Item {
	items := []list.Item{}
	for _, account := range api.AccountsByType("asset") {
		items = append(items, newAccountListItem(
			account,
			"Balance",
			api.AccountBalance(account.ID),
		))
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
