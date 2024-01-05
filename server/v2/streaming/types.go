package streaming

import (
	streamingabci "cosmossdk.io/server/v2/streaming/abci"
)

// StreamingManager is the struct that maintains a list of ABCIListeners and configuration settings.
type StreamingManager struct {
	// Listeners for hooking into the message processing of the server
	// and exposing the requests and responses to external consumers
	Listeners []streamingabci.Listener

	// StopNodeOnErr halts the node when ABCI streaming service listening results in an error.
	StopNodeOnErr bool
}
