package grpcgateway

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

func TestMatchURI(t *testing.T) {
	testCases := []struct {
		name     string
		uri      string
		mapping  map[string]string
		expected *uriMatch
	}{
		{
			name:     "simple match, no wildcards",
			uri:      "/foo/bar",
			mapping:  map[string]string{"/foo/bar": "bar"},
			expected: &uriMatch{QueryInputName: "bar"},
		},
		{
			name:    "wildcard match at the end",
			uri:     "/foo/bar/buzz",
			mapping: map[string]string{"/foo/bar/{baz}": "bar"},
			expected: &uriMatch{
				QueryInputName: "bar",
				Params:         map[string]string{"baz": "buzz"},
			},
		},
		{
			name:    "wildcard match in the middle",
			uri:     "/foo/buzz/bar",
			mapping: map[string]string{"/foo/{baz}/bar": "bar"},
			expected: &uriMatch{
				QueryInputName: "bar",
				Params:         map[string]string{"baz": "buzz"},
			},
		},
		{
			name:    "multiple wild cards",
			uri:     "/foo/bar/baz/buzz",
			mapping: map[string]string{"/foo/bar/{q1}/{q2}": "bar"},
			expected: &uriMatch{
				QueryInputName: "bar",
				Params:         map[string]string{"q1": "baz", "q2": "buzz"},
			},
		},
		{
			name:    "catch-all wildcard",
			uri:     "/foo/bar/ibc/token/stuff",
			mapping: map[string]string{"/foo/bar/{ibc_token=**}": "bar"},
			expected: &uriMatch{
				QueryInputName: "bar",
				Params:         map[string]string{"ibc_token": "ibc/token/stuff"},
			},
		},
		{
			name:     "no match should return nil",
			uri:      "/foo/bar",
			mapping:  map[string]string{"/bar/foo": "bar"},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := matchURI(tc.uri, tc.mapping)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestURIMatch_HasParams(t *testing.T) {
	u := uriMatch{Params: map[string]string{"foo": "bar"}}
	require.True(t, u.HasParams())

	u = uriMatch{}
	require.False(t, u.HasParams())
}

const dummyProtoName = "dummy"

type DummyProto struct {
	Foo string `protobuf:"bytes,1,opt,name=foo,proto3" json:"foo,omitempty"`
	Bar bool   `protobuf:"varint,2,opt,name=bar,proto3" json:"bar,omitempty"`
	Baz int    `protobuf:"varint,3,opt,name=baz,proto3" json:"baz,omitempty"`
}

func (d DummyProto) Reset() {}

func (d DummyProto) String() string { return dummyProtoName }

func (d DummyProto) ProtoMessage() {}

func TestCreateMessage(t *testing.T) {
	gogoproto.RegisterType(&DummyProto{}, dummyProtoName)

	testCases := []struct {
		name     string
		uri      uriMatch
		expected gogoproto.Message
		expErr   bool
	}{
		{
			name:     "simple, empty message",
			uri:      uriMatch{QueryInputName: dummyProtoName},
			expected: &DummyProto{},
		},
		{
			name: "message with params",
			uri: uriMatch{
				QueryInputName: dummyProtoName,
				Params:         map[string]string{"foo": "blah", "bar": "true", "baz": "1352"},
			},
			expected: &DummyProto{
				Foo: "blah",
				Bar: true,
				Baz: 1352,
			},
		},
		{
			name: "invalid params should error out",
			uri: uriMatch{
				QueryInputName: dummyProtoName,
				Params:         map[string]string{"foo": "blah", "bar": "235235", "baz": "true"},
			},
			expErr: true,
		},
		{
			name: "unknown input type",
			uri: uriMatch{
				QueryInputName: "foobar",
			},
			expErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := createMessage(&tc.uri)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestCreateMessageFromJson(t *testing.T) {
	gogoproto.RegisterType(&DummyProto{}, dummyProtoName)
	testCases := []struct {
		name     string
		uri      uriMatch
		request  func() *http.Request
		expected gogoproto.Message
	}{
		{
			name: "simple, empty message",
			uri:  uriMatch{QueryInputName: dummyProtoName},
			request: func() *http.Request {
				return &http.Request{Body: io.NopCloser(bytes.NewReader([]byte("{}")))}
			},
			expected: &DummyProto{},
		},
		{
			name: "message with json input",
			uri:  uriMatch{QueryInputName: dummyProtoName},
			request: func() *http.Request {
				d := DummyProto{
					Foo: "hello",
					Bar: true,
					Baz: 320,
				}
				bz, err := json.Marshal(d)
				require.NoError(t, err)
				return &http.Request{Body: io.NopCloser(bytes.NewReader(bz))}
			},
			expected: &DummyProto{
				Foo: "hello",
				Bar: true,
				Baz: 320,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := createMessageFromJSON(&tc.uri, tc.request())
			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}
