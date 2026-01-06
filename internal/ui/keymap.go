/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

type UIKeyMap struct {
	Quit          key.Binding
	ShowShortHelp key.Binding

	PreviousPeriod key.Binding
	NextPeriod     key.Binding
}

type AssetKeyMap struct {
	ShowFullHelp key.Binding
	Quit         key.Binding
	Filter       key.Binding
	ResetFilter  key.Binding
	Select       key.Binding
	New          key.Binding
	Refresh      key.Binding

	ViewTransactions key.Binding
	ViewAssets       key.Binding
	ViewCategories   key.Binding
	ViewExpenses     key.Binding
	ViewRevenues     key.Binding
	ViewLiabilities  key.Binding
}

type ExpenseKeyMap struct {
	ShowFullHelp key.Binding
	Quit         key.Binding
	Filter       key.Binding
	ResetFilter  key.Binding
	New          key.Binding
	Refresh      key.Binding
	Sort         key.Binding

	ViewTransactions key.Binding
	ViewAssets       key.Binding
	ViewCategories   key.Binding
	ViewExpenses     key.Binding
	ViewRevenues     key.Binding
	ViewLiabilities  key.Binding
}

type RevenueKeyMap struct {
	ShowFullHelp key.Binding
	Quit         key.Binding
	Filter       key.Binding
	ResetFilter  key.Binding
	New          key.Binding
	Refresh      key.Binding
	Sort         key.Binding

	ViewTransactions key.Binding
	ViewAssets       key.Binding
	ViewCategories   key.Binding
	ViewExpenses     key.Binding
	ViewRevenues     key.Binding
	ViewLiabilities  key.Binding
}

type CategoryKeyMap struct {
	ShowFullHelp key.Binding
	Quit         key.Binding
	Filter       key.Binding
	ResetFilter  key.Binding
	New          key.Binding
	Refresh      key.Binding
	Sort         key.Binding

	ViewTransactions key.Binding
	ViewAssets       key.Binding
	ViewCategories   key.Binding
	ViewExpenses     key.Binding
	ViewRevenues     key.Binding
	ViewLiabilities  key.Binding
}

type LiabilityKeyMap struct {
	ShowFullHelp key.Binding
	Quit         key.Binding
	Filter       key.Binding
	ResetFilter  key.Binding
	New          key.Binding
	Refresh      key.Binding

	ViewTransactions key.Binding
	ViewAssets       key.Binding
	ViewCategories   key.Binding
	ViewExpenses     key.Binding
	ViewRevenues     key.Binding
	ViewLiabilities  key.Binding
}

type TransactionFormKeyMap struct {
	Reset        key.Binding
	Cancel       key.Binding
	Submit       key.Binding
	NewElement   key.Binding
	Refresh      key.Binding
	AddSplit     key.Binding
	DeleteSplit  key.Binding
	ChangeLayout key.Binding
}

type TransactionsKeyMap struct {
	ShowFullHelp       key.Binding
	Quit               key.Binding
	Refresh            key.Binding
	Filter             key.Binding
	ResetFilter        key.Binding
	Search             key.Binding
	New                key.Binding
	Select             key.Binding
	NewFromTransaction key.Binding
	Delete             key.Binding
	ToggleFullView     key.Binding

	ViewAssets      key.Binding
	ViewCategories  key.Binding
	ViewExpenses    key.Binding
	ViewRevenues    key.Binding
	ViewLiabilities key.Binding
}

func DefaultUIKeyMap() UIKeyMap {
	return UIKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		ShowShortHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "less"),
		),
		PreviousPeriod: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[", "previous period"),
		),
		NextPeriod: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]", "next period"),
		),
	}
}

func DefaultAssetKeyMap() AssetKeyMap {
	return AssetKeyMap{
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more/less"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "Back"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter assets(twice for exclusive)"),
		),
		ResetFilter: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "reset filter"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select asset"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new asset"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh assets"),
		),
		ViewTransactions: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "view transactions"),
		),
		ViewAssets: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "view assets"),
			key.WithDisabled(),
		),
		ViewCategories: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "view categories"),
		),
		ViewExpenses: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "view expenses"),
		),
		ViewRevenues: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "view revenues"),
		),
		ViewLiabilities: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "view liabilities"),
		),
	}
}

func DefaultExpenseKeyMap() ExpenseKeyMap {
	return ExpenseKeyMap{
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more/less"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "Back"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter expenses(twice for exclusive)"),
		),
		ResetFilter: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "reset filter"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new expense"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh expenses"),
		),
		Sort: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sort expenses"),
		),
		ViewTransactions: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "view transactions"),
		),
		ViewAssets: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "view assets"),
		),
		ViewCategories: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "view categories"),
		),
		ViewExpenses: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "view expenses"),
			key.WithDisabled(),
		),
		ViewRevenues: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "view revenues"),
		),
		ViewLiabilities: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "view liabilities"),
		),
	}
}

func DefaultRevenueKeyMap() RevenueKeyMap {
	return RevenueKeyMap{
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more/less"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "Back"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter revenues(twice for exclusive)"),
		),
		ResetFilter: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "reset filter"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new revenue"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh revenues"),
		),
		Sort: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sort revenues"),
		),
		ViewTransactions: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "view transactions"),
		),
		ViewAssets: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "view assets"),
		),
		ViewCategories: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "view categories"),
		),
		ViewExpenses: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "view expenses"),
		),
		ViewRevenues: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "view revenues"),
			key.WithDisabled(),
		),
		ViewLiabilities: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "view liabilities"),
		),
	}
}

