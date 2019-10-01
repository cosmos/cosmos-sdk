package utils

import (
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type JSONState struct {
	Connection      connection.Connection `json:"connection"`
	ConnectionProof commitment.Proof      `json:"connection_proof,omitempty"`
	Available       bool                  `json:"available"`
	AvailableProof  commitment.Proof      `json:"available_proof,omitempty"`
	Kind            string                `json:"kind"`
	KindProof       commitment.Proof      `json:"kind_proof,omitempty"`

	State                   byte             `json:"state,omitempty"`
	StateProof              commitment.Proof `json:"state_proof,omitempty"`
	CounterpartyClient      string           `json:"counterparty_client,omitempty"`
	CounterpartyClientProof commitment.Proof `json:"counterparty_client_proof,omitempty"`
}

func NewJSONState(
	conn connection.Connection, connp commitment.Proof,
	avail bool, availp commitment.Proof,
	kind string, kindp commitment.Proof,
) JSONState {
	return JSONState{
		Connection:      conn,
		ConnectionProof: connp,
		Available:       avail,
		AvailableProof:  availp,
		Kind:            kind,
		KindProof:       kindp,
	}
}

func NewHandshakeJSONState(
	conn connection.Connection, connp commitment.Proof,
	avail bool, availp commitment.Proof,
	kind string, kindp commitment.Proof,
	state byte, statep commitment.Proof,
	cpclient string, cpclientp commitment.Proof,
) JSONState {
	return JSONState{
		Connection:      conn,
		ConnectionProof: connp,
		Available:       avail,
		AvailableProof:  availp,
		Kind:            kind,
		KindProof:       kindp,

		State:                   state,
		StateProof:              statep,
		CounterpartyClient:      cpclient,
		CounterpartyClientProof: cpclientp,
	}
}
