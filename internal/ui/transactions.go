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
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FilterMsg struct {
	account  string
	category string
	query    string
	reset    bool
}
type SearchMsg struct {
	query string
}

type RefreshTransactionsMsg struct{}

type modelTransactions struct {
	table           table.Model
	transactions    []firefly.Transaction
	api             *firefly.Api
	currentAccount  string
	currentCategory string
	currentSearch   string
	currentFilter   string
	focus           bool
}

func newModelTransactions(api *firefly.Api) modelTransactions {
	transactions, err := api.ListTransactions()
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

	switch msg := msg.(type) {
	case SearchMsg:
		if msg.query == "None" && m.currentSearch == "" {
			return m, nil
		}
		if msg.query == "None" {
			m.currentSearch = ""
		} else {
			m.currentSearch = msg.query
		}
		return m, Cmd(RefreshTransactionsMsg{})
	case FilterMsg:
		// Reset flag
		if msg.reset {
			m.currentAccount = ""
			m.currentCategory = ""
			m.currentFilter = ""
		}

		// Clear filters if user set to "None"
		if msg.account == "None" {
			m.currentAccount = ""
		}
		if msg.category == "None" {
			m.currentCategory = ""
		}
		if msg.query == "None" {
			m.currentFilter = ""
		}

		// Reset other filters if the same filter is applied again
		if msg.account != "" && msg.account != "None" {
			if msg.account == m.currentAccount {
				m.currentCategory = ""
				m.currentFilter = ""
			}
			m.currentAccount = msg.account
		}
		if msg.category != "" && msg.category != "None" {
			if msg.category == m.currentCategory {
				m.currentAccount = ""
				m.currentFilter = ""
			}
			m.currentCategory = msg.category
		}
		if msg.query != "" && msg.query != "None" {
			if msg.query == m.currentFilter {
				m.currentAccount = ""
				m.currentCategory = ""
			}
			m.currentFilter = msg.query
		}

		transactions := m.transactions

		value := m.currentAccount
		if value != "" {
			txs := []firefly.Transaction{}
			for _, tx := range transactions {
				if tx.Source.Name == value || tx.Destination.Name == value {
					txs = append(txs, tx)
				}
			}
			transactions = txs
		}

		value = m.currentCategory
		if value != "" {
			txs := []firefly.Transaction{}
			for _, tx := range transactions {
				if tx.Category.Name == value {
					txs = append(txs, tx)
				}
			}
			transactions = txs
		}

		value = m.currentFilter
		if value != "" {
			txs := []firefly.Transaction{}
			for _, tx := range transactions {
				if CaseInsensitiveContains(tx.Description, value) ||
					CaseInsensitiveContains(tx.Source.Name, value) ||
					CaseInsensitiveContains(tx.Destination.Name, value) ||
					CaseInsensitiveContains(tx.Category.Name, value) ||
					CaseInsensitiveContains(tx.Currency, value) ||
					strings.Contains(fmt.Sprintf("%.2f", tx.Amount), value) ||
					CaseInsensitiveContains(tx.ForeignCurrency, value) ||
					strings.Contains(fmt.Sprintf("%.2f", tx.ForeignAmount), value) {
					txs = append(txs, tx)
				}
			}
			transactions = txs
		}

		rows, columns := getRows(transactions)
		m.table.SetRows(rows)
		m.table.SetColumns(columns)

	case RefreshTransactionsMsg:
		return m, Cmd(func() tea.Msg {
			var err error
			transactions := []firefly.Transaction{}
			if m.currentSearch != "" {
				transactions, err = m.api.SearchTransactions(m.currentSearch)
				if err != nil {
					return nil
				}
			} else {
				transactions, err = m.api.ListTransactions()
				if err != nil {
					return nil
				}
			}
			m.transactions = transactions
			return FilterMsg{
				account:  m.currentAccount,
				category: m.currentCategory,
				query:    m.currentFilter,
			}
		}())

	case tea.WindowSizeMsg:
		h, v := baseStyle.GetFrameSize()
		m.table.SetWidth(msg.Width - h)
		m.table.SetHeight(msg.Height - v - topSize)
	}

	if !m.focus {
		return m, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.table.Focused() {
			switch msg.String() {
			case "r":
				return m, tea.Sequence(
					Cmd(RefreshTransactionsMsg{}),
					Cmd(RefreshAssetsMsg{}),
					Cmd(RefreshExpensesMsg{}),
					Cmd(RefreshRevenuesMsg{}),
					Cmd(RefreshCategoriesMsg{}),
				)
			case "f":
				return m, Cmd(PromptMsg{
					Prompt: "Filter query: ",
					Value:  m.currentFilter,
					Callback: func(value string) tea.Cmd {
						var cmds []tea.Cmd
						cmds = append(cmds,
							Cmd(FilterMsg{query: value}),
							Cmd(ViewTransactionsMsg{}))
						return tea.Sequence(cmds...)
					}})
			case "s":
				return m, Cmd(PromptMsg{
					Prompt: "Search query: ",
					Value:  m.currentSearch,
					Callback: func(value string) tea.Cmd {
						var cmds []tea.Cmd
						cmds = append(cmds,
							Cmd(SearchMsg{query: value}),
							Cmd(ViewTransactionsMsg{}),
						)
						return tea.Sequence(cmds...)
					}})
			case "n":
				return m, Cmd(ViewNewMsg{})
			case "N":
				row := m.table.SelectedRow()
				id, err := strconv.Atoi(row[0])
				if err != nil {
					return m, nil
				}
				trx := m.transactions[id]
				return m, tea.Sequence(
					Cmd(NewTransactionMsg{transaction: trx}),
					Cmd(ViewNewMsg{}))
			case "a":
				return m, Cmd(ViewAssetsMsg{})
			case "ctrl+a":
				return m, Cmd(FilterMsg{reset: true})
			case "c":
				return m, Cmd(ViewCategoriesMsg{})
			case "e":
				return m, Cmd(ViewExpensesMsg{})
			case "i":
				return m, Cmd(ViewRevenuesMsg{})
			case "t":
				return m, Cmd(ViewFullTransactionViewMsg{})
				// enter
				// case "enter":
				// 	return m, tea.Batch(
				// 		tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
				// 	)
			case "q":
				return m, tea.Quit
			}
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
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
	for _, tx := range transactions {
		amount = fmt.Sprintf("%.2f", tx.Amount)
		foreignAmount = fmt.Sprintf("%.2f", tx.ForeignAmount)

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
			fmt.Sprintf("%d", tx.ID),
			Type,
			date.Format("2006-01-02"),
			tx.Source.Name,
			tx.Destination.Name,
			tx.Category.Name,
			tx.Currency,
			amount,
			tx.ForeignCurrency,
			foreignAmount,
			tx.Description,
			tx.TransactionID,
		}
		rows = append(rows, row)

		// Update max widths
		sourceLen := len(tx.Source.Name)
		if sourceLen > sourceWidth {
			sourceWidth = sourceLen
		}
		destinationLen := len(tx.Destination.Name)
		if destinationLen > destinationWidth {
			destinationWidth = destinationLen
		}
		categoryLen := len(tx.Category.Name)
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
		{Title: "Currency", Width: 3},
		{Title: "Amount", Width: amountWidth},
		{Title: "Foreign Currency", Width: 4},
		{Title: "Foreign Amount", Width: foreignAmountWidth},
		{Title: "Description", Width: descriptionWidth},
		{Title: "TxID", Width: transactionIDWidth},
	}

}
