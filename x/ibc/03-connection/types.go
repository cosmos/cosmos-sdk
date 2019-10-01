package connection

import (
	/*
		"errors"
		"strings"
	*/
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type Connection struct {
	Client       string          `json:"client"`
	Counterparty string          `json:"counterParty"`
	Path         commitment.Path `json:"path"`
}

/*
func (conn Connection) MarshalAmino() (string, error) {
	return strings.Join([]string{conn.Client, conn.Counterparty}, "/"), nil
}

func (conn *Connection) UnmarshalAmino(text string) (err error) {
	fields := strings.Split(text, "/")
	if len(fields) < 2 {
		return errors.New("not enough number of fields")
	}
	conn.Client = fields[0]
	conn.Counterparty = fields[1]
	return nil
}
*/
var kinds = map[string]Kind{
	"handshake": Kind{true, true},
}

type Kind struct {
	Sendable   bool
	Receivable bool
}
