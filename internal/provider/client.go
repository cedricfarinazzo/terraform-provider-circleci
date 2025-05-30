package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// APIError represents an error response from the CircleCI API
type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code,omitempty"`
}

func (e APIError) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("CircleCI API error (code %d): %s", e.Code, e.Message)
	}
	return fmt.Sprintf("CircleCI API error: %s", e.Message)
}

// MakeRequest makes an HTTP request to the CircleCI API
func (c *CircleCIClient) MakeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var requestBody io.Reader

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		requestBody = bytes.NewBuffer(jsonData)
	}

	url := c.BaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Circle-Token", c.ApiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "terraform-provider-circleci")

	tflog.Debug(ctx, "Making CircleCI API request", map[string]interface{}{
		"method":   method,
		"url":      url,
		"endpoint": endpoint,
	})

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	// Check for API errors
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)

		var apiErr APIError
		if err := json.Unmarshal(bodyBytes, &apiErr); err != nil {
			// If we can't parse the error response, create a generic error
			apiErr = APIError{
				Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(bodyBytes)),
				Code:    resp.StatusCode,
			}
		}

		return nil, apiErr
	}

	return resp, nil
}

// Get makes a GET request to the CircleCI API
func (c *CircleCIClient) Get(ctx context.Context, endpoint string, result interface{}) error {
	resp, err := c.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}

	return nil
}

// Post makes a POST request to the CircleCI API
func (c *CircleCIClient) Post(ctx context.Context, endpoint string, body interface{}, result interface{}) error {
	resp, err := c.MakeRequest(ctx, "POST", endpoint, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}

	return nil
}

// Put makes a PUT request to the CircleCI API
func (c *CircleCIClient) Put(ctx context.Context, endpoint string, body interface{}, result interface{}) error {
	resp, err := c.MakeRequest(ctx, "PUT", endpoint, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}

	return nil
}

// Delete makes a DELETE request to the CircleCI API
func (c *CircleCIClient) Delete(ctx context.Context, endpoint string) error {
	resp, err := c.MakeRequest(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// BuildURL constructs a URL with query parameters
func BuildURL(endpoint string, params map[string]string) string {
	if len(params) == 0 {
		return endpoint
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}

	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Items         []interface{} `json:"items"`
	NextPageToken string        `json:"next_page_token,omitempty"`
}

// GetAllPages retrieves all pages of a paginated API response
func (c *CircleCIClient) GetAllPages(ctx context.Context, endpoint string, params map[string]string) ([]interface{}, error) {
	var allItems []interface{}
	nextPageToken := ""

	for {
		queryParams := make(map[string]string)
		for k, v := range params {
			queryParams[k] = v
		}

		if nextPageToken != "" {
			queryParams["page-token"] = nextPageToken
		}

		url := BuildURL(endpoint, queryParams)

		var response PaginatedResponse
		if err := c.Get(ctx, url, &response); err != nil {
			return nil, err
		}

		allItems = append(allItems, response.Items...)

		if response.NextPageToken == "" {
			break
		}
		nextPageToken = response.NextPageToken
	}

	return allItems, nil
}

// ParseID parses various ID formats used in CircleCI (UUID, slug, etc.)
func ParseID(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("id cannot be empty")
	}

	// Handle VCS slug format (e.g., "gh/owner/repo")
	if strings.Contains(id, "/") && len(strings.Split(id, "/")) >= 2 {
		return id, nil
	}

	// Handle UUID format
	if len(id) == 36 && strings.Count(id, "-") == 4 {
		return id, nil
	}

	// Handle simple string IDs
	return id, nil
}

// EscapeProjectSlug URL encodes a project slug for use in API endpoints
func EscapeProjectSlug(slug string) string {
	return url.PathEscape(slug)
}

// ValidateUUID checks if a string is a valid UUID
func ValidateUUID(id string) bool {
	return len(id) == 36 && strings.Count(id, "-") == 4
}

// ConvertBoolToString converts a boolean to string for API parameters
func ConvertBoolToString(b bool) string {
	return strconv.FormatBool(b)
}
