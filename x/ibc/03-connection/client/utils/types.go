package utils

import (
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type JSONState struct {
	Connection      connection.Connection `json:"connection"`
	ConnectionProof ics23.Proof           `json:"connection_proof,omitempty"`
	Available       bool                  `json:"available"`
	AvailableProof  ics23.Proof           `json:"available_proof,omitempty"`
	Kind            string                `json:"kind"`
	KindProof       ics23.Proof           `json:"kind_proof,omitempty"`

	State                   byte        `json:"state,omitempty"`
	StateProof              ics23.Proof `json:"state_proof,omitempty"`
	CounterpartyClient      string      `json:"counterparty_client,omitempty"`
	CounterpartyClientProof ics23.Proof `json:"counterparty_client_proof,omitempty"`
}

func NewJSONState(
	conn connection.Connection, connp ics23.Proof,
	avail bool, availp ics23.Proof,
	kind string, kindp ics23.Proof,
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
	conn connection.Connection, connp ics23.Proof,
	avail bool, availp ics23.Proof,
	kind string, kindp ics23.Proof,
	state byte, statep ics23.Proof,
	cpclient string, cpclientp ics23.Proof,
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
