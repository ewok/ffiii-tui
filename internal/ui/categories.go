/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"ffiii-tui/internal/firefly"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type RefreshCategoriesMsg struct{}
type NewCategoryMsg struct {
	category string
}

type categoryItem struct {
	id, name, notes, spent, currency string
}

func (i categoryItem) Title() string { return i.name }
func (i categoryItem) Description() string {
	if i.spent != "" && i.currency != "" {
		return fmt.Sprintf("%s %s", i.spent, i.currency)
	}
	return ""
}
func (i categoryItem) FilterValue() string { return i.name }

type modelCategories struct {
	list  list.Model
	api   *firefly.Api
	focus bool
}

func newModelCategories(api *firefly.Api) modelCategories {
	items := getCategoriesItems(api)

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
	case RefreshCategoriesMsg:
		m.api.UpdateCategories()
		return m, m.list.SetItems(getCategoriesItems(m.api))
	case NewCategoryMsg:
		err := m.api.CreateCategory(msg.category, "")
		if err == nil {
			return m, Cmd(RefreshCategoriesMsg{})

		}
		// TODO: report
		return m, nil
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
			case "esc", "q", "ctrl+c":
				return m, Cmd(ViewTransactionsMsg{})
			case "n":
				return m, Cmd(PromptMsg{
					Prompt: "New Category: ",
					Value:  "",
					Callback: func(value string) tea.Cmd {
						var cmds []tea.Cmd
						if value != "" {
							cmds = append(cmds, Cmd(NewCategoryMsg{category: value}))
						}
						cmds = append(cmds, Cmd(ViewCategoriesMsg{}))
						return tea.Sequence(cmds...)
					}})
			case "f":
				i, ok := m.list.SelectedItem().(categoryItem)
				if ok {
					return m, Cmd(FilterItemMsg{category: i.name})
				}
				// TODO: report
				return m, nil
			case "a":
				return m, Cmd(ViewAssetsMsg{})
			case "e":
				return m, Cmd(ViewExpensesMsg{})
			case "i":
				return m, Cmd(ViewRevenuesMsg{})
			case "r":
				return m, Cmd(RefreshCategoriesMsg{})
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

func getCategoriesItems(api *firefly.Api) []list.Item {
	items := []list.Item{}
	for _, i := range api.Categories {
		spent := ""
		currency := ""

		// Get the first spent entry if available
		if len(i.Spent) > 0 {
			spent = fmt.Sprintf("%.2f", i.Spent[0].Amount)
			currency = i.Spent[0].CurrencyCode
		}

		items = append(items, categoryItem{
			id:       i.ID,
			name:     i.Name,
			notes:    i.Notes,
			spent:    spent,
			currency: currency,
		})
	}
	return items
}
