/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package ui

import (
	"fmt"

	"ffiii-tui/internal/firefly"
)

// ListEntity is a constraint for types that can be displayed in account lists
type ListEntity interface {
	firefly.Account
	GetName() string
}

// accountListItem is a generic list item for accounts and categories
type accountListItem[T ListEntity] struct {
	Entity       T
	PrimaryVal   float64
	primaryLabel string
}

// Accessors for backward compatibility with tests
func (i accountListItem[T]) GetPrimaryVal() float64 {
	return i.PrimaryVal
}

func (i accountListItem[T]) Title() string {
	var name string
	switch entity := any(i.Entity).(type) {
	case firefly.Account:
		name = entity.Name
	}
	return name
}

func (i accountListItem[T]) Description() string {
	var currencyCode string
	switch entity := any(i.Entity).(type) {
	case firefly.Account:
		currencyCode = entity.CurrencyCode
	}

	desc := fmt.Sprintf("%s: %.2f %s", i.primaryLabel, i.PrimaryVal, currencyCode)
	return desc
}

func (i accountListItem[T]) FilterValue() string {
	var name string
	switch entity := any(i.Entity).(type) {
	case firefly.Account:
		name = entity.Name
	}
	return name
}

// newAccountListItem creates a new account list item with a single value
func newAccountListItem[T ListEntity](entity T, primaryLabel string, primaryVal float64) accountListItem[T] {
	return accountListItem[T]{
		Entity:       entity,
		PrimaryVal:   primaryVal,
		primaryLabel: primaryLabel,
	}
}
