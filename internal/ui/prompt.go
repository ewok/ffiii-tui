/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type PromptMsg struct {
	Prompt   string
	Value    string
	Callback func(value string) []tea.Cmd
}

type modelPrompt struct {
	input    textinput.Model
	callback func(value string) []tea.Cmd
	focus    bool
}

func newPrompt(msg PromptMsg) modelPrompt {
	m := textinput.New()
	m.Prompt = msg.Prompt
	m.SetValue(msg.Value)

	prompt := modelPrompt{
		input:    m,
		callback: msg.Callback,
	}

	return prompt
}

func (m modelPrompt) Init() tea.Cmd {
	return nil
}

func (m modelPrompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case PromptMsg:
		m = newPrompt(msg)
		cmds = append(cmds, cmd, Cmd(ViewPromptMsg{}))
		return m, tea.Batch(cmds...)
	}

	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			cmds = m.callback(m.input.Value())
		case "esc", "ctrl+c":
			cmds = m.callback("")
		default:
			m.input, cmd = m.input.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m modelPrompt) View() string {
	return m.input.View()
}

func (m *modelPrompt) Focus() {
	m.input.Focus()
	m.focus = true
}

func (m *modelPrompt) Blur() {
	m.input.Blur()
	m.focus = false
}
