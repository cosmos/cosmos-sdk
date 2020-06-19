package testing

import (
	"strconv"
)

var (
	ChannelIDPrefix = "channelid"
)

// TestConnections is a testing helper struct to keep track of the connectionID, source clientID,
// and counterparty clientID used in creating and interacting with a connection.
type TestConnection struct {
	ID                   string
	ClientID             string
	CounterpartyClientID string
	Channels             []TestChannel
}

// AddTestChannel appends a new TestChannel which contains references to the port and channel ID
// used for channel creation and interaction. The channel id and port id format:
// channelid<index>
// portid<index>
func (conn *TestConnection) AddTestChannel() TestChannel {
	channel := conn.NextTestChannel()
	conn.Channels = append(conn.Channels, channel)
	return channel
}

// NextTestChannel returns the next test channel to be created on this connection
func (conn *TestConnection) NextTestChannel() TestChannel {
	// TODO: pass as arg so application developers can provide their port
	portID := "transfer"
	// TODO: come up with better naming scheme, will colide with mulitple client creations
	channelID := conn.ID + strconv.Itoa(len(conn.Channels))
	return TestChannel{
		PortID:    portID,
		ChannelID: channelID,
	}
}

// TestChannel is a testing helper struct to keep track of the portID and channelID
// used in creating and interacting with a channel.
type TestChannel struct {
	PortID    string
	ChannelID string
}
