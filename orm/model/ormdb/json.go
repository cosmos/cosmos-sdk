package ormdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"

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

func (r RawJSONSource) JSONReader(tableName protoreflect.FullName) (io.Reader, error) {
	return bytes.NewReader(r.m[string(tableName)]), nil
}

var _ JSONSource = RawJSONSource{}

type FSJSONSouce struct {
	fs fs.FS
}

func (F FSJSONSouce) JSONWriter(tableName protoreflect.FullName) (io.Writer, error) {
}

func (F FSJSONSouce) JSONReader(tableName protoreflect.FullName) (io.Reader, error) {
	return F.fs.Open(fmt.Sprintf("%s.json", tableName))
}

var _ JSONSource = FSJSONSouce{}
var _ JSONSink = FSJSONSouce{}
