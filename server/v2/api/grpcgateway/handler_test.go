package grpcgateway

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/core/transaction"
)

func Test_fixCatchAll(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want string
	}{
		{
			name: "replaces catch all",
			uri:  "/foo/bar/{baz=**}",
			want: "/foo/bar/{baz...}",
		},
		{
			name: "returns original",
			uri:  "/foo/bar/baz",
			want: "/foo/bar/baz",
		},
		{
			name: "doesn't tamper with normal wildcard",
			uri:  "/foo/{baz}",
			want: "/foo/{baz}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, fixCatchAll(tt.uri))
		})
	}
}

func Test_extractWildcardKeyNames(t *testing.T) {
	tests := []struct {
		name string
		uri  string
		want []string
	}{
		{
			name: "single",
			uri:  "/foo/bar/{baz}",
			want: []string{"baz"},
		},
		{
			name: "multiple",
			uri:  "/foo/{bar}/baz/{buzz}",
			want: []string{"bar", "buzz"},
		},
		{
			name: "catch-all wildcard",
			uri:  "/foo/{buzz...}",
			want: []string{"buzz"},
		},
		{
			name: "none",
			uri:  "/foo/bar",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, extractWildcardKeyNames(tt.uri))
		})
	}
}

func TestPopulateMessage(t *testing.T) {
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
			name: "simple query params and body",
			request: func() *http.Request {
				body := `{"denoms": ["hello", "there"]}`

				req := httptest.NewRequest(
					http.MethodGet,
					"/foo", // this doesn't really matter
					bytes.NewReader([]byte(body)),
				)
				return req
			},
			wildcardValues: map[string]string{
				"foo": "wildFooValue", // from path wildcard e.g. /dummy/{foo}
			},
			expected: &DummyProto{
				Foo:    "wildFooValue",
				Denoms: []string{"hello", "there"},
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

	// We only need a minimal gatewayInterceptor instance to call populateMessage,
	// so it's fine to leave most fields nil for this unit test.
	g := &protoHandler[transaction.Tx]{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.request()
			inputMsg := &DummyProto{}
			gotMsg, err := g.populateMessage(
				req,
				&runtime.JSONPb{},
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
