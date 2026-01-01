package ui

import (
	"github.com/charmbracelet/bubbles/key"
)

type UIKeyMap struct {
	Quit          key.Binding
	ShowShortHelp key.Binding

	PreviousPeriod key.Binding
	NextPeriod     key.Binding

	ViewAssets       key.Binding
	ViewExpenses     key.Binding
	ViewRevenue      key.Binding
	ViewCategories   key.Binding
	ViewTransactions key.Binding
}

type AssetKeyMap struct {
	ShowFullHelp key.Binding
	Quit         key.Binding
	Filter       key.Binding
	ResetFilter  key.Binding
	Select       key.Binding
	New          key.Binding
	Refresh      key.Binding
}

type ExpenseKeyMap struct {
	ShowFullHelp key.Binding
	Quit         key.Binding
	Filter       key.Binding
	ResetFilter  key.Binding
	New          key.Binding
	Refresh      key.Binding
	Sort         key.Binding
}

type RevenueKeyMap struct {
	ShowFullHelp key.Binding
	Quit         key.Binding
	Filter       key.Binding
	ResetFilter  key.Binding
	New          key.Binding
	Refresh      key.Binding
	Sort         key.Binding
}

type CategoryKeyMap struct {
	ShowFullHelp key.Binding
	Quit         key.Binding
	Filter       key.Binding
	ResetFilter  key.Binding
	New          key.Binding
	Refresh      key.Binding
	Sort         key.Binding
}

type TransactionFormKeyMap struct {
	Reset      key.Binding
	Cancel     key.Binding
	Submit     key.Binding
	NewElement key.Binding
}

type TransactionsKeyMap struct {
	ShowFullHelp       key.Binding
	Quit               key.Binding
	Refresh            key.Binding
	Filter             key.Binding
	ResetFilter        key.Binding
	Search             key.Binding
	New                key.Binding
	NewFromTransaction key.Binding
	ToggleFullView     key.Binding
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
		ViewAssets: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "view assets"),
		),
		ViewExpenses: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "view expenses"),
		),
		ViewRevenue: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "view revenues"),
		),
		ViewCategories: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "view categories"),
		),
		ViewTransactions: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "view transactions"),
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
	}
}

func DefaultExpenseKeyMap() ExpenseKeyMap {
	return ExpenseKeyMap{
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more/less"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "Back"),
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
	}
}

func DefaultRevenueKeyMap() RevenueKeyMap {
	return RevenueKeyMap{
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more/less"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "Back"),
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
	}
}

func DefaultCategoryKeyMap() CategoryKeyMap {
	return CategoryKeyMap{
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more/less"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "Back"),
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
	}
}

func DefaultTransactionFormKeyMap() TransactionFormKeyMap {
	return TransactionFormKeyMap{
		Reset: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "reset form"),
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
	}
}

func DefaultTransactionsKeyMap() TransactionsKeyMap {
	return TransactionsKeyMap{
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more/less"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
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
			key.WithHelp("n", "new transaction"),
		),
		NewFromTransaction: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "new from transaction"),
		),
		ToggleFullView: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle full view"),
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
		k.Refresh,
	}
}

func (k TransactionFormKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Submit,
		k.Cancel,
		k.Reset,
		k.NewElement,
	}
}

func (k UIKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.PreviousPeriod,
			k.NextPeriod,
		},
		{
			k.ViewAssets,
			k.ViewTransactions,
			k.ViewExpenses,
			k.ViewRevenue,
			k.ViewCategories,
		},
	}
}

func (k AssetKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

func (k ExpenseKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

func (k RevenueKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}
func (k CategoryKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

func (k TransactionsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}

func (k TransactionFormKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		k.ShortHelp(),
	}
}
