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

var (
	year                      string
	month                     string
	day                       string
	triggerCategoryCounter    byte
	triggerSourceCounter      byte
	triggerDestinationCounter byte
	transactionType           string
	source                    firefly.Account
	destination               firefly.Account
	splits                    []*split
	groupTitle                string
	lastWindowWidth           int
	fullNewForm               bool

	dateGroup *huh.Group
)

type RefreshNewCategoryMsg struct{}
type RefreshNewAssetMsg struct{}
type RefreshNewExpenseMsg struct{}
type RefreshNewRevenueMsg struct{}
type RedrawFormMsg struct{}

type NewTransactionMsg struct {
	Transaction firefly.Transaction
}

type modelNewTransaction struct {
	form   *huh.Form
	api    *firefly.Api
	focus  bool
	keymap TransactionFormKeyMap
}

type split struct {
	source        firefly.Account
	destination   firefly.Account
	category      firefly.Category
	amount        string
	foreignAmount string
	description   string
}

func newModelNewTransaction(api *firefly.Api, trx firefly.Transaction) modelNewTransaction {
	now := time.Now()

	if trx.Type != "" {
		transactionType = trx.Type
		year = trx.Date[0:4]
		month = trx.Date[5:7]
		day = trx.Date[8:10]
		groupTitle = trx.GroupTitle

		splits = []*split{}
		for _, s := range trx.Splits {
			splits = append(splits, &split{
				source:        s.Source,
				destination:   s.Destination,
				category:      s.Category,
				amount:        fmt.Sprintf("%.2f", s.Amount),
				foreignAmount: fmt.Sprintf("%.2f", s.ForeignAmount),
				description:   s.Description,
			})
		}
	} else {
		transactionType = "withdrawal"
		year = fmt.Sprintf("%d", now.Year())
		month = fmt.Sprintf("%02d", now.Month())
		day = fmt.Sprintf("%02d", now.Day())
		groupTitle = ""
		splits = []*split{
			{
				source:        firefly.Account{},
				destination:   firefly.Account{},
				category:      firefly.Category{},
				amount:        "",
				foreignAmount: "",
				description:   "",
			},
		}
	}

	years := []string{}
	startYear := now.Year() - 9
	for y := range 10 {
		years = append(years, fmt.Sprintf("%d", startYear+y))
	}
	dateGroup = huh.NewGroup(
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
	)

	return modelNewTransaction{
		api:    api,
		keymap: DefaultTransactionFormKeyMap(),
		form:   BuildForm(api),
	}
}

func (m modelNewTransaction) Init() tea.Cmd {
	return tea.WindowSize()
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
	case RedrawFormMsg:
		m.form = BuildForm(m.api)
		return m, tea.WindowSize()
	case tea.WindowSizeMsg:
		lastWindowWidth = msg.Width
		return m, nil
	}

	if !m.focus {
		return m, nil
	}

	var cmd tea.Cmd

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
			return m, RedrawForm()
		case key.Matches(msg, m.keymap.AddSplit):
			splits = append(splits, &split{})
			return m, RedrawForm()
		case key.Matches(msg, m.keymap.DeleteSplit):
			return m, DeleteSplit()
		case key.Matches(msg, m.keymap.ChangeLayout):
			fullNewForm = !fullNewForm
			return m, RedrawForm()
		case key.Matches(msg, m.keymap.Submit):
			// if m.form.State == huh.StateCompleted {
			// 	description := m.form.GetString("description")
			// 	if description == "" {
			// 		description = fmt.Sprintf("%s, %s -> %s", category.Name, source.Name, destination.Name)
			// 	}
			//
			// 	currencyCode := ""
			// 	switch source.Type {
			// 	case "asset", "liabilities":
			// 		currencyCode = source.CurrencyCode
			// 	case "revenue":
			// 		currencyCode = destination.CurrencyCode
			// 	}
			//
			// 	foreignCurrencyCode := ""
			// 	if (source.Type == "asset" || source.Type == "liabilities") && (destination.Type == "asset" || destination.Type == "liabilities") {
			// 		if source.CurrencyCode != destination.CurrencyCode {
			// 			foreignCurrencyCode = destination.CurrencyCode
			// 		}
			// 	}
			//
			// 	if err := m.api.CreateTransaction(firefly.NewTransaction{
			// 		ApplyRules:           true,
			// 		ErrorIfDuplicateHash: false,
			// 		FireWebhooks:         true,
			// 		Transactions: []firefly.NewSubTransaction{
			// 			{
			// 				Type:                transactionType,
			// 				Date:                fmt.Sprintf("%s-%s-%s", year, month, day),
			// 				Amount:              amount,
			// 				ForeignAmount:       foreignAmount,
			// 				Description:         description,
			// 				CurrencyCode:        currencyCode,
			// 				ForeignCurrencyCode: foreignCurrencyCode,
			// 				SourceID:            source.ID,
			// 				DestinationID:       destination.ID,
			// 				CategoryID:          category.ID,
			// 			},
			// 		},
			// 	}); err != nil {
			// 		newModel := newModelNewTransaction(m.api, firefly.Transaction{})
			// 		return newModel, tea.Sequence(
			// 			Notify(err.Error(), Warning),
			// 			SetView(transactionsView))
			//
			// 	} else {
			// 		newModel := newModelNewTransaction(m.api, firefly.Transaction{})
			// 		cmds = append(cmds,
			// 			SetView(transactionsView),
			// 			Cmd(RefreshAssetsMsg{}),
			// 			Cmd(RefreshLiabilitiesMsg{}),
			// 			Cmd(RefreshTransactionsMsg{}),
			// 			Cmd(RefreshExpenseInsightsMsg{}),
			// 			Cmd(RefreshRevenueInsightsMsg{}),
			// 		)
			// 		return newModel, tea.Batch(cmds...)
			// 	}
			// }
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

