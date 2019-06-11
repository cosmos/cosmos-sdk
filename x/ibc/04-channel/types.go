package channel

type State = byte

const (
	Idle State = iota
	Init
	OpenTry
	Open
	CloseTry
	Closed
)

/*
type Packet struct {
	Sequence      uint64
	TimeoutHeight uint64

	SourceConnection string
	SourceChannel    string
	DestConnection   string
	DestChannel      string

	Data []byte
}
*/

type Packet interface {
	Timeout() uint64
	Commit() []byte // Can be a commit message
}

type Channel struct {
	Module             string
	Counterparty       string
	CounterpartyModule string
}
