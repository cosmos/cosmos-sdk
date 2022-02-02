package ormdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type JSONSource interface {
	// JSONReader returns an io.ReadCloser for the named table. If there
	// is no JSON for this table, this method will return nil.
	JSONReader(tableName protoreflect.FullName) (io.ReadCloser, error)
}

type JSONSink interface {
	JSONWriter(tableName protoreflect.FullName) (io.WriteCloser, error)
}

func (m moduleDB) DefaultJSON(sink JSONSink) error {
	for name, table := range m.tablesByName {
		w, err := sink.JSONWriter(name)
		if err != nil {
			return err
		}

		_, err = w.Write(table.DefaultJSON())
		if err != nil {
			return err
		}
	}
	return nil
}

func (m moduleDB) ValidateJSON(source JSONSource) error {
	var errors map[protoreflect.FullName]error
	for name, table := range m.tablesByName {
		r, err := source.JSONReader(name)
		if err != nil {
			return err
		}

		err = table.ValidateJSON(r)
		if err != nil {
			errors[name] = err
		}

		err = r.Close()
		if err != nil {
			return err
		}
	}

	if len(errors) != 0 {
		panic("TODO")
	}
	return nil
}

func (m moduleDB) ImportJSON(ctx context.Context, source JSONSource) error {
	var names []string
	for name := range m.tablesByName {
		names = append(names, string(name))
	}
	sort.Strings(names)

	for _, name := range names {
		fullName := protoreflect.FullName(name)
		table := m.tablesByName[fullName]

		r, err := source.JSONReader(fullName)
		if err != nil {
			return err
		}

		if r == nil {
			continue
		}

		err = table.ImportJSON(ctx, r)
		if err != nil {
			return err
		}

		err = r.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (m moduleDB) ExportJSON(ctx context.Context, sink JSONSink) error {
	for name, table := range m.tablesByName {
		w, err := sink.JSONWriter(name)
		if err != nil {
			return err
		}

		err = table.ExportJSON(ctx, w)
		if err != nil {
			return err
		}

		err = w.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

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
	for s, message := range r.m {
		fmt.Printf("%s -> %s\n", s, message)
	}
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
