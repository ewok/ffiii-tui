/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package notify

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	NotifyLog  lipgloss.Style
	NotifyWarn lipgloss.Style
	NotifyErr  lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		NotifyLog:  lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")),
		NotifyWarn: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFAF00")),
		NotifyErr:  lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")),
	}
}
