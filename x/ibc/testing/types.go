package testing

// TestConnections is a testing helper struct to keep track of the connectionID, source clientID,
// and counterparty clientID used in creating and interacting with a connection.
type TestConnection struct {
	ID                   string
	ClientID             string
	CounterpartyClientID string
}

// TestChannel is a testing helper struct to keep track of the portID and channelID
// used in creating and interacting with a channel.
type TestChannel struct {
	PortID    string
	ChannelID string
}
