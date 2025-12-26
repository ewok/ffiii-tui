/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"ffiii-tui/internal/firefly"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type RefreshCategoriesMsg struct{}
type NewCategoryMsg struct {
	category string
}

type categoryItem struct {
	id, name, notes string
}

func (i categoryItem) FilterValue() string { return i.name }

type categoryItemDelegate struct{}

func (d categoryItemDelegate) Height() int                             { return 1 }
func (d categoryItemDelegate) Spacing() int                            { return 0 }
func (d categoryItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d categoryItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(categoryItem)
	if !ok {
		return
	}

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("| " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(i.name))
}

type modelCategories struct {
	list  list.Model
	api   *firefly.Api
	focus bool
}

func newModelCategories(api *firefly.Api) modelCategories {
	items := getCategoriesItems(api)

	m := modelCategories{list: list.New(items, categoryItemDelegate{}, 0, 0), api: api}
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

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := baseStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		if m.focus {
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				cmds = append(cmds, Cmd(ViewTransactionsMsg{}))
			case "n":
				cmds = append(cmds, Cmd(PromptMsg{
					Prompt: "New Category: ",
					Value:  "",
					Callback: func(value string) []tea.Cmd {
						var cmds []tea.Cmd
						if value != "" {
							cmds = append(cmds, Cmd(NewCategoryMsg{category: value}))
						}
						cmds = append(cmds, Cmd(ViewCategoriesMsg{}))
						return cmds
					}}))
				return m, tea.Batch(cmds...)
			case "a":
				cmds = append(cmds, Cmd(ViewAccountsMsg{}))
			case "e":
				cmds = append(cmds, Cmd(ViewExpensesMsg{}))
			case "r":
				cmds = append(cmds, Cmd(RefreshCategoriesMsg{}))
			}
		}
	}

	if m.focus {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
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
		items = append(items, categoryItem{id: i.ID, name: i.Name, notes: i.Notes})
	}
	return items
}
