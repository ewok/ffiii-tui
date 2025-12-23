/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

// TODO: Use last date as input, and key for resetting to today.
// TODO: Add category, destination creation.
// TODO: Add foreign currency for transfer.

import (
	"errors"
	"ffiii-tui/internal/firefly"
	"fmt"
	"strconv"
	"time"

	// "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type modelNewTransaction struct {
	form *huh.Form // huh.Form is just a tea.Model
	api  *firefly.Api
}

var (
	source          firefly.Account
	destination     firefly.Account
	transactionType string
	amount          string
	category        firefly.Category
	year            string
	month           string
	day             string
)

// Type
// Date
// Amount
// Description
// CurrencyCode
// SourceName
// DestinationName
// CategoryName
func newModelNewTransaction(api *firefly.Api) modelNewTransaction {
	// Initialize default values
	now := time.Now()
	year = now.Format("2006")
	month = now.Format("01")
	day = now.Format("02")

	years := []string{}
	for y := range 10 {
		years = append(years, fmt.Sprintf("%d", now.Year()-y))
	}
	months := []string{"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12"}

	source = firefly.Account{}
	destination = firefly.Account{}
	transactionType = "withdrawal"
	amount = ""
	category = firefly.Category{}

	return modelNewTransaction{
		api: api,
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("type").
					Options(huh.NewOptions("withdrawal", "deposit", "transfer")...).
					Title("Type").
					Value(&transactionType),

				huh.NewSelect[firefly.Account]().
					Key("source_name").
					Title("Source Account").
					DescriptionFunc(func() string {
						switch transactionType {
						case "withdrawal", "transfer":
							return fmt.Sprintf("Balance: %.2f %s", source.Balance, source.CurrencyCode)
						}
						return ""
					}, []any{&source, &transactionType}).
					Value(&source).
					OptionsFunc(func() []huh.Option[firefly.Account] {
						options := []huh.Option[firefly.Account]{}

						switch transactionType {
						case "withdrawal":
							for _, account := range api.Assets {
								options = append(options, huh.NewOption(account.Name, account))
							}
						case "transfer":
							for _, account := range api.Assets {
								options = append(options, huh.NewOption(account.Name, account))
							}
							for _, account := range api.Liabilities {
								options = append(options, huh.NewOption(account.Name, account))
							}
						case "deposit":
							for _, account := range api.Revenues {
								options = append(options, huh.NewOption(account.Name, account))
							}
						}
						return options
					}, &transactionType).WithHeight(5),

				huh.NewSelect[firefly.Account]().
					Key("destination_name").
					Title("Destination Account Name").
					Value(&destination).
					OptionsFunc(func() []huh.Option[firefly.Account] {
						options := []huh.Option[firefly.Account]{}
						switch transactionType {
						case "withdrawal":
							for _, account := range api.Expenses {
								options = append(options, huh.NewOption(account.Name, account))
							}
						case "transfer":
							for _, account := range api.Assets {
								options = append(options, huh.NewOption(account.Name, account))
							}
							for _, account := range api.Liabilities {
								options = append(options, huh.NewOption(account.Name, account))
							}
						case "deposit":
							for _, account := range api.Assets {
								options = append(options, huh.NewOption(account.Name, account))
							}
						}
						return options
					}, &transactionType).WithHeight(5),

				huh.NewSelect[firefly.Category]().
					Key("category_name").
					Title("Category Name").
					Value(&category).
					OptionsFunc(func() []huh.Option[firefly.Category] {
						options := []huh.Option[firefly.Category]{}
						for _, category := range api.Categories {
							options = append(options, huh.NewOption(category.Name, category))
						}
						return options
					}, nil).WithHeight(5),

				huh.NewInput().
					Key("amount").
					Title("Amount").
					Value(&amount).
					DescriptionFunc(func() string {
						switch transactionType {
						case "withdrawal":
							return source.CurrencyCode
						case "deposit":
							return destination.CurrencyCode
						case "transfer":
							return source.CurrencyCode
						default:
							return ""
						}
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
					WithHeight(1),
				huh.NewSelect[string]().
					Key("month").
					Title("Month").
					Options(huh.NewOptions(months...)...).
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
		).WithLayout(huh.LayoutGrid(2, 2)),
	}
}

func (m modelNewTransaction) Init() tea.Cmd {
	return m.form.Init()
}

func (m modelNewTransaction) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			cmds = append(cmds, func() tea.Msg { return viewTransactionsMsg{} })
			// case "n":
			// 	customInput = !customInput
			// 	return m, nil
			// 	// switch m.form.GetFocusedField().GetKey() {
			// 	// case "destination_name":
			// 	//
			// 	// }
			//
		case "ctrl+n":
			form := newModelNewTransaction(m.api)
			return form, nil
		case "enter":
			if m.form.State == huh.StateCompleted {
				description := m.form.GetString("description")
				if description == "" {
					description = fmt.Sprintf("%s, %s -> %s", category.Name, source.Name, destination.Name)
				}
				currencyCode := ""
				switch transactionType {
				case "withdrawal", "transfer":
					currencyCode = source.CurrencyCode
				case "deposit":
					currencyCode = destination.CurrencyCode
				default:
					currencyCode = source.CurrencyCode
				}

				if err := m.api.CreateTransaction(firefly.NewTransaction{
					ApplyRules:           true,
					ErrorIfDuplicateHash: true,
					FireWebhooks:         true,
					Transactions: []firefly.NewSubTransaction{
						{
							Type:          transactionType,
							Date:          fmt.Sprintf("%s-%s-%s", year, month, day),
							Amount:        amount,
							Description:   description,
							CurrencyCode:  currencyCode,
							SourceID:      source.ID,
							DestinationID: destination.ID,
							CategoryID:    category.ID,
						},
					},
				}); err != nil {
					m.form.State = huh.StateAborted
				} else {
					form := newModelNewTransaction(m.api)
					cmds = append(cmds, func() tea.Msg { return RefreshMsg{} })
					cmds = append(cmds, func() tea.Msg { return viewTransactionsMsg{} })
					return form, tea.Batch(cmds...)
				}

			}
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m modelNewTransaction) View() string {
	if m.form.State == huh.StateCompleted {
		return "Transaction created successfully! Press Enter to submit, Ctrl+N to create another, or Esc to go back."
	}
	return m.form.View()
}
