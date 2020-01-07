package relayer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	clientExported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clientTypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connTypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	chanTypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	xferTypes "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

// Chain represents the necessary data for connecting to and indentifying a chain and its counterparites
// TODO: See what existing abstractions we can reuse. Maybe this should be CLIContext and the
// Counterparties should be just a list of chain-ids?
type Chain struct {
	ChainID        string         `json:"chain-id"`
	Endpoint       string         `json:"endpoint"`
	Counterparties []Chain        `json:"counterparties"`
	From           sdk.AccAddress `json:"from"`
}

// SubmitDatagrams sends the standard transactions to the individual chain
// NOTE: If we reused CLIContext here this would be done
func (c Chain) SubmitDatagrams(datagram []sdk.Msg) {
	// TODO: reuse CLI code and the data from Chain to send the signed transactions to the individual chains
	// TODO: figure out key management for the relayer
}

// LatestHeight uses the CLI utilities to pull the latest height from a given chain
// NOTE: If we reused CLIContext here this would be done
func (c Chain) LatestHeight() uint64 {
	return 0
}

// LatestHeader returns the header to be used for client creation
// NOTE: If we reused CLIContext here this would be done
func (c Chain) LatestHeader() clientExported.Header {
	return nil
}

// QueryConsensusState returns a consensus state for a given chain to be used as a
// client in another chain
// NOTE: If we reused CLIContext here this would be done
func (c Chain) QueryConsensusState() clientTypes.ConsensusStateResponse {
	return clientTypes.ConsensusStateResponse{}
}

// GetConnectionsUsingClient gets any connections that exist between chain and counterparty
// NOTE: If we reused CLIContext here this would be done
func (c Chain) GetConnectionsUsingClient(counterparty Chain) []connTypes.ConnectionEnd {
	return []connTypes.ConnectionEnd{}
}

// GetConnection returns the remote end of a given connection
// NOTE: If we reused CLIContext here this would be done
func (c Chain) GetConnection(connectionID string) connTypes.ConnectionEnd {
	return connTypes.ConnectionEnd{}
}

// GetChannelsUsingConnections returns all channels associated with a given set of connections
// NOTE: If we reused CLIContext here this would be done
func (c Chain) GetChannelsUsingConnections(connections []connTypes.ConnectionEnd) []chanTypes.Channel {
	return []chanTypes.Channel{}
}

// GetChannel returns the channel associated with a channelID
// NOTE: If we reused CLIContext here this would be done
func (c Chain) GetChannel(channelID string) chanTypes.Channel {
	return chanTypes.Channel{}
}

// QueryTxs returns an array of transactions given a tag
// NOTE: If we reused CLIContext here this would be done
func (c Chain) QueryTxs(height uint64, tag string) []auth.StdTx {
	return []auth.StdTx{}
}

// Relay implements the algorithm described in ICS18 (https://github.com/cosmos/ics/tree/master/spec/ics-018-relayer-algorithms)
func Relay(chains []Chain) {
	for _, chain := range chains {
		for _, cp := range chain.Counterparties {
			if cp.ChainID != chain.ChainID {
				datagrams := chain.PendingDatagrams(chain, cp)
				chain.SubmitDatagrams(datagrams[0])
				cp.SubmitDatagrams(datagrams[1])
			}
		}
	}
}

