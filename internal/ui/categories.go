/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"fmt"
	"slices"

	"ffiii-tui/internal/firefly"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	RefreshCategoriesMsg       struct{}
	RefreshCategoryInsightsMsg struct{}
	CategoriesUpdateMsg        struct{}
	NewCategoryMsg             struct {
		Category string
	}
)

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
	keymap CategoryKeyMap
}

func newModelCategories(api *firefly.Api) modelCategories {
	items := getCategoriesItems(api, 0)

	m := modelCategories{
		list:   list.New(items, list.NewDefaultDelegate(), 0, 0),
		api:    api,
		keymap: DefaultCategoryKeyMap(),
	}
	m.list.Title = "Categories"
	m.list.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowStatusBar(false)
	m.list.SetShowHelp(false)
	m.list.DisableQuitKeybindings()

	return m
}

func (m modelCategories) Init() tea.Cmd {
	return nil
}

func (m modelCategories) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case RefreshCategoryInsightsMsg:
		return m, func() tea.Msg {
			m.api.UpdateCategoriesInsights()
			return CategoriesUpdateMsg{}
		}
	case RefreshCategoriesMsg:
		return m, func() tea.Msg {
			m.api.UpdateCategories()
			return CategoriesUpdateMsg{}
		}
	case CategoriesUpdateMsg:
		return m, tea.Batch(
			m.list.SetItems(getCategoriesItems(m.api, m.sorted)),
			Cmd(DataLoadCompletedMsg{DataType: "categories"}),
		)
	case NewCategoryMsg:
		err := m.api.CreateCategory(msg.Category, "")
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
		switch {
		case key.Matches(msg, m.keymap.Quit):
			return m, SetView(transactionsView)
		case key.Matches(msg, m.keymap.New):
			return m, CmdPromptNewCategory(SetView(categoriesView))
		case key.Matches(msg, m.keymap.Filter):
			i, ok := m.list.SelectedItem().(categoryItem)
			if ok {
				return m, Cmd(FilterMsg{Category: i.category})
			}
			return m, nil
		case key.Matches(msg, m.keymap.ResetFilter):
			return m, Cmd(FilterMsg{Reset: true})
		case key.Matches(msg, m.keymap.Refresh):
			return m, Cmd(RefreshCategoriesMsg{})
		case key.Matches(msg, m.keymap.Sort):
			switch m.sorted {
			case 0:
				m.sorted = -1
			case -1:
				m.sorted = 1
			case 1:
				m.sorted = 0
			}
			return m, m.list.SetItems(getCategoriesItems(m.api, m.sorted))
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
			// case "R":
			// 	return m, Cmd(RefreshCategoriesMsg{})
		}
	}

	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m modelCategories) View() string {
	return lipgloss.NewStyle().PaddingRight(1).Render(m.list.View())
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

func CmdPromptNewCategory(backCmd tea.Cmd) tea.Cmd {
	return Cmd(PromptMsg{
		Prompt: "New Category(<name>): ",
		Value:  "",
		Callback: func(value string) tea.Cmd {
			var cmds []tea.Cmd
			if value != "None" {
				cmds = append(cmds, Cmd(NewCategoryMsg{Category: value}))
			}
			cmds = append(cmds, backCmd)
			return tea.Sequence(cmds...)
		},
	})
}
