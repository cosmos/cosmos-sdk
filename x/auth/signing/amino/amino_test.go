package amino_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing/amino"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"
)

func TestLegacyAminoJSONHandler(t *testing.T) {
	priv1 := secp256k1.GenPrivKey()
	addr1 := sdk.AccAddress(priv1.PubKey().Address())
	priv2 := secp256k1.GenPrivKey()
	addr2 := sdk.AccAddress(priv2.PubKey().Address())

	coins := sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}

	fee := types.StdFee{
		Amount: coins,
		Gas:    10000,
	}
	memo := "foo"
	msgs := []sdk.Msg{
		&bank.MsgSend{
			FromAddress: addr1,
			ToAddress:   addr2,
			Amount:      coins,
		},
	}

	tx := auth.StdTx{
		Msgs:       msgs,
		Fee:        fee,
		Signatures: nil,
		Memo:       memo,
	}

	var (
		chainId        = "test-chain"
		accNum  uint64 = 7
		seqNum  uint64 = 7
	)

	handler := amino.LegacyAminoJSONHandler{}
	signBz, err := handler.GetSignBytes(types2.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, signing.SigningData{
		ChainID:         chainId,
		AccountNumber:   accNum,
		AccountSequence: seqNum,
		PublicKey:       priv1.PubKey(),
	}, tx)
	require.NoError(t, err)

	expectedSignBz := types.StdSignBytes(chainId, accNum, seqNum, fee, msgs, memo)

	require.Equal(t, expectedSignBz, signBz)

}
