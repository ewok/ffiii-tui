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

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

var totalCategory = firefly.Category{Name: "Total", CurrencyCode: ""}

type (
	RefreshCategoriesMsg       struct{}
	RefreshCategoryInsightsMsg struct{}
	CategoriesUpdateMsg        struct{}
	NewCategoryMsg             struct {
		Category string
	}
)

type categoryItem struct {
	category firefly.Category
	spent    float64
	earned   float64
}

func (i categoryItem) Title() string { return i.category.Name }
func (i categoryItem) Description() string {
	s := ""
	if i.spent != 0 {
		s += fmt.Sprintf("Spent: %.2f %s", i.spent, i.category.CurrencyCode)
	}
	if i.earned != 0 {
		if s != "" {
			s += " | "
		}
		s += fmt.Sprintf("Earned: %.2f %s", i.earned, i.category.CurrencyCode)
	}
	if s == "" {
		s = "No transactions"
	}
	return s
}
func (i categoryItem) FilterValue() string { return i.category.Name }

type modelCategories struct {
	list   list.Model
	api    CategoryAPI
	focus  bool
	sorted int
	keymap CategoryKeyMap
	styles Styles
}

func newModelCategories(api CategoryAPI) modelCategories {
	// Set the currency code for the total category
	totalCategory.CurrencyCode = api.PrimaryCurrency().Code

	items := getCategoriesItems(api, 0)

	m := modelCategories{
		list:   list.New(items, list.NewDefaultDelegate(), 0, 0),
		api:    api,
		keymap: DefaultCategoryKeyMap(),
		styles: DefaultStyles(),
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
			startLoading("Loading category insights...")
			defer stopLoading()
			err := m.api.UpdateCategoriesInsights()
			if err != nil {
				return notify.NotifyWarn(err.Error())()
			}
			return CategoriesUpdateMsg{}
		}
	case RefreshCategoriesMsg:
		return m, func() tea.Msg {
			startLoading("Loading categories...")
			defer stopLoading()
			err := m.api.UpdateCategories()
			if err != nil {
				return notify.NotifyWarn(err.Error())()
			}
			return CategoriesUpdateMsg{}
		}
	case CategoriesUpdateMsg:
		tSpent, tEarned := m.api.GetTotalSpentEarnedCategories()
		return m, tea.Batch(
			m.list.SetItems(getCategoriesItems(m.api, m.sorted)),
			m.list.InsertItem(0, categoryItem{
				category: totalCategory,
				spent:    tSpent,
				earned:   tEarned,
			}),
			Cmd(DataLoadCompletedMsg{DataType: "categories"}),
		)
	case NewCategoryMsg:
		startLoading("Creating category...")
		defer stopLoading()
		err := m.api.CreateCategory(msg.Category, "")
		if err != nil {
			return m, notify.NotifyWarn(err.Error())
		}
		return m, tea.Batch(
			Cmd(RefreshCategoriesMsg{}),
			notify.NotifyLog(fmt.Sprintf("Category '%s' created", msg.Category)),
		)
	case UpdatePositions:
		if msg.layout != nil {
			h, v := m.styles.Base.GetFrameSize()
			m.list.SetSize(
				msg.layout.Width-h,
				msg.layout.Height-v-msg.layout.TopSize,
			)
		}
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
				if i.category == totalCategory {
					return m, nil
				}
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
			return m, Cmd(CategoriesUpdateMsg{})
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
	return m.styles.LeftPanel.Render(m.list.View())
}

func (m *modelCategories) Focus() {
	m.list.FilterInput.Focus()
	m.focus = true
}

func (m *modelCategories) Blur() {
	m.list.FilterInput.Blur()
	m.focus = false
}

func getCategoriesItems(api CategoriesAPI, sorted int) []list.Item {
	items := []list.Item{}
	for _, category := range api.CategoriesList() {
		spent := api.CategorySpent(category.ID)
		earned := api.CategoryEarned(category.ID)
		if sorted < 0 && spent == 0 {
			continue
		}
		if sorted > 0 && earned == 0 {
			continue
		}
		items = append(items, categoryItem{
			category: category,
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
	return prompt.Ask(
		"New Category(<name>): ",
		"",
		func(value string) tea.Cmd {
			var cmds []tea.Cmd
			if value != "None" {
				cmds = append(cmds, Cmd(NewCategoryMsg{Category: value}))
			}
			cmds = append(cmds, backCmd)
			return tea.Sequence(cmds...)
		},
	)
}
