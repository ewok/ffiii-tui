/*
Copyright © 2025 Artur Taranchiev <artur.taranchiev@gmail.com>
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

type Category struct {
	ID    string
	Name  string
	Notes string
}

type apiCategory struct {
	Type       string          `json:"type"`
	ID         string          `json:"id"`
	Attributes apiCategoryAttr `json:"attributes"`
}

type apiCategoryAttr struct {
	Name  string `json:"name"`
	Notes string `json:"notes"`
}

type apiCategoriesResponse struct {
	Data []apiCategory `json:"data"`
}

const categoriesEndpoint = "%s/categories?page=%d"

func (api *Api) CreateCategory(name, notes string) error {
	endpoint := fmt.Sprintf("%s/categories", api.Config.ApiUrl)

	payload := map[string]any{
		"name":  name,
		"notes": notes,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Config.ApiKey))

	client := &http.Client{Timeout: time.Duration(api.Config.TimeoutSeconds) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// TODO: Make nice error handling
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}
	// if returned .data.id is present then is good
	var response map[string]any
	err = json.Unmarshal(body, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		message, ok := response["message"].(string)
		if ok && message != "" {
			return fmt.Errorf("API error: %s", message)
		}
		return fmt.Errorf("failed status code : %d", resp.StatusCode)
	}

	data, ok := response["data"].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid response format: missing data field")
	}

	id, ok := data["id"].(string)
	if !ok || id == "" {
		return fmt.Errorf("invalid response format: missing transaction id")
	}

	return nil
}

func (api *Api) UpdateCategories() error {
	categories, err := api.ListCategories()
	if err != nil {
		return err
	}
	api.Categories = categories
	return nil
}

func (api *Api) ListCategories() ([]Category, error) {
	categories := []Category{}
	page := 1

	for {
		catsPage, err := api.listCategories(page)
		if err != nil {
			return nil, err
		}
		if len(catsPage) == 0 {
			break
		}
		categories = append(categories, catsPage...)
		page++
	}

	return categories, nil
}

func (api *Api) listCategories(page int) ([]Category, error) {
	endpoint := fmt.Sprintf(categoriesEndpoint, api.Config.ApiUrl, page)

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

	var apiResp apiCategoriesResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	categories := []Category{}
	for _, apiCat := range apiResp.Data {
		categories = append(categories, Category{
			ID:    apiCat.ID,
			Name:  apiCat.Attributes.Name,
			Notes: apiCat.Attributes.Notes,
		})
	}

	return categories, nil
}

func (api *Api) GetCategoryByName(name string) Category {
	for _, category := range api.Categories {
		if category.Name == name {
			return category
		}
	}
	return Category{}
}

func (api *Api) GetCategoryByID(ID string) Category {
	for _, category := range api.Categories {
		if category.ID == ID {
			return category
		}
	}
	return Category{}
}
