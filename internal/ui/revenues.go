/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"ffiii-tui/internal/firefly"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type RefreshRevenuesMsg struct{}
type RefreshRevenueInsightsMsg struct{}
type NewRevenueMsg struct {
	account string
}

type revenueItem struct {
	account, currency string
	earned            float64
}

func (i revenueItem) Title() string { return i.account }
func (i revenueItem) Description() string {
	return fmt.Sprintf("Earned: %.2f %s", i.earned, i.currency)
}
func (i revenueItem) FilterValue() string { return i.account }

type modelRevenues struct {
	list   list.Model
	api    *firefly.Api
	focus  bool
	sorted bool
}

func newModelRevenues(api *firefly.Api) modelRevenues {
	items := getRevenuesItems(api, false)

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
	case RefreshRevenueInsightsMsg:
		return m, tea.Sequence(
			Cmd(m.api.UpdateRevenueInsights()),
			m.list.SetItems(getRevenuesItems(m.api, m.sorted)))
	case RefreshRevenuesMsg:
		return m, tea.Sequence(
			Cmd(m.api.UpdateAccounts("revenue")),
			m.list.SetItems(getRevenuesItems(m.api, m.sorted)))
	case NewRevenueMsg:
		err := m.api.CreateAccount(msg.account, "revenue", "")
		if err != nil {
			return m, Notify(err.Error(), Warning)
		}
		return m, Cmd(RefreshRevenuesMsg{})
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
						if value != "None" {
							cmds = append(cmds, Cmd(NewRevenueMsg{account: value}))
						}
						cmds = append(cmds, Cmd(ViewRevenuesMsg{}))
						return tea.Sequence(cmds...)
					}})
			case "f":
				i, ok := m.list.SelectedItem().(revenueItem)
				if ok {
					return m, Cmd(FilterMsg{account: i.account})
				}
				return m, nil
			case "a":
				return m, Cmd(ViewAssetsMsg{})
			case "c":
				return m, Cmd(ViewCategoriesMsg{})
			case "e":
				return m, Cmd(ViewExpensesMsg{})
			case "r":
				return m, Cmd(RefreshRevenueInsightsMsg{})
			case "R":
				return m, Cmd(RefreshRevenuesMsg{})
			case "s":
				m.sorted = !m.sorted
				return m, m.list.SetItems(getRevenuesItems(m.api, m.sorted))
			case "t":
				return m, Cmd(ViewTransactionsMsg{})
			case "ctrl+a":
				return m, Cmd(FilterMsg{reset: true})
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

func getRevenuesItems(api *firefly.Api, sorted bool) []list.Item {
	items := []list.Item{}
	for _, i := range api.Accounts["revenue"] {
		earned := api.GetRevenueDiff(i.ID)
		if sorted && earned == 0 {
			continue
		}
		items = append(items, revenueItem{
			account:  i.Name,
			currency: i.CurrencyCode,
			earned:   earned,
		})
	}
	if sorted {
		slices.SortFunc(items, func(a, b list.Item) int {
			return int(b.(revenueItem).earned) - int(a.(revenueItem).earned)
		})
	}
	return items
}
