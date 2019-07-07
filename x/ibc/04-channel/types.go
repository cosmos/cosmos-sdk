package channel

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
	Port             string
	Counterparty     string
	CounterpartyPort string
}
