/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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

func NewCurrency(code string) Currency {
	return Currency{Code: strings.ToUpper(code)}
}

// {
//   "data": [
//     {
//       "id": "2",
//       "attributes": {
//         "enabled": true,
//         "code": "AMS",
//         "name": "Ankh-Morpork dollar",
//       }
//     }
//   ]
// }

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
	currencies := []Currency{}
	page := 1

	for {
		cursPage, err := api.listCurrencies(page)
		if err != nil {
			return nil, err
		}
		if len(cursPage) == 0 {
			break
		}
		currencies = append(currencies, cursPage...)
		page++
	}

	return currencies, nil
}

func (api *Api) listCurrencies(page int) ([]Currency, error) {
	endpoint := fmt.Sprintf(currenciesEndpoint, api.Config.ApiUrl, page)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Config.ApiKey))

	client := &http.Client{Timeout: time.Duration(api.Config.TimeoutSeconds) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var response map[string]any
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
		}

		message, ok := response["message"].(string)
		if ok && message != "" {
			return nil, fmt.Errorf("API error: %s", message)
		}
		return nil, fmt.Errorf("failed status code : %d", resp.StatusCode)
	}

	var apiResp apiCurrenciesResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	currencies := []Currency{}
	for _, apiCur := range apiResp.Data {
		if !apiCur.Attributes.Enabled {
			continue
		}
		currency := Currency{
			ID:   apiCur.ID,
			Code: apiCur.Attributes.Code,
			Name: apiCur.Attributes.Name,
		}
		currencies = append(currencies, currency)
	}

	return currencies, nil
}
