/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
SPDX-License-Identifier: Apache-2.0
*/
package firefly

import (
	"fmt"
)

type Category struct {
	ID           string
	Name         string
	Notes        string
	CurrencyCode string
}

type apiCategory struct {
	Type       string          `json:"type"`
	ID         string          `json:"id"`
	Attributes apiCategoryAttr `json:"attributes"`
}

type apiCategoryAttr struct {
	Name         string `json:"name"`
	Notes        string `json:"notes"`
	CurrencyCode string `json:"primary_currency_code"`
}

func (api *Api) CreateCategory(name, notes string) error {
	endpoint := fmt.Sprintf("%s/categories", api.Config.ApiUrl)

	payload := map[string]any{
		"name":  name,
		"notes": notes,
	}

	response, err := api.postRequest(endpoint, payload)
	if err != nil {
		return err
	}
	data, ok := response.Data.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid response format: missing data field")
	}
	id, ok := data["id"].(string)
	if !ok || id == "" {
		return fmt.Errorf("invalid response format: missing category id")
	}

	return nil
}

func (api *Api) UpdateCategoriesInsights() error {
	// TODO: Need error reporting
	insights := make(map[string]categoryInsight)

	spentInsights, err := api.GetInsights("expense/category")
	if err == nil {
		for _, item := range spentInsights {
			insights[item.ID] = categoryInsight{
				Spent:  (-1) * item.DifferenceFloat,
				Earned: 0,
			}
		}
	}

	earnedInsights, err := api.GetInsights("income/category")
	if err == nil {
		for _, item := range earnedInsights {
			if val, ok := insights[item.ID]; ok {
				val.Earned = item.DifferenceFloat
				insights[item.ID] = val
			} else {
				insights[item.ID] = categoryInsight{
					Spent:  0,
					Earned: item.DifferenceFloat,
				}
			}
		}
	}

	api.categoryInsights = insights

	return nil
}

func (api *Api) UpdateCategories() error {
	categories, err := api.ListCategories()
	if err != nil {
		return err
	}
	api.Categories = categories

	api.UpdateCategoriesInsights()

	return nil
}

func (api *Api) ListCategories() ([]Category, error) {
	allData, err := api.fetchPaginated("%s/categories?page=%d", api.Config.ApiUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch paginated categories: %v", err)
	}

	cats, err := unmarshalItems[apiCategory](allData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal categories: %v", err)
	}

	categories := []Category{}
	for _, cat := range cats {
		categories = append(categories, Category{
			ID:           cat.ID,
			Name:         cat.Attributes.Name,
			Notes:        cat.Attributes.Notes,
			CurrencyCode: cat.Attributes.CurrencyCode,
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

// CategoriesList returns the cached categories.
// It returns a copy of the slice to avoid accidental mutation by callers.
func (api *Api) CategoriesList() []Category {
	return append([]Category(nil), api.Categories...)
}

// CategorySpent returns the cached spent amount for a category.
func (api *Api) CategorySpent(categoryID string) float64 {
	if insight, ok := api.categoryInsights[categoryID]; ok {
		return insight.Spent
	}
	return 0
}

// CategoryEarned returns the cached earned amount for a category.
func (api *Api) CategoryEarned(categoryID string) float64 {
	if insight, ok := api.categoryInsights[categoryID]; ok {
		return insight.Earned
	}
	return 0
}

func (c *Category) GetSpent(api *Api) float64 {
	return api.CategorySpent(c.ID)
}

func (c *Category) GetEarned(api *Api) float64 {
	return api.CategoryEarned(c.ID)
}

func (api *Api) GetTotalSpentEarnedCategories() (spent, earned float64) {
	for _, insight := range api.categoryInsights {
		spent += insight.Spent
		earned += insight.Earned
	}
	return
}

func (c *Category) IsEmpty() bool {
	return *c == Category{}
}
