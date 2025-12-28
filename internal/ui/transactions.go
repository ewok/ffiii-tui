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
type FilterAssetMsg struct {
	account string
}

type RefreshTransactionsMsg struct{}
type RefreshExpensesMsg struct{}

type modelTransactions struct {
	table        table.Model
	transactions []firefly.Transaction
	api          *firefly.Api
	currentAsset string
	focus        bool
}

func newModelTransactions(api *firefly.Api) modelTransactions {
	transactions, err := api.ListTransactions("", "", "")
	if err != nil {
		fmt.Println("Error fetching transactions:", err)
		os.Exit(1)
	}

	rows, columns := getRows(transactions)
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
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

	m := modelTransactions{table: t, transactions: transactions, api: api}
	return m
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
			rows, columns := getRows(m.transactions)
			m.table.SetRows(rows)
			m.table.SetColumns(columns)
			return m, nil
		}
		transactions, err := m.api.SearchTransactions(value)
		if err != nil {
			return m, nil
		}
		rows, columns := getRows(transactions)
		m.table.SetRows(rows)
		m.table.SetColumns(columns)
		cmds = append(cmds, tea.Printf("Filtered"))
	case FilterAssetMsg:
		value := msg.account
		m.currentAsset = value
		if value != "" {
			transactions := []firefly.Transaction{}
			for _, tx := range m.transactions {
				if tx.Source == value || tx.Destination == value {
					transactions = append(transactions, tx)
				}
			}
			rows, columns := getRows(transactions)
			m.table.SetRows(rows)
			m.table.SetColumns(columns)
		} else {
			rows, columns := getRows(m.transactions)
			m.table.SetRows(rows)
			m.table.SetColumns(columns)
		}
	case RefreshTransactionsMsg:
		transactions, err := m.api.ListTransactions("", "", "")
		if err != nil {
			return m, nil
		}
		m.transactions = transactions
		cmds = append(cmds, Cmd(FilterAssetMsg{account: m.currentAsset}))
	case tea.WindowSizeMsg:
		h, v := baseStyle.GetFrameSize()
		m.table.SetWidth(msg.Width - h)
		m.table.SetHeight(msg.Height - v - topSize)
	}

	if !m.focus {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.table.Focused() {
			switch msg.String() {
			case "r":
				cmds = append(cmds,
					Cmd(RefreshTransactionsMsg{}),
					Cmd(RefreshAssetsMsg{}),
					Cmd(RefreshExpensesMsg{}))
				return m, tea.Batch(cmds...)
			case "f":
				cmds = append(cmds, Cmd(PromptMsg{
					Prompt: "Filter query: ",
					Value:  "",
					Callback: func(value string) []tea.Cmd {
						var cmds []tea.Cmd
						cmds = append(cmds,
							Cmd(FilterMsg{query: value}),
							Cmd(ViewTransactionsMsg{}))
						return cmds
					}}))
				return m, tea.Batch(cmds...)
			case "n":
				cmds = append(cmds, Cmd(ViewNewMsg{}))
			case "N":
				row := m.table.SelectedRow()
                id, err := strconv.Atoi(row[0])
                if err != nil {
                    break
                }
				trx := m.transactions[id-1]
				cmds = append(cmds, Cmd(NewTransactionMsg{
					transaction: trx,
				}), Cmd(ViewNewMsg{}))
			case "a":
				cmds = append(cmds, Cmd(ViewAssetsMsg{}))
			case "ctrl+a":
				cmds = append(cmds, Cmd(FilterAssetMsg{account: ""}))
			case "c":
				cmds = append(cmds, Cmd(ViewCategoriesMsg{}))
			case "e":
				cmds = append(cmds, Cmd(ViewExpensesMsg{}))
			case "i":
				cmds = append(cmds, Cmd(ViewRevenuesMsg{}))
			case "t":
				cmds = append(cmds, Cmd(ViewFullTransactionViewMsg{}))
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

func (m *modelTransactions) Blur() {
	m.table.Blur()
	m.focus = false
}

func (m *modelTransactions) Focus() {
	m.table.Focus()
	m.focus = true
}

func getRows(transactions []firefly.Transaction) ([]table.Row, []table.Column) {

	// Determine max widths for dynamic columns
	sourceWidth := 5
	destinationWidth := 5
	categoryWidth := 5
	amountWidth := 5
	foreignAmountWidth := 5
	descriptionWidth := 10
	transactionIDWidth := 4

	rows := []table.Row{}
	// Populate rows from transactions
	for id, tx := range transactions {

		// Parse amounts
		fAmount, err := strconv.ParseFloat(tx.Amount, 64)
		if err != nil {
			amount = "N/A"
		} else {
			amount = fmt.Sprintf("%.2f", fAmount)
		}

		fForeignAmount, err := strconv.ParseFloat(tx.ForeignAmount, 64)
		if err != nil {
			foreignAmount = "N/A"
		} else {
			foreignAmount = fmt.Sprintf("%.2f", fForeignAmount)
		}

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
			fmt.Sprintf("%d", id+1),
			Type,
			date.Format("2006-01-02"),
			tx.Source,
			tx.Destination,
			tx.Category,
			tx.Currency,
			amount,
			tx.ForeignCurrency,
			foreignAmount,
			tx.Description,
			tx.TransactionID,
		}
		rows = append(rows, row)

		// Update max widths
		sourceLen := len(tx.Source)
		if sourceLen > sourceWidth {
			sourceWidth = sourceLen
		}
		destinationLen := len(tx.Destination)
		if destinationLen > destinationWidth {
			destinationWidth = destinationLen
		}
		categoryLen := len(tx.Category)
		if categoryLen > categoryWidth {
			categoryWidth = categoryLen
		}
		amountLen := len(amount)
		if amountLen > amountWidth {
			amountWidth = amountLen
		}
		foreignAmountLen := len(foreignAmount)
		if foreignAmountLen > foreignAmountWidth {
			foreignAmountWidth = foreignAmountLen
		}
		descriptionLen := len(tx.Description)
		if descriptionLen > descriptionWidth {
			descriptionWidth = descriptionLen
		}
		transactionIDLen := len(tx.TransactionID)
		if transactionIDLen > transactionIDWidth {
			transactionIDWidth = transactionIDLen
		}
	}

	return rows, []table.Column{
		{Title: "ID", Width: 4},
		{Title: "Type", Width: 2},
		{Title: "Date", Width: 10},
		{Title: "Source", Width: sourceWidth},
		{Title: "Destination", Width: destinationWidth},
		{Title: "Category", Width: categoryWidth},
		{Title: "Currency", Width: 5},
		{Title: "Amount", Width: amountWidth},
		{Title: "Foreign Currency", Width: 5},
		{Title: "Foreign Amount", Width: foreignAmountWidth},
		{Title: "Description", Width: descriptionWidth},
		{Title: "TxID", Width: transactionIDWidth},
	}

}
