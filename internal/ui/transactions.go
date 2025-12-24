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
type FilterAccountMsg struct {
	account string
}

type RefreshAssetsMsg struct{}
type RefreshTransactionsMsg struct{}
type RefreshExpensesMsg struct{}

type modelTransactions struct {
	table          table.Model
	transactions   []firefly.Transaction
	fireflyApi     *firefly.Api
	currentAccount string
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
	{Title: "TxID", Width: 5},
}

func InitList(api *firefly.Api) modelTransactions {
	transactions, err := api.ListTransactions("", "", "")
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

	m := modelTransactions{table: t, transactions: transactions, fireflyApi: api}
	return m
}

func getRows(transactions []firefly.Transaction) []table.Row {

	rows := []table.Row{}
	// Populate rows from transactions
	for _, tx := range transactions {

		// Parse amounts
		amount, _ := strconv.ParseFloat(tx.Amount, 64)
		foreignAmount, _ := strconv.ParseFloat(tx.ForeignAmount, 64)

		// Convert from string to desired date format
		// YYYY-MM-DD HH:MM:SS
		date, _ := time.Parse(time.RFC3339, tx.Date)

		// Determine type icon
		Type := ""
		switch tx.Type {
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
			tx.Source,
			tx.Destination,
			tx.Category,
			tx.Currency,
			fmt.Sprintf("%.2f", amount),
			tx.ForeignCurrency,
			fmt.Sprintf("%.2f", foreignAmount),
			tx.Description,
			tx.TransactionID,
		}
		rows = append(rows, row)
	}

	return rows
}

func (m modelTransactions) Init() tea.Cmd {
	return nil
}

func (m modelTransactions) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		cmds = append(cmds, tea.Printf("Filtered"))
	case FilterAccountMsg:
		value := msg.account
		m.currentAccount = value
		if value != "" {
			transactions := []firefly.Transaction{}
			for _, tx := range m.transactions {
				if tx.Source == value || tx.Destination == value {
					transactions = append(transactions, tx)
				}
			}
			m.table.SetRows(getRows(transactions))
		} else {
			m.table.SetRows(getRows(m.transactions))
		}
	case RefreshTransactionsMsg:
		transactions, err := m.fireflyApi.ListTransactions("", "", "")
		if err != nil {
			return m, nil
		}
		m.transactions = transactions
		cmds = append(cmds, func() tea.Msg { return FilterAccountMsg{account: m.currentAccount} })
		cmds = append(cmds, tea.Printf("Refreshed"))
	case RefreshAssetsMsg:
		m.fireflyApi.UpdateAssets()
		cmds = append(cmds, func() tea.Msg { return RefreshBalanceMsg{} })
	case RefreshExpensesMsg:
		m.fireflyApi.UpdateExpenses()
	case tea.WindowSizeMsg:
		h, v := baseStyle.GetFrameSize()
		m.table.SetWidth(msg.Width - h)
		m.table.SetHeight(msg.Height - v)
	case tea.KeyMsg:
		if m.table.Focused() {

			switch msg.String() {
			case "r":
				cmds = append(cmds, func() tea.Msg { return RefreshTransactionsMsg{} })
				cmds = append(cmds, func() tea.Msg { return RefreshAssetsMsg{} })
				cmds = append(cmds, func() tea.Msg { return RefreshExpensesMsg{} })
				return m, tea.Batch(cmds...)
			case "f":
				cmds = append(cmds, func() tea.Msg { return viewFilterMsg{} })
			case "n":
				cmds = append(cmds, func() tea.Msg { return viewNewMsg{} })
			case "a":
				cmds = append(cmds, func() tea.Msg { return viewAccountsMsg{} })
			case "ctrl+a":
				cmds = append(cmds, func() tea.Msg { return FilterAccountMsg{account: ""} })
			// enter
			// case "enter":
			// 	return m, tea.Batch(
			// 		tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			// 	)
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}

	}
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m modelTransactions) View() string {
	return m.table.View()
}
