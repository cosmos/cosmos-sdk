package client

import "io"

// Export allows to export client configuration to the given reader
func Export(c *Client, w io.Reader) error {
	return nil
}

// Import instantiates a new *Client from an import
func Import(desc io.Reader) *Client {
	return nil
}
