package otx

import (
	"context"

	"github.com/vijayvenkatj/cti-miner/pkg/commons"
)

const apiKeyHeader = "X-OTX-API-KEY"

type Client struct {
	apiKey string
	http   *commons.HTTPClient
}

func NewClient(apiKey string, httpClient *commons.HTTPClient) *Client {
	return &Client{apiKey: apiKey, http: httpClient}
}

func (c *Client) Get(ctx context.Context, endpoint string) ([]byte, error) {
	return c.http.Get(ctx, endpoint, map[string]string{apiKeyHeader: c.apiKey})
}
