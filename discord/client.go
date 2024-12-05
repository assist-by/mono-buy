package discord

import (
	"net/http"
	"time"
)

type Client struct {
	webhookURL string
	httpClient *http.Client
}

// NewClient creates a new Discord webhook client
func NewClient(webhookURL string) *Client {
	return &Client{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}
