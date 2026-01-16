/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"ffiii-tui/internal/firefly"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type AccountListConfig[T firefly.Account] struct {
	// Data specific
	AccountType string

	// Display
	Title string

	// Data functions
	GetItems     func(api any, sorted bool) []list.Item
	RefreshItems func(api any, accountType string) error

	// Messages
	RefreshMsgType any
	UpdateMsgType  any

	// UI Behavior
	PromptNewFunc func() tea.Cmd
	HasSort       bool
	HasTotalRow   bool
	HasSummary    bool
	GetTotalFunc  func(api any) float64 // for totals

	FilterFunc func(item list.Item) tea.Cmd
	SelectFunc func(item list.Item) tea.Cmd
}
