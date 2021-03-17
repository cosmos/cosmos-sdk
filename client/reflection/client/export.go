package client

import (
	"io"

	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
)

type exporter struct {
	FileRegistry           map[string][]byte             `json:"file_registry" yaml:"file_registry"`
	Services               []*reflection.QueryDescriptor `json:"services" yaml:"services"`
	ServiceDescriptors     map[string][]byte             `json:"service_descriptors" yaml:"service_descriptors"`
	MsgImplementers        map[string][]string           `json:"mgs_implementers" yaml:"msg_implementers"`
	ServiceMsgImplementers map[string][]string           `json:"service_msg_implementers" yaml:"service_msg_implementers"`
	TypeDescriptors        map[string][]byte             `json:"type_descriptors" yaml:"type_descriptors"`
	InterfaceImplementers  map[string][]string           `json:"interface_implementers" yaml:"interface_implementers"`
}

// Export allows to export client configuration to the given reader
func Export(c *Client, w io.Writer) error {
	return nil
}

// Import instantiates a new *Client from an import
func Import(desc io.Reader) (*Client, error) {
	return nil, nil
}
