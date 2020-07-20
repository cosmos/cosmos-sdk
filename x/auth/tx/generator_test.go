package tx

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/testutil"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
)

func TestGenerator(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &testdata.TestMsg{})
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	pubKeyCodec := std.DefaultPublicKeyCodec{}
	signModeHandler := DefaultSignModeHandler()
	suite.Run(t, testutil.NewTxConfigTestSuite(NewTxConfig(marshaler, pubKeyCodec, signModeHandler)))
}
