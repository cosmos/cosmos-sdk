package cometbft

import (
	"testing"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	"cosmossdk.io/core/event"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func TestSplitABCIQueryPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "empty path",
			path:     "",
			expected: []string{},
		},
		{
			name:     "root path",
			path:     "/",
			expected: []string{""},
		},
		{
			name:     "single component",
			path:     "/store",
			expected: []string{"store"},
		},
		{
			name:     "multiple components",
			path:     "/store/key/value",
			expected: []string{"store", "key", "value"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := splitABCIQueryPath(tc.path)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestIntoABCIEvents(t *testing.T) {
	indexSet := map[string]struct{}{
		"transfer.sender": {},
	}

	tests := []struct {
		name      string
		events    []event.Event
		indexSet  map[string]struct{}
		indexNone bool
		wantErr   bool
	}{
		{
			name: "basic events",
			events: []event.Event{
				event.NewEvent("transfer", event.NewAttribute("sender", "addr1")),
			},
			indexSet:  indexSet,
			indexNone: false,
			wantErr:   false,
		},
		{
			name:      "empty events",
			events:    []event.Event{},
			indexSet:  indexSet,
			indexNone: false,
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := intoABCIEvents(tc.events, tc.indexSet, tc.indexNone)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)
		})
	}
}

func TestUint64ToInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected int64
	}{
		{
			name:     "zero",
			input:    0,
			expected: 0,
		},
		{
			name:     "max int64",
			input:    uint64(9223372036854775807),
			expected: 9223372036854775807,
		},
		{
			name:     "overflow",
			input:    18446744073709551615,
			expected: 9223372036854775807,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := uint64ToInt64(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestValidateExtendedCommitAgainstLastCommit(t *testing.T) {
	tests := []struct {
		name      string
		ec        abci.ExtendedCommitInfo
		lc        abci.CommitInfo
		expectErr bool
	}{
		{
			name: "valid commit info",
			ec: abci.ExtendedCommitInfo{
				Round: 1,
				Votes: []abci.ExtendedVoteInfo{
					{
						Validator: abci.Validator{
							Address: []byte("val1"),
							Power:   100,
						},
					},
				},
			},
			lc: abci.CommitInfo{
				Round: 1,
				Votes: []abci.VoteInfo{
					{
						Validator: abci.Validator{
							Address: []byte("val1"),
							Power:   100,
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "mismatched rounds",
			ec: abci.ExtendedCommitInfo{
				Round: 1,
			},
			lc: abci.CommitInfo{
				Round: 2,
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateExtendedCommitAgainstLastCommit(tc.ec, tc.lc)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGRPCErrorToSDKError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode uint32
	}{
		{
			name:         "not found error",
			err:          grpcstatus.Error(codes.NotFound, "not found"),
			expectedCode: sdkerrors.ErrKeyNotFound.ABCICode(),
		},
		{
			name:         "invalid argument",
			err:          grpcstatus.Error(codes.InvalidArgument, "invalid"),
			expectedCode: sdkerrors.ErrInvalidRequest.ABCICode(),
		},
		{
			name:         "unauthenticated",
			err:          grpcstatus.Error(codes.Unauthenticated, "unauthorized"),
			expectedCode: sdkerrors.ErrUnauthorized.ABCICode(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := gRPCErrorToSDKError(tc.err)
			require.Equal(t, tc.expectedCode, resp.Code)
		})
	}
}
