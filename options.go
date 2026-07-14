package veris

import (
	"net/http"
)

func WithHTTPClient(client http.Client) Option {
	return func(c *Client) {
		c.HTTPClient = client
	}
}

// Override the default base URL.
// Must follow proto://host[:port]
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}
