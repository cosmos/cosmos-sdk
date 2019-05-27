package connection

type State = byte

const (
	Idle State = iota
	Init
	TryOpen
	CloseTry
	Open
	Closed
)

type Connection struct {
	Counterparty       string
	Client             string
	CounterpartyClient string
	NextTimeoutHeight  uint64
}
