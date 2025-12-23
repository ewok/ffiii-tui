/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"ffiii-tui/internal/firefly"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/bubbles/textinput"
)

type state uint

const (
	transactionView state = iota
	filterView
	periodView
	newView
	accountView
	categoryView
)

type (
	viewTransactionsMsg struct{}
	viewFilterMsg       struct{}
	viewNewMsg          struct{}
)

type modelUI struct {
	state      state
	list       modelList
	filter     textinput.Model
	fireflyApi *firefly.Api
	new        modelNewTransaction
}

func Show(api *firefly.Api) {

	ti := textinput.New()
	ti.Placeholder = "Filter"
	ti.CharLimit = 156
	ti.Width = 20

	n := newModelNewTransaction(api)

	m := modelUI{filter: ti, fireflyApi: api, list: InitList(api), new: n}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func (m modelUI) Init() tea.Cmd {
	return nil
}

func (m modelUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	case viewTransactionsMsg:
		m.state = transactionView
		m.filter.Blur()
		m.list.table.Focus()
	case viewFilterMsg:
		m.state = filterView
		m.filter.Focus()
		m.list.table.Blur()
	case viewNewMsg:
		m.state = newView
		m.filter.Blur()
		m.list.table.Blur()
	}

	switch m.state {
	case transactionView:
		nModel, nCmd := m.list.Update(msg)
		listModel, ok := nModel.(modelList)
		if !ok {
			panic("Somthing bad happened")
		}
		m.list = listModel
		cmd = nCmd
	case filterView:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				cmds = append(cmds, func() tea.Msg { return FilterMsg{query: ""} })
				cmds = append(cmds, func() tea.Msg { return viewTransactionsMsg{} })
			case "enter":
				value := m.filter.Value()
				cmds = append(cmds, func() tea.Msg { return FilterMsg{query: value} })
				cmds = append(cmds, func() tea.Msg { return viewTransactionsMsg{} })
			}
		}
		m.filter, cmd = m.filter.Update(msg)
	case newView:
		// m.new = newModelNewTransaction(m.fireflyApi)
		nModel, nCmd := m.new.Update(msg)
		newModel, ok := nModel.(modelNewTransaction)
		if !ok {
			panic("Somthing bad happened")
		}
		m.new = newModel
		cmd = nCmd
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m modelUI) View() string {
	switch m.state {
	case transactionView:
		return baseStyle.Render(m.list.View()) + "\n"
	case filterView:
		return baseStyle.Render(fmt.Sprintf("filter: %s", m.filter.View()) + "\n" + m.list.View())
	case newView:
		return baseStyle.Render(m.new.View())
	}
	return baseStyle.Render(m.list.View()) + "\n"
}
