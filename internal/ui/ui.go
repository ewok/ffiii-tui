package ui

import (
	"ffiii-tui/internal/firefly"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/charmbracelet/bubbles/textinput"

	"github.com/charmbracelet/lipgloss"
)

type model struct {
	table        table.Model
	transactions []firefly.Transaction
	filter       textinput.Model
	fireflyApi   *firefly.Api
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var tableColumns = []table.Column{
	{Title: "Type", Width: 2},
	{Title: "Date", Width: 20},
	{Title: "Source", Width: 30},
	{Title: "Destination", Width: 30},
	{Title: "Category", Width: 40},
	{Title: "Currency", Width: 8},
	{Title: "Amount", Width: 14},
	{Title: "Foreign Currency", Width: 8},
	{Title: "Foreign Amount", Width: 14},
	{Title: "Description", Width: 30},
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

func Show(transactions []firefly.Transaction, api *firefly.Api) {

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

	ti := textinput.New()
	ti.Placeholder = "Filter"
	ti.CharLimit = 156
	ti.Width = 20

	m := model{table: t, transactions: transactions, filter: ti, fireflyApi: api}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.table.Focused() {

		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.table.SetWidth(msg.Width - 2)
			m.table.SetHeight(msg.Height - 5)
			return m, nil
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				if m.table.Focused() {
					m.table.Blur()
				} else {
					m.table.Focus()
					m.filter.Blur()
				}
			case "backspace":
				if len(m.table.Rows()) > 0 {
					index := m.table.Cursor()
					rows := m.table.Rows()
					rows = append(rows[:index], rows[index+1:]...)
					m.table.SetRows(rows)
					if index >= len(rows) && index > 0 {
						m.table.SetCursor(index - 1)
					}
				}
			// refresh
			case "r":
				m.table.SetRows(getRows(m.transactions))
				return m, tea.Batch(
					tea.Printf("Refreshed"),
				)
			// filter
			case "f":
				m.filter.Focus()
				m.table.Blur()
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
		return m, cmd
	}

	if m.filter.Focused() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.table.Focus()
				m.filter.Blur()
			case "enter":
				m.table.Focus()
				m.filter.Blur()

				value := m.filter.Value()

				if value == "" {
					m.table.SetRows(getRows(m.transactions))
					return m, nil
				}

				transactions := []firefly.Transaction{}
				page := 1
				for {
					txs, err := m.fireflyApi.SearchTransactions(page, 20, value)
					if err != nil {
						return m, nil
					}
					if len(txs) == 0 {
						break
					}
					transactions = append(transactions, txs...)
					page++
				}

				m.table.SetRows(getRows(transactions))

			}
		}

		m.filter, cmd = m.filter.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("filter: %s", m.filter.View()) + "\n" +
		baseStyle.Render(m.table.View()) + "\n"
}
