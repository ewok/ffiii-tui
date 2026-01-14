/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"ffiii-tui/internal/firefly"
	"ffiii-tui/internal/ui/notify"
	"ffiii-tui/internal/ui/prompt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FilterMsg struct {
	Account  firefly.Account
	Category firefly.Category
	Query    string
	Reset    bool
}
type SearchMsg struct {
	Query string
}

type (
	RefreshTransactionsMsg struct{}
	TransactionsUpdateMsg  struct {
		Transactions []firefly.Transaction
	}
	DeleteTransactionMsg struct {
		Transaction firefly.Transaction
	}
)

type modelTransactions struct {
	table           table.Model
	transactions    []firefly.Transaction
	api             *firefly.Api
	currentAccount  firefly.Account
	currentCategory firefly.Category
	currentSearch   string
	currentFilter   string
	focus           bool
	keymap          TransactionsKeyMap
	styles          Styles
}

func newModelTransactions(api *firefly.Api) modelTransactions {
	transactions := []firefly.Transaction{}

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
		styles:       DefaultStyles(),
	}
	return m
}

func (m modelTransactions) Init() tea.Cmd {
	return nil
}

func (m modelTransactions) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SearchMsg:
		if msg.Query == "None" {
			if m.currentSearch == "" {
				return m, nil
			}
			m.currentSearch = ""
			// m.currentSearch, m.currentAccount, m.currentCategory, m.currentFilter = "", "", "", ""
		} else {
			// if m.currentSearch == "" {
			// 	m.currentAccount, m.currentCategory, m.currentFilter = "", "", ""
			// }
			m.currentSearch = msg.Query
		}
		return m, Cmd(RefreshTransactionsMsg{})

	case FilterMsg:
		// Reset flag
		if msg.Reset {
			m.currentAccount = firefly.Account{}
			m.currentCategory = firefly.Category{}
			m.currentFilter = ""
		}

		// if msg.Account == "None" {
		// 	m.currentAccount = firefly.Account{}
		// }
		// if msg.Category == "None" {
		// 	m.currentCategory = firefly.Category{}
		// }

		// Clear filters if user set to "None"
		if msg.Query == "None" {
			m.currentFilter = ""
		}

		// Reset other filters if the same filter is applied again
		if !msg.Account.IsEmpty() {
			if msg.Account == m.currentAccount {
				m.currentCategory = firefly.Category{}
				m.currentFilter = ""
			}
			m.currentAccount = msg.Account
		}
		if !msg.Category.IsEmpty() {
			if msg.Category == m.currentCategory {
				m.currentAccount = firefly.Account{}
				m.currentFilter = ""
			}
			m.currentCategory = msg.Category
		}
		if msg.Query != "" && msg.Query != "None" {
			if msg.Query == m.currentFilter {
				m.currentAccount = firefly.Account{}
				m.currentCategory = firefly.Category{}
			}
			m.currentFilter = msg.Query
		}

		transactions := m.transactions

		if !m.currentAccount.IsEmpty() {
			txs := []firefly.Transaction{}
			for _, tx := range transactions {
				for _, split := range tx.Splits {
					if split.Source == m.currentAccount || split.Destination == m.currentAccount {
						txs = append(txs, tx)
						break
					}
				}
			}
			transactions = txs
		}

		if !m.currentCategory.IsEmpty() {
			txs := []firefly.Transaction{}
			for _, tx := range transactions {
				for _, split := range tx.Splits {
					if split.Category == m.currentCategory {
						txs = append(txs, tx)
						break
					}
				}
			}
			transactions = txs
		}

		value := m.currentFilter
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
		return m, func() tea.Msg {
			var err error
			transactions := []firefly.Transaction{}
			if m.currentSearch != "" {
				transactions, err = m.api.ListTransactions(url.QueryEscape(m.currentSearch))
				if err != nil {
					return notify.NotifyWarn(err.Error())
				}
			} else {
				transactions, err = m.api.ListTransactions("")
				if err != nil {
					return notify.NotifyWarn(err.Error())
				}
			}
			return TransactionsUpdateMsg{
				Transactions: transactions,
			}
		}

	case TransactionsUpdateMsg:
		m.transactions = msg.Transactions
		return m, tea.Batch(Cmd(FilterMsg{
			Account:  m.currentAccount,
			Category: m.currentCategory,
			Query:    m.currentFilter,
		}), notify.NotifyLog("Transactions loaded"))

	case DeleteTransactionMsg:
		id := msg.Transaction.TransactionID
		if id != "" {
			err := m.api.DeleteTransaction(id)
			if err != nil {
				return m, tea.Batch(
					notify.NotifyError(fmt.Sprint("Error deleting transaction, ", err.Error())),
					SetView(transactionsView))
			}
			return m, tea.Batch(
				notify.NotifyLog("Transaction deleted successfully."),
				SetView(transactionsView),
				Cmd(RefreshAssetsMsg{}),
				Cmd(RefreshLiabilitiesMsg{}),
				Cmd(RefreshSummaryMsg{}),
				Cmd(RefreshTransactionsMsg{}),
				Cmd(RefreshExpenseInsightsMsg{}),
				Cmd(RefreshRevenueInsightsMsg{}))
		}
		return m, SetView(transactionsView)
	case UpdatePositions:
		h, v := m.styles.Base.GetFrameSize()
		m.table.SetWidth(globalWidth - leftSize - h)
		m.table.SetHeight(globalHeight - topSize - v)
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
			return m, Cmd(RefreshAllMsg{})
		case key.Matches(msg, m.keymap.Filter):
			return m, prompt.Ask(
				"Filter query: ",
				m.currentFilter,
				func(value string) tea.Cmd {
					var cmds []tea.Cmd
					cmds = append(cmds,
						Cmd(FilterMsg{Query: value}),
						SetView(transactionsView))
					return tea.Sequence(cmds...)
				},
			)
		case key.Matches(msg, m.keymap.Search):
			return m, prompt.Ask(
				"Search query: ",
				m.currentSearch,
				func(value string) tea.Cmd {
					var cmds []tea.Cmd
					cmds = append(cmds,
						Cmd(SearchMsg{Query: value}),
						SetView(transactionsView),
					)
					return tea.Sequence(cmds...)
				},
			)
		case key.Matches(msg, m.keymap.NewView):
			return m, Cmd(NewTransactionMsg{
				Transaction: firefly.Transaction{
					Splits: []firefly.Split{
						{
							Source:      m.currentAccount,
							Destination: firefly.Account{},
							Category:    m.currentCategory,
						},
					},
				},
			})
		case key.Matches(msg, m.keymap.NewTransactionFrom):
			trx, err := m.GetCurrentTransaction()
			if err != nil {
				return m, notify.NotifyWarn(err.Error())
			}
			return m, Cmd(NewTransactionFromMsg{Transaction: trx})
		case key.Matches(msg, m.keymap.Select):
			trx, err := m.GetCurrentTransaction()
			if err != nil {
				return m, notify.NotifyWarn(err.Error())
			}
			return m, tea.Sequence(
				Cmd(EditTransactionMsg{Transaction: trx}),
				SetView(newView))
		case key.Matches(msg, m.keymap.Delete):
			if len(m.table.Rows()) < 1 {
				return m, notify.NotifyWarn("No transactions.")
			}
			row := m.table.SelectedRow()
			if row == nil {
				return m, notify.NotifyWarn("Transaction not selected.")
			}
			id, err := strconv.Atoi(row[0])
			if err != nil {
				return m, nil
			}
			trx := m.transactions[id]
			return m, prompt.Ask(
				fmt.Sprintf("Are you sure you want to delete the transaction? Type 'yes!' to confirm. Transaction: %s - %s: ", trx.TransactionID, trx.Description()),
				"no",
				func(value string) tea.Cmd {
					var cmd tea.Cmd
					if value == "yes!" {
						cmd = Cmd(DeleteTransactionMsg{Transaction: trx})
					}
					return tea.Sequence(SetView(transactionsView), cmd)
				},
			)
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

func (m *modelTransactions) GetCurrentTransaction() (firefly.Transaction, error) {
	if len(m.table.Rows()) < 1 {
		return firefly.Transaction{}, fmt.Errorf("No transactions in the list")
	}
	row := m.table.SelectedRow()
	if row == nil {
		return firefly.Transaction{}, fmt.Errorf("Transaction not selected")
	}
	id, err := strconv.Atoi(row[0])
	if err != nil {
		return firefly.Transaction{}, fmt.Errorf("Wrong transaction id: %s", row[0])
	}

	if id >= len(m.transactions) {
		return firefly.Transaction{}, fmt.Errorf("Index out of range: %d", id)
	}

	return m.transactions[id], nil
}
