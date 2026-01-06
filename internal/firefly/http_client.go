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
	return &http.Client{
		Timeout: time.Duration(api.Config.TimeoutSeconds) * time.Second,
	}
}

func (api *Api) makeRequest(method, endpoint string, payload any, okStatus int) (*APIResponse, error) {
	if okStatus == 0 {
		okStatus = 200
	}
	var body io.Reader
	if payload != nil {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(payloadBytes)
	}

	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set common headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Config.ApiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp APIResponse

	if resp.StatusCode != okStatus {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	if okStatus != http.StatusNoContent {
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		if apiResp.Message != "" {
			return nil, fmt.Errorf("API error: %s", apiResp.Message)
		}
	}

	return &apiResp, nil
}

func (api *Api) getRequest(endpoint string) (*APIResponse, error) {
	return api.makeRequest("GET", endpoint, nil, http.StatusOK)
}

func (api *Api) postRequest(endpoint string, payload any) (*APIResponse, error) {
	return api.makeRequest("POST", endpoint, payload, http.StatusOK)
}

func (api *Api) putRequest(endpoint string, payload any) (*APIResponse, error) {
	return api.makeRequest("PUT", endpoint, payload, http.StatusOK)
}

func (api *Api) deleteRequest(endpoint string) (*APIResponse, error) {
	return api.makeRequest("DELETE", endpoint, nil, http.StatusNoContent)
}

func (api *Api) fetchPaginated(endpointTemplate string, args ...any) ([]any, error) {
	var allData []any
	page := 1

	for {
		endpoint := fmt.Sprintf(endpointTemplate, append(args, page)...)
		resp, err := api.getRequest(endpoint)
		if err != nil {
			return nil, err
		}

		data, ok := resp.Data.([]any)
		if !ok {
			return nil, fmt.Errorf("invalid data format in response")
		}

		if len(data) == 0 {
			break
		}

		allData = append(allData, data...)

		if resp.Meta.Pagination.CurrentPage >= resp.Meta.Pagination.TotalPages {
			break
		}

		page++
	}

	return allData, nil
}

func unmarshalItems[T any](items []any) ([]T, error) {
	result := make([]T, 0, len(items))

	for _, item := range items {
		itemBytes, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal item: %w", err)
		}

		var typed T
		if err := json.Unmarshal(itemBytes, &typed); err != nil {
			return nil, fmt.Errorf("failed to unmarshal item: %w", err)
		}

		result = append(result, typed)
	}

	return result, nil
}
