package amino_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing/amino"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

func TestLegacyAminoJSONHandler_GetSignBytes(t *testing.T) {
	priv1 := secp256k1.GenPrivKey()
	addr1 := sdk.AccAddress(priv1.PubKey().Address())
	priv2 := secp256k1.GenPrivKey()
	addr2 := sdk.AccAddress(priv2.PubKey().Address())

	coins := sdk.Coins{sdk.NewInt64Coin("foocoin", 10)}

	fee := auth.StdFee{
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
	signingData := signing.SignerData{
		ChainID:         chainId,
		AccountNumber:   accNum,
		AccountSequence: seqNum,
	}
	signBz, err := handler.GetSignBytes(txtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, signingData, tx)
	require.NoError(t, err)

	expectedSignBz := auth.StdSignBytes(chainId, accNum, seqNum, fee, msgs, memo)

	require.Equal(t, expectedSignBz, signBz)

	// expect error with wrong sign mode
	_, err = handler.GetSignBytes(txtypes.SignMode_SIGN_MODE_DIRECT, signingData, tx)
	require.Error(t, err)
}

func TestLegacyAminoJSONHandler_DefaultMode(t *testing.T) {
	handler := amino.LegacyAminoJSONHandler{}
	require.Equal(t, txtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, handler.DefaultMode())
}

func TestLegacyAminoJSONHandler_Modes(t *testing.T) {
	handler := amino.LegacyAminoJSONHandler{}
	require.Equal(t, []txtypes.SignMode{txtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON}, handler.Modes())
}
