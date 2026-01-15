/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"time"
	"unicode/utf8"

	"go.uber.org/zap"
)

type SummaryItem struct {
	Key                   string  `json:"key"`
	Title                 string  `json:"title"`
	MonetaryValue         float64 `json:"monetary_value,string"`
	CurrencyID            string  `json:"currency_id"`
	CurrencyCode          string  `json:"currency_code"`
	CurrencySymbol        string  `json:"currency_symbol,omitempty"`
	CurrencyDecimalPlaces int     `json:"currency_decimal_places"`
	NoAvailableBudgets    bool    `json:"no_available_budgets,omitempty"`
	ValueParsed           string  `json:"value_parsed"`
	LocalIcon             string  `json:"local_icon"`
	SubTitle              string  `json:"sub_title"`
}

func (api *Api) GetSummary() (map[string]SummaryItem, error) {
	endpoint := fmt.Sprintf("%s/summary/basic?start=%s&end=%s",
		api.Config.ApiUrl,
		api.StartDate.Format("2006-01-02"),
		api.EndDate.Format("2006-01-02"))

	zap.L().Debug("Fetching summary data",
		zap.String("endpoint", endpoint),
		zap.String("start_date", api.StartDate.Format("2006-01-02")),
		zap.String("end_date", api.EndDate.Format("2006-01-02")))

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		zap.L().Error("Failed to create HTTP request", zap.Error(err))
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Config.ApiKey))

	startTime := time.Now()
	client := &http.Client{Timeout: time.Duration(api.Config.TimeoutSeconds) * time.Second}
	resp, err := client.Do(req)
	requestDuration := time.Since(startTime)

	if err != nil {
		zap.L().Error("Failed to send HTTP request",
			zap.Error(err),
			zap.Duration("request_duration", requestDuration))
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	zap.L().Debug("Summary API request completed",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("request_duration", requestDuration),
		zap.String("content_type", resp.Header.Get("Content-Type")))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error("Failed to read response body", zap.Error(err))
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	zap.S().Debugf("Summary API response body size: %d bytes", len(body))

	if resp.StatusCode != http.StatusOK {
		zap.L().Warn("Summary API returned non-OK status",
			zap.Int("status_code", resp.StatusCode),
			zap.ByteString("response_body", body))

		var response map[string]any
		err = json.Unmarshal(body, &response)
		if err != nil {
			zap.L().Error("Failed to unmarshal error response", zap.Error(err))
			return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
		}

		message, ok := response["message"].(string)
		if ok && message != "" {
			zap.L().Error("Summary API error", zap.String("api_message", message))
			return nil, fmt.Errorf("API error: %s", message)
		}
		return nil, fmt.Errorf("failed status code : %d", resp.StatusCode)
	}

	var items map[string]SummaryItem
	err = json.Unmarshal(body, &items)
	if err != nil {
		zap.L().Error("Failed to unmarshal summary items",
			zap.Error(err),
			zap.ByteString("response_body", body))
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	for key, item := range items {
		zap.L().Debug("Summary item retrieved",
			zap.String("key", key),
			zap.String("title", item.Title),
			zap.Float64("monetary_value", item.MonetaryValue),
			zap.String("currency_code", item.CurrencyCode),
			zap.String("value_parsed", item.ValueParsed))
	}

	// zap.L().Info("Summary data retrieved successfully",
	// 	zap.Int("item_count", itemCount),
	// 	zap.Duration("total_duration", time.Since(startTime)))

	return items, nil
}

func (api *Api) UpdateSummary() error {
	summary, err := api.GetSummary()
	if err != nil {
		return fmt.Errorf("failed to get summary: %v", err)
	}
	api.Summary = summary
	return nil
}

// SummaryItems returns a shallow copy of cached summary items.
func (api *Api) SummaryItems() map[string]SummaryItem {
	return maps.Clone(api.Summary)
}

func (api *Api) GetMaxWidth() int {
	if len(api.Summary) < 1 {
		api.UpdateSummary()
	}
	maxLength := 0
	for _, s := range api.Summary {
		l := utf8.RuneCountInString(s.Title) + utf8.RuneCountInString(s.ValueParsed)
		if l > maxLength {
			maxLength = l
		}
	}
	return maxLength + 1
}
