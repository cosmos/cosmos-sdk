package handshake

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
)

type State = byte

const (
	Idle State = iota
	Init
	OpenTry
	Open
	CloseTry
	Closed
)

var _ connection.Connection = Connection{}

// TODO: Connection is amino marshaled currently, need to implement MarshalBinary manually
type Connection struct {
	Counterparty       string
	Client             string
	CounterpartyClient string
	State              State
}

func (conn Connection) GetCounterparty() string {
	return conn.Counterparty
}

func (conn Connection) GetClient() string {
	return conn.Client
}

func (conn Connection) GetCounterpartyClient() string {
	return conn.CounterpartyClient
}

func (conn Connection) Available() bool {
	return conn.State == Open
}

func (conn Connection) Equal(conn0 Connection) bool {
	return conn.Counterparty == conn0.Counterparty &&
		conn.Client == conn0.Client &&
		conn.CounterpartyClient == conn0.CounterpartyClient
}

func (conn Connection) Symmetric(id string, conn0 Connection) bool {
	return conn0.Equal(conn.Symmetry(id))
}

func (conn Connection) Symmetry(id string) Connection {
	return Connection{
		Counterparty:       id,
		Client:             conn.CounterpartyClient,
		CounterpartyClient: conn.Client,
	}
}
