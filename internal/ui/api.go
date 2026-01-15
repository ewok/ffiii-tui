/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"time"

	"ffiii-tui/internal/firefly"
)

// Note: Interfaces in this file are consumer-driven (UI-owned).
// Each UI model should depend on the smallest interface it needs.

// PeriodAPI provides date-range navigation.
type PeriodAPI interface {
	PreviousPeriod()
	NextPeriod()
}

// CurrencyAPI provides access to currency configuration used in UI.
type CurrencyAPI interface {
	PrimaryCurrency() firefly.Currency
}

// SummaryAPI provides summary refresh and read access.
type SummaryAPI interface {
	UpdateSummary() error
	GetMaxWidth() int
	SummaryItems() map[string]firefly.SummaryItem
}

// AccountsAPI provides account refresh and read access.
type AccountsAPI interface {
	UpdateAccounts(accountType string) error
	AccountsByType(accountType string) []firefly.Account
	AccountBalance(accountID string) float64
}

// AssetAPI is the minimal API used by the assets UI.
type AssetAPI interface {
	AccountsAPI
	CreateAssetAccount(name, currencyCode string) error
}

// AccountCreateAPI provides account creation operations.
type AccountCreateAPI interface {
	CreateAssetAccount(name, currencyCode string) error
	CreateExpenseAccount(name string) error
	CreateRevenueAccount(name string) error
	CreateLiabilityAccount(nl firefly.NewLiability) error
}

// ExpenseInsightsAPI provides expense insights used by the UI.
type ExpenseInsightsAPI interface {
	UpdateExpenseInsights() error
	GetExpenseDiff(accountID string) float64
	GetTotalExpenseDiff() float64
}

// ExpenseAPI is the minimal API used by the expenses UI.
type ExpenseAPI interface {
	AccountsAPI
	CurrencyAPI
	ExpenseInsightsAPI
	CreateExpenseAccount(name string) error
}

// RevenueInsightsAPI provides revenue insights used by the UI.
type RevenueInsightsAPI interface {
	UpdateRevenueInsights() error
	GetRevenueDiff(accountID string) float64
	GetTotalRevenueDiff() float64
}

// RevenueAPI is the minimal API used by the revenues UI.
type RevenueAPI interface {
	AccountsAPI
	CurrencyAPI
	RevenueInsightsAPI
	CreateRevenueAccount(name string) error
}

// LiabilityAPI is the minimal API used by the liabilities UI.
type LiabilityAPI interface {
	AccountsAPI
	CreateLiabilityAccount(nl firefly.NewLiability) error
}

// CategoriesAPI provides category refresh and read access.
type CategoriesAPI interface {
	UpdateCategories() error
	UpdateCategoriesInsights() error
	CategoriesList() []firefly.Category
	GetTotalSpentEarnedCategories() (spent, earned float64)
	CategorySpent(categoryID string) float64
	CategoryEarned(categoryID string) float64
	CreateCategory(name, notes string) error
}

// CategoryAPI is the minimal API used by the categories UI.
type CategoryAPI interface {
	CategoriesAPI
	CurrencyAPI
}

// TransactionAPI provides read/delete operations for the transaction list.
type TransactionAPI interface {
	ListTransactions(query string) ([]firefly.Transaction, error)
	DeleteTransaction(transactionID string) error
}

// TransactionWriteAPI provides create/update operations used by the transaction form.
type TransactionWriteAPI interface {
	CreateTransaction(tx firefly.RequestTransaction) error
	UpdateTransaction(transactionID string, tx firefly.RequestTransaction) error
}

// TransactionFormAPI is the minimal API used by the transaction form UI.
type TransactionFormAPI interface {
	AccountsAPI
	CategoriesAPI
	TransactionWriteAPI
}

// UIAPI is the minimal API used by the root UI model.
// It is intentionally larger since it wires multiple sub-models.
type UIAPI interface {
	PeriodAPI
	SummaryAPI
	AssetAPI
	CategoryAPI
	ExpenseAPI
	RevenueAPI
	LiabilityAPI
	TransactionAPI
	TransactionFormAPI

	TimeoutSeconds() int
	PeriodStart() time.Time
	PeriodEnd() time.Time
}
