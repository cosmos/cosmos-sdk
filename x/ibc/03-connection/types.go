package connection

import (
	"errors"
	"strings"
)

type Connection struct {
	Client       string
	Counterparty string
}

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
