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

type RefreshRevenuesMsg struct{}
type NewRevenueMsg struct {
	account string
}

type revenuesItem struct {
	id, name string
}

func (i revenuesItem) FilterValue() string { return i.name }

type revenuesItemDelegate struct{}

func (d revenuesItemDelegate) Height() int                             { return 1 }
func (d revenuesItemDelegate) Spacing() int                            { return 0 }
func (d revenuesItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d revenuesItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(revenuesItem)
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

type modelRevenues struct {
	list  list.Model
	api   *firefly.Api
	focus bool
}

func newModelRevenues(api *firefly.Api) modelRevenues {
	items := getRevenuesItems(api)

	m := modelRevenues{list: list.New(items, revenuesItemDelegate{}, 0, 0), api: api}
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

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case RefreshRevenuesMsg:
		m.api.UpdateRevenues()
		cmds = append(cmds, m.list.SetItems(getRevenuesItems(m.api)))
		return m, tea.Batch(cmds...)
	case NewRevenueMsg:
		err := m.api.CreateAccount(msg.account, "revenue", "")
        // TODO: Report error to user
		if err == nil {
			cmds = append(cmds, Cmd(RefreshRevenuesMsg{}))
		}
		return m, tea.Batch(cmds...)
	case tea.WindowSizeMsg:
		h, v := baseStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-topSize)
	}

	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.focus {
			switch msg.String() {
			case "esc", "q", "ctrl+c":
				cmds = append(cmds, Cmd(ViewTransactionsMsg{}))
			case "n":
				cmds = append(cmds, Cmd(PromptMsg{
					Prompt: "New Revenue: ",
					Value:  "",
					Callback: func(value string) []tea.Cmd {
						var cmds []tea.Cmd
						if value != "" {
							cmds = append(cmds, Cmd(NewRevenueMsg{account: value}))
						}
						cmds = append(cmds, Cmd(ViewRevenuesMsg{}))
						return cmds
					}}))
				return m, tea.Batch(cmds...)
			case "a":
				cmds = append(cmds, Cmd(ViewAssetsMsg{}))
			case "c":
				cmds = append(cmds, Cmd(ViewCategoriesMsg{}))
			case "e":
				cmds = append(cmds, Cmd(ViewExpensesMsg{}))
			case "r":
				cmds = append(cmds, Cmd(RefreshRevenuesMsg{}))
			}
		}
	}

	if m.focus {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
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

func getRevenuesItems(api *firefly.Api) []list.Item {
	items := []list.Item{}
	for _, i := range api.Revenues {
		items = append(items, revenuesItem{id: i.ID, name: i.Name})
	}
	return items
}

