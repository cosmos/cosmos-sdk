package grpc

import "math"

var ConfigTemplate = `
###############################################################################
###                           gRPC Configuration                            ###
###############################################################################

[grpc]

# Enable defines if the gRPC server should be enabled.
enable = {{ .GRPC.Enable }}

# Address defines the gRPC server address to bind to.
address = "{{ .GRPC.Address }}"

# MaxRecvMsgSize defines the max message size in bytes the server can receive.
# The default value is 10MB.
max-recv-msg-size = "{{ .GRPC.MaxRecvMsgSize }}"

# MaxSendMsgSize defines the max message size in bytes the server can send.
# The default value is math.MaxInt32.
max-send-msg-size = "{{ .GRPC.MaxSendMsgSize }}"
`

const (
	// DefaultGRPCAddress defines the default address to bind the gRPC server to.
	DefaultGRPCAddress = "localhost:9090"

	// DefaultGRPCMaxRecvMsgSize defines the default gRPC max message size in
	// bytes the server can receive.
	DefaultGRPCMaxRecvMsgSize = 1024 * 1024 * 10

	// DefaultGRPCMaxSendMsgSize defines the default gRPC max message size in
	// bytes the server can send.
	DefaultGRPCMaxSendMsgSize = math.MaxInt32
)

// GRPCConfig defines configuration for the gRPC server.
type Config struct {
	// Enable defines if the gRPC server should be enabled.
	Enable bool `mapstructure:"enable"`

	// Address defines the API server to listen on
	Address string `mapstructure:"address"`

	// MaxRecvMsgSize defines the max message size in bytes the server can receive.
	// The default value is 10MB.
	MaxRecvMsgSize int `mapstructure:"max-recv-msg-size"`

	// MaxSendMsgSize defines the max message size in bytes the server can send.
	// The default value is math.MaxInt32.
	MaxSendMsgSize int `mapstructure:"max-send-msg-size"`
}
