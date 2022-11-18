package mock

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestGetPubKey(t *testing.T) {
	pv := NewPV()
	pb, err := pv.GetPubKey()
	require.NoError(t, err)
	require.NotNil(t, pb)
}

func TestSignVote(t *testing.T) {
	pv := NewPV()
	v := tmproto.Vote{}
	err := pv.SignVote("chain-id", &v)
	require.NoError(t, err)
	require.NotNil(t, v.Signature)
}

func TestSignProposal(t *testing.T) {
	pv := NewPV()
	p := tmproto.Proposal{}
	err := pv.SignProposal("chain-id", &p)
	require.NoError(t, err)
	require.NotNil(t, p.Signature)
}
