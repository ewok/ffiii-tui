/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package period

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	Border   lipgloss.Style
	Item     lipgloss.Style
	Selected lipgloss.Style
	Current  lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("#5F5FD7")),
		Item: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#585858")),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#D75F87")),
		Current: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5F5FD7")),
	}
}
