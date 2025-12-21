/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/

//	if err := fireflyApi.CreateTransaction(firefly.NewTransaction{
//		ApplyRules:           true,
//		ErrorIfDuplicateHash: true,
//		GroupTitle:           "Test Transaction Group",
//		Transactions: []firefly.NewSubTransaction{
//			{
//				Type:            "withdrawal",
//				Date:            "2025-12-21",
//				Amount:          "100.00",
//				Description:     "XXXX",
//				CurrencyCode:    "KGS",
//				SourceName:      "💳 mbank:visa:kgs",
//				DestinationName: "Capito1234",
//				CategoryName:    "Groceries!4",
//			},
//		},
//	}); err != nil {
//
//		return err
//	}

package ui

import (
	"errors"
	"ffiii-tui/internal/firefly"
	"fmt"

	// "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type modelNewTransaction struct {
	form *huh.Form // huh.Form is just a tea.Model
	api  *firefly.Api
}

// var (
// 	name         string
// 	instructions string
// 	discount     bool
// )

// Type
// Date
// Amount
// Description
// CurrencyCode
// SourceName
// DestinationName
// CategoryName
func newModelNewTransaction(api *firefly.Api) modelNewTransaction {
	suggestionsSource := []string{}
	if accounts, err := api.ListAccounts("asset"); err == nil {
		for _, account := range accounts {
			suggestionsSource = append(suggestionsSource, account.Attributes.Name)
		}
	}

	return modelNewTransaction{
		api: api,
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("class").
					Options(huh.NewOptions("withdrawal", "deposit", "transfer")...).
					Title("Type"),

				// Validate(func(str string) error {
				//     var year, month, day int
				//     n, err := fmt.Sscanf(str, "%d-%d-%d", &year, &month, &day)
				//     if err != nil || n != 3 || month < 1 || month > 12 || day < 1 || day > 31 {
				//         return errors.New("Please enter a valid date in YYYY-MM-DD format.")
				//     }
				//     return nil
				// }),

				huh.NewInput().
					Key("source_name").
					Title("Source Account Name").
					Suggestions(suggestionsSource),

				huh.NewInput().
					Key("destination_name").
					Title("Destination Account Name"),

				huh.NewInput().
					Key("category_name").
					Title("Category Name"),

				huh.NewInput().
					Key("amount").
					Title("Amount").
					Validate(func(str string) error {
						var amount float64
						_, err := fmt.Sscanf(str, "%f", &amount)
						if err != nil || amount <= 0 {
							return errors.New("Please enter a valid positive number for amount.")
						}
						return nil
					}),
				huh.NewSelect[string]().
					Key("currency_code").
					Title("Currency Code (e.g., USD, EUR)").
					Options(huh.NewOptions("USD", "EUR", "GBP", "JPY", "CNY")...),

				huh.NewInput().
					Key("description").
					Title("Description"),
			),
			huh.NewGroup(
				huh.NewSelect[string]().
					Key("year").
					Title("Year").
					// OptionsFunc(func() []huh.Option[string] {
					//                    years := []huh.Option[string]{}
					//                    for y := 2025; y >= 2000; y-- {
					//                        years = append(years, huh.Option[string]{Value: fmt.Sprintf("%d", y)})
					//                    }
					//                    return years
					//                }, nil).WithHeight(1),
					Options(huh.NewOptions("2025", "2024", "2023", "2022")...).WithHeight(1),

				huh.NewSelect[string]().
					Key("month").
					Title("Month").
					// OptionsFunc(func() []huh.Option[string] {
					//     months := []huh.Option[string]{}
					//     for m := 1; m <= 12; m++ {
					//         months = append(months, huh.Option[string]{Value: fmt.Sprintf("%02d", m)})
					//     }
					//     return months
					// }, nil).WithHeight(4),
					Options(huh.NewOptions(
						"01", "02", "03", "04", "05", "06",
						"07", "08", "09", "10", "11", "12",
					)...).WithHeight(4),
				huh.NewSelect[string]().
					Key("day").
					Title("Day").
					// OptionsFunc(func() []huh.Option[string] {
					//                 days := []huh.Option[string]{}
					//                 for d := 1; d <= 31; d++ {
					//                     days = append(days, huh.Option[string]{Value: fmt.Sprintf("%02d", d)})
					//                 }
					//                 return days
					//             }, nil).WithHeight(4),
					Options(huh.NewOptions(
						"01", "02", "03", "04", "05", "06",
						"07", "08", "09", "10", "11", "12",
						"13", "14", "15", "16", "17", "18",
						"19", "20", "21", "22", "23", "24",
						"25", "26", "27", "28", "29", "30",
						"31",
					)...).WithHeight(4),
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
	// if m.form.State == huh.StateCompleted {
	// 	class := m.form.GetString("class")
	// 	level := m.form.GetInt("level")
	// 	return fmt.Sprintf("You selected: %s, Lvl. %d", class, level)
	// }
	return m.form.View()
}
