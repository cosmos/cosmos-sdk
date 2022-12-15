package module

import (
	"bytes"
	"encoding/json"
	"io"

	"cosmossdk.io/core/appmodule"
)

func genesisSource(message json.RawMessage) (appmodule.GenesisSource, error) {
	var m map[string]json.RawMessage
	err := json.Unmarshal(message, &m)
	if err != nil {
		return nil, err
	}
	return func(field string) (io.ReadCloser, error) {
		j, ok := m[field]
		if !ok {
			return nil, nil
		}
		return readCloserWrapper{bytes.NewReader(j)}, nil
	}, nil
}

type readCloserWrapper struct {
	io.Reader
}

func (r readCloserWrapper) Close() error { return nil }

type genesisTarget struct {
	m map[string]json.RawMessage
}

func newGenesisTarget() *genesisTarget {
	return &genesisTarget{m: map[string]json.RawMessage{}}
}

func (r *genesisTarget) target() func(field string) (io.WriteCloser, error) {
	return func(field string) (io.WriteCloser, error) {
		if r.m == nil {
			r.m = map[string]json.RawMessage{}
		}

		return &genesisWriter{Buffer: &bytes.Buffer{}, sink: r, field: field}, nil
	}
}

func (r *genesisTarget) json() (json.RawMessage, error) {
	return json.MarshalIndent(r.m, "", "  ")
}

type genesisWriter struct {
	*bytes.Buffer
	field string
	sink  *genesisTarget
}

func (r genesisWriter) Close() error {
	r.sink.m[r.field] = r.Buffer.Bytes()
	return nil
}
