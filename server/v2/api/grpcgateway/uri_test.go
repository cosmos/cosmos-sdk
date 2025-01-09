package grpcgateway

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
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
			uri:      "https://localhost:8080/foo/bar",
			mapping:  map[string]string{"/foo/bar": "query.Bank"},
			expected: &uriMatch{QueryInputName: "query.Bank", Params: map[string][]string{}},
		},
		{
			name:     "match with query parameters",
			uri:      "https://localhost:8080/foo/bar?baz=qux",
			mapping:  map[string]string{"/foo/bar": "query.Bank"},
			expected: &uriMatch{QueryInputName: "query.Bank", Params: map[string][]string{"baz": {"qux"}}},
		},
		{
			name:     "match with multiple query parameters",
			uri:      "https://localhost:8080/foo/bar?baz=qux&foo=/msg.type.bank.send",
			mapping:  map[string]string{"/foo/bar": "query.Bank"},
			expected: &uriMatch{QueryInputName: "query.Bank", Params: map[string][]string{"baz": {"qux"}, "foo": {"/msg.type.bank.send"}}},
		},
		{
			name:    "wildcard match at the end",
			uri:     "https://localhost:8080/foo/bar/buzz",
			mapping: map[string]string{"/foo/bar/{baz}": "bar"},
			expected: &uriMatch{
				QueryInputName: "bar",
				Params:         map[string][]string{"baz": {"buzz"}},
			},
		},
		{
			name:    "wildcard match in the middle",
			uri:     "https://localhost:8080/foo/buzz/bar",
			mapping: map[string]string{"/foo/{baz}/bar": "bar"},
			expected: &uriMatch{
				QueryInputName: "bar",
				Params:         map[string][]string{"baz": {"buzz"}},
			},
		},
		{
			name:    "multiple wild cards",
			uri:     "https://localhost:8080/foo/bar/baz/buzz",
			mapping: map[string]string{"/foo/bar/{q1}/{q2}": "bar"},
			expected: &uriMatch{
				QueryInputName: "bar",
				Params:         map[string][]string{"q1": {"baz"}, "q2": {"buzz"}},
			},
		},
		{
			name:    "match with multiple query parameters",
			uri:     "https://localhost:8080/bank/supply/by_denom?denom=foo&denom=bar",
			mapping: map[string]string{"/bank/supply/by_denom": "queryDenom"},
			expected: &uriMatch{
				QueryInputName: "queryDenom",
				Params:         map[string][]string{"denom": {"foo", "bar"}},
			},
		},
		{
			name:    "catch-all wildcard",
			uri:     "https://localhost:8080/foo/bar/ibc/token/stuff",
			mapping: map[string]string{"/foo/bar/{ibc_token=**}": "bar"},
			expected: &uriMatch{
				QueryInputName: "bar",
				Params:         map[string][]string{"ibc_token": {"ibc/token/stuff"}},
			},
		},
		{
			name:     "no match should return nil",
			uri:      "https://localhost:8080/foo/bar",
			mapping:  map[string]string{"/bar/foo": "bar"},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := url.Parse(tc.uri)
			require.NoError(t, err)
			regexpMapping := createRegexMapping(tc.mapping)
			actual := matchURL(u, regexpMapping)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestURIMatch_HasParams(t *testing.T) {
	u := uriMatch{Params: map[string][]string{"foo": {"bar"}}}
	require.True(t, u.HasParams())

	u = uriMatch{}
	require.False(t, u.HasParams())
}

type Nested struct {
	Foo int `protobuf:"varint,1,opt,name=foo,proto3" json:"foo,omitempty"`
}

type Pagination struct {
	Limit int     `protobuf:"varint,1,opt,name=limit,proto3" json:"limit,omitempty"`
	Nest  *Nested `protobuf:"bytes,2,opt,name=nest,proto3" json:"nest,omitempty"`
}

const dummyProtoName = "dummy"

type DummyProto struct {
	Foo    string      `protobuf:"bytes,1,opt,name=foo,proto3" json:"foo,omitempty"`
	Bar    bool        `protobuf:"varint,2,opt,name=bar,proto3" json:"bar,omitempty"`
	Baz    int         `protobuf:"varint,3,opt,name=baz,proto3" json:"baz,omitempty"`
	Denoms []string    `protobuf:"bytes,4,rep,name=denoms,proto3" json:"denoms,omitempty"`
	Page   *Pagination `protobuf:"bytes,4,opt,name=page,proto3" json:"page,omitempty"`
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
				Params: map[string][]string{
					"foo": {"blah"},
					"bar": {"true"},
					"baz": {"1352"},
				},
			},
			expected: &DummyProto{
				Foo: "blah",
				Bar: true,
				Baz: 1352,
			},
		},
		{
			name: "message with slice param",
			uri: uriMatch{
				QueryInputName: dummyProtoName,
				Params: map[string][]string{
					"foo":    {"blah"},
					"bar":    {"true"},
					"baz":    {"1352"},
					"denoms": {"atom", "stake"},
				},
			},
			expected: &DummyProto{
				Foo:    "blah",
				Bar:    true,
				Baz:    1352,
				Denoms: []string{"atom", "stake"},
			},
		},
		{
			name: "message with multiple param for single value field should fail",
			uri: uriMatch{
				QueryInputName: dummyProtoName,
				Params: map[string][]string{
					"foo":    {"blah", "blahhh"}, // foo is a single value field.
					"bar":    {"true"},
					"baz":    {"1352"},
					"denoms": {"atom", "stake"},
				},
			},
			expErr: true,
		},
		{
			name: "message with nested params",
			uri: uriMatch{
				QueryInputName: dummyProtoName,
				Params: map[string][]string{
					"foo":        {"blah"},
					"bar":        {"true"},
					"baz":        {"1352"},
					"page.limit": {"3"},
				},
			},
			expected: &DummyProto{
				Foo:  "blah",
				Bar:  true,
				Baz:  1352,
				Page: &Pagination{Limit: 3},
			},
		},
		{
			name: "message with multi nested params",
			uri: uriMatch{
				QueryInputName: dummyProtoName,
				Params: map[string][]string{
					"foo":           {"blah"},
					"bar":           {"true"},
					"baz":           {"1352"},
					"page.limit":    {"3"},
					"page.nest.foo": {"5"},
				},
			},
			expected: &DummyProto{
				Foo:  "blah",
				Bar:  true,
				Baz:  1352,
				Page: &Pagination{Limit: 3, Nest: &Nested{Foo: 5}},
			},
		},
		{
			name: "invalid params should error out",
			uri: uriMatch{
				QueryInputName: dummyProtoName,
				Params: map[string][]string{
					"foo": {"blah"},
					"bar": {"235235"},
					"baz": {"true"},
				},
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
		expErr   bool
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
		{
			name: "message with invalid json",
			uri:  uriMatch{QueryInputName: dummyProtoName},
			request: func() *http.Request {
				return &http.Request{Body: io.NopCloser(bytes.NewReader([]byte(`{"foo":12,dfi3}"`)))}
			},
			expErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := createMessageFromJSON(&tc.uri, tc.request())
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			}
		})
	}
}

