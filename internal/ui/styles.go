/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	ListItem         lipgloss.Style
	ListSelectedItem lipgloss.Style

	Base        lipgloss.Style
	BaseFocused lipgloss.Style

	LeftPanel lipgloss.Style

	Prompt        lipgloss.Style
	PromptFocused lipgloss.Style
	PromptNewTr   lipgloss.Style
	PromptEditTr  lipgloss.Style

	HelpFullKey  lipgloss.Style
	HelpShortKey lipgloss.Style

	NotifyLog  lipgloss.Style
	NotifyWarn lipgloss.Style
	NotifyErr  lipgloss.Style

	Withdrawal lipgloss.Style
	Deposit    lipgloss.Style
	Normal     lipgloss.Style

	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
}

func DefaultStyles() Styles {
	// Base styles for consistent theming
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#585858"))

	baseStyleFocused := baseStyle.
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("#5F5FD7"))

	return Styles{
		// List styles
		ListItem: lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2),
		ListSelectedItem: lipgloss.NewStyle().
			PaddingLeft(0).
			Foreground(lipgloss.Color("#D75F87")),

		// Base component styles
		Base:        baseStyle,
		BaseFocused: baseStyleFocused,
		LeftPanel:   lipgloss.NewStyle().PaddingRight(1),

		// Prompt styles
		Prompt:        baseStyle,
		PromptFocused: baseStyleFocused.BorderForeground(lipgloss.Color("#FF5555")),
		PromptNewTr:   baseStyle.BorderForeground(lipgloss.Color("#00AF00")),
		PromptEditTr:  baseStyle.BorderForeground(lipgloss.Color("#FFAF00")),

		// Help styles
		HelpFullKey:  lipgloss.NewStyle().PaddingLeft(1),
		HelpShortKey: lipgloss.NewStyle().PaddingLeft(1),

		// Notification styles
		NotifyLog:  lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")),
		NotifyWarn: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAF00")),
		NotifyErr:  lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")),

		// Transaction type styles
		Withdrawal: lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")),
		Deposit:    lipgloss.NewStyle().Foreground(lipgloss.Color("#00AF00")),
		Normal:     lipgloss.NewStyle().Foreground(lipgloss.Color("#DDDADA")),

		// Tab bar styles
		TabActive:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5F5FD7")),
		TabInactive: lipgloss.NewStyle().Foreground(lipgloss.Color("#585858")),
	}
}
