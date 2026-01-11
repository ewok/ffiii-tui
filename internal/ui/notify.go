/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"fmt"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	MaxQueueSize = 20

	LogDuration   = 5 * time.Second
	WarnDuration  = 7 * time.Second
	ErrorDuration = 10 * time.Second

	ShowQueueCounter = true // Show "(X more)" counter
	MessageIDPrefix  = "notify_"
)

type NotifyMsg struct {
	Message  string
	Level    NotifyLevel
	Duration *time.Duration
}

type NotifyLevel uint

const (
	Log NotifyLevel = iota
	Warn
	Err
)

type (
	NotifyShowNextMsg   struct{}
	NotifyExpireMsg     struct{ ID string }
	NotifyClearQueueMsg struct{}
)

type MessageState uint

const (
	Queued MessageState = iota
	Displaying
	Expired
)

type QueuedMessage struct {
	ID        string
	Message   string
	Level     NotifyLevel
	Duration  time.Duration
	Timestamp time.Time
	State     MessageState
}

type notifyQueue struct {
	messages []QueuedMessage
	current  *QueuedMessage
	nextID   uint64
	maxSize  int
}

type modelNotify struct {
	queue        notifyQueue
	styles       Styles
	Width        int
	isDisplaying bool
}

var globalMessageID uint64

func Notify(message string, level NotifyLevel) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		return NotifyMsg{
			Message:  message,
			Level:    level,
			Duration: nil,
		}
	})
}

func NotifyWithDuration(message string, level NotifyLevel, duration time.Duration) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		return NotifyMsg{
			Message:  message,
			Level:    level,
			Duration: &duration,
		}
	})
}

func ShowNextNotification() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		return NotifyShowNextMsg{}
	})
}

func NotifyLog(message string) tea.Cmd {
	return Notify(message, Log)
}

func NotifyWarn(message string) tea.Cmd {
	return Notify(message, Warn)
}

func NotifyError(message string) tea.Cmd {
	return Notify(message, Err)
}

func newNotify() modelNotify {
	return modelNotify{
		queue:        newNotifyQueue(),
		styles:       DefaultStyles(),
		isDisplaying: false,
	}
}

func (m modelNotify) Init() tea.Cmd {
	if m.queue.Size() > 0 && !m.isDisplaying {
		return tea.Cmd(func() tea.Msg {
			return NotifyShowNextMsg{}
		})
	}
	return nil
}

func (m modelNotify) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NotifyMsg:
		return m.enqueueMessage(msg)

	case NotifyShowNextMsg:
		return m.startDisplaying()

	case NotifyExpireMsg:
		return m.expireMessage(msg.ID)

	case UpdatePositions:
		h, _ := m.styles.Base.GetFrameSize()
		m.Width = globalWidth - h
		return m, nil

	default:
		return m, nil
	}
}

func (m modelNotify) View() string {
	if m.queue.current == nil || !m.isDisplaying {
		return ""
	}

	msg := m.queue.current
	baseText := " Notification: " + msg.Message

	remaining := m.queue.Remaining()
	if ShowQueueCounter && remaining > 0 {
		baseText += fmt.Sprintf(" (%d more)", remaining)
	}

	return m.styleMessage(baseText, msg.Level)
}

// Private methods for queue management and state machine

func (m modelNotify) enqueueMessage(msg NotifyMsg) (tea.Model, tea.Cmd) {
	m.queue.Enqueue(msg)

	if m.queue.current == nil && !m.isDisplaying {
		return m.startDisplaying()
	}

	return m, nil
}

func (m modelNotify) startDisplaying() (tea.Model, tea.Cmd) {
	msg := m.queue.Dequeue()
	if msg == nil {
		m.isDisplaying = false
		return m, nil
	}

	m.isDisplaying = true

	expireCmd := tea.Tick(msg.Duration, func(t time.Time) tea.Msg {
		return NotifyExpireMsg{ID: msg.ID}
	})

	return m, expireCmd
}

func (m modelNotify) expireMessage(id string) (tea.Model, tea.Cmd) {
	if m.queue.current != nil && m.queue.current.ID == id {
		m.queue.current.State = Expired
		m.queue.current = nil
		m.isDisplaying = false

		return m.startDisplaying()
	}

	return m, nil
}

func (m modelNotify) styleMessage(text string, level NotifyLevel) string {
	style := m.styles.NotifyLog

	switch level {
	case Log:
		style = m.styles.NotifyLog
	case Warn:
		style = m.styles.NotifyWarn
	case Err:
		style = m.styles.NotifyErr
	}

	return style.Width(m.Width).Render(text)
}

// Queue implementation methods

func newNotifyQueue() notifyQueue {
	return notifyQueue{
		messages: make([]QueuedMessage, 0, MaxQueueSize),
		maxSize:  MaxQueueSize,
		nextID:   0,
	}
}

func (q *notifyQueue) Enqueue(msg NotifyMsg) QueuedMessage {
	id := fmt.Sprintf("%s%d", MessageIDPrefix, atomic.AddUint64(&globalMessageID, 1))

	duration := getDurationForLevel(msg.Level)
	if msg.Duration != nil {
		duration = *msg.Duration
	}

	queuedMsg := QueuedMessage{
		ID:        id,
		Message:   msg.Message,
		Level:     msg.Level,
		Duration:  duration,
		Timestamp: time.Now(),
		State:     Queued,
	}

	// Handle queue overflow - remove oldest messages
	totalCapacity := q.maxSize
	if q.current != nil {
		totalCapacity-- // Reserve one slot for current message
	}

	if len(q.messages) >= totalCapacity {
		q.messages = q.messages[1:]
	}

	q.messages = append(q.messages, queuedMsg)

	return queuedMsg
}

func (q *notifyQueue) Dequeue() *QueuedMessage {
	if len(q.messages) == 0 {
		return nil
	}

	msg := q.messages[0]
	q.messages = q.messages[1:]

	msg.State = Displaying
	q.current = &msg

	return &msg
}

func (q *notifyQueue) Size() int {
	size := len(q.messages)
	if q.current != nil {
		size++ // Include currently displayed message
	}
	return size
}

func (q *notifyQueue) Remaining() int {
	return len(q.messages)
}

// Helper functions

func getDurationForLevel(level NotifyLevel) time.Duration {
	switch level {
	case Log:
		return LogDuration
	case Warn:
		return WarnDuration
	case Err:
		return ErrorDuration
	default:
		return WarnDuration
	}
}
