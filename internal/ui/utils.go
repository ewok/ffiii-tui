/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func daysIn(m int, year int) int {
	month := time.Month(m)
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func Cmd(msg tea.Msg) tea.Cmd {
	return func() tea.Msg { return msg }
}
