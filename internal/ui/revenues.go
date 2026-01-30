/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"fmt"
	"slices"

	"ffiii-tui/internal/firefly"
	"ffiii-tui/internal/ui/notify"
	"ffiii-tui/internal/ui/prompt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type (
	RefreshRevenuesMsg        struct{}
	RefreshRevenueInsightsMsg struct{}
	RevenuesUpdateMsg         struct{}
	NewRevenueMsg             struct {
		Account string
	}
)

type revenueItem = accountListItem[firefly.Account]

type modelRevenues struct {
	AccountListModel[firefly.Account]
}

func newModelRevenues(api RevenueAPI) modelRevenues {
	config := &AccountListConfig[firefly.Account]{
		AccountType: "revenue",
		Title:       "Revenue accounts",
		GetItems: func(apiInterface any, sorted bool) []list.Item {
			return getRevenuesItems(apiInterface.(RevenueAPI), sorted)
		},
		RefreshItems: func(apiInterface any, accountType string) error {
			return apiInterface.(RevenueAPI).UpdateAccounts(accountType)
		},
		RefreshMsgType: RefreshRevenuesMsg{},
		UpdateMsgType:  RevenuesUpdateMsg{},
		PromptNewFunc: func() tea.Cmd {
			return CmdPromptNewRevenue(SetView(revenuesView))
		},
		HasSort:     true,
		HasTotalRow: true,
		GetTotalFunc: func(api any) float64 {
			return api.(RevenueAPI).GetTotalRevenueDiff()
		},
		FilterFunc: func(item list.Item) tea.Cmd {
			i, ok := item.(revenueItem)
			if ok {
				return Cmd(FilterMsg{Account: i.Entity})
			}
			return nil
		},
		SelectFunc: func(item list.Item) tea.Cmd {
			var cmds []tea.Cmd
			i, ok := item.(revenueItem)
			if ok {
				cmds = append(cmds, Cmd(FilterMsg{Account: i.Entity}))
			}
			cmds = append(cmds, SetView(transactionsView))
			return tea.Sequence(cmds...)
		},
	}
	return modelRevenues{
		AccountListModel: NewAccountListModel(api, config),
	}
}

func (m modelRevenues) Init() tea.Cmd {
	return m.AccountListModel.Init()
}

func (m modelRevenues) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if newMsg, ok := msg.(NewRevenueMsg); ok {
		api := m.api.(RevenueAPI)
		err := api.CreateRevenueAccount(newMsg.Account)
		if err != nil {
			return m, notify.NotifyWarn(err.Error())
		}
		return m, tea.Batch(
			Cmd(RefreshRevenuesMsg{}),
			notify.NotifyLog(fmt.Sprintf("Revenue account '%s' created", newMsg.Account)),
		)
	}

	switch msg.(type) {
	case RefreshRevenueInsightsMsg:
		return m, func() tea.Msg {
			startLoading("Loading revenue insights...")
			defer stopLoading()
			err := m.api.(RevenueAPI).UpdateRevenueInsights()
			if err != nil {
				return notify.NotifyWarn(err.Error())()
			}
			return RevenuesUpdateMsg{}
		}
	}
	updated, cmd := m.AccountListModel.Update(msg)
	m.AccountListModel = updated.(AccountListModel[firefly.Account])
	return m, cmd
}

func getRevenuesItems(api RevenueAPI, sorted bool) []list.Item {
	items := []list.Item{}
	for _, account := range api.AccountsByType("revenue") {
		earned := api.GetRevenueDiff(account.ID)
		if sorted && earned == 0 {
			continue
		}
		items = append(items, newAccountListItem(
			account,
			"Earned",
			earned,
		))
	}
	if sorted {
		slices.SortFunc(items, func(a, b list.Item) int {
			return int(b.(revenueItem).PrimaryVal) - int(a.(revenueItem).PrimaryVal)
		})
	}
	return items
}

func CmdPromptNewRevenue(backCmd tea.Cmd) tea.Cmd {
	return prompt.Ask(
		"New Revenue(<name>): ",
		"",
		func(value string) tea.Cmd {
			var cmds []tea.Cmd
			if value != "None" {
				cmds = append(cmds, Cmd(NewRevenueMsg{Account: value}))
			}
			cmds = append(cmds, backCmd)
			return tea.Sequence(cmds...)
		},
	)
}
