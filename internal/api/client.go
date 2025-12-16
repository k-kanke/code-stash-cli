package api

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

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken  string  `json:"access_token"`
	TokenType    string  `json:"token_type"`
	ExpiresIn    int     `json:"expires_in"`
	RefreshToken *string `json:"refresh_token"`
	Scope        string  `json:"scope"`
}

type APIError struct {
	Code        string `json:"error"`
	Description string `json:"error_description"`
}

type Client struct {
	baseURL      *url.URL
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func NewClient(baseURL, clientID, clientSecret string) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("api base url is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid api base url: %w", err)
	}

	return &Client{
		baseURL:      parsed,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}, nil
}

func (c *Client) StartDeviceCode(ctx context.Context) (*DeviceCodeResponse, error) {
	payload := map[string]string{
		"client_id": c.clientID,
	}
	var resp DeviceCodeResponse
	if err := c.post(ctx, "/oauth/device/code", payload, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ExchangeDeviceCode(ctx context.Context, deviceCode string) (*TokenResponse, *APIError, error) {
	payload := map[string]string{
		"grant_type":    "device_code",
		"device_code":   deviceCode,
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	}

	req, err := c.newRequest(ctx, "POST", "/oauth/token", payload)
	if err != nil {
		return nil, nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		var body TokenResponse
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			return nil, nil, fmt.Errorf("decode token response: %w", err)
		}
		return &body, nil, nil
	}

	apiErr, err := decodeAPIError(res.Body)
	if err != nil {
		return nil, nil, err
	}
	return nil, apiErr, nil
}

func (c *Client) post(ctx context.Context, target string, payload any, out any) error {
	req, err := c.newRequest(ctx, "POST", target, payload)
	if err != nil {
		return err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		apiErr, err := decodeAPIError(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("api error: %s", apiErr.Code)
	}

	if out == nil {
		return nil
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) newRequest(ctx context.Context, method, endpoint string, payload any) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode payload: %w", err)
	}

	u := *c.baseURL
	u.Path = joinPath(c.baseURL.Path, endpoint)

	req, err := http.NewRequestWithContext(ctx, method, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func joinPath(basePath, endpoint string) string {
	if endpoint == "" {
		return basePath
	}
	base := strings.TrimSuffix(basePath, "/")
	if strings.HasPrefix(endpoint, "/") {
		return base + endpoint
	}
	return base + "/" + endpoint
}

func decodeAPIError(r io.Reader) (*APIError, error) {
	var apiErr APIError
	if err := json.NewDecoder(r).Decode(&apiErr); err != nil {
		return nil, fmt.Errorf("decode api error: %w", err)
	}
	if apiErr.Code == "" {
		apiErr.Code = "unknown_error"
	}
	return &apiErr, nil
}
