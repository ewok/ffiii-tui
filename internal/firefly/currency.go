/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import "strings"

type Currency struct {
	Code string
}

func (c Currency) String() string {
	return string(c.Code)
}

func (c Currency) GetLCode() string {
	return strings.ToLower(c.Code)
}

func (c Currency) GetCode() string {
	return strings.ToUpper(c.Code)
}

func NewCurrency(code string) Currency {
	return Currency{Code: strings.ToUpper(code)}
}