func BuildForm(api *firefly.Api) *huh.Form {
	var allGroups []*huh.Group

	for i, s := range splits {
		allGroups = append(allGroups, huh.NewGroup(
			huh.NewNote().
				Title(fmt.Sprint("Split: ", i)).
				TitleFunc(func() string {
					if i == 0 {

						source = s.source
						destination = s.destination

						stx := s.source.Type
						dtx := s.destination.Type
						transactionType = ""

						switch {
						case stx == "asset" && (dtx == "expense" || dtx == "liabilities"):
							transactionType = "withdrawal"
						case stx == "asset" && dtx == "asset":
							transactionType = "transfer"
						case stx == "revenue":
							transactionType = "deposit"
						case stx == "liabilities" && dtx == "expense":
							transactionType = "withdrawal"
						case stx == "liabilities" && dtx == "asset":
							transactionType = "deposit"
						case stx == "liabilities" && dtx == "liabilities":
							transactionType = "transfer"
						default:
							transactionType = "unknown"
						}

						return fmt.Sprintf("Current Type: %s", transactionType)
					}
					return fmt.Sprint("Split: ", i)

				}, []any{&s.source, &s.destination}),
			huh.NewSelect[firefly.Account]().
				Title("Source").
				DescriptionFunc(func() string {
					switch s.source.Type {
					case "asset", "liabilities":
						return fmt.Sprintf("Balance: %.2f %s", s.source.Balance, s.source.CurrencyCode)
					}
					return ""
				}, &s.source).
				Value(&s.source).
				Options(huh.NewOption(s.source.Name, s.source)).
				OptionsFunc(func() []huh.Option[firefly.Account] {
					options := []huh.Option[firefly.Account]{}

					if (transactionType == "withdrawal" || transactionType == "transfer") && i > 0 {
						options = append(options, huh.NewOption(source.Name, source))
					} else {
						for _, account := range api.Accounts["asset"] {
							options = append(options, huh.NewOption(account.Name, account))
						}
						for _, account := range api.Accounts["revenue"] {
							options = append(options, huh.NewOption(account.Name, account))
						}
						for _, account := range api.Accounts["liabilities"] {
							options = append(options, huh.NewOption(account.Name, account))
						}
					}
					return options
				}, []any{&triggerSourceCounter, &source}).WithHeight(5),
			huh.NewSelect[firefly.Account]().
				Title("Destination").
				DescriptionFunc(func() string {
					switch s.destination.Type {
					case "asset", "liabilities":
						return fmt.Sprintf("Balance: %.2f %s", s.destination.Balance, s.destination.CurrencyCode)
					}
					return ""
				}, &s.destination).
				Value(&s.destination).
				Options(huh.NewOption(s.destination.Name, s.destination)).
				OptionsFunc(func() []huh.Option[firefly.Account] {
					options := []huh.Option[firefly.Account]{}

					if (transactionType == "deposit" || transactionType == "transfer") && i > 0 {
						options = append(options, huh.NewOption(destination.Name, destination))
					} else {
						switch s.source.Type {
						case "asset":
							for _, account := range api.Accounts["expense"] {
								options = append(options, huh.NewOption(account.Name, account))
							}
							if i == 0 {
								for _, account := range api.Accounts["asset"] {
									options = append(options, huh.NewOption(account.Name, account))
								}
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
					}
					return options
				}, []any{&s.source, &triggerDestinationCounter, &destination}).WithHeight(5),
			huh.NewSelect[firefly.Category]().
				Title("Category").
				Value(&s.category).
				Options(huh.NewOption(s.category.Name, s.category)).
				OptionsFunc(func() []huh.Option[firefly.Category] {
					options := []huh.Option[firefly.Category]{}
					for _, category := range api.Categories {
						options = append(options, huh.NewOption(category.Name, category))
					}
					return options
				}, &triggerCategoryCounter).WithHeight(5),
			huh.NewInput().
				Title("Amount").
				Value(&s.amount).
				DescriptionFunc(func() string {
					switch s.source.Type {
					case "asset", "liabilities":
						return s.source.CurrencyCode
					case "revenue":
						return s.destination.CurrencyCode
					}
					return ""

				}, []any{&s.source, &s.destination}).
				Validate(func(str string) error {
					var amount float64
					_, err := strconv.ParseFloat(str, 64)
					if err != nil || amount < 0 {
						return errors.New("please enter a valid positive number for amount")
					}
					return nil
				}),
			huh.NewInput().
				Title("Foreign Amount").
				Value(&s.foreignAmount).
				DescriptionFunc(func() string {
					sType := s.source.Type
					dType := s.destination.Type
					if (sType == "asset" || sType == "liabilities") && (dType == "asset" || dType == "liabilities") {
						if s.source.CurrencyCode == s.destination.CurrencyCode {
							return "N/A"
						}
						return s.destination.CurrencyCode
					}
					return "N/A"
				}, []any{&s.source, &s.destination}).
				Validate(func(str string) error {
					sType := s.source.Type
					dType := s.destination.Type
					if (sType == "asset" || sType == "liabilities") && (dType == "asset" || dType == "liabilities") {
						if s.source.CurrencyCode == s.destination.CurrencyCode {
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
				},
				),
			huh.NewInput().
				Title("Description").
				Value(&s.description).
				PlaceholderFunc(func() string {
					return fmt.Sprintf("%s, %s -> %s", s.category.Name, s.source.Name, s.destination.Name)
				}, []any{&s.category, &s.source, &s.destination}),
		))
	}

	allGroups = append(allGroups, dateGroup, huh.NewGroup(
		huh.NewInput().
			Title("Group Title").
			Value(&groupTitle).
			PlaceholderFunc(func() string {
				if len(splits) > 0 {
					return fmt.Sprintf("%s", splits[0].description)
				}
				return ""
			}, &splits),
	))

	form := huh.NewForm(allGroups...)

	if fullNewForm {
		return form.WithLayout(huh.LayoutDefault)
	}

	return form.WithLayout(huh.LayoutGrid(2, len(splits)+1))
}

func DeleteSplit() tea.Cmd {
	return Cmd(PromptMsg{
		Prompt: "Delete split number: ",
		Value:  "",
		Callback: func(value string) tea.Cmd {
			if value != "None" {
				index, err := strconv.Atoi(value)
				if err == nil && index >= 1 && index < len(splits) {
					splits = append(splits[:index], splits[index+1:]...)
					return tea.Sequence(RedrawForm(), SetView(newView))
				}
				return tea.Sequence(Notify("Invalid split index", Warning), SetView(newView))
			}
			return SetView(newView)
		}})
}

func RedrawForm() tea.Cmd {
	return Cmd(RedrawFormMsg{})
}
