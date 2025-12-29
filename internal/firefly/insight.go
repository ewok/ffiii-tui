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
	"time"
)

type InsightItem struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Difference      string  `json:"difference"`
	DifferenceFloat float64 `json:"difference_float"`
	CurrencyID      string  `json:"currency_id"`
	CurrencyCode    string  `json:"currency_code"`
}

func (api *Api) GetInsights(ep string) ([]InsightItem, error) {
	endpoint := fmt.Sprintf(
		"%s/insight/%s?start=%s&end=%s",
		api.Config.ApiUrl,
		ep,
		api.StartDate.Format("2006-01-02"),
		api.EndDate.Format("2006-01-02"))

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

	var items []InsightItem
	err = json.Unmarshal(body, &items)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	return items, nil
}
