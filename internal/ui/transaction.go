/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

// TODO: Use last date as input, and key for resetting to today.

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"ffiii-tui/internal/firefly"
	"ffiii-tui/internal/ui/notify"
	"ffiii-tui/internal/ui/prompt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"go.uber.org/zap"
)

var (
	triggerCategoryCounter    byte
	triggerSourceCounter      byte
	triggerDestinationCounter byte
	fullNewForm               bool
)

type (
	RefreshNewCategoryMsg struct{}
	RefreshNewAssetMsg    struct{}
	RefreshNewExpenseMsg  struct{}
	RefreshNewRevenueMsg  struct{}
	RedrawFormMsg         struct{}
	DeleteSplitMsg        struct{ Index int }
	NewTransactionMsg     struct{ Transaction firefly.Transaction }
	NewTransactionFromMsg struct{ Transaction firefly.Transaction }
	EditTransactionMsg    struct{ Transaction firefly.Transaction }
	ResetTransactionMsg   struct{}
)

type modelTransaction struct {
	form   *huh.Form
	api    TransactionFormAPI
	keymap TransactionFormKeyMap
	focus  bool

	new     bool
	created bool

	splits []*split
	attr   *transactionAttr
}

type split struct {
	source        firefly.Account
	destination   firefly.Account
	category      firefly.Category
	amount        string
	foreignAmount string
	description   string

	trxJID string // For editing existing transactions
}

type transactionAttr struct {
	year            string
	month           string
	day             string
	transactionType string
	source          firefly.Account
	destination     firefly.Account
	groupTitle      string

	trxID string // For editing existing transactions
}

func newModelTransaction(api TransactionFormAPI) modelTransaction {
	return modelTransaction{
		api:    api,
		keymap: DefaultTransactionFormKeyMap(),
		attr:   &transactionAttr{},
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewNote().Title("Loading..."),
			),
		).WithLayout(huh.LayoutDefault),
		splits: []*split{},
	}
}

