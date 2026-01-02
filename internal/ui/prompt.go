/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type PromptMsg struct {
	Prompt   string
	Value    string
	Callback func(value string) tea.Cmd
}

type modelPrompt struct {
	input    textinput.Model
	callback func(value string) tea.Cmd
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

	switch msg := msg.(type) {
	case PromptMsg:
		m = newPrompt(msg)
		return m, SetView(promptView)
	}

	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			value := strings.TrimSpace(m.input.Value())
			if value == "" {
				value = "None"
			}
			return m, m.callback(value)
		case "esc":
			return m, m.callback("None")
		default:
			m.input, cmd = m.input.Update(msg)
		}
	}
	return m, cmd
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