// PendingDatagrams returns the set of transactions that needs to be run to relay between
// `chain` and `counterparty`.
func (c Chain) PendingDatagrams(chain Chain, counterparty Chain) [][]sdk.Msg {
	localDatagrams := make([]sdk.Msg, 0)
	cpDatagrams := make([]sdk.Msg, 0)

	// ICS2 : Clients
	// Determine if light client needs to be updated on counterparty
	if counterparty.QueryConsensusState().ProofHeight < chain.LatestHeight() {
		cpDatagrams = append(cpDatagrams,
			clientTypes.NewMsgUpdateClient("client-id", chain.LatestHeader(), chain.From))
	}

	// Determine if light client needs to be updated locally
	if chain.QueryConsensusState().ProofHeight < counterparty.LatestHeight() {
		localDatagrams = append(localDatagrams,
			clientTypes.NewMsgUpdateClient("client-id", counterparty.LatestHeader(), counterparty.From))
	}

	// ICS3 : Connections
	// - Determine if any connection handshakes are in progress
	connections := chain.GetConnectionsUsingClient(counterparty)
	for _, localEnd := range connections {
		remoteEnd := counterparty.GetConnection(localEnd.Counterparty.ConnectionID)

		// Handshake has started locally (1 step done), relay `connOpenTry` to the remote end
		if localEnd.State == connTypes.INIT && remoteEnd.State == connTypes.UNINITIALIZED {
			// TODO: move to NewMsgOpenTry and apply correct args
			cpDatagrams = append(cpDatagrams, connTypes.MsgConnectionOpenTry{})
		}

		// Handshake has started on the other end (2 steps done), relay `connOpenAck` to the local end
		if localEnd.State == connTypes.INIT && remoteEnd.State == connTypes.TRYOPEN {
			// TODO: move to NewMsgOpenAck and apply correct args
			localDatagrams = append(localDatagrams, connTypes.MsgConnectionOpenAck{})
		}

		// Handshake has confirmed locally (3 steps done), relay `connOpenConfirm` to the remote end
		if localEnd.State == connTypes.OPEN && remoteEnd.State == connTypes.TRYOPEN {
			// TODO: move to NewMsgOpenConfirm and apply correct args
			cpDatagrams = append(cpDatagrams, connTypes.MsgConnectionOpenConfirm{})
		}
	}

	// ICS4 : Channels & Packets
	// - Determine if any channel handshakes are in progress
	// - Determine if any packets, acknowledgements, or timeouts need to be relayed
	channels := chain.GetChannelsUsingConnections(connections)
	for _, localEnd := range channels {
		remoteEnd := counterparty.GetChannel(localEnd.Counterparty.ChannelID)
		// Deal with handshakes in progress

		// Handshake has started locally (1 step done), relay `chanOpenTry` to the remote end
		if localEnd.State == chanTypes.INIT && remoteEnd.State == chanTypes.UNINITIALIZED {
			// TODO: move to NewMsgOpenTry and apply correct args
			cpDatagrams = append(cpDatagrams, chanTypes.MsgChannelOpenTry{})
		}

		// Handshake has started on the other end (2 steps done), relay `chanOpenAck` to the local end
		if localEnd.State == chanTypes.INIT && remoteEnd.State == chanTypes.TRYOPEN {
			// TODO: move to NewMsgOpenAck and apply correct args
			localDatagrams = append(localDatagrams, chanTypes.MsgChannelOpenAck{})
		}

		// Handshake has confirmed locally (3 steps done), relay `chanOpenConfirm` to the remote end
		if localEnd.State == chanTypes.OPEN && remoteEnd.State == chanTypes.TRYOPEN {
			// TODO: move to NewMsgOpenConfirm and apply correct args
			cpDatagrams = append(cpDatagrams, chanTypes.MsgChannelOpenConfirm{})
		}

		// Deal with packets
		// TODO: Once ADR15 is merged this section needs to be completed cc @mossid @fedekunze @cwgoes

		// First, scan logs for sent packets and relay all of them
		// TODO: This is currently incorrect and will change
		for _, tx := range chain.QueryTxs(chain.LatestHeight(), "type:transfer") {
			for _, msg := range tx.Msgs {
				if msg.Type() == "transfer" {
					cpDatagrams = append(cpDatagrams, xferTypes.MsgRecvPacket{})
				}
			}
		}

		// Then, scan logs for received packets and relay acknowledgements
		// TODO: This is currently incorrect and will change
		for _, tx := range chain.QueryTxs(chain.LatestHeight(), "type:recv_packet") {
			for _, msg := range tx.Msgs {
				if msg.Type() == "recv_packet" {
					cpDatagrams = append(cpDatagrams, xferTypes.MsgRecvPacket{})
				}
			}
		}
	}

	//   Return for pending datagrams
	return [][]sdk.Msg{localDatagrams, cpDatagrams}
}
