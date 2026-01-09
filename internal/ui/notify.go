/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

// TODO: Make queue of messages.

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type NotifyMsg struct {
	Message string
	Level   NotifyLevel
}
type NotifyClearMsg time.Time

type NotifyLevel uint

const (
	Log NotifyLevel = iota
	Warn
	Err
)

func Notify(message string, level NotifyLevel) tea.Cmd {
	return Cmd(NotifyMsg{
		Message: message,
		Level:   level,
	})
}

type modelNotify struct {
	text   string
	level  NotifyLevel
	styles Styles
	Width  int
}

func newNotify(msg NotifyMsg) modelNotify {
	return modelNotify{
		text: msg.Message, level: msg.Level, styles: DefaultStyles(),
	}
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
	case UpdatePositions:
		h, _ := m.styles.Base.GetFrameSize()
		m.Width = globalWidth - h
	}
	return m, nil
}

func (m modelNotify) View() string {
	s := " Notification: " + m.text
	fn := m.styles.NotifyLog
	switch m.level {
	case Warn:
		fn = m.styles.NotifyWarn
	case Err:
		fn = m.styles.NotifyErr
	}
	return fn.Width(m.Width).Render(s)
}
