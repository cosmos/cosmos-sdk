package types_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptoAmino "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func testCodec() *codec.LegacyAmino {
	cdc := codec.NewLegacyAmino()
	sdk.RegisterLegacyAminoCodec(cdc)
	cryptoAmino.RegisterCrypto(cdc)
	cdc.RegisterConcrete(&testdata.TestMsg{}, "cosmos-sdk/Test", nil)
	return cdc
}

func TestStdTxConfig(t *testing.T) {
	cdc := testCodec()
	txGen := types.StdTxConfig{Cdc: cdc}
	suite.Run(t, testutil.NewTxConfigTestSuite(txGen))
}
