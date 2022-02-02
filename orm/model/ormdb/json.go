package ormdb

import (
	"bytes"
	"encoding/json"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type RawJSONSource struct {
	m map[string]json.RawMessage
}

func NewRawJSONSource(message json.RawMessage) (*RawJSONSource, error) {
	var m map[string]json.RawMessage
	err := json.Unmarshal(message, &m)
	if err != nil {
		return nil, err
	}
	return &RawJSONSource{m}, err
}

func (r RawJSONSource) JSONReader(tableName protoreflect.FullName) (io.ReadCloser, error) {
	j, ok := r.m[string(tableName)]
	if !ok {
		return nil, nil
	}
	return readCloserWrapper{bytes.NewReader(j)}, nil
}

type readCloserWrapper struct {
	io.Reader
}

func (r readCloserWrapper) Close() error { return nil }

var _ JSONSource = RawJSONSource{}

type RawJSONSink struct {
	m map[string]json.RawMessage
}

func (r *RawJSONSink) JSONWriter(tableName protoreflect.FullName) (io.WriteCloser, error) {
	if r.m == nil {
		r.m = map[string]json.RawMessage{}
	}

	return &rawWriter{Buffer: &bytes.Buffer{}, sink: r, table: tableName}, nil
}

func (r *RawJSONSink) JSON() (json.RawMessage, error) {
	return json.Marshal(r.m)
}

type rawWriter struct {
	*bytes.Buffer
	table protoreflect.FullName
	sink  *RawJSONSink
}

func (r rawWriter) Close() error {
	r.sink.m[string(r.table)] = r.Buffer.Bytes()
	return nil
}

var _ JSONSink = &RawJSONSink{}

//type FSJSONSouce struct {
//	fs fs.FS
//}
//
//func (F FSJSONSouce) JSONWriter(tableName protoreflect.FullName) (io.Writer, error) {
//}
//
//func (F FSJSONSouce) JSONReader(tableName protoreflect.FullName) (io.Reader, error) {
//	return F.fs.Open(fmt.Sprintf("%s.json", tableName))
//}
//
//var _ JSONSource = FSJSONSouce{}
//var _ JSONSink = FSJSONSouce{}
