/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

// TODO: Use last date as input, and key for resetting to today.

import (
	"errors"
	"ffiii-tui/internal/firefly"
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type RefreshNewCategoryMsg struct{}
type RefreshNewAssetMsg struct{}
type RefreshNewExpenseMsg struct{}
type RefreshNewRevenueMsg struct{}

type NewTransactionMsg struct {
	transaction firefly.Transaction
}

type modelNewTransaction struct {
	form  *huh.Form // huh.Form is just a tea.Model
	api   *firefly.Api
	focus bool
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
		}
		if trx.ForeignAmount != 0 {
			foreignAmount = fmt.Sprintf("%.2f", trx.ForeignAmount)
		}
	} else {
		transactionType = "withdrawal"
		year = fmt.Sprint(now.Year())
		month = fmt.Sprintf("%02d", now.Month())
		day = fmt.Sprintf("%02d", now.Day())
		source = firefly.Account{}
		destination = firefly.Account{}
		category = firefly.Category{}
		amount = ""
		foreignAmount = ""
	}

	years := []string{}
	for y := range 10 {
		years = append(years, fmt.Sprintf("%d", now.Year()-y))
	}

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
					}, []any{&transactionType, &triggerSourceCounter}).WithHeight(5),
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
					}, []any{&transactionType, &triggerDestinationCounter}).WithHeight(5),
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
						switch transactionType {
						case "withdrawal", "transfer":
							return source.CurrencyCode
						case "deposit":
							return destination.CurrencyCode
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
					Key("foreignAmount").
					Title("Foreign Amount").
					Value(&foreignAmount).
					DescriptionFunc(func() string {
						if transactionType == "transfer" {
							if source.CurrencyCode == destination.CurrencyCode {
								return "N/A"
							}
							return destination.CurrencyCode
						}
						return "N/A"
					}, []any{&source, &destination}).
					Validate(func(str string) error {
						if transactionType == "transfer" {
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
							return errors.New("foreign amount is only applicable for transfers")
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
		switch transactionType {
		case "withdrawal":
			triggerSourceCounter++
		case "deposit":
			triggerDestinationCounter++
		case "transfer":
			triggerSourceCounter++
			triggerDestinationCounter++
		}
	case RefreshNewExpenseMsg:
		switch transactionType {
		case "withdrawal":
			triggerDestinationCounter++
		}
	case RefreshNewRevenueMsg:
		switch transactionType {
		case "deposit":
			triggerSourceCounter++
		}
	case RefreshNewCategoryMsg:
		triggerCategoryCounter++
	case NewTransactionMsg:
		newModel := newModelNewTransaction(m.api, msg.transaction)
		return newModel, Cmd(ViewNewMsg{})
	}

	if !m.focus {
		return m, nil
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			switch m.form.GetFocusedField().GetKey() {
			case "source_name":

				switch transactionType {
				case "withdrawal", "transfer":
					return m, Cmd(PromptMsg{
						Prompt: "New Asset(name,currency): ",
						Value:  "",
						Callback: func(value string) tea.Cmd {
							var cmds []tea.Cmd
							if value != "" {
								split := strings.SplitN(value, ",", 2)
								if len(split) >= 2 {
									acc := strings.TrimSpace(split[0])
									cur := strings.TrimSpace(split[1])
									if acc != "" && cur != "" {
										cmds = append(cmds, Cmd(NewAssetMsg{account: acc, currency: cur}))
									} else {
										// TODO: Report error to user
									}
								}
							}
							cmds = append(cmds,
								Cmd(RefreshNewAssetMsg{}),
								Cmd(ViewNewMsg{}))
							return tea.Sequence(cmds...)
						}})
				case "deposit":
					return m, Cmd(PromptMsg{
						Prompt: "New Revenue: ",
						Value:  "",
						Callback: func(value string) tea.Cmd {
							var cmds []tea.Cmd
							if value != "" {
								cmds = append(cmds, Cmd(NewRevenueMsg{account: value}))
							}
							cmds = append(cmds,
								Cmd(RefreshNewRevenueMsg{}),
								Cmd(ViewNewMsg{}))
							return tea.Sequence(cmds...)
						}})
				default:
					return m, nil
				}

			case "destination_name":

				switch transactionType {
				case "withdrawal":
					return m, Cmd(PromptMsg{
						Prompt: "New Expense: ",
						Value:  "",
						Callback: func(value string) tea.Cmd {
							var cmds []tea.Cmd
							if value != "" {
								cmds = append(cmds, Cmd(NewExpenseMsg{account: value}))
							}
							cmds = append(cmds,
								Cmd(RefreshNewExpenseMsg{}),
								Cmd(ViewNewMsg{}))
							return tea.Sequence(cmds...)
						}})
				case "transfer", "deposit":
					return m, Cmd(PromptMsg{
						Prompt: "New Asset(name,currency): ",
						Value:  "",
						Callback: func(value string) tea.Cmd {
							var cmds []tea.Cmd
							if value != "" {
								split := strings.SplitN(value, ",", 2)
								if len(split) >= 2 {
									acc := strings.TrimSpace(split[0])
									cur := strings.TrimSpace(split[1])
									if acc != "" && cur != "" {
										cmds = append(cmds, Cmd(NewAssetMsg{account: acc, currency: cur}))
									} else {
										// TODO: Report error to user
									}
								}
							}
							cmds = append(cmds,
								Cmd(RefreshNewAssetMsg{}),
								Cmd(ViewNewMsg{}))
							return tea.Sequence(cmds...)
						}})
				default:
					return m, nil
				}

			case "category_name":
				return m, Cmd(PromptMsg{
					Prompt: "New Category: ",
					Value:  "",
					Callback: func(value string) tea.Cmd {
						var cmds []tea.Cmd
						if value != "" {
							cmds = append(cmds, Cmd(NewCategoryMsg{category: value}))
						}
						cmds = append(cmds,
							Cmd(RefreshNewCategoryMsg{}),
							Cmd(ViewNewMsg{}))
						return tea.Sequence(cmds...)
					}})
			}
		case "esc":
			return m, Cmd(ViewTransactionsMsg{})
		case "ctrl+r":
			newModel := newModelNewTransaction(m.api, firefly.Transaction{})
			return newModel, Cmd(ViewNewMsg{})
		case "enter":
			if m.form.State == huh.StateCompleted {
				description := m.form.GetString("description")
				if description == "" {
					description = fmt.Sprintf("%s, %s -> %s", category.Name, source.Name, destination.Name)
				}

				currencyCode := ""
				foreignCurrencyCode := ""

				switch transactionType {
				case "withdrawal":
					currencyCode = source.CurrencyCode
				case "deposit":
					currencyCode = destination.CurrencyCode
				case "transfer":
					currencyCode = source.CurrencyCode
					if source.CurrencyCode != destination.CurrencyCode {
						foreignCurrencyCode = destination.CurrencyCode
					}
				default:
					newModel := newModelNewTransaction(m.api, firefly.Transaction{})
					return newModel, nil
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
					return newModel, nil

				} else {
					newModel := newModelNewTransaction(m.api, firefly.Transaction{})
					cmds = append(cmds,
						Cmd(RefreshAssetsMsg{}),
						Cmd(RefreshTransactionsMsg{}),
						Cmd(ViewTransactionsMsg{}))
					return newModel, tea.Sequence(cmds...)
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
