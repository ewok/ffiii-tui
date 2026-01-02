/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

// TODO: Use last date as input, and key for resetting to today.

import (
	"errors"
	"ffiii-tui/internal/firefly"
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type RefreshNewCategoryMsg struct{}
type RefreshNewAssetMsg struct{}
type RefreshNewExpenseMsg struct{}
type RefreshNewRevenueMsg struct{}

type NewTransactionMsg struct {
	Transaction firefly.Transaction
}

type modelNewTransaction struct {
	form   *huh.Form // huh.Form is just a tea.Model
	api    *firefly.Api
	focus  bool
	keymap TransactionFormKeyMap
}

var (
	transactionType           string
	year                      string
	month                     string
	day                       string
	source                    firefly.Account
	destination               firefly.Account
	category                  firefly.Category
	amount                    string
	foreignAmount             string
	triggerCategoryCounter    byte
	triggerSourceCounter      byte
	triggerDestinationCounter byte
)

func newModelNewTransaction(api *firefly.Api, trx firefly.Transaction) modelNewTransaction {
	now := time.Now()

	if trx.Type != "" {
		transactionType = trx.Type
		year = trx.Date[0:4]
		month = trx.Date[5:7]
		day = trx.Date[8:10]
		source = trx.Source
		destination = trx.Destination
		category = trx.Category
		if trx.Amount != 0 {
			amount = fmt.Sprintf("%.2f", trx.Amount)
		} else {
			amount = "0"
		}
		if trx.ForeignAmount != 0 {
			foreignAmount = fmt.Sprintf("%.2f", trx.ForeignAmount)
		} else {
			foreignAmount = ""
		}
	} else {
		transactionType = "withdrawal"
		year = fmt.Sprintf("%d", now.Year())
		month = fmt.Sprintf("%02d", now.Month())
		day = fmt.Sprintf("%02d", now.Day())
		source = firefly.Account{}
		destination = firefly.Account{}
		category = firefly.Category{}
		amount = ""
		foreignAmount = ""
	}

	years := []string{}
	startYear := now.Year() - 9
	for y := range 10 {
		years = append(years, fmt.Sprintf("%d", startYear+y))
	}

	return modelNewTransaction{
		api:    api,
		keymap: DefaultTransactionFormKeyMap(),
		form: huh.NewForm(
			huh.NewGroup(

				huh.NewNote().
					TitleFunc(func() string {
						s := source.Type
						d := destination.Type
						transactionType = ""

						switch {
						case s == "asset" && (d == "expense" || d == "liabilities"):
							transactionType = "withdrawal"
						case s == "asset" && d == "asset":
							transactionType = "transfer"
						case s == "revenue":
							transactionType = "deposit"
						case s == "liabilities" && d == "expense":
							transactionType = "withdrawal"
						case s == "liabilities" && d == "asset":
							transactionType = "deposit"
						case s == "liabilities" && d == "liabilities":
							transactionType = "transfer"
						default:
							transactionType = "unknown"
						}

						return fmt.Sprintf("Current Type: %s", transactionType)

					}, []any{&source, &destination}),

				huh.NewSelect[firefly.Account]().
					Key("source_name").
					Title("Source Account").
					DescriptionFunc(func() string {
						switch source.Type {
						case "asset", "liabilities":
							return fmt.Sprintf("Balance: %.2f %s", source.Balance, source.CurrencyCode)
						}
						return ""
					}, &source).
					Value(&source).
					OptionsFunc(func() []huh.Option[firefly.Account] {
						options := []huh.Option[firefly.Account]{}

						for _, account := range api.Accounts["asset"] {
							options = append(options, huh.NewOption(account.Name, account))
						}
						for _, account := range api.Accounts["revenue"] {
							options = append(options, huh.NewOption(account.Name, account))
						}
						for _, account := range api.Accounts["liabilities"] {
							options = append(options, huh.NewOption(account.Name, account))
						}
						return options
					}, []any{&triggerSourceCounter}).WithHeight(5),
				huh.NewSelect[firefly.Account]().
					Key("destination_name").
					Title("Destination Account Name").
					DescriptionFunc(func() string {
						switch source.Type {
						case "asset", "liabilities":
							return fmt.Sprintf("Balance: %.2f %s", source.Balance, source.CurrencyCode)
						}
						return ""
					}, &source).
					Value(&destination).
					OptionsFunc(func() []huh.Option[firefly.Account] {
						options := []huh.Option[firefly.Account]{}

						switch source.Type {
						case "asset":
							for _, account := range api.Accounts["asset"] {
								options = append(options, huh.NewOption(account.Name, account))
							}
							for _, account := range api.Accounts["expense"] {
								options = append(options, huh.NewOption(account.Name, account))
							}
							for _, account := range api.Accounts["liabilities"] {
								options = append(options, huh.NewOption(account.Name, account))
							}
						case "revenue":
							for _, account := range api.Accounts["asset"] {
								options = append(options, huh.NewOption(account.Name, account))
							}
							for _, account := range api.Accounts["liabilities"] {
								options = append(options, huh.NewOption(account.Name, account))
							}
						case "liabilities":
							for _, account := range api.Accounts["asset"] {
								options = append(options, huh.NewOption(account.Name, account))
							}
							for _, account := range api.Accounts["expense"] {
								options = append(options, huh.NewOption(account.Name, account))
							}
							for _, account := range api.Accounts["liabilities"] {
								options = append(options, huh.NewOption(account.Name, account))
							}
						}
						return options
					}, []any{&source, &triggerDestinationCounter}).WithHeight(5),
				huh.NewSelect[firefly.Category]().
					Key("category_name").
					Title("Category").
					Value(&category).
					OptionsFunc(func() []huh.Option[firefly.Category] {
						options := []huh.Option[firefly.Category]{}
						for _, category := range api.Categories {
							options = append(options, huh.NewOption(category.Name, category))
						}
						return options
					}, &triggerCategoryCounter).WithHeight(5),
				huh.NewInput().
					Key("amount").
					Title("Amount").
					Value(&amount).
					DescriptionFunc(func() string {
						switch source.Type {
						case "asset", "liabilities":
							return source.CurrencyCode
						case "revenue":
							return destination.CurrencyCode
						}
						return ""

					}, []any{&source, &destination}).
					Validate(func(str string) error {
						var amount float64
						_, err := strconv.ParseFloat(str, 64)
						if err != nil || amount < 0 {
							return errors.New("please enter a valid positive number for amount")
						}
						return nil
					}),
				huh.NewInput().
					Key("foreignAmount").
					Title("Foreign Amount").
					Value(&foreignAmount).
					DescriptionFunc(func() string {
						s := source.Type
						d := destination.Type
						if (s == "asset" || s == "liabilities") && (d == "asset" || d == "liabilities") {
							if source.CurrencyCode == destination.CurrencyCode {
								return "N/A"
							}
							return destination.CurrencyCode
						}
						return "N/A"
					}, []any{&source, &destination}).
					Validate(func(str string) error {
						s := source.Type
						d := destination.Type
						if (s == "asset" || s == "liabilities") && (d == "asset" || d == "liabilities") {
							if source.CurrencyCode == destination.CurrencyCode {
								if str != "" {
									return errors.New("for transfers between same currency accounts, foreign amount should be empty")
								}
								return nil
							}
							var amount float64
							_, err := strconv.ParseFloat(str, 64)
							if err != nil || amount < 0 {
								return errors.New("please enter a valid positive number for amount")
							}
							return nil
						}
						if str != "" {
							return errors.New("foreign amount is only applicable for transactions between asset/liability accounts")
						}
						return nil
					}),
				huh.NewInput().
					Key("description").
					Title("Description").
					PlaceholderFunc(func() string {
						return fmt.Sprintf("%s, %s -> %s", category.Name, source.Name, destination.Name)
					}, []any{&transactionType, &category, &source, &destination}).
					WithWidth(60),
			),
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("year").
					Title("Year").
					Options(huh.NewOptions(years...)...).
					Value(&year).
					WithHeight(3),
				huh.NewSelect[string]().
					Key("month").
					Title("Month").
					Options(huh.NewOptions("01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12")...).
					Value(&month).
					WithHeight(4),
				huh.NewSelect[string]().
					Key("day").
					Title("Day").
					Value(&day).
					Options(huh.NewOptions(day)...).
					OptionsFunc(func() []huh.Option[string] {
						days := []string{}
						// According to month and year, determine number of days
						monthInt, _ := strconv.Atoi(month)
						yearInt, _ := strconv.Atoi(year)
						numDays := daysIn(monthInt, yearInt)
						for d := range numDays {
							days = append(days, fmt.Sprintf("%02d", d+1))
						}
						return huh.NewOptions(days...)
					}, []any{&month, &year}).WithHeight(4),
			),
		).WithLayout(huh.LayoutColumns(2)),
	}
}

