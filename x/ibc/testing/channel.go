package testing

// Channel is a testing helper struct to keep track of the portID and channelID used in creating
// and interacting with a channel.
type Channel struct {
	PortID    string
	ChannelID string
}

// NewChannel returns a new channel instance.
func NewChannel(portID, channelID string) Channel {
	return Channel{
		PortID:    portID,
		ChannelID: channelID,
	}
}
