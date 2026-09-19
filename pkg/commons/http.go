package commons

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient struct {
	httpClient *http.Client
}

func NewHTTPClient(httpClient *http.Client) *HTTPClient {
	return &HTTPClient{httpClient: httpClient}
}

func (c *HTTPClient) Get(ctx context.Context, endpoint string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http request failed: status=%d body=%s", resp.StatusCode, string(body))
	}
	return body, nil
}