func DefaultCategoryKeyMap() CategoryKeyMap {
	return CategoryKeyMap{
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more/less"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "Back"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter categories(twice for exclusive)"),
		),
		ResetFilter: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "reset filter"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new category"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh categories"),
		),
		Sort: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sort categories"),
		),
		ViewTransactions: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "view transactions"),
		),
		ViewAssets: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "view assets"),
		),
		ViewCategories: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "view categories"),
			key.WithDisabled(),
		),
		ViewExpenses: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "view expenses"),
		),
		ViewRevenues: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "view revenues"),
		),
		ViewLiabilities: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "view liabilities"),
		),
	}
}

func DefaultLiabilityKeyMap() LiabilityKeyMap {
	return LiabilityKeyMap{
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more/less"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "Back"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter liabilities(twice for exclusive)"),
		),
		ResetFilter: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "reset filter"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh liabilities"),
		),
		ViewTransactions: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "view transactions"),
		),
		ViewAssets: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "view assets"),
		),
		ViewCategories: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "view categories"),
		),
		ViewExpenses: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "view expenses"),
		),
		ViewRevenues: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "view revenues"),
		),
		ViewLiabilities: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "view liabilities"),
			key.WithDisabled(),
		),
	}
}

func DefaultTransactionFormKeyMap() TransactionFormKeyMap {
	return TransactionFormKeyMap{
		Reset: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "reset form"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh data"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
		NewElement: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new element"),
		),
		AddSplit: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "add split"),
		),
		DeleteSplit: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "delete split"),
		),
		ChangeLayout: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("ctrl+f", "change layout(if too many splits)"),
		),
	}
}

func DefaultTransactionsKeyMap() TransactionsKeyMap {
	return TransactionsKeyMap{
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more/less"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter(twice for exclusive)"),
		),
		ResetFilter: key.NewBinding(
			key.WithKeys("ctrl+a"),
			key.WithHelp("ctrl+a", "reset filter"),
		),
		Search: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "global search"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "transaction view"),
		),
		NewFromTransaction: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "new transaction from..."),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "edit current transaction"),
		),
		Delete: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "delete transaction"),
		),
		ToggleFullView: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle full view"),
		),
		ViewAssets: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "view assets"),
		),
		ViewCategories: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "view categories"),
		),
		ViewExpenses: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "view expenses"),
		),
		ViewRevenues: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "view revenues"),
		),
		ViewLiabilities: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "view liabilities"),
		),
	}
}

func (k UIKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.ShowShortHelp,
		k.Quit,
		k.PreviousPeriod,
		k.NextPeriod,
	}
}

func (k AssetKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.ShowFullHelp,
		k.Quit,
		k.Filter,
		k.ResetFilter,
		k.Select,
		k.New,
		k.Refresh,
	}
}

func (k ExpenseKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.ShowFullHelp,
		k.Quit,
		k.Filter,
		k.ResetFilter,
		k.New,
		k.Refresh,
		k.Sort,
	}
}

func (k RevenueKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.ShowFullHelp,
		k.Quit,
		k.Filter,
		k.ResetFilter,
		k.New,
		k.Refresh,
		k.Sort,
	}
}

func (k CategoryKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.ShowFullHelp,
		k.Quit,
		k.Filter,
		k.ResetFilter,
		k.New,
		k.Refresh,
		k.Sort,
	}
}

func (k LiabilityKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.ShowFullHelp,
		k.Quit,
		k.Filter,
		k.ResetFilter,
		k.New,
		k.Refresh,
	}
}

func (k TransactionsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.ShowFullHelp,
		k.Quit,
		k.ToggleFullView,
		k.Search,
		k.Filter,
		k.ResetFilter,
		k.New,
		k.NewFromTransaction,
		k.Select,
		k.Delete,
		k.Refresh,
	}
}

func (k TransactionFormKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.NewElement,
		k.AddSplit,
		k.DeleteSplit,
		k.Submit,
		k.Cancel,
		k.Reset,
		k.Refresh,
		k.ChangeLayout,
	}
}

func (k UIKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.PreviousPeriod,
			k.NextPeriod,
		},
	}
}

func (k AssetKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
		{
			k.ViewTransactions,
			k.ViewAssets,
			k.ViewCategories,
			k.ViewExpenses,
			k.ViewRevenues,
			k.ViewLiabilities,
		},
	}
}

func (k ExpenseKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
		{
			k.ViewTransactions,
			k.ViewAssets,
			k.ViewCategories,
			k.ViewExpenses,
			k.ViewRevenues,
			k.ViewLiabilities,
		},
	}
}

func (k RevenueKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
		{
			k.ViewTransactions,
			k.ViewAssets,
			k.ViewCategories,
			k.ViewExpenses,
			k.ViewRevenues,
			k.ViewLiabilities,
		},
	}
}

func (k CategoryKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
		{
			k.ViewTransactions,
			k.ViewAssets,
			k.ViewCategories,
			k.ViewExpenses,
			k.ViewRevenues,
			k.ViewLiabilities,
		},
	}
}

func (k LiabilityKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
		{
			k.ViewTransactions,
			k.ViewAssets,
			k.ViewCategories,
			k.ViewExpenses,
			k.ViewRevenues,
			k.ViewLiabilities,
		},
	}
}

func (k TransactionsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
		{
			k.ViewAssets,
			k.ViewCategories,
			k.ViewExpenses,
			k.ViewRevenues,
			k.ViewLiabilities,
		},
	}
}

func (k TransactionFormKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}
