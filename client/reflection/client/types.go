package client

import (
	"context"
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

type AccountInfoProvider interface {
	Address(ctx context.Context, pubKey cryptotypes.PubKey) (string, error)
	SigningInfo(ctx context.Context, pubKey cryptotypes.PubKey) (accountNumber, sequence uint64, err error)
	Sign(ctx context.Context, pubKey cryptotypes.PubKey, b []byte) (signedBytes []byte, err error)
}

type Query struct {
	Service  string
	Method   string
	Request  string
	Response string
}

func (q Query) String() string {
	return fmt.Sprintf("service: %s method: %s request: %s response: %s", q.Service, q.Method, q.Request, q.Response)
}

type Deliverable struct {
	MsgName string
}

func (d Deliverable) String() string {
	return fmt.Sprintf("deliverable: %s", d.MsgName)
}