func (m modelTransaction) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m modelTransaction) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case RefreshNewAssetMsg:
		triggerSourceCounter++
		triggerDestinationCounter++
		return m, RedrawForm()
	case RefreshNewExpenseMsg:
		triggerDestinationCounter++
		return m, RedrawForm()
	case RefreshNewRevenueMsg:
		triggerSourceCounter++
		return m, RedrawForm()
	case RefreshNewCategoryMsg:
		triggerCategoryCounter++
		return m, RedrawForm()
	case NewTransactionMsg:
		if !m.created {
			m.SetTransaction(msg.Transaction, true)
			m.created = true
		}
		return m, tea.Batch(
			RedrawForm(),
			SetView(newView),
		)
	case NewTransactionFromMsg:
		m.SetTransaction(msg.Transaction, true)
		m.created = true
		return m, tea.Batch(
			RedrawForm(),
			SetView(newView),
		)
	case EditTransactionMsg:
		m.SetTransaction(msg.Transaction, false)
		m.created = true
		return m, tea.Batch(
			RedrawForm(),
			SetView(newView),
		)
	case ResetTransactionMsg:
		trx := firefly.Transaction{}
		m.SetTransaction(trx, true)
		m.created = true
		return m, nil
	case RedrawFormMsg:
		m.UpdateForm()
		return m, tea.WindowSize()
	case DeleteSplitMsg:
		return m, m.DeleteSplit(msg.Index)
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
			return m, tea.Batch(
				SetView(newView),
				Cmd(ResetTransactionMsg{}),
			)
		case key.Matches(msg, m.keymap.Refresh):
			triggerCategoryCounter++
			triggerSourceCounter++
			triggerDestinationCounter++
			return m, RedrawForm()
		case key.Matches(msg, m.keymap.AddSplit):
			m.splits = append(m.splits, &split{})
			return m, RedrawForm()
		case key.Matches(msg, m.keymap.DeleteSplit):
			return m, prompt.Ask(
				"Delete split number: ",
				"",
				func(value string) tea.Cmd {
					if value != "None" {
						index, err := strconv.Atoi(value)
						if err == nil {
							return Cmd(DeleteSplitMsg{Index: index})
						}
					}
					return SetView(newView)
				},
			)
		case key.Matches(msg, m.keymap.ChangeLayout):
			fullNewForm = !fullNewForm
			return m, RedrawForm()
		case key.Matches(msg, m.keymap.Submit):
			if m.form.State == huh.StateCompleted {
				if m.new {
					return m, m.CreateTransaction()
				} else {
					return m, m.UpdateTransaction()
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

func (m modelTransaction) View() string {
	if m.form.State == huh.StateCompleted {
		return "Press Enter to submit, Ctrl+N to reset current form, Ctrl+R to edit current form again, or Esc to go back."
	}
	return m.form.View()
}

func (m *modelTransaction) Focus() {
	m.focus = true
}

func (m *modelTransaction) Blur() {
	m.focus = false
}

func (m *modelTransaction) UpdateForm() {
	var allGroups []*huh.Group

	for i, s := range m.splits {
		allGroups = append(allGroups, huh.NewGroup(
			huh.NewNote().
				Title(fmt.Sprint("Split: ", i)).
				TitleFunc(m.trxTitle(i, s)),
			huh.NewSelect[firefly.Account]().
				Title("Source").
				Value(&s.source).
				Options(huh.NewOption(s.source.Name, s.source)).
				OptionsFunc(m.trxSourceOptions(i, s)).WithHeight(5),
			huh.NewSelect[firefly.Account]().
				Title("Destination").
				Value(&s.destination).
				Options(huh.NewOption(s.destination.Name, s.destination)).
				OptionsFunc(m.trxDestinationOptions(i, s)).WithHeight(4),
			huh.NewSelect[firefly.Category]().
				Title("Category").
				Value(&s.category).
				Options(huh.NewOption(s.category.Name, s.category)).
				OptionsFunc(func() []huh.Option[firefly.Category] {
					options := []huh.Option[firefly.Category]{}
					for _, category := range m.api.CategoriesList() {
						options = append(options, huh.NewOption(category.Name, category))
					}
					return options
				}, &triggerCategoryCounter).WithHeight(4),
			huh.NewInput().
				Title("Amount").
				Value(&s.amount).
				TitleFunc(func() string {
					title := "Amount "
					switch s.source.Type {
					case "asset", "liabilities":
						return title + s.source.CurrencyCode
					case "revenue":
						return title + s.destination.CurrencyCode
					}
					return title
				}, []any{&s.source, &s.destination}).
				Validate(func(str string) error {
					var amount float64
					amount, err := strconv.ParseFloat(str, 64)
					if err != nil || amount < 0 {
						return errors.New("please enter a valid positive number for amount")
					}
					return nil
				}),
			huh.NewInput().
				Title("Foreign Amount").
				Value(&s.foreignAmount).
				TitleFunc(func() string {
					title := "Foreign Amount "
					sType := s.source.Type
					dType := s.destination.Type
					if (sType == "asset" || sType == "liabilities") && (dType == "asset" || dType == "liabilities") {
						if s.source.CurrencyCode == s.destination.CurrencyCode {
							return title + "N/A"
						}
						return title + s.destination.CurrencyCode
					}
					return title + "N/A"
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
						amount, err := strconv.ParseFloat(str, 64)
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
				PlaceholderFunc(s.Description, []any{&s.category, &s.source, &s.destination}).
				WithWidth(30),
		))
	}

	now := time.Now()
	years := []string{}
	startYear := now.Year() - 9
	for y := range 10 {
		years = append(years, fmt.Sprintf("%d", startYear+y))
	}
	allGroups = append(allGroups, huh.NewGroup(
		huh.NewSelect[string]().
			Key("year").
			Title("Year").
			Options(huh.NewOptions(years...)...).
			Value(&m.attr.year).
			WithHeight(3),
		huh.NewSelect[string]().
			Key("month").
			Title("Month").
			Options(huh.NewOptions("01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12")...).
			Value(&m.attr.month).
			WithHeight(4),
		huh.NewSelect[string]().
			Key("day").
			Title("Day").
			Value(&m.attr.day).
			Options(huh.NewOptions(m.attr.day)...).
			OptionsFunc(func() []huh.Option[string] {
				days := []string{}
				// According to month and year, determine number of days
				monthInt, _ := strconv.Atoi(m.attr.month)
				yearInt, _ := strconv.Atoi(m.attr.year)
				numDays := daysIn(monthInt, yearInt)
				for d := range numDays {
					days = append(days, fmt.Sprintf("%02d", d+1))
				}
				return huh.NewOptions(days...)
			}, []any{&m.attr.month, &m.attr.year}).WithHeight(4),
	))

	if len(m.splits) > 1 {
		allGroups = append(allGroups, huh.NewGroup(
			huh.NewInput().
				Title("Group Title").
				Value(&m.attr.groupTitle).
				PlaceholderFunc(m.GroupTitle, &m.splits).
				WithWidth(30),
		))
	}

	if fullNewForm {
		m.form = huh.NewForm(allGroups...).WithLayout(huh.LayoutDefault)
	} else {
		m.form = huh.NewForm(allGroups...).WithLayout(huh.LayoutGrid(2, len(m.splits)+1))
	}
}

func (m *modelTransaction) DeleteSplit(index int) tea.Cmd {
	if index >= 1 && index < len(m.splits) {
		m.splits = append(m.splits[:index], m.splits[index+1:]...)
		return tea.Sequence(RedrawForm(), SetView(newView))
	}
	return tea.Sequence(notify.NotifyWarn("Invalid split index"), SetView(newView))
}

func (m *modelTransaction) CreateTransaction() tea.Cmd {
	trx := []firefly.RequestTransactionSplit{}
	for _, s := range m.splits {
		trx = append(trx, firefly.RequestTransactionSplit{
			Type:                m.attr.transactionType,
			Date:                fmt.Sprintf("%s-%s-%s", m.attr.year, m.attr.month, m.attr.day),
			SourceID:            s.source.ID,
			DestinationID:       s.destination.ID,
			CategoryID:          s.category.ID,
			CurrencyCode:        s.CurrencyCode(),
			ForeignCurrencyCode: s.ForeignCurrencyCode(),
			Amount:              s.amount,
			ForeignAmount:       s.foreignAmount,
			Description:         s.Description(),
		})
	}

	id, err := m.api.CreateTransaction(firefly.RequestTransaction{
		ApplyRules:           true,
		ErrorIfDuplicateHash: false,
		FireWebhooks:         true,
		GroupTitle:           m.GroupTitle(),
		Transactions:         trx,
	})
	if err != nil {
		return tea.Sequence(
			notify.NotifyError(err.Error()),
			SetView(transactionsView))
	}

	m.created = false

	return tea.Batch(
		SetView(transactionsView),
		notify.NotifyLog("Transaction created successfully"),
		Cmd(RefreshAssetsMsg{}),
		Cmd(RefreshLiabilitiesMsg{}),
		Cmd(RefreshSummaryMsg{}),
		Cmd(RefreshTransactionsMsg{TrxID: id}),
		Cmd(RefreshExpenseInsightsMsg{}),
		Cmd(RefreshRevenueInsightsMsg{}),
		Cmd(RefreshCategoryInsightsMsg{}))
}

func (m *modelTransaction) UpdateTransaction() tea.Cmd {
	trx := []firefly.RequestTransactionSplit{}
	for _, s := range m.splits {
		trx = append(trx, firefly.RequestTransactionSplit{
			TransactionJournalID: s.trxJID,
			Type:                 m.attr.transactionType,
			Date:                 fmt.Sprintf("%s-%s-%s", m.attr.year, m.attr.month, m.attr.day),
			SourceID:             s.source.ID,
			DestinationID:        s.destination.ID,
			CategoryID:           s.category.ID,
			CurrencyCode:         s.CurrencyCode(),
			ForeignCurrencyCode:  s.ForeignCurrencyCode(),
			Amount:               s.amount,
			ForeignAmount:        s.foreignAmount,
			Description:          s.Description(),
		})
	}

	id, err := m.api.UpdateTransaction(m.attr.trxID, firefly.RequestTransaction{
		ApplyRules:   true,
		FireWebhooks: true,
		GroupTitle:   m.GroupTitle(),
		Transactions: trx,
	})
	if err != nil {
		return tea.Sequence(
			notify.NotifyError(err.Error()),
			SetView(transactionsView))
	}

	m.created = false

	return tea.Batch(
		SetView(transactionsView),
		notify.NotifyLog("Transaction updated successfully"),
		Cmd(RefreshAssetsMsg{}),
		Cmd(RefreshLiabilitiesMsg{}),
		Cmd(RefreshSummaryMsg{}),
		Cmd(RefreshTransactionsMsg{TrxID: id}),
		Cmd(RefreshExpenseInsightsMsg{}),
		Cmd(RefreshRevenueInsightsMsg{}),
		Cmd(RefreshCategoryInsightsMsg{}))
}

func (m *modelTransaction) SetTransaction(trx firefly.Transaction, newT bool) {
	zap.L().Debug("newModelTransaction", zap.Any("trx", trx))

	m.new = newT

	now := time.Now()

	if trx.TransactionID != "" {
		m.attr.transactionType = trx.Type
		m.attr.year = trx.Date[0:4]
		m.attr.month = trx.Date[5:7]
		m.attr.day = trx.Date[8:10]
		m.attr.groupTitle = trx.GroupTitle
		m.attr.trxID = trx.TransactionID

		m.splits = []*split{}
		for _, s := range trx.Splits {
			amount := ""
			if s.Amount != 0 {
				amount = fmt.Sprintf("%.2f", s.Amount)
			}
			foreignAmount := ""
			if s.ForeignAmount != 0 {
				foreignAmount = fmt.Sprintf("%.2f", s.ForeignAmount)
			}
			m.splits = append(m.splits, &split{
				source:        s.Source,
				destination:   s.Destination,
				category:      s.Category,
				amount:        amount,
				foreignAmount: foreignAmount,
				description:   s.Description,
				trxJID:        s.TransactionJournalID,
			})
		}
	} else {
		m.attr.transactionType = "withdrawal"
		m.attr.year = fmt.Sprintf("%d", now.Year())
		m.attr.month = fmt.Sprintf("%02d", now.Month())
		m.attr.day = fmt.Sprintf("%02d", now.Day())
		m.attr.groupTitle = ""
		source := firefly.Account{}
		destination := firefly.Account{}
		category := firefly.Category{}
		if len(trx.Splits) > 0 {
			source = trx.Splits[0].Source
			destination = trx.Splits[0].Destination
			category = trx.Splits[0].Category
		}
		m.splits = []*split{
			{
				source:        source,
				destination:   destination,
				category:      category,
				amount:        "",
				foreignAmount: "",
				description:   "",
				trxJID:        "",
			},
		}
		m.new = true
	}
}

func RedrawForm() tea.Cmd {
	return Cmd(RedrawFormMsg{})
}

// Helpers
func (m *modelTransaction) trxTitle(i int, s *split) (func() string, any) {
	bindings := []any{&s.source, &s.destination}

	if i == 0 {
		return func() string {
			m.attr.source = s.source
			m.attr.destination = s.destination

			stx := s.source.Type
			dtx := s.destination.Type
			m.attr.transactionType = ""

			switch {
			case stx == "asset" && (dtx == "expense" || dtx == "liabilities" || dtx == "cash"):
				m.attr.transactionType = "withdrawal"
			case stx == "asset" && dtx == "asset":
				m.attr.transactionType = "transfer"
			case stx == "revenue":
				m.attr.transactionType = "deposit"
			case stx == "liabilities" && dtx == "expense":
				m.attr.transactionType = "withdrawal"
			case stx == "liabilities" && dtx == "asset":
				m.attr.transactionType = "deposit"
			case stx == "liabilities" && dtx == "liabilities":
				m.attr.transactionType = "transfer"
			default:
				m.attr.transactionType = "unknown"
			}
			return fmt.Sprintf("Current Type: %s", m.attr.transactionType)
		}, bindings
	}

	return func() string { return fmt.Sprint("Split: ", i) }, bindings
}

func (m *modelTransaction) trxSourceOptions(i int, s *split) (func() []huh.Option[firefly.Account], any) {
	bindings := []any{&triggerSourceCounter}

	if i > 0 {
		bindings = append(bindings, &m.attr.source)
		return func() []huh.Option[firefly.Account] {
			options := []huh.Option[firefly.Account]{}
			if m.attr.transactionType == "withdrawal" || m.attr.transactionType == "transfer" {
				options = append(options, huh.NewOption(m.attr.source.Name, m.attr.source))
				s.source = m.attr.source
			} else {
				for _, account := range m.api.AccountsByType("revenue") {
					options = append(options, huh.NewOption(account.Name, account))
				}
				for _, account := range m.api.AccountsByType("liabilities") {
					options = append(options, huh.NewOption(account.Name, account))
				}
			}
			return options
		}, bindings
	}

	return func() []huh.Option[firefly.Account] {
		options := []huh.Option[firefly.Account]{}
		for _, account := range m.api.AccountsByType("asset") {
			options = append(options, huh.NewOption(account.Name, account))
		}
		for _, account := range m.api.AccountsByType("revenue") {
			options = append(options, huh.NewOption(account.Name, account))
		}
		for _, account := range m.api.AccountsByType("liabilities") {
			options = append(options, huh.NewOption(account.Name, account))
		}
		return options
	}, bindings
}

func (m *modelTransaction) trxDestinationOptions(i int, s *split) (func() []huh.Option[firefly.Account], any) {
	bindings := []any{&s.source.Type, &triggerDestinationCounter}

	if i > 0 {
		bindings = append(bindings, &m.attr.destination)
		return func() []huh.Option[firefly.Account] {
			options := []huh.Option[firefly.Account]{}
			if m.attr.transactionType == "deposit" || m.attr.transactionType == "transfer" {
				options = append(options, huh.NewOption(m.attr.destination.Name, m.attr.destination))
				s.destination = m.attr.destination
			} else {
				switch s.source.Type {
				case "asset":
					for _, account := range m.api.AccountsByType("expense") {
						options = append(options, huh.NewOption(account.Name, account))
					}
					for _, account := range m.api.AccountsByType("liabilities") {
						options = append(options, huh.NewOption(account.Name, account))
					}
				case "revenue":
					for _, account := range m.api.AccountsByType("asset") {
						options = append(options, huh.NewOption(account.Name, account))
					}
					for _, account := range m.api.AccountsByType("liabilities") {
						options = append(options, huh.NewOption(account.Name, account))
					}
				case "liabilities":
					for _, account := range m.api.AccountsByType("asset") {
						options = append(options, huh.NewOption(account.Name, account))
					}
					for _, account := range m.api.AccountsByType("expense") {
						options = append(options, huh.NewOption(account.Name, account))
					}
				}
			}
			return options
		}, bindings
	}

	return func() []huh.Option[firefly.Account] {
		options := []huh.Option[firefly.Account]{}
		switch s.source.Type {
		case "asset":
			for _, account := range m.api.AccountsByType("expense") {
				options = append(options, huh.NewOption(account.Name, account))
			}
			for _, account := range m.api.AccountsByType("asset") {
				options = append(options, huh.NewOption(account.Name, account))
			}
			for _, account := range m.api.AccountsByType("liabilities") {
				options = append(options, huh.NewOption(account.Name, account))
			}
		case "revenue":
			for _, account := range m.api.AccountsByType("asset") {
				options = append(options, huh.NewOption(account.Name, account))
			}
			for _, account := range m.api.AccountsByType("liabilities") {
				options = append(options, huh.NewOption(account.Name, account))
			}
		case "liabilities":
			for _, account := range m.api.AccountsByType("asset") {
				options = append(options, huh.NewOption(account.Name, account))
			}
			for _, account := range m.api.AccountsByType("expense") {
				options = append(options, huh.NewOption(account.Name, account))
			}
			for _, account := range m.api.AccountsByType("liabilities") {
				options = append(options, huh.NewOption(account.Name, account))
			}
		}
		return options
	}, bindings
}

func (m *modelTransaction) GroupTitle() string {
	if len(m.splits) > 1 {
		if m.attr.groupTitle != "" {
			return m.attr.groupTitle
		}
		acc := ""
		switch m.attr.transactionType {
		case "withdrawal":
			acc = m.attr.source.Name
		case "deposit":
			acc = m.attr.destination.Name
		case "transfer":
			acc = fmt.Sprintf("%s -> %s", m.attr.source.Name, m.attr.destination.Name)
		}
		return fmt.Sprintf("%s, splits: %d, %s", m.attr.transactionType, len(m.splits), acc)
	}
	return ""
}

func (s *split) Description() string {
	if s.description == "" {
		return fmt.Sprintf("%s, %s -> %s", s.category.Name, s.source.Name, s.destination.Name)
	}
	return s.description
}

func (s *split) CurrencyCode() string {
	switch s.source.Type {
	case "asset", "liabilities":
		return s.source.CurrencyCode
	case "revenue":
		return s.destination.CurrencyCode
	}
	return ""
}

func (s *split) ForeignCurrencyCode() string {
	if (s.source.Type == "asset" || s.source.Type == "liabilities") && (s.destination.Type == "asset" || s.destination.Type == "liabilities") {
		if s.source.CurrencyCode != s.destination.CurrencyCode {
			return s.destination.CurrencyCode
		}
	}
	return ""
}
