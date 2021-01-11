package ibctesting

import (
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
)

// TestConnection is a testing helper struct to keep track of the connectionID, source clientID,
// counterparty clientID, and the next channel version used in creating and interacting with a
// connection.
type TestConnection struct {
	ID                   string
	ClientID             string
	CounterpartyClientID string
	NextChannelVersion   string
	Channels             []TestChannel
}

// FirstOrNextTestChannel returns the first test channel if it exists, otherwise it
// returns the next test channel to be created. This function is expected to be used
// when the caller does not know if the channel has or has not been created in app
// state, but would still like to refer to it to test existence or non-existence.
func (conn *TestConnection) FirstOrNextTestChannel(portID string) TestChannel {
	if len(conn.Channels) > 0 {
		return conn.Channels[0]
	}
	return TestChannel{
		PortID:               portID,
		ID:                   channeltypes.FormatChannelIdentifier(0),
		ClientID:             conn.ClientID,
		CounterpartyClientID: conn.CounterpartyClientID,
		Version:              conn.NextChannelVersion,
	}
}

// TestChannel is a testing helper struct to keep track of the portID and channelID
// used in creating and interacting with a channel. The clientID and counterparty
// client ID are also tracked to cut down on querying and argument passing.
type TestChannel struct {
	PortID               string
	ID                   string
	ClientID             string
	CounterpartyClientID string
	Version              string
}
