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
	//	Commit() []byte // Can be a commit message
	Route() string
}

type Channel struct {
	Port             string `json:"port"`
	Counterparty     string `json:"counterparty"`
	CounterpartyPort string `json:"counterparty_port"`
}
