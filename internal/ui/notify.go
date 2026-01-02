/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

// TODO: Make queue of messages.

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type NotifyMsg struct {
	Message string
	Level   NotifyLevel
}
type NotifyClearMsg time.Time

type NotifyLevel uint

const (
	Log NotifyLevel = iota
	Warning
	Critical
	Panic
)

var (
	notifyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
)

func Notify(message string, level NotifyLevel) tea.Cmd {
	return Cmd(NotifyMsg{
		Message: message,
		Level:   level,
	})
}

type modelNotify struct {
	text  string
	level NotifyLevel
}

func newNotify(msg NotifyMsg) modelNotify {
	return modelNotify{text: msg.Message, level: msg.Level}
}

func (m modelNotify) Init() tea.Cmd {
	return nil
}

func (m modelNotify) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NotifyClearMsg:
		return m, Notify("", Log)
	case NotifyMsg:
		var cmd tea.Cmd
		if msg.Message != "" {
			cmd = tea.Tick(time.Second*10, func(t time.Time) tea.Msg {
				return NotifyClearMsg(t)
			})
		}
		return newNotify(msg), cmd
	}
	return m, nil
}

func (m modelNotify) View() string {
	switch m.level {
	case Warning:
		return notifyStyle.Foreground(lipgloss.Color("214")).Render(m.text) // orange
	case Critical:
		return notifyStyle.Foreground(lipgloss.Color("196")).Render(m.text) // red
	case Panic:
		return notifyStyle.Foreground(lipgloss.Color("199")).Bold(true).Render(m.text) // bright red
	}
	return notifyStyle.Render(m.text)
}