func (m modelNewTransaction) Init() tea.Cmd {
	return m.form.Init()
}

func (m modelNewTransaction) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case RefreshNewAssetMsg:
		triggerSourceCounter++
		triggerDestinationCounter++
	case RefreshNewExpenseMsg:
		triggerDestinationCounter++
	case RefreshNewRevenueMsg:
		triggerSourceCounter++
	case RefreshNewCategoryMsg:
		triggerCategoryCounter++
	case NewTransactionMsg:
		newModel := newModelNewTransaction(m.api, msg.Transaction)
		return newModel, SetView(newView)
	}

	if !m.focus {
		return m, nil
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.NewElement):
			field := m.form.GetFocusedField()
			switch field.(type) {
			case *huh.Select[firefly.Category]:
				f, ok := m.form.GetFocusedField().(*huh.Select[firefly.Category])
				if ok && !f.GetFiltering() {
					return m, CmdPromptNewCategory(
						tea.Sequence(
							SetView(newView),
							Cmd(RefreshNewCategoryMsg{})))
				}

			case *huh.Select[firefly.Account]:
				f, ok := m.form.GetFocusedField().(*huh.Select[firefly.Account])
				if ok && !f.GetFiltering() {

					a, ok := f.GetValue().(firefly.Account)
					if ok {
						switch a.Type {
						case "asset":
							return m, CmdPromptNewAsset(
								tea.Sequence(
									SetView(newView),
									Cmd(RefreshNewAssetMsg{})))
						case "expense":
							return m, CmdPromptNewExpense(
								tea.Sequence(
									SetView(newView),
									Cmd(RefreshNewExpenseMsg{})))
						case "revenue":
							return m, CmdPromptNewRevenue(
								tea.Sequence(
									SetView(newView),
									Cmd(RefreshNewRevenueMsg{})))
						case "liability":
							return m, CmdPromptNewLiability(
								tea.Sequence(
									SetView(newView),
									Cmd(RefreshNewAssetMsg{})))
						}
						return m, nil
					}
				}
			}

		case key.Matches(msg, m.keymap.Cancel):
			return m, SetView(transactionsView)
		case key.Matches(msg, m.keymap.Reset):
			newModel := newModelNewTransaction(m.api, firefly.Transaction{})
			return newModel, SetView(newView)
		case key.Matches(msg, m.keymap.Refresh):
			triggerCategoryCounter++
			triggerSourceCounter++
			triggerDestinationCounter++
			return m, nil
		case key.Matches(msg, m.keymap.Submit):
			if m.form.State == huh.StateCompleted {
				description := m.form.GetString("description")
				if description == "" {
					description = fmt.Sprintf("%s, %s -> %s", category.Name, source.Name, destination.Name)
				}

				currencyCode := ""
				switch source.Type {
				case "asset", "liabilities":
					currencyCode = source.CurrencyCode
				case "revenue":
					currencyCode = destination.CurrencyCode
				}

				foreignCurrencyCode := ""
				if (source.Type == "asset" || source.Type == "liabilities") && (destination.Type == "asset" || destination.Type == "liabilities") {
					if source.CurrencyCode != destination.CurrencyCode {
						foreignCurrencyCode = destination.CurrencyCode
					}
				}

				if err := m.api.CreateTransaction(firefly.NewTransaction{
					ApplyRules:           true,
					ErrorIfDuplicateHash: false,
					FireWebhooks:         true,
					Transactions: []firefly.NewSubTransaction{
						{
							Type:                transactionType,
							Date:                fmt.Sprintf("%s-%s-%s", year, month, day),
							Amount:              amount,
							ForeignAmount:       foreignAmount,
							Description:         description,
							CurrencyCode:        currencyCode,
							ForeignCurrencyCode: foreignCurrencyCode,
							SourceID:            source.ID,
							DestinationID:       destination.ID,
							CategoryID:          category.ID,
						},
					},
				}); err != nil {
					newModel := newModelNewTransaction(m.api, firefly.Transaction{})
					return newModel, tea.Sequence(
						Notify(err.Error(), Warning),
						SetView(transactionsView))

				} else {
					newModel := newModelNewTransaction(m.api, firefly.Transaction{})
					cmds = append(cmds,
						SetView(transactionsView),
						Cmd(RefreshAssetsMsg{}),
						Cmd(RefreshLiabilitiesMsg{}),
						Cmd(RefreshTransactionsMsg{}),
						Cmd(RefreshExpenseInsightsMsg{}),
						Cmd(RefreshRevenueInsightsMsg{}),
					)
					return newModel, tea.Batch(cmds...)
				}
			}
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	return m, cmd
}

func (m modelNewTransaction) View() string {
	if m.form.State == huh.StateCompleted {
		return "Transaction created successfully! Press Enter to submit, Ctrl+R to reset curren form, or Esc to go back."
	}
	return m.form.View()
}

func (m *modelNewTransaction) Focus() {
	m.focus = true
}

func (m *modelNewTransaction) Blur() {
	m.focus = false
}
