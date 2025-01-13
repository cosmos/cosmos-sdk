package grpcgateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
)

func Test_createRegexMapping(t *testing.T) {
	tests := []struct {
		name           string
		annotations    map[string]string
		expectedRegex  int
		expectedSimple int
		wantWarn       bool
	}{
		{
			name: "no annotations should not warn",
		},
		{
			name: "expected correct amount of regex and simple matchers",
			annotations: map[string]string{
				"/foo/bar/baz":   "",
				"/foo/{bar}/baz": "",
				"/foo/bar/bell":  "",
			},
			expectedRegex:  1,
			expectedSimple: 2,
		},
		{
			name: "different annotations should not warn",
			annotations: map[string]string{
				"/foo/bar/{baz}":     "",
				"/crypto/{currency}": "",
			},
			expectedRegex: 2,
		},
		{
			name: "duplicate annotations should warn",
			annotations: map[string]string{
				"/hello/{world}":      "",
				"/hello/{developers}": "",
			},
			expectedRegex: 2,
			wantWarn:      true,
		},
	}
	buf := bytes.NewBuffer(nil)
	logger := log.NewLogger(buf)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, simple := createRegexMapping(logger, tt.annotations)
			if tt.wantWarn {
				require.NotEmpty(t, buf.String())
			} else {
				require.Empty(t, buf.String())
			}
			require.Equal(t, tt.expectedRegex, len(regex))
			require.Equal(t, tt.expectedSimple, len(simple))
		})
	}
}

