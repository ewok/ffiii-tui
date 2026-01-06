/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import (
	"fmt"
	"strings"
)

type Currency struct {
	ID   string
	Code string
	Name string
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

type apiCurrency struct {
	ID         string          `json:"id"`
	Attributes apiCurrencyAttr `json:"attributes"`
}

type apiCurrencyAttr struct {
	Enabled bool   `json:"enabled"`
	Code    string `json:"code"`
	Name    string `json:"name"`
}

type apiCurrenciesResponse struct {
	Data []apiCurrency `json:"data"`
}

const currenciesEndpoint = "%s/currencies?page=%d"

func (api *Api) UpdateCurrencies() error {
	currencies, err := api.ListCurrencies()
	if err != nil {
		return err
	}
	api.Currencies = currencies
	return nil
}

func (api *Api) ListCurrencies() ([]Currency, error) {
	allData, err := api.fetchPaginated("%s/currencies?page=%d", api.Config.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch paginated currencies: %v", err)
	}

	currs, err := unmarshalItems[apiCurrency](allData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal currencies: %v", err)
	}

	currencies := []Currency{}
	for _, cur := range currs {
		if !cur.Attributes.Enabled {
			continue
		}
		currency := Currency{
			ID:   cur.ID,
			Code: cur.Attributes.Code,
			Name: cur.Attributes.Name,
		}
		currencies = append(currencies, currency)
	}

	return currencies, nil
}

func (api *Api) GetCurrencyByCode(code string) Currency {
	for _, cur := range api.Currencies {
		if strings.EqualFold(cur.Code, code) {
			return cur
		}
	}
	return Currency{}
}
