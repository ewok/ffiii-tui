/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type APIResponse struct {
	Data    any    `json:"data"`
	Message string `json:"message,omitempty"`
	Meta    struct {
		Pagination struct {
			CurrentPage int `json:"current_page"`
			TotalPages  int `json:"total_pages"`
			Total       int `json:"total"`
		} `json:"pagination"`
	} `json:"meta"`
}

func (api *Api) httpClient() *http.Client {
	timeout := time.Duration(api.Config.TimeoutSeconds) * time.Second
	zap.L().Debug("Creating HTTP client",
		zap.Duration("timeout", timeout))

	return &http.Client{
		Timeout: timeout,
	}
}

func (api *Api) makeRequest(method, endpoint string, payload any, okStatus int) (*APIResponse, error) {
	if okStatus == 0 {
		okStatus = 200
	}

	startTime := time.Now()

	zap.L().Debug("Starting HTTP request",
		zap.String("method", method),
		zap.String("endpoint", endpoint),
		zap.Int("expected_status", okStatus),
		zap.Bool("has_payload", payload != nil))

	var body io.Reader
	var payloadSize int
	if payload != nil {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			zap.L().Error("Failed to marshal request payload",
				zap.Error(err),
				zap.String("method", method),
				zap.String("endpoint", endpoint))
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		payloadSize = len(payloadBytes)
		body = bytes.NewBuffer(payloadBytes)

		zap.L().Debug("Request payload prepared",
			zap.Int("payload_size_bytes", payloadSize),
			zap.String("endpoint", endpoint))
	}

	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		zap.L().Error("Failed to create HTTP request",
			zap.Error(err),
			zap.String("method", method),
			zap.String("endpoint", endpoint))
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set common headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Config.ApiKey))
	req.Header.Set("Content-Type", "application/json")

	zap.L().Debug("HTTP request headers set",
		zap.String("content_type", req.Header.Get("Content-Type")),
		zap.String("accept", req.Header.Get("Accept")),
		zap.Bool("has_auth", req.Header.Get("Authorization") != ""))

	resp, err := api.httpClient().Do(req)
	requestDuration := time.Since(startTime)

	if err != nil {
		zap.L().Error("HTTP request failed",
			zap.Error(err),
			zap.String("method", method),
			zap.String("endpoint", endpoint),
			zap.Duration("request_duration", requestDuration))
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			zap.L().Warn("Failed to close response body",
				zap.Error(closeErr),
				zap.String("endpoint", endpoint))
		}
	}()

	zap.L().Debug("HTTP response received",
		zap.String("method", method),
		zap.String("endpoint", endpoint),
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("request_duration", requestDuration),
		zap.String("content_type", resp.Header.Get("Content-Type")),
		zap.Int64("content_length", resp.ContentLength))

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error("Failed to read response body",
			zap.Error(err),
			zap.String("endpoint", endpoint),
			zap.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	responseSize := len(respBody)
	zap.S().Debugf("Response body read: %d bytes from %s", responseSize, endpoint)

	var apiResp APIResponse

	if resp.StatusCode != okStatus {
		zap.L().Warn("HTTP request returned unexpected status",
			zap.String("method", method),
			zap.String("endpoint", endpoint),
			zap.Int("actual_status", resp.StatusCode),
			zap.Int("expected_status", okStatus),
			zap.Duration("request_duration", requestDuration),
			zap.Int("response_size_bytes", responseSize))

		// Try to parse error response for more context
		if responseSize > 0 {
			var errorResp map[string]any
			if json.Unmarshal(respBody, &errorResp) == nil {
				if message, ok := errorResp["message"].(string); ok {
					zap.L().Error("API error details",
						zap.String("api_message", message),
						zap.String("endpoint", endpoint))
				}
			}
		}

		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	if okStatus != http.StatusNoContent {
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			zap.L().Error("Failed to unmarshal API response",
				zap.Error(err),
				zap.String("endpoint", endpoint),
				zap.Int("response_size_bytes", responseSize),
				zap.ByteString("response_preview", respBody[:min(len(respBody), 200)]))
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		if apiResp.Message != "" {
			zap.L().Error("API returned error message",
				zap.String("api_message", apiResp.Message),
				zap.String("endpoint", endpoint))
			return nil, fmt.Errorf("API error: %s", apiResp.Message)
		}

		// Log pagination info if available
		if apiResp.Meta.Pagination.Total > 0 {
			zap.L().Debug("API response with pagination",
				zap.String("endpoint", endpoint),
				zap.Int("current_page", apiResp.Meta.Pagination.CurrentPage),
				zap.Int("total_pages", apiResp.Meta.Pagination.TotalPages),
				zap.Int("total_items", apiResp.Meta.Pagination.Total))
		}
	}

	// zap.L().Info("HTTP request completed successfully",
	// 	zap.String("method", method),
	// 	zap.String("endpoint", endpoint),
	// 	zap.Int("status_code", resp.StatusCode),
	// 	zap.Duration("total_duration", time.Since(startTime)),
	// 	zap.Int("response_size_bytes", responseSize))

	return &apiResp, nil
}

func (api *Api) getRequest(endpoint string) (*APIResponse, error) {
	zap.L().Debug("Executing GET request", zap.String("endpoint", endpoint))
	return api.makeRequest("GET", endpoint, nil, http.StatusOK)
}

func (api *Api) postRequest(endpoint string, payload any) (*APIResponse, error) {
	zap.L().Debug("Executing POST request",
		zap.String("endpoint", endpoint),
		zap.Bool("has_payload", payload != nil))
	return api.makeRequest("POST", endpoint, payload, http.StatusOK)
}

func (api *Api) putRequest(endpoint string, payload any) (*APIResponse, error) {
	zap.L().Debug("Executing PUT request",
		zap.String("endpoint", endpoint),
		zap.Bool("has_payload", payload != nil))
	return api.makeRequest("PUT", endpoint, payload, http.StatusOK)
}

func (api *Api) deleteRequest(endpoint string) (*APIResponse, error) {
	zap.L().Debug("Executing DELETE request", zap.String("endpoint", endpoint))
	return api.makeRequest("DELETE", endpoint, nil, http.StatusNoContent)
}

func (api *Api) fetchPaginated(endpointTemplate string, args ...any) ([]any, error) {
	zap.L().Debug("Starting paginated fetch",
		zap.String("endpoint_template", endpointTemplate),
		zap.Int("args_count", len(args)))

	var allData []any
	page := 1
	totalItems := 0

	for {
		endpoint := fmt.Sprintf(endpointTemplate, append(args, page)...)

		zap.L().Debug("Fetching page",
			zap.Int("page", page),
			zap.String("endpoint", endpoint))

		resp, err := api.getRequest(endpoint)
		if err != nil {
			zap.L().Error("Failed to fetch page",
				zap.Error(err),
				zap.Int("page", page),
				zap.String("endpoint", endpoint))
			return nil, err
		}

		data, ok := resp.Data.([]any)
		if !ok {
			zap.L().Error("Invalid data format in paginated response",
				zap.Int("page", page),
				zap.String("endpoint", endpoint),
				zap.String("data_type", fmt.Sprintf("%T", resp.Data)))
			return nil, fmt.Errorf("invalid data format in response")
		}

		pageItemCount := len(data)
		zap.L().Debug("Page fetched successfully",
			zap.Int("page", page),
			zap.Int("items_on_page", pageItemCount),
			zap.Int("current_page", resp.Meta.Pagination.CurrentPage),
			zap.Int("total_pages", resp.Meta.Pagination.TotalPages))

		if pageItemCount == 0 {
			zap.L().Debug("Empty page received, ending pagination",
				zap.Int("page", page))
			break
		}

		allData = append(allData, data...)
		totalItems += pageItemCount

		if resp.Meta.Pagination.CurrentPage >= resp.Meta.Pagination.TotalPages {
			zap.L().Debug("Reached last page",
				zap.Int("current_page", resp.Meta.Pagination.CurrentPage),
				zap.Int("total_pages", resp.Meta.Pagination.TotalPages))
			break
		}

		page++

		// Safety check to prevent infinite loops
		if page > 1000 {
			zap.L().Warn("Pagination safety limit reached",
				zap.Int("max_pages", 1000),
				zap.String("endpoint_template", endpointTemplate))
			break
		}
	}

	// zap.L().Info("Paginated fetch completed",
	// 	zap.String("endpoint_template", endpointTemplate),
	// 	zap.Int("total_pages_fetched", page),
	// 	zap.Int("total_items", totalItems),
	// 	zap.Duration("total_duration", time.Since(startTime)))

	return allData, nil
}

func unmarshalItems[T any](items []any) ([]T, error) {
	itemCount := len(items)
	zap.L().Debug("Starting item unmarshaling",
		zap.Int("item_count", itemCount),
		zap.String("target_type", fmt.Sprintf("%T", *new(T))))

	result := make([]T, 0, itemCount)

	for i, item := range items {
		itemBytes, err := json.Marshal(item)
		if err != nil {
			zap.L().Error("Failed to marshal item for unmarshaling",
				zap.Error(err),
				zap.Int("item_index", i))
			return nil, fmt.Errorf("failed to marshal item: %w", err)
		}

		var typed T
		if err := json.Unmarshal(itemBytes, &typed); err != nil {
			zap.L().Error("Failed to unmarshal item to target type",
				zap.Error(err),
				zap.Int("item_index", i),
				zap.String("target_type", fmt.Sprintf("%T", typed)),
				zap.ByteString("item_data", itemBytes[:min(len(itemBytes), 100)]))
			return nil, fmt.Errorf("failed to unmarshal item: %w", err)
		}

		result = append(result, typed)
	}

	zap.L().Debug("Item unmarshaling completed",
		zap.Int("input_items", itemCount),
		zap.Int("successful_items", len(result)),
		zap.String("target_type", fmt.Sprintf("%T", *new(T))))

	return result, nil
}
