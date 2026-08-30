package pocketid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultTimeout = 15 * time.Second

// Client talks to the Pocket ID Admin API using an X-API-KEY header.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// UserCreate is the body for POST /api/users.
type UserCreate struct {
	Username      string `json:"username"`
	Email         string `json:"email,omitempty"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName,omitempty"`
	DisplayName   string `json:"displayName"`
	IsAdmin       bool   `json:"isAdmin"`
	Disabled      bool   `json:"disabled,omitempty"`
	EmailVerified bool   `json:"emailVerified,omitempty"`
}

// User is a Pocket ID user record.
type User struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email,omitempty"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName,omitempty"`
	DisplayName string `json:"displayName"`
	IsAdmin     bool   `json:"isAdmin"`
}

// SignupToken is returned by signup-token endpoints.
type SignupToken struct {
	ID         string `json:"id"`
	Token      string `json:"token"`
	ExpiresAt  string `json:"expiresAt"`
	UsageLimit int    `json:"usageLimit"`
	UsageCount int    `json:"usageCount"`
	CreatedAt  string `json:"createdAt"`
}

// NewClient builds a client for the given base URL and API key.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		APIKey:  strings.TrimSpace(apiKey),
		HTTPClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// SetupURL builds a one-time access link for a token.
func (c *Client) SetupURL(token string) string {
	return c.BaseURL + "/lc/" + url.PathEscape(token)
}

// CreateUser creates a user via POST /api/users.
func (c *Client) CreateUser(ctx context.Context, u UserCreate) (User, error) {
	var out User
	if err := c.doJSON(ctx, http.MethodPost, "/api/users", u, &out); err != nil {
		return User{}, err
	}
	return out, nil
}

// DeleteUser removes a user via DELETE /api/users/{id}.
func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("pocket-id: user id is required")
	}
	return c.doJSON(ctx, http.MethodDelete, "/api/users/"+url.PathEscape(userID), nil, nil)
}

// UpdateUserGroups sets group membership for a user.
func (c *Client) UpdateUserGroups(ctx context.Context, userID string, groupIDs []string) error {
	body := map[string]any{"userGroupIds": groupIDs}
	return c.doJSON(ctx, http.MethodPut, "/api/users/"+url.PathEscape(userID)+"/user-groups", body, nil)
}

// CreateOneTimeAccessToken creates an OTA and returns the raw token string.
func (c *Client) CreateOneTimeAccessToken(ctx context.Context, userID, ttl string) (string, error) {
	body := map[string]any{}
	if ttl != "" {
		body["ttl"] = ttl
	}
	var out struct {
		Token string `json:"token"`
	}
	path := "/api/users/" + url.PathEscape(userID) + "/one-time-access-token"
	if err := c.doJSON(ctx, http.MethodPost, path, body, &out); err != nil {
		return "", err
	}
	if out.Token == "" {
		return "", fmt.Errorf("pocket-id: empty one-time access token")
	}
	return out.Token, nil
}

// RequestOneTimeAccessEmail asks Pocket ID to email a one-time access link.
func (c *Client) RequestOneTimeAccessEmail(ctx context.Context, userID, ttl string) error {
	body := map[string]any{}
	if ttl != "" {
		body["ttl"] = ttl
	}
	path := "/api/users/" + url.PathEscape(userID) + "/one-time-access-email"
	return c.doJSON(ctx, http.MethodPost, path, body, nil)
}

// CreateSignupToken creates a signup token.
func (c *Client) CreateSignupToken(ctx context.Context, ttl string, usageLimit int, groupIDs []string) (SignupToken, error) {
	body := map[string]any{
		"ttl":          ttl,
		"usageLimit":   usageLimit,
		"userGroupIds": groupIDs,
	}
	var out SignupToken
	if err := c.doJSON(ctx, http.MethodPost, "/api/signup-tokens", body, &out); err != nil {
		return SignupToken{}, err
	}
	return out, nil
}

// ListSignupTokens lists signup tokens (first page).
func (c *Client) ListSignupTokens(ctx context.Context) ([]SignupToken, error) {
	var out struct {
		Data []SignupToken `json:"data"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/api/signup-tokens", nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// DeleteSignupToken deletes a signup token by id.
func (c *Client) DeleteSignupToken(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/api/signup-tokens/"+url.PathEscape(id), nil, nil)
}

// ListUsers fetches a small page of users to verify connectivity.
func (c *Client) ListUsers(ctx context.Context) error {
	return c.doJSON(ctx, http.MethodGet, "/api/users?pagination[page]=1&pagination[limit]=1", nil, nil)
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any, out any) error {
	if c.BaseURL == "" {
		return fmt.Errorf("pocket-id: base url is required")
	}
	if c.APIKey == "" {
		return fmt.Errorf("pocket-id: api key is required")
	}
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-KEY", c.APIKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		msg := strings.TrimSpace(string(data))
		if msg == "" {
			msg = res.Status
		}
		return fmt.Errorf("pocket-id: %s", msg)
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}
