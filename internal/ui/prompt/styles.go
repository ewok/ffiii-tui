/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package prompt

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	PromptFocused lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		PromptFocused: lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("#FF5555")),
	}
}
