package legacytx_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/pointnetwork/cosmos-point-sdk/codec"
	cryptoAmino "github.com/pointnetwork/cosmos-point-sdk/crypto/codec"
	"github.com/pointnetwork/cosmos-point-sdk/testutil/testdata"
	sdk "github.com/pointnetwork/cosmos-point-sdk/types"
	"github.com/pointnetwork/cosmos-point-sdk/x/auth/migrations/legacytx"
	"github.com/pointnetwork/cosmos-point-sdk/x/auth/testutil"
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
	txGen := legacytx.StdTxConfig{Cdc: cdc}
	suite.Run(t, testutil.NewTxConfigTestSuite(txGen))
}
