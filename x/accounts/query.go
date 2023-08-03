package accounts

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"google.golang.org/protobuf/proto"
)

func (a Accounts[H]) Query(ctx context.Context, addr []byte, msg proto.Message) (proto.Message, error) {
	accountImpl, err := a.getAccountImpl(ctx, addr)
	if err != nil {
		return nil, err
	}
	ctx = a.createContext(ctx, nil, addr)
	resp, err := accountImpl.Query(ctx, msg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// TODO: remove me
func (a Accounts[H]) DumpState(ctx context.Context, addr []byte) ([]byte, error) {
	accountImpl, err := a.getAccountImpl(ctx, addr)
	if err != nil {
		return nil, err
	}
	ctx = a.createContext(ctx, nil, addr)
	state := map[string]*bytes.Buffer{}

	err = accountImpl.Schemas.State.ExportGenesis(ctx, func(field string) (io.WriteCloser, error) {
		buf := bytes.NewBuffer(nil)
		state[field] = buf
		return nopCloser{buf}, nil
	})
	if err != nil {
		return nil, err
	}

	resp := map[string]json.RawMessage{}
	for field, buf := range state {
		resp[field] = buf.Bytes()
	}

	return json.Marshal(resp)
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }
