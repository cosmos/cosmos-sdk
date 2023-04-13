package signing

import (
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	groupv1 "cosmossdk.io/api/cosmos/group/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"cosmossdk.io/x/tx/internal/testpb"
)

func TestGetSigners(t *testing.T) {
	ctx, err := NewGetSignersContext(GetSignersOptions{})
	require.NoError(t, err)
	tests := []struct {
		name    string
		msg     proto.Message
		want    []string
		wantErr bool
	}{
		{
			name: "MsgSend",
			msg: &bankv1beta1.MsgSend{
				FromAddress: "foo",
			},
			want: []string{"foo"},
		},
		{
			name: "MsgMultiSend",
			msg: &bankv1beta1.MsgMultiSend{
				Inputs: []*bankv1beta1.Input{
					{Address: "foo"},
					{Address: "bar"},
				},
			},
			want: []string{"foo", "bar"},
		},
		{
			name: "MsgSubmitProposal",
			msg: &groupv1.MsgSubmitProposal{
				Proposers: []string{"foo", "bar"},
			},
			want: []string{"foo", "bar"},
		},
		{
			name: "simple",
			msg:  &testpb.SimpleSigner{Signer: "foo"},
			want: []string{"foo"},
		},
		{
			name: "repeated",
			msg:  &testpb.RepeatedSigner{Signer: []string{"foo", "bar"}},
			want: []string{"foo", "bar"},
		},
		{
			name: "nested",
			msg:  &testpb.NestedSigner{Inner: &testpb.NestedSigner_Inner{Signer: "foo"}},
			want: []string{"foo"},
		},
		{
			name: "nested repeated",
			msg:  &testpb.NestedRepeatedSigner{Inner: &testpb.NestedRepeatedSigner_Inner{Signer: []string{"foo", "bar"}}},
			want: []string{"foo", "bar"},
		},
		{
			name: "repeated nested",
			msg: &testpb.RepeatedNestedSigner{Inner: []*testpb.RepeatedNestedSigner_Inner{
				{Signer: "foo"},
				{Signer: "bar"},
			}},
			want: []string{"foo", "bar"},
		},
		{
			name: "nested repeated",
			msg: &testpb.NestedRepeatedSigner{Inner: &testpb.NestedRepeatedSigner_Inner{
				Signer: []string{"foo", "bar"},
			}},
			want: []string{"foo", "bar"},
		},
		{
			name: "repeated nested repeated",
			msg: &testpb.RepeatedNestedRepeatedSigner{Inner: []*testpb.RepeatedNestedRepeatedSigner_Inner{
				{Signer: []string{"foo", "bar"}},
				{Signer: []string{"baz", "bam"}},
				{Signer: []string{"blah"}},
			}},
			want: []string{"foo", "bar", "baz", "bam", "blah"},
		},
		{
			name:    "bad",
			msg:     &testpb.BadSigner{},
			wantErr: true,
		},
		{
			name:    "no signer",
			msg:     &testpb.NoSignerOption{},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			signers, err := ctx.GetSigners(test.msg)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.want, signers)
		})
	}
}
