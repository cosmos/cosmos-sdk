package connection

type State = byte

const (
	Idle State = iota
	Init
	Try
	Open
	CloseTry
	Closed
)

type Connection struct {
	Counterparty       string
	Client             string
	CounterpartyClient string
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
