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
	"ffiii-tui/internal/ui/notify"
	"ffiii-tui/internal/ui/prompt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

var promptValue string

type (
	RefreshLiabilitiesMsg struct{}
	LiabilitiesUpdateMsg  struct{}
	NewLiabilityMsg       struct {
		Account   string
		Currency  string
		Type      string
		Direction string
	}
)

type liabilityItem = accountListItem[firefly.Account]

type modelLiabilities struct {
	AccountListModel[firefly.Account]
}

func newModelLiabilities(api LiabilityAPI) modelLiabilities {
	config := &AccountListConfig[firefly.Account]{
		AccountType: "liability",
		Title:       "Liabilities",
		GetItems: func(apiInterface any, sorted bool) []list.Item {
			return getLiabilitiesItems(apiInterface.(LiabilityAPI))
		},
		RefreshItems: func(apiInterface any, accountType string) error {
			return apiInterface.(LiabilityAPI).UpdateAccounts(accountType)
		},
		RefreshMsgType: RefreshLiabilitiesMsg{},
		UpdateMsgType:  LiabilitiesUpdateMsg{},
		PromptNewFunc: func() tea.Cmd {
			return CmdPromptNewLiability(SetView(liabilitiesView))
		},
		HasSort:     false,
		HasTotalRow: false,
		FilterFunc: func(item list.Item) tea.Cmd {
			i, ok := item.(liabilityItem)
			if ok {
				return Cmd(FilterMsg{Account: i.Entity})
			}
			return nil
		},
		SelectFunc: func(item list.Item) tea.Cmd {
			var cmds []tea.Cmd
			i, ok := item.(liabilityItem)
			if ok {
				cmds = append(cmds, Cmd(FilterMsg{Account: i.Entity}))
			}
			cmds = append(cmds, SetView(transactionsView))
			return tea.Sequence(cmds...)
		},
	}
	return modelLiabilities{
		AccountListModel: NewAccountListModel[firefly.Account](api, config),
	}
}

func (m modelLiabilities) Init() tea.Cmd {
	return m.AccountListModel.Init()
}

func (m modelLiabilities) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if newMsg, ok := msg.(NewLiabilityMsg); ok {
		api := m.api.(LiabilityAPI)
		err := api.CreateLiabilityAccount(
			firefly.NewLiability{
				Name:         newMsg.Account,
				CurrencyCode: newMsg.Currency,
				Type:         newMsg.Type,
				Direction:    newMsg.Direction,
			})
		if err != nil {
			return m, notify.NotifyWarn(err.Error())
		}
		// Reset prompt on accaunt creation
		promptValue = ""
		return m, tea.Batch(
			Cmd(RefreshLiabilitiesMsg{}),
			notify.NotifyLog(fmt.Sprintf("Liability account '%s' created", newMsg.Account)),
		)
	}
	updated, cmd := m.AccountListModel.Update(msg)
	m.AccountListModel = updated.(AccountListModel[firefly.Account])
	return m, cmd
}

func getLiabilitiesItems(api AccountsAPI) []list.Item {
	items := []list.Item{}
	for _, account := range api.AccountsByType("liabilities") {
		label := "They owe us"
		balance := api.AccountBalance(account.ID)
		if account.LiabilityDirection == "debit" {
			label = "We owe"
			balance = (-1) * balance
		}
		items = append(items, newAccountListItem(
			account,
			label,
			balance,
		))
	}
	return items
}

func CmdPromptNewLiability(backCmd tea.Cmd) tea.Cmd {
	return prompt.Ask(
		"New Liabity(<name>,<currency>,<type:loan|debt|mortage>,<direction:credit|debit>): ",
		promptValue,
		func(value string) tea.Cmd {
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
						cmds = append(cmds, notify.NotifyWarn("Invalid liability name or currency"))
					}
				} else {
					cmds = append(cmds, notify.NotifyWarn("Invalid liability request"))
				}

			}
			cmds = append(cmds, backCmd)
			return tea.Sequence(cmds...)
		},
	)
}
