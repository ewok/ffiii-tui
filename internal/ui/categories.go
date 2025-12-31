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

type RefreshCategoriesMsg struct{}
type RefreshCategoryInsightsMsg struct{}
type NewCategoryMsg struct {
	category string
}

type categoryItem struct {
	category, currency string
	spent              float64
	earned             float64
}

func (i categoryItem) Title() string { return i.category }
func (i categoryItem) Description() string {
	s := ""
	if i.spent != 0 {
		s += fmt.Sprintf("Spent: %.2f %s", i.spent, i.currency)
	}
	if i.earned != 0 {
		if s != "" {
			s += " | "
		}
		s += fmt.Sprintf("Earned: %.2f %s", i.earned, i.currency)
	}
	if s == "" {
		s = "No transactions"
	}
	return s
}
func (i categoryItem) FilterValue() string { return i.category }

type modelCategories struct {
	list   list.Model
	api    *firefly.Api
	focus  bool
	sorted int
}

func newModelCategories(api *firefly.Api) modelCategories {
	items := getCategoriesItems(api, 0)

	m := modelCategories{list: list.New(items, list.NewDefaultDelegate(), 0, 0), api: api}
	m.list.Title = "Categories"
	m.list.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowStatusBar(false)
	m.list.DisableQuitKeybindings()

	return m
}

func (m modelCategories) Init() tea.Cmd {
	return nil
}

func (m modelCategories) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case RefreshCategoryInsightsMsg:
		return m, tea.Sequence(Cmd(m.api.UpdateCategoriesInsights()),
			m.list.SetItems(getCategoriesItems(m.api, m.sorted)))
	case RefreshCategoriesMsg:
		return m, tea.Sequence(
			Cmd(m.api.UpdateCategories()),
			m.list.SetItems(getCategoriesItems(m.api, m.sorted)))
	case NewCategoryMsg:
		err := m.api.CreateCategory(msg.category, "")
		if err != nil {
			return m, Notify(err.Error(), Warning)
		}
		return m, Cmd(RefreshCategoriesMsg{})
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
					Prompt: "New Category: ",
					Value:  "",
					Callback: func(value string) tea.Cmd {
						var cmds []tea.Cmd
						if value != "None" {
							cmds = append(cmds, Cmd(NewCategoryMsg{category: value}))
						}
						cmds = append(cmds, Cmd(ViewCategoriesMsg{}))
						return tea.Sequence(cmds...)
					}})
			case "f":
				i, ok := m.list.SelectedItem().(categoryItem)
				if ok {
					return m, Cmd(FilterMsg{category: i.category})
				}
				return m, nil
			case "a":
				return m, Cmd(ViewAssetsMsg{})
			case "e":
				return m, Cmd(ViewExpensesMsg{})
			case "i":
				return m, Cmd(ViewRevenuesMsg{})
			case "r":
				return m, Cmd(RefreshCategoryInsightsMsg{})
			case "R":
				return m, Cmd(RefreshCategoriesMsg{})
			case "s":
				switch m.sorted {
				case 0:
					m.sorted = -1
				case -1:
					m.sorted = 1
				case 1:
					m.sorted = 0
				}
				return m, m.list.SetItems(getCategoriesItems(m.api, m.sorted))
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

func (m modelCategories) View() string {
	return m.list.View()
}

func (m *modelCategories) Focus() {
	m.list.FilterInput.Focus()
	m.focus = true
}

func (m *modelCategories) Blur() {
	m.list.FilterInput.Blur()
	m.focus = false
}

func getCategoriesItems(api *firefly.Api, sorted int) []list.Item {
	items := []list.Item{}
	for _, i := range api.Categories {
		spent := i.GetSpent(api)
		earned := i.GetEarned(api)
		if sorted < 0 && spent == 0 {
			continue
		}
		if sorted > 0 && earned == 0 {
			continue
		}
		items = append(items, categoryItem{
			category: i.Name,
			currency: i.CurrencyCode,
			spent:    spent,
			earned:   earned,
		})
	}
	if sorted < 0 {
		slices.SortFunc(items, func(a, b list.Item) int {
			return int(b.(categoryItem).spent) - int(a.(categoryItem).spent)
		})
	} else if sorted > 0 {
		slices.SortFunc(items, func(a, b list.Item) int {
			return int(b.(categoryItem).earned) - int(a.(categoryItem).earned)
		})
	}

	return items
}
