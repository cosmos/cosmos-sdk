package tendermint

import (
	"fmt"
)

type Client struct {
	endpoint string
}

func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
	}
}

func (c Client) getEndpoint(path string) string {
	return fmt.Sprintf("%s/%s", c.endpoint, path)
}
