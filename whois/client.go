package whois

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	DefaultTimeout = 30 * time.Second
)

type Client struct {
	HTTPClient *http.Client
	Timeout    time.Duration
}

var DefaultClient = NewClient(DefaultTimeout)

// Create a new Client with the specified timeout
func NewClient(timeout time.Duration) *Client {
	return &Client{
		Timeout: timeout,
	}
}

func (c *Client) Fetch(req *http.Request) (*Response, error) {
	ctx := context.Background()

	if c.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.Timeout)
		defer cancel()
	}

	return c.fetchHTTP(req.WithContext(ctx))
}

func (c *Client) fetchHTTP(req *http.Request) (*Response, error) {
	var res DomainResponse

	// If HTTPClient is nil, DefaultClient will be used
	hc := c.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}

	hres, err := hc.Do(req)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(hres.Body)
	if err := decoder.Decode(&res); err != nil {
		return nil, err
	}
	defer hres.Body.Close()

	if res.Code == "1" {
		return nil, fmt.Errorf("domain %s does not exist", res.DomainName)
	}

	return &Response{[]byte(fmt.Sprintf("Expiration Date: %s", res.ExpirationDate))}, nil
}
