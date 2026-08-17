/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

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

var fullNewForm bool

type (
	RedrawFormMsg                  struct{}
	ContinueTransactionMsg         struct{}
	DeleteSplitMsg                 struct{ Index int }
	NewTransactionMsg              struct{ Transaction firefly.Transaction }
	NewTransactionFromMsg          struct{ Transaction firefly.Transaction }
	NewTransactionFromConfirmedMsg struct{ Transaction firefly.Transaction }
	EditTransactionMsg             struct{ Transaction firefly.Transaction }
	EditTransactionConfirmedMsg    struct{ Transaction firefly.Transaction }
	ResetTransactionMsg            struct{ Transaction firefly.Transaction }
	TransactionSaveResultMsg       struct {
		ID      string
		Err     error
		Updated bool
	}
)

type modelTransaction struct {
	form   *huh.Form
	api    TransactionFormAPI
	keymap TransactionFormKeyMap
	focus  bool

	new      bool
	created  bool
	lastDate string

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
	case NewTransactionMsg:
		if m.created {
			trx := msg.Transaction
			return m, prompt.Ask(
				"Unsaved form data will be lost. Discard? (y - yes/ any key - no): ",
				"",
				func(value string) tea.Cmd {
					if value == "y" {
						return Cmd(NewTransactionFromConfirmedMsg{Transaction: trx})
					}
					return SetView(transactionsView)
				},
			)
		}
		m.SetTransaction(msg.Transaction, true)
		m.created = true
		return m, tea.Batch(
			RedrawForm(),
			SetView(newView),
		)
	case ContinueTransactionMsg:
		if m.created {
			return m, tea.Batch(
				RedrawForm(),
				SetView(newView),
			)
		}
		return m, nil
	case NewTransactionFromMsg:
		if m.created {
			trx := msg.Transaction
			return m, prompt.Ask(
				"Unsaved form data will be lost. Discard? (y - yes/ any key - no): ",
				"",
				func(value string) tea.Cmd {
					if value == "y" {
						return Cmd(NewTransactionFromConfirmedMsg{Transaction: trx})
					}
					return SetView(transactionsView)
				},
			)
		}
		m.SetTransaction(msg.Transaction, true)
		m.created = true
		return m, tea.Batch(
			RedrawForm(),
			SetView(newView),
		)
	case NewTransactionFromConfirmedMsg:
		m.SetTransaction(msg.Transaction, true)
		m.created = true
		return m, tea.Batch(
			RedrawForm(),
			SetView(newView),
		)
	case EditTransactionMsg:
		if m.created {
			trx := msg.Transaction
			return m, prompt.Ask(
				"Unsaved form data will be lost. Discard? (y - yes/ any key - no): ",
				"",
				func(value string) tea.Cmd {
					if value == "y" {
						return Cmd(EditTransactionConfirmedMsg{Transaction: trx})
					}
					return SetView(transactionsView)
				},
			)
		}
		m.SetTransaction(msg.Transaction, false)
		m.created = true
		return m, tea.Batch(
			RedrawForm(),
			SetView(newView),
		)
	case EditTransactionConfirmedMsg:
		m.SetTransaction(msg.Transaction, false)
		m.created = true
		return m, tea.Batch(
			RedrawForm(),
			SetView(newView),
		)
	case ResetTransactionMsg:
		m.SetTransaction(msg.Transaction, true)
		now := time.Now()
		m.attr.year = fmt.Sprintf("%d", now.Year())
		m.attr.month = fmt.Sprintf("%02d", now.Month())
		m.attr.day = fmt.Sprintf("%02d", now.Day())
		m.created = true
		return m, RedrawForm()
	case RedrawFormMsg:
		m.UpdateForm()
		return m, tea.WindowSize()
	case DeleteSplitMsg:
		return m, m.DeleteSplit(msg.Index)
	case TransactionSaveResultMsg:
		if msg.Err != nil {
			return m, tea.Sequence(
				notify.NotifyError(msg.Err.Error()),
				SetView(transactionsView))
		}

		m.created = false

		action := "created"
		if msg.Updated {
			action = "updated"
		} else {
			m.lastDate = fmt.Sprintf("%s-%s-%s", m.attr.year, m.attr.month, m.attr.day)
		}
		return m, tea.Batch(
			SetView(transactionsView),
			notify.NotifyLog(fmt.Sprintf("Transaction %s successfully", action)),
			Cmd(RefreshAssetsMsg{}),
			Cmd(RefreshLiabilitiesMsg{}),
			Cmd(RefreshSummaryMsg{}),
			Cmd(RefreshTransactionsMsg{TrxID: msg.ID}),
			Cmd(RefreshExpenseInsightsMsg{}),
			Cmd(RefreshRevenueInsightsMsg{}),
			Cmd(RefreshCategoryInsightsMsg{}))
	}

	if !m.focus {
		return m, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Cancel):
			if m.created {
				return m, tea.Batch(
					SetView(transactionsView),
					notify.NotifyLog("Form saved. Press esc to continue editing."),
				)
			}
			return m, SetView(transactionsView)
		case key.Matches(msg, m.keymap.Reset):
			return m, tea.Batch(
				SetView(newView),
				Cmd(ResetTransactionMsg{}),
			)
		case key.Matches(msg, m.keymap.Refresh):
			return m, RedrawForm()
		case key.Matches(msg, m.keymap.EditFormAgain):
			return m, RedrawForm()
		case key.Matches(msg, m.keymap.AddSplit):
			if len(m.splits) >= 5 {
				return m, notify.NotifyWarn("Maximum of 5 splits allowed")
			}
			m.splits = append(m.splits, &split{})
			return m, RedrawForm()
		case key.Matches(msg, m.keymap.DeleteSplit):
			if len(m.splits) <= 1 {
				return m, notify.NotifyWarn("Cannot delete the only split")
			}
			if len(m.splits) == 2 {
				// Only one deletable split (index 1), delete it directly
				return m, Cmd(DeleteSplitMsg{Index: 1})
			}
			// Build a descriptive prompt listing deletable splits
			options := ""
			for i := 1; i < len(m.splits); i++ {
				s := m.splits[i]
				desc := s.Description()
				if s.amount != "" {
					desc += " (" + s.amount + ")"
				}
				options += fmt.Sprintf(" %d: %s,", i, desc)
			}
			return m, prompt.Ask(
				fmt.Sprintf("Delete split [%s ]: ", options),
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
		case key.Matches(msg, m.keymap.ToggleLastDate):
			return m, m.toggleLastDate()
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
		return "Press Ctrl+S to submit, Ctrl+N to reset current form, Ctrl+E to edit current form again, or Esc to go back."
	}
	return m.form.View()
}

