/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"ffiii-tui/internal/firefly"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FilterMsg struct {
	query string
}

type modelList struct {
	table        table.Model
	transactions []firefly.Transaction
	fireflyApi   *firefly.Api
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var tableColumns = []table.Column{
	{Title: "Type", Width: 2},
	{Title: "Date", Width: 20},
	{Title: "Source", Width: 25},
	{Title: "Destination", Width: 25},
	{Title: "Category", Width: 30},
	{Title: "Currency", Width: 5},
	{Title: "Amount", Width: 10},
	{Title: "Foreign Currency", Width: 8},
	{Title: "Foreign Amount", Width: 10},
	{Title: "Description", Width: 30},
}

func InitList(api *firefly.Api) modelList {
	transactions, err := api.ListTransactions("", "")
	if err != nil {
		fmt.Println("Error fetching transactions:", err)
		os.Exit(1)
	}

	t := table.New(
		table.WithColumns(tableColumns),
		table.WithRows(getRows(transactions)),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	m := modelList{table: t, transactions: transactions, fireflyApi: api}
	return m
}

func getRows(transactions []firefly.Transaction) []table.Row {

	rows := []table.Row{}
	// Populate rows from transactions
	for _, tx := range transactions {
		for _, split := range tx.Attributes.Transactions {

			// Parse amounts
			amount, _ := strconv.ParseFloat(split.Amount, 64)
			foreignAmount, _ := strconv.ParseFloat(split.ForeignAmount, 64)

			// Convert from string to desired date format
			// YYYY-MM-DD HH:MM:SS
			date, _ := time.Parse(time.RFC3339, split.Date)

			// Determine type icon
			Type := ""
			switch split.Type {
			case "withdrawal":
				Type = "➖"
			case "deposit":
				Type = "➕"
			case "transfer":
				Type = "➡️"
			}

			row := table.Row{
				Type,
				date.Format("2006-01-02 15:04:05"),
				split.SourceName,
				split.DestinationName,
				split.CategoryName,
				split.CurrencyCode,
				fmt.Sprintf("%.2f", amount),
				split.ForeignCurrencyCode,
				fmt.Sprintf("%.2f", foreignAmount),
				split.Description,
			}
			rows = append(rows, row)
		}
	}

	return rows
}

func (m modelList) Init() tea.Cmd {
	return nil
}

func (m modelList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case FilterMsg:
		value := msg.query
		if value == "" {
			m.table.SetRows(getRows(m.transactions))
			return m, nil
		}
		transactions, err := m.fireflyApi.SearchTransactions(value)
		if err != nil {
			return m, nil
		}
		m.table.SetRows(getRows(transactions))
	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width - 2)
		m.table.SetHeight(msg.Height - 5)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		// case "esc":
		// 	if m.table.Focused() {
		// 		m.table.Blur()
		// 	} else {
		// 		m.table.Focus()
		// 		m.filter.Blur()
		// 	}
		// case "backspace":
		// 	if len(m.table.Rows()) > 0 {
		// 		index := m.table.Cursor()
		// 		rows := m.table.Rows()
		// 		rows = append(rows[:index], rows[index+1:]...)
		// 		m.table.SetRows(rows)
		// 		if index >= len(rows) && index > 0 {
		// 			m.table.SetCursor(index - 1)
		// 		}
		// 	}
		// refresh
		case "r":
			transactions, err := m.fireflyApi.ListTransactions("", "")
			if err != nil {
				return m, nil
			}
			m.table.SetRows(getRows(transactions))
			return m, tea.Batch(
				tea.Printf("Refreshed"),
			)
		// filter
		case "f":
			cmds = append(cmds, func() tea.Msg { return viewFilterMsg{} })
		case "n":
			cmds = append(cmds, func() tea.Msg { return viewNewMsg{} })
		// enter
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			)
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	}
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m modelList) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}
