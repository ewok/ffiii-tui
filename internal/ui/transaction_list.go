/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"ffiii-tui/internal/firefly"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FilterMsg struct {
	Account  string
	Category string
	Query    string
	Reset    bool
}
type SearchMsg struct {
	Query string
}

type (
	RefreshTransactionsMsg struct{}
	DeleteTransactionMsg   struct {
		Transaction firefly.Transaction
	}
)

type modelTransactions struct {
	table           table.Model
	transactions    []firefly.Transaction
	api             *firefly.Api
	currentAccount  string
	currentCategory string
	currentSearch   string
	currentFilter   string
	focus           bool
	keymap          TransactionsKeyMap
}

func newModelTransactions(api *firefly.Api) modelTransactions {
	transactions, err := api.ListTransactions("")
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

	m := modelTransactions{
		table:        t,
		transactions: transactions,
		api:          api,
		keymap:       DefaultTransactionsKeyMap(),
	}
	return m
}

func (m modelTransactions) Init() tea.Cmd {
	return nil
}

func (m modelTransactions) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SearchMsg:
		if msg.Query == "None" && m.currentSearch == "" {
			return m, nil
		}
		if msg.Query == "None" {
			m.currentSearch = ""
		} else {
			m.currentSearch = msg.Query
		}
		return m, Cmd(RefreshTransactionsMsg{})
	case FilterMsg:
		// Reset flag
		if msg.Reset {
			m.currentAccount = ""
			m.currentCategory = ""
			m.currentFilter = ""
		}

		// Clear filters if user set to "None"
		if msg.Account == "None" {
			m.currentAccount = ""
		}
		if msg.Category == "None" {
			m.currentCategory = ""
		}
		if msg.Query == "None" {
			m.currentFilter = ""
		}

		// Reset other filters if the same filter is applied again
		if msg.Account != "" && msg.Account != "None" {
			if msg.Account == m.currentAccount {
				m.currentCategory = ""
				m.currentFilter = ""
			}
			m.currentAccount = msg.Account
		}
		if msg.Category != "" && msg.Category != "None" {
			if msg.Category == m.currentCategory {
				m.currentAccount = ""
				m.currentFilter = ""
			}
			m.currentCategory = msg.Category
		}
		if msg.Query != "" && msg.Query != "None" {
			if msg.Query == m.currentFilter {
				m.currentAccount = ""
				m.currentCategory = ""
			}
			m.currentFilter = msg.Query
		}

		transactions := m.transactions

		value := m.currentAccount
		if value != "" {
			txs := []firefly.Transaction{}
			for _, tx := range transactions {
				for _, split := range tx.Splits {
					if split.Source.Name == value || split.Destination.Name == value {
						txs = append(txs, tx)
						break
					}
				}
			}
			transactions = txs
		}

		value = m.currentCategory
		if value != "" {
			txs := []firefly.Transaction{}
			for _, tx := range transactions {
				for _, split := range tx.Splits {
					if split.Category.Name == value {
						txs = append(txs, tx)
						break
					}
				}
			}
			transactions = txs
		}

		value = m.currentFilter
		if value != "" {
			txs := []firefly.Transaction{}
			for _, tx := range transactions {
				if CaseInsensitiveContains(tx.GroupTitle, value) {
					txs = append(txs, tx)
					continue
				}
				for _, split := range tx.Splits {
					if CaseInsensitiveContains(split.Description, value) ||
						CaseInsensitiveContains(split.Source.Name, value) ||
						CaseInsensitiveContains(split.Destination.Name, value) ||
						CaseInsensitiveContains(split.Category.Name, value) ||
						CaseInsensitiveContains(split.Currency, value) ||
						strings.Contains(fmt.Sprintf("%.2f", split.Amount), value) ||
						CaseInsensitiveContains(split.ForeignCurrency, value) ||
						strings.Contains(fmt.Sprintf("%.2f", split.ForeignAmount), value) {
						txs = append(txs, tx)
						break
					}
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
				transactions, err = m.api.ListTransactions(url.QueryEscape(m.currentSearch))
				if err != nil {
					return NotifyMsg{
						Message: err.Error(),
						Level:   Warning,
					}
				}
			} else {
				transactions, err = m.api.ListTransactions("")
				if err != nil {
					return NotifyMsg{
						Message: err.Error(),
						Level:   Warning,
					}
				}
			}
			m.transactions = transactions
			return FilterMsg{
				Account:  m.currentAccount,
				Category: m.currentCategory,
				Query:    m.currentFilter,
			}
		}())

	case DeleteTransactionMsg:
		id := msg.Transaction.TransactionID
		if id != "" {
			err := m.api.DeleteTransaction(id)
			if err != nil {
				return m, tea.Batch(
					Notify(fmt.Sprint("Error deleting transaction, ", err.Error()), Warning),
					SetView(transactionsView))
			}
			return m, tea.Batch(
				Notify("Transaction deleted successfully.", Log),
				SetView(transactionsView),
				Cmd(RefreshAssetsMsg{}),
				Cmd(RefreshLiabilitiesMsg{}),
				Cmd(RefreshTransactionsMsg{}),
				Cmd(RefreshExpenseInsightsMsg{}),
				Cmd(RefreshRevenueInsightsMsg{}))
		}
		return m, SetView(transactionsView)
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
		switch {
		case key.Matches(msg, m.keymap.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keymap.Refresh):
			return m, tea.Batch(
				Cmd(RefreshTransactionsMsg{}),
				Cmd(RefreshAssetsMsg{}),
				Cmd(RefreshExpensesMsg{}),
				Cmd(RefreshRevenuesMsg{}),
				Cmd(RefreshCategoriesMsg{}),
			)
		case key.Matches(msg, m.keymap.Filter):
			return m, Cmd(PromptMsg{
				Prompt: "Filter query: ",
				Value:  m.currentFilter,
				Callback: func(value string) tea.Cmd {
					var cmds []tea.Cmd
					cmds = append(cmds,
						Cmd(FilterMsg{Query: value}),
						SetView(transactionsView))
					return tea.Sequence(cmds...)
				},
			})
		case key.Matches(msg, m.keymap.Search):
			return m, Cmd(PromptMsg{
				Prompt: "Search query: ",
				Value:  m.currentSearch,
				Callback: func(value string) tea.Cmd {
					var cmds []tea.Cmd
					cmds = append(cmds,
						Cmd(SearchMsg{Query: value}),
						SetView(transactionsView),
					)
					return tea.Sequence(cmds...)
				},
			})
		case key.Matches(msg, m.keymap.New):
			return m, SetView(newView)
		case key.Matches(msg, m.keymap.NewFromTransaction):
			if len(m.table.Rows()) < 1 {
				return m, Notify("No transactions.", Warning)
			}
			row := m.table.SelectedRow()
			if row == nil {
				return m, Notify("Transaction not selected.", Warning)
			}
			id, err := strconv.Atoi(row[0])
			if err != nil {
				return m, nil
			}
			trx := m.transactions[id]
			return m, tea.Sequence(
				Cmd(NewTransactionMsg{Transaction: trx}),
				SetView(newView))
		case key.Matches(msg, m.keymap.ResetFilter):
			return m, Cmd(FilterMsg{Reset: true})
		case key.Matches(msg, m.keymap.ToggleFullView):
			return m, Cmd(ViewFullTransactionViewMsg{})
		case key.Matches(msg, m.keymap.ViewAssets):
			return m, SetView(assetsView)
		case key.Matches(msg, m.keymap.ViewCategories):
			return m, SetView(categoriesView)
		case key.Matches(msg, m.keymap.ViewExpenses):
			return m, SetView(expensesView)
		case key.Matches(msg, m.keymap.ViewRevenues):
			return m, SetView(revenuesView)
		case key.Matches(msg, m.keymap.ViewLiabilities):
			return m, SetView(liabilitiesView)
		case key.Matches(msg, m.keymap.Select):
			if len(m.table.Rows()) < 1 {
				return m, Notify("No transactions.", Warning)
			}
			row := m.table.SelectedRow()
			if row == nil {
				return m, Notify("Transaction not selected.", Warning)
			}
			id, err := strconv.Atoi(row[0])
			if err != nil {
				return m, nil
			}
			trx := m.transactions[id]
			return m, tea.Sequence(
				Cmd(EditTransactionMsg{Transaction: trx}),
				SetView(newView))
		case key.Matches(msg, m.keymap.Delete):
			if len(m.table.Rows()) < 1 {
				return m, Notify("No transactions.", Warning)
			}
			row := m.table.SelectedRow()
			if row == nil {
				return m, Notify("Transaction not selected.", Warning)
			}
			id, err := strconv.Atoi(row[0])
			if err != nil {
				return m, nil
			}
			trx := m.transactions[id]
			return m, Cmd(PromptMsg{
				Prompt: fmt.Sprintf("Are you sure you want to delete transaction(type 'yes!' if yes) %s:%s: ", trx.TransactionID, trx.Description()),
				Value:  "no",
				Callback: func(value string) tea.Cmd {
					var cmd tea.Cmd
					if value == "yes!" {
						cmd = Cmd(DeleteTransactionMsg{Transaction: trx})
					}
					return tea.Sequence(SetView(transactionsView), cmd)
				},
			})
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
	sourceWidth := 5
	destinationWidth := 5
	categoryWidth := 5
	amountWidth := 5
	foreignAmountWidth := 5
	descriptionWidth := 10
	currencyWidth := 3
	foreignCurrencyWidth := 4
	transactionIDWidth := 4

	rows := []table.Row{}

	for _, tx := range transactions {
		date, _ := time.Parse(time.RFC3339, tx.Date)

		Type := ""
		switch tx.Type {
		case "withdrawal":
			Type = "←"
		case "deposit":
			Type = "→"
		case "transfer":
			Type = "⇄"
		}

		for idx, split := range tx.Splits {
			icon := Type
			if len(tx.Splits) > 1 && idx > 0 {
				icon = " ↳"
			}
			amount := fmt.Sprintf("%.2f", split.Amount)
			foreignAmount := fmt.Sprintf("%.2f", split.ForeignAmount)

			row := table.Row{
				fmt.Sprintf("%d", tx.ID),
				icon,
				date.Format("2006-01-02"),
				split.Source.Name,
				split.Destination.Name,
				split.Category.Name,
				split.Currency,
				amount,
				split.ForeignCurrency,
				foreignAmount,
				split.Description,
				tx.TransactionID,
			}
			rows = append(rows, row)

			sourceLen := len(split.Source.Name)
			if sourceLen > sourceWidth {
				sourceWidth = sourceLen
			}
			destinationLen := len(split.Destination.Name)
			if destinationLen > destinationWidth {
				destinationWidth = destinationLen
			}
			categoryLen := len(split.Category.Name)
			if categoryLen > categoryWidth {
				categoryWidth = categoryLen
			}
			currencyLen := len(split.Currency)
			if currencyLen > currencyWidth {
				currencyWidth = currencyLen
			}
			amountLen := len(amount)
			if amountLen > amountWidth {
				amountWidth = amountLen
			}
			foreignCurrencyLen := len(split.ForeignCurrency)
			if foreignCurrencyLen > foreignCurrencyWidth {
				foreignCurrencyWidth = foreignCurrencyLen
			}
			foreignAmountLen := len(foreignAmount)
			if foreignAmountLen > foreignAmountWidth {
				foreignAmountWidth = foreignAmountLen
			}
			descriptionLen := len(split.Description)
			if descriptionLen > descriptionWidth {
				descriptionWidth = descriptionLen
			}
			transactionIDLen := len(tx.TransactionID)
			if transactionIDLen > transactionIDWidth {
				transactionIDWidth = transactionIDLen
			}
		}
	}

	return rows, []table.Column{
		{Title: "ID", Width: 0},
		{Title: "Type", Width: 2},
		{Title: "Date", Width: 10},
		{Title: "Source", Width: sourceWidth},
		{Title: "Destination", Width: destinationWidth},
		{Title: "Category", Width: categoryWidth},
		{Title: "Currency", Width: currencyWidth},
		{Title: "Amount", Width: amountWidth},
		{Title: "Foreign Currency", Width: foreignCurrencyWidth},
		{Title: "Foreign Amount", Width: foreignAmountWidth},
		{Title: "Description", Width: descriptionWidth},
		{Title: "TxID", Width: transactionIDWidth},
	}
}