func TestCreateMessageFromGetRequest(t *testing.T) {
	gogoproto.RegisterType(&DummyProto{}, dummyProtoName)

	testCases := []struct {
		name           string
		request        func() *http.Request
		wildcardValues map[string]string
		expected       *DummyProto
		wantErr        bool
		errCode        codes.Code
	}{
		{
			name: "simple wildcard + query params",
			request: func() *http.Request {
				// GET with query params: ?bar=true&baz=42&denoms=atom&denoms=osmo
				// Also nested pagination params: page.limit=100, page.nest.foo=999
				req := httptest.NewRequest(
					http.MethodGet,
					"/dummy?bar=true&baz=42&denoms=atom&denoms=osmo&page.limit=100&page.nest.foo=999",
					nil,
				)
				return req
			},
			wildcardValues: map[string]string{
				"foo": "wildFooValue", // from path wildcard e.g. /dummy/{foo}
			},
			expected: &DummyProto{
				Foo:    "wildFooValue",
				Bar:    true,
				Baz:    42,
				Denoms: []string{"atom", "osmo"},
				Page: &Pagination{
					Limit: 100,
					Nest: &Nested{
						Foo: 999,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid integer in query param",
			request: func() *http.Request {
				req := httptest.NewRequest(
					http.MethodGet,
					"/dummy?baz=notanint",
					nil,
				)
				return req
			},
			wildcardValues: map[string]string{},
			expected:       &DummyProto{}, // won't get populated
			wantErr:        true,
			errCode:        codes.InvalidArgument,
		},
		{
			name: "no query params, but wildcard set",
			request: func() *http.Request {
				// No query params. Only the wildcard.
				req := httptest.NewRequest(
					http.MethodGet,
					"/dummy",
					nil,
				)
				return req
			},
			wildcardValues: map[string]string{
				"foo": "barFromWildcard",
			},
			expected: &DummyProto{
				Foo: "barFromWildcard",
			},
			wantErr: false,
		},
	}

	// We only need a minimal gatewayInterceptor instance to call createMessageFromGetRequest,
	// so it's fine to leave most fields nil for this unit test.
	g := &gatewayInterceptor[transaction.Tx]{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.request()

			inputMsg := &DummyProto{}
			gotMsg, err := g.createMessageFromGetRequest(
				req,
				inputMsg,
				tc.wildcardValues,
			)

			if tc.wantErr {
				require.Error(t, err, "expected error but got none")
				st, ok := status.FromError(err)
				if ok && tc.errCode != codes.OK {
					require.Equal(t, tc.errCode, st.Code())
				}
			} else {
				require.NoError(t, err, "unexpected error")
				require.Equal(t, tc.expected, gotMsg, "message contents do not match expected")
			}
		})
	}
}

func TestCreateMessageFromPostRequest(t *testing.T) {
	gogoproto.RegisterType(&DummyProto{}, dummyProtoName)
	gogoproto.RegisterType(&Pagination{}, "pagination")
	gogoproto.RegisterType(&Nested{}, "nested")

	testCases := []struct {
		name     string
		body     any
		wantErr  bool
		errCode  codes.Code
		expected *DummyProto
	}{
		{
			name: "valid JSON body with nested fields",
			body: map[string]any{
				"foo":    "postFoo",
				"bar":    true,
				"baz":    42,
				"denoms": []string{"atom", "osmo"},
				"page": map[string]any{
					"limit": 100,
					"nest": map[string]any{
						"foo": 999,
					},
				},
			},
			wantErr: false,
			expected: &DummyProto{
				Foo:    "postFoo",
				Bar:    true,
				Baz:    42,
				Denoms: []string{"atom", "osmo"},
				Page: &Pagination{
					Limit: 100,
					Nest: &Nested{
						Foo: 999,
					},
				},
			},
		},
		{
			name: "invalid JSON structure",
			// Provide a broken JSON string:
			body:    `{"foo": "bad json", "extra": "not closed"`,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name:     "empty JSON object",
			body:     map[string]any{},
			wantErr:  false,
			expected: &DummyProto{}, // all fields remain zeroed
		},
	}

	g := &gatewayInterceptor[transaction.Tx]{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var reqBody []byte
			switch typedBody := tc.body.(type) {
			case string:
				// This might be invalid JSON we intentionally want to test
				reqBody = []byte(typedBody)
			default:
				// Marshal the given any into JSON
				b, err := json.Marshal(typedBody)
				require.NoError(t, err, "failed to marshal test body to JSON")
				reqBody = b
			}

			req := httptest.NewRequest(http.MethodPost, "/dummy", bytes.NewReader(reqBody))

			inputMsg := &DummyProto{}
			gotMsg, err := g.createMessageFromPostRequest(
				&runtime.JSONPb{}, // JSONPb marshaler
				req,
				inputMsg,
			)

			if tc.wantErr {
				require.Error(t, err, "expected an error but got none")
				// Optionally verify the gRPC status code
				st, ok := status.FromError(err)
				if ok && tc.errCode != codes.OK {
					require.Equal(t, tc.errCode, st.Code())
				}
			} else {
				require.NoError(t, err, "did not expect an error")
				require.Equal(t, tc.expected, gotMsg)
			}
		})
	}
}

/*
--- Testing Types ---
*/
type Nested struct {
	Foo int32 `protobuf:"varint,1,opt,name=foo,proto3" json:"foo,omitempty"`
}

func (n Nested) Reset() {}

func (n Nested) String() string { return "" }

func (n Nested) ProtoMessage() {}

type Pagination struct {
	Limit int32   `protobuf:"varint,1,opt,name=limit,proto3" json:"limit,omitempty"`
	Nest  *Nested `protobuf:"bytes,2,opt,name=nest,proto3" json:"nest,omitempty"`
}

func (p Pagination) Reset() {}

func (p Pagination) String() string { return "" }

func (p Pagination) ProtoMessage() {}

const dummyProtoName = "dummy"

type DummyProto struct {
	Foo    string      `protobuf:"bytes,1,opt,name=foo,proto3" json:"foo,omitempty"`
	Bar    bool        `protobuf:"varint,2,opt,name=bar,proto3" json:"bar,omitempty"`
	Baz    int32       `protobuf:"varint,3,opt,name=baz,proto3" json:"baz,omitempty"`
	Denoms []string    `protobuf:"bytes,4,rep,name=denoms,proto3" json:"denoms,omitempty"`
	Page   *Pagination `protobuf:"bytes,5,opt,name=page,proto3" json:"page,omitempty"`
}

func (d DummyProto) Reset() {}

func (d DummyProto) String() string { return dummyProtoName }

func (d DummyProto) ProtoMessage() {}
