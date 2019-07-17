package utils

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type JSONObject struct {
	Connection      connection.Connection `json:"connection"`
	ConnectionProof commitment.Proof      `json:"connection_proof,omitempty"`
	Available       bool                  `json:"available"`
	AvailableProof  commitment.Proof      `json:"available_proof,omitempty"`
	Kind            string                `json:"kind"`
	KindProof       commitment.Proof      `json:"kind_proof,omitempty"`
}

func NewJSONObject(
	conn connection.Connection, connp commitment.Proof,
	avail bool, availp commitment.Proof,
	kind string, kindp commitment.Proof,
) JSONObject {
	return JSONObject{
		Connection:      conn,
		ConnectionProof: connp,
		Available:       avail,
		AvailableProof:  availp,
		Kind:            kind,
		KindProof:       kindp,
	}
}

type HandshakeJSONObject struct {
	JSONObject              `json:"connection"`
	State                   byte             `json:"state"`
	StateProof              commitment.Proof `json:"state_proof,omitempty"`
	CounterpartyClient      string           `json:"counterparty_client"`
	CounterpartyClientProof commitment.Proof `json:"counterparty_client_proof,omitempty"`
	NextTimeout             uint64           `json:"next_timeout"`
	NextTimeoutProof        commitment.Proof `json:"next_timeout_proof,omitempty"`
}

func NewHandshakeJSONObject(
	conn connection.Connection, connp commitment.Proof,
	avail bool, availp commitment.Proof,
	kind string, kindp commitment.Proof,
	state byte, statep commitment.Proof,
	cpclient string, cpclientp commitment.Proof,
	timeout uint64, timeoutp commitment.Proof,
) HandshakeJSONObject {
	return HandshakeJSONObject{
		JSONObject:              NewJSONObject(conn, connp, avail, availp, kind, kindp),
		State:                   state,
		StateProof:              statep,
		CounterpartyClient:      cpclient,
		CounterpartyClientProof: cpclientp,
		NextTimeout:             timeout,
		NextTimeoutProof:        timeoutp,
	}
}
