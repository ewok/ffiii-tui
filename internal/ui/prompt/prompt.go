/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package prompt

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

type Model struct {
	input    textinput.Model
	callback func(value string) tea.Cmd
	focus    bool
	styles   Styles
	Width    int
}

func New() Model {
	m := textinput.New()

	prompt := Model{
		input:  m,
		styles: DefaultStyles(),
		Width:  80,
	}

	return prompt
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case PromptMsg:
		m.input.Prompt = msg.Prompt
		m.input.SetValue(msg.Value)
		m.callback = msg.Callback
		m.Focus()
		return m, nil
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
			m.Blur()
			return m, m.callback(value)
		case "esc":
			m.Blur()
			return m, m.callback("None")
		default:
			m.input, cmd = m.input.Update(msg)
		}
	}
	return m, cmd
}

func (m Model) View() string {
	return m.styles.PromptFocused.Width(m.Width).Render(" " + m.input.View())
}

func (m *Model) Focus() {
	m.input.Focus()
	m.focus = true
}

func (m *Model) Blur() {
	m.input.Blur()
	m.focus = false
}

func (m *Model) Focused() bool {
	return m.focus
}

func (m *Model) WithWidth(width int) *Model {
	m.Width = width
	return m
}

func (m *Model) WithStyles(styles Styles) *Model {
	m.styles = styles
	return m
}

func Ask(prompt, value string, callback func(value string) tea.Cmd) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		return PromptMsg{
			Prompt:   prompt,
			Value:    value,
			Callback: callback,
		}
	})
}
