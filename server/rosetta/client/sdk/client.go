package sdk

import (
	"fmt"
)

type Client struct {
	endpoint string
}

// NewClient returns the client to call Cosmos RPC.
func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
	}
}

func (c Client) getEndpoint(path string) string {
	return fmt.Sprintf("%s%s", c.endpoint, path)
}
