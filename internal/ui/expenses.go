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

type RefreshEspensesMsg struct{}

type expensesItem struct {
	id, name string
}

func (i expensesItem) FilterValue() string { return i.name }

type expensesItemDelegate struct{}

func (d expensesItemDelegate) Height() int                             { return 1 }
func (d expensesItemDelegate) Spacing() int                            { return 0 }
func (d expensesItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d expensesItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(expensesItem)
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

type modelExpenses struct {
	list  list.Model
	api   *firefly.Api
	focus bool
}

func newModelExpenses(api *firefly.Api) modelExpenses {
	items := getExpensesItems(api)

	m := modelExpenses{list: list.New(items, expensesItemDelegate{}, 0, 0), api: api}
	m.list.Title = "Expenses"
	m.list.Styles.HelpStyle = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowStatusBar(false)
	m.list.DisableQuitKeybindings()

	return m
}

func (m modelExpenses) Init() tea.Cmd {
	return nil
}

func (m modelExpenses) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

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
			case "a":
				cmds = append(cmds, Cmd(ViewAccountsMsg{}))
			case "c":
				cmds = append(cmds, Cmd(ViewCategoriesMsg{}))
			}
		}
	}

	if m.focus {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m modelExpenses) View() string {
	return m.list.View()
}

func (m *modelExpenses) Focus() {
	m.list.FilterInput.Focus()
	m.focus = true
}

func (m *modelExpenses) Blur() {
	m.list.FilterInput.Blur()
	m.focus = false
}

func getExpensesItems(api *firefly.Api) []list.Item {
	items := []list.Item{}
	for _, i := range api.Expenses {
		items = append(items, expensesItem{id: i.ID, name: i.Name})
	}
	return items
}