func Test_patternToRegex(t *testing.T) {
	tests := []struct {
		name           string
		pattern        string
		wildcards      []string
		wildcardValues []string
		shouldMatch    string
		shouldNotMatch []string
	}{
		{
			name:           "simple match, no wildcards",
			pattern:        "/foo/bar/baz",
			shouldMatch:    "/foo/bar/baz",
			shouldNotMatch: []string{"/foo/bar", "/foo", "/foo/bar/baz/boo"},
		},
		{
			name:           "match with wildcard",
			pattern:        "/foo/bar/{baz}",
			wildcards:      []string{"baz"},
			shouldMatch:    "/foo/bar/hello",
			wildcardValues: []string{"hello"},
			shouldNotMatch: []string{"/foo/bar", "/foo/bar/baz/boo"},
		},
		{
			name:           "match with multiple wildcards",
			pattern:        "/foo/{bar}/{baz}/meow",
			wildcards:      []string{"bar", "baz"},
			shouldMatch:    "/foo/hello/world/meow",
			wildcardValues: []string{"hello", "world"},
			shouldNotMatch: []string{"/foo/bar/baz/boo", "/foo/bar/baz"},
		},
		{
			name:           "match catch-all wildcard",
			pattern:        `/foo/bar/{baz=**}`,
			wildcards:      []string{"baz"},
			shouldMatch:    `/foo/bar/this/is/a/long/wildcard`,
			wildcardValues: []string{"this/is/a/long/wildcard"},
			shouldNotMatch: []string{"/foo/bar", "/foo", "/foo/baz/bar/long/wild/card"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regString, wildcards := patternToRegex(tt.pattern)
			// should produce the same wildcard keys
			require.Equal(t, tt.wildcards, wildcards)
			reg := regexp.MustCompile(regString)

			// handle the "should match" case.
			matches := reg.FindStringSubmatch(tt.shouldMatch)
			require.True(t, len(matches) > 0) // there should always be a match.
			// when matches > 1, this means we got wildcard values to handle. the test should have wildcard values.
			if len(matches) > 1 {
				require.Greater(t, len(tt.wildcardValues), 0)
			}
			// matches[0] is the URL, everything else should be those wildcard values.
			if len(tt.wildcardValues) > 0 {
				require.Equal(t, matches[1:], tt.wildcardValues)
			}

			// should never match these.
			for _, notMatch := range tt.shouldNotMatch {
				require.Len(t, reg.FindStringSubmatch(notMatch), 0)
			}
		})
	}
}
