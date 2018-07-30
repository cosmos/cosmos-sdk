package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestStdTx(t *testing.T) {
	priv := ed25519.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())
	msgs := []sdk.Msg{sdk.NewTestMsg(addr)}
	fee := newStdFee()
	sigs := []StdSignature{}

	tx := NewStdTx(msgs, fee, sigs, "")
	require.Equal(t, msgs, tx.GetMsgs())
	require.Equal(t, sigs, tx.GetSignatures())

	feePayer := FeePayer(tx)
	require.Equal(t, addr, feePayer)
}

func TestStdSignBytes(t *testing.T) {
	priv := ed25519.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())
	msgs := []sdk.Msg{sdk.NewTestMsg(addr)}
	fee := newStdFee()
	signMsg := StdSignMsg{
		"1234",
		3,
		6,
		fee,
		msgs,
		"memo",
	}
	require.Equal(t, fmt.Sprintf("{\"account_number\":\"3\",\"chain_id\":\"1234\",\"fee\":{\"amount\":[{\"amount\":\"150\",\"denom\":\"atom\"}],\"gas\":\"5000\"},\"memo\":\"memo\",\"msgs\":[[\"%s\"]],\"sequence\":\"6\"}", addr), string(signMsg.Bytes()))
}