func (m *modelTransaction) Focus() {
	m.focus = true
}

func (m *modelTransaction) Blur() {
	m.focus = false
}

func (m *modelTransaction) Focused() bool {
	return m.focus
}

func (m *modelTransaction) UpdateForm() {
	opts := m.buildFormOptions()
	var allGroups []*huh.Group

	for i, s := range m.splits {
		note := huh.NewNote().Title(fmt.Sprint("Split: ", i))
		if i == 0 {
			note = note.TitleFunc(func() string {
				return fmt.Sprintf("Current Type: %s", deriveTransactionType(s.source, s.destination))
			}, []any{&s.source.Type, &s.destination.Type})
		}

		sourceSelect := huh.NewSelect[firefly.Account]().
			Title("Source").
			Value(&s.source)
		if i == 0 {
			sourceSelect = sourceSelect.Options(firstSourceOptions(opts)...)
		} else {
			sourceSelect = sourceSelect.
				Options(huh.NewOption(s.source.Name, s.source)).
				OptionsFunc(m.trxSourceOptions(opts))
		}

		allGroups = append(allGroups, huh.NewGroup(
			note,
			sourceSelect.WithHeight(5),
			huh.NewSelect[firefly.Account]().
				Title("Destination").
				Value(&s.destination).
				Options(huh.NewOption(s.destination.Name, s.destination)).
				OptionsFunc(m.trxDestinationOptions(opts, i, s)).WithHeight(4),
			huh.NewSelect[firefly.Category]().
				Title("Category").
				Value(&s.category).
				Options(opts.categories...).WithHeight(4),
			huh.NewInput().
				Title("Amount").
				Value(&s.amount).
				TitleFunc(func() string {
					title := "Amount "
					switch s.source.Type {
					case "asset", "liabilities":
						return title + s.source.CurrencyCode
					case "revenue", "cash":
						return title + s.destination.CurrencyCode
					}
					return title
				}, []any{&s.source.Type, &s.source.CurrencyCode, &s.destination.CurrencyCode}).
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
				}, []any{&s.source.Type, &s.source.CurrencyCode, &s.destination.Type, &s.destination.CurrencyCode}).
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
				PlaceholderFunc(s.Description, []any{&s.category.Name, &s.source.Name, &s.destination.Name}).
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
		first := m.firstSplit()
		allGroups = append(allGroups, huh.NewGroup(
			huh.NewInput().
				Title("Group Title").
				Value(&m.attr.groupTitle).
				PlaceholderFunc(m.GroupTitle, []any{&first.source.ID, &first.destination.ID}).
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

func (m *modelTransaction) buildRequestSplits() []firefly.RequestTransactionSplit {
	ttype := m.transactionType()
	date := fmt.Sprintf("%s-%s-%s", m.attr.year, m.attr.month, m.attr.day)
	first := m.firstSplit()

	trx := []firefly.RequestTransactionSplit{}
	for i, s := range m.splits {
		if i > 0 {
			switch ttype {
			case "withdrawal":
				s.source = first.source
			case "deposit":
				s.destination = first.destination
			case "transfer":
				s.source = first.source
				s.destination = first.destination
			}
		}
		trx = append(trx, firefly.RequestTransactionSplit{
			TransactionJournalID: s.trxJID,
			Type:                 ttype,
			Date:                 date,
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
	return trx
}

func (m *modelTransaction) CreateTransaction() tea.Cmd {
	req := firefly.RequestTransaction{
		ApplyRules:           true,
		ErrorIfDuplicateHash: false,
		FireWebhooks:         true,
		GroupTitle:           m.GroupTitle(),
		Transactions:         m.buildRequestSplits(),
	}

	api := m.api
	return func() tea.Msg {
		opID := startLoading("Creating transaction...")
		defer stopLoading(opID)
		id, err := api.CreateTransaction(req)
		return TransactionSaveResultMsg{ID: id, Err: err}
	}
}

func (m *modelTransaction) UpdateTransaction() tea.Cmd {
	req := firefly.RequestTransaction{
		ApplyRules:   true,
		FireWebhooks: true,
		GroupTitle:   m.GroupTitle(),
		Transactions: m.buildRequestSplits(),
	}

	api := m.api
	trxID := m.attr.trxID
	return func() tea.Msg {
		opID := startLoading("Updating transaction...")
		defer stopLoading(opID)
		id, err := api.UpdateTransaction(trxID, req)
		return TransactionSaveResultMsg{ID: id, Err: err, Updated: true}
	}
}

func (m *modelTransaction) SetTransaction(trx firefly.Transaction, newT bool) {
	zap.L().Debug("newModelTransaction", zap.Any("trx", trx))

	m.new = newT

	now := time.Now()

	if trx.TransactionID != "" {
		m.attr.transactionType = trx.Type
		m.attr.year, m.attr.month, m.attr.day = splitTransactionDate(trx.Date, now)
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

func (m *modelTransaction) toggleLastDate() tea.Cmd {
	if !m.new {
		return nil
	}
	if m.lastDate == "" {
		return notify.NotifyWarn("No saved date yet")
	}
	now := time.Now()
	current := fmt.Sprintf("%s-%s-%s", m.attr.year, m.attr.month, m.attr.day)
	if current == m.lastDate {
		m.attr.year = fmt.Sprintf("%d", now.Year())
		m.attr.month = fmt.Sprintf("%02d", now.Month())
		m.attr.day = fmt.Sprintf("%02d", now.Day())
	} else {
		m.attr.year, m.attr.month, m.attr.day = splitTransactionDate(m.lastDate, now)
	}
	return RedrawForm()
}

func RedrawForm() tea.Cmd {
	return Cmd(RedrawFormMsg{})
}

func splitTransactionDate(date string, fallback time.Time) (year, month, day string) {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		t, err = time.Parse("2006-01-02", date)
	}
	if err != nil {
		zap.S().Warnf("Failed to parse transaction date %q, using current date: %v", date, err)
		t = fallback
	}
	return fmt.Sprintf("%d", t.Year()), fmt.Sprintf("%02d", t.Month()), fmt.Sprintf("%02d", t.Day())
}

func deriveTransactionType(source, destination firefly.Account) string {
	stx := source.Type
	dtx := destination.Type

	switch {
	case stx == "asset" && (dtx == "expense" || dtx == "liabilities" || dtx == "cash"):
		return "withdrawal"
	case stx == "asset" && dtx == "asset":
		return "transfer"
	case stx == "revenue" || stx == "cash":
		return "deposit"
	case stx == "liabilities" && (dtx == "expense" || dtx == "cash"):
		return "withdrawal"
	case stx == "liabilities" && dtx == "asset":
		return "deposit"
	case stx == "liabilities" && dtx == "liabilities":
		return "transfer"
	default:
		return "unknown"
	}
}

func (m *modelTransaction) firstSplit() *split {
	if len(m.splits) > 0 {
		return m.splits[0]
	}
	return &split{}
}

func (m *modelTransaction) transactionType() string {
	if len(m.splits) > 0 {
		return deriveTransactionType(m.splits[0].source, m.splits[0].destination)
	}
	return m.attr.transactionType
}

// Helpers

// formOptions holds account and category options prebuilt once per form
// redraw, so the dynamic huh OptionsFuncs only slice prebuilt lists instead
// of rebuilding them on every evaluation.
type formOptions struct {
	asset       []huh.Option[firefly.Account]
	expense     []huh.Option[firefly.Account]
	revenue     []huh.Option[firefly.Account]
	liabilities []huh.Option[firefly.Account]
	cash        []huh.Option[firefly.Account]
	categories  []huh.Option[firefly.Category]
}

func (m *modelTransaction) buildFormOptions() *formOptions {
	accountOptions := func(accountType string) []huh.Option[firefly.Account] {
		accounts := m.api.AccountsByType(accountType)
		options := make([]huh.Option[firefly.Account], 0, len(accounts))
		for _, account := range accounts {
			options = append(options, huh.NewOption(account.Name, account))
		}
		return options
	}

	categories := m.api.CategoriesList()
	categoryOptions := make([]huh.Option[firefly.Category], 0, len(categories))
	for _, category := range categories {
		categoryOptions = append(categoryOptions, huh.NewOption(category.Name, category))
	}

	return &formOptions{
		asset:       accountOptions("asset"),
		expense:     accountOptions("expense"),
		revenue:     accountOptions("revenue"),
		liabilities: accountOptions("liabilities"),
		cash:        accountOptions("cash"),
		categories:  categoryOptions,
	}
}

func concatOptions(lists ...[]huh.Option[firefly.Account]) []huh.Option[firefly.Account] {
	total := 0
	for _, list := range lists {
		total += len(list)
	}
	options := make([]huh.Option[firefly.Account], 0, total)
	for _, list := range lists {
		options = append(options, list...)
	}
	return options
}

// firstSourceOptions returns the static source options for the first split.
func firstSourceOptions(opts *formOptions) []huh.Option[firefly.Account] {
	return concatOptions(opts.asset, opts.revenue, opts.liabilities, opts.cash)
}

// trxSourceOptions returns the source options for splits after the first one:
// they follow the first split for withdrawals/transfers and offer the full
// deposit source list otherwise.
func (m *modelTransaction) trxSourceOptions(opts *formOptions) (func() []huh.Option[firefly.Account], any) {
	first := m.firstSplit()
	bindings := []any{&first.source.ID, &first.destination.ID}

	return func() []huh.Option[firefly.Account] {
		ttype := deriveTransactionType(first.source, first.destination)
		if ttype == "withdrawal" || ttype == "transfer" {
			return []huh.Option[firefly.Account]{huh.NewOption(first.source.Name, first.source)}
		}
		return concatOptions(opts.revenue, opts.liabilities, opts.cash)
	}, bindings
}

func (m *modelTransaction) trxDestinationOptions(opts *formOptions, i int, s *split) (func() []huh.Option[firefly.Account], any) {
	destinationsBySourceType := func(includeSameType bool) []huh.Option[firefly.Account] {
		switch s.source.Type {
		case "asset":
			if includeSameType {
				return concatOptions(opts.expense, opts.asset, opts.liabilities)
			}
			return concatOptions(opts.expense, opts.liabilities)
		case "revenue", "cash":
			return concatOptions(opts.asset, opts.liabilities)
		case "liabilities":
			if includeSameType {
				return concatOptions(opts.asset, opts.expense, opts.liabilities)
			}
			return concatOptions(opts.asset, opts.expense)
		}
		return nil
	}

	if i > 0 {
		first := m.firstSplit()
		bindings := []any{&s.source.Type, &first.source.ID, &first.destination.ID}
		return func() []huh.Option[firefly.Account] {
			ttype := deriveTransactionType(first.source, first.destination)
			if ttype == "deposit" || ttype == "transfer" {
				return []huh.Option[firefly.Account]{huh.NewOption(first.destination.Name, first.destination)}
			}
			return destinationsBySourceType(false)
		}, bindings
	}

	bindings := []any{&s.source.Type}
	return func() []huh.Option[firefly.Account] {
		return destinationsBySourceType(true)
	}, bindings
}

func (m *modelTransaction) GroupTitle() string {
	if len(m.splits) > 1 {
		if m.attr.groupTitle != "" {
			return m.attr.groupTitle
		}
		first := m.firstSplit()
		ttype := m.transactionType()
		acc := ""
		switch ttype {
		case "withdrawal":
			acc = first.source.Name
		case "deposit":
			acc = first.destination.Name
		case "transfer":
			acc = fmt.Sprintf("%s -> %s", first.source.Name, first.destination.Name)
		}
		return fmt.Sprintf("%s, splits: %d, %s", ttype, len(m.splits), acc)
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
	case "revenue", "cash":
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
