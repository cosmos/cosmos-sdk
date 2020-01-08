package relayer

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	clientExported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clientTypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connTypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	chanTypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// Chain represents the necessary data for connecting to and indentifying a chain and its counterparites
// TODO: See what existing abstractions we can reuse. Maybe this should be CLIContext and the
// Counterparties should be just a list of chain-ids?
type Chain struct {
	Context        context.CLIContext
	Counterparties []string
}

// SendMsgs sends the standard transactions to the individual chain
func (c Chain) SendMsgs(datagram []sdk.Msg) {
	// TODO: reuse CLI code and the data from Chain to send the signed transactions to the individual chains
	// TODO: figure out key management for the relayer
}

// LatestHeight uses the CLI utilities to pull the latest height from a given chain
func (c Chain) LatestHeight() uint64 {
	return 0
}

// LatestHeader returns the header to be used for client creation
func (c Chain) LatestHeader() clientExported.Header {
	return nil
}

// QueryConsensusState returns a consensus state for a given chain to be used as a
// client in another chain
func (c Chain) QueryConsensusState() clientTypes.ConsensusStateResponse {
	return clientTypes.ConsensusStateResponse{}
}

// GetConnectionsUsingClient gets any connections that exist between chain and counterparty
func (c Chain) GetConnectionsUsingClient(counterparty Chain) []connTypes.ConnectionEnd {
	return []connTypes.ConnectionEnd{}
}

// GetConnection returns the remote end of a given connection
func (c Chain) GetConnection(connectionID string) connTypes.ConnectionEnd {
	return connTypes.ConnectionEnd{}
}

// GetChannelsUsingConnections returns all channels associated with a given set of connections
func (c Chain) GetChannelsUsingConnections(connections []connTypes.ConnectionEnd) []chanTypes.Channel {
	return []chanTypes.Channel{}
}

// GetChannel returns the channel associated with a channelID
func (c Chain) GetChannel(channelID string) chanTypes.Channel {
	return chanTypes.Channel{}
}

// QueryTxs returns an array of transactions given a tag
func (c Chain) QueryTxs(height uint64, tag string) []auth.StdTx {
	return []auth.StdTx{}
}

// GetChain returns the chain associated with a chain-id from a []Chain
func GetChain(chainID string, chains []Chain) Chain {
	for _, chain := range chains {
		if chain.Context.ChainID == chainID {
			return chain
		}
	}
	return Chain{}
}
