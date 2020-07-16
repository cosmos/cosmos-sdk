package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/testdata"

	"github.com/cosmos/cosmos-sdk/client/testutil"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func testCodec() *codec.Codec {
	cdc := codec.New()
	sdk.RegisterCodec(cdc)
	cdc.RegisterConcrete(&testdata.TestMsg{}, "cosmos-sdk/Test", nil)
	return cdc
}

func TestStdTxConfig(t *testing.T) {
	cdc := testCodec()
	txGen := types.StdTxConfig{Cdc: cdc}
	suite.Run(t, testutil.NewTxConfigTestSuite(txGen))
}
