package types

// query routes supported by the IBC channel Querier
const (
	QueryChannel = "channel"
)

// QueryChannelParams defines the params for the following queries:
// - 'custom/ibc/channel'
type QueryChannelParams struct {
	PortID    string
	ChannelID string
}

// NewQueryChannelParams creates a new QueryChannelParams instance
func NewQueryChannelParams(portID, channelID string) QueryChannelParams {
	return QueryChannelParams{
		PortID:    portID,
		ChannelID: channelID,
	}
}
