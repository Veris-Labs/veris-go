package veris

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const DefaultBaseURL = "https://api.verislabs.io"

type Option func(*Client)

type Client struct {
	BaseURL    string
	ApiKey     string
	HTTPClient http.Client
}

func NewClient(apiKey string, options ...Option) *Client {
	client := &Client{
		BaseURL:    DefaultBaseURL,
		ApiKey:     apiKey,
		HTTPClient: http.Client{Timeout: 5 * time.Second},
	}

	for _, option := range options {
		option(client)
	}

	return client
}

func (client *Client) post(ctx context.Context, endpoint string, query url.Values, request any, response any) error {
	return client.do(ctx, http.MethodPost, endpoint, query, request, response)
}

func (client *Client) get(ctx context.Context, endpoint string, query url.Values, response any) error {
	return client.do(ctx, http.MethodGet, endpoint, query, nil, response)
}

func (client *Client) do(ctx context.Context, method, endpoint string, query url.Values, request any, response any) error {
	var bodyReader io.Reader

	if request != nil {
		reqBytes, err := json.Marshal(request)
		if err != nil {
			return fmt.Errorf("error while marshalling request: %w", err)
		}

		bodyReader = bytes.NewReader(reqBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, client.BaseURL+endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("error while creating request: %w", err)
	}

	req.Header.Set("X-Api-Key", client.ApiKey)
	req.Header.Set("User-Agent", "veris-go")

	if request != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if len(query) > 0 {
		req.URL.RawQuery = query.Encode()
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error while reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResponse struct {
			Error string `json:"error"`
		}

		if err := json.Unmarshal(respBytes, &errorResponse); err != nil || errorResponse.Error == "" {
			return &APIError{
				StatusCode: resp.StatusCode,
			}
		}

		return &APIError{Message: errorResponse.Error, StatusCode: resp.StatusCode}
	}

	if response == nil {
		return nil
	}

	if err := json.Unmarshal(respBytes, response); err != nil {
		return fmt.Errorf("error while unmarshalling response: %w", err)
	}

	return nil
}
