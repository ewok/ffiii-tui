/*
Copyright © 2025-2026 Artur Taranchiev <artur.taranchiev@gmail.com>
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

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type apiUser struct {
	ID         string      `json:"id"`
	Attributes apiUserAttr `json:"attributes"`
}

type apiUserAttr struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (api *Api) GetCurrentUser() (string, error) {
	endpoint := fmt.Sprintf("%s/about/user", api.Config.ApiUrl)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/vnd.api+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.Config.ApiKey))

	client := &http.Client{Timeout: time.Duration(api.Config.TimeoutSeconds) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var response map[string]any
		err = json.Unmarshal(body, &response)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal response body: %v", err)
		}

		message, ok := response["message"].(string)
		if ok && message != "" {
			return "", fmt.Errorf("API error: %s", message)
		}
		return "", fmt.Errorf("failed status code : %d", resp.StatusCode)
	}

	var userResponse apiUser
	err = json.Unmarshal(body, &userResponse)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	user := User{
		ID:    userResponse.ID,
		Email: userResponse.Attributes.Email,
		Role:  userResponse.Attributes.Role,
	}

	return user.Email, nil
}
