package utils

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type JSONObject struct {
	Channel              channel.Channel  `json:"channel"`
	ChannelProof         commitment.Proof `json:"channel_proof,omitempty"`
	Available            bool             `json:"available"`
	AvailableProof       commitment.Proof `json:"available_proof,omitempty"`
	SequenceSend         uint64           `json:"sequence_send"`
	SequenceSendProof    commitment.Proof `json:"sequence_send_proof,omitempty"`
	SequenceReceive      uint64           `json:"sequence_receive"`
	SequenceReceiveProof commitment.Proof `json:"sequence_receive_proof,omitempty"`
	//	Kind                 string           `json:"kind"`
	//	KindProof            commitment.Proof `json:"kind_proof,omitempty"`
}

func NewJSONObject(
	channel channel.Channel, channelp commitment.Proof,
	avail bool, availp commitment.Proof,
	//	kind string, kindp commitment.Proof,
	seqsend uint64, seqsendp commitment.Proof,
	seqrecv uint64, seqrecvp commitment.Proof,
) JSONObject {
	return JSONObject{
		Channel:        channel,
		ChannelProof:   channelp,
		Available:      avail,
		AvailableProof: availp,
		//	Kind:           kind,
		//	KindProof:      kindp,
	}
}

type HandshakeJSONObject struct {
	JSONObject              `json:"channel"`
	State                   byte             `json:"state"`
	StateProof              commitment.Proof `json:"state_proof,omitempty"`
	CounterpartyClient      string           `json:"counterparty_client"`
	CounterpartyClientProof commitment.Proof `json:"counterparty_client_proof,omitempty"`
	NextTimeout             uint64           `json:"next_timeout"`
	NextTimeoutProof        commitment.Proof `json:"next_timeout_proof,omitempty"`
}

func NewHandshakeJSONObject(
	channel channel.Channel, channelp commitment.Proof,
	avail bool, availp commitment.Proof,
	//	kind string, kindp commitment.Proof,
	seqsend uint64, seqsendp commitment.Proof,
	seqrecv uint64, seqrecvp commitment.Proof,

	state byte, statep commitment.Proof,
	timeout uint64, timeoutp commitment.Proof,
) HandshakeJSONObject {
	return HandshakeJSONObject{
		JSONObject:       NewJSONObject(channel, channelp, avail, availp, seqsend, seqsendp, seqrecv, seqrecvp),
		State:            state,
		StateProof:       statep,
		NextTimeout:      timeout,
		NextTimeoutProof: timeoutp,
	}
}
