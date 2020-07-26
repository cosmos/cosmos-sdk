package tx

import (
	"testing"

	"github.com/KiraCore/cosmos-sdk/testutil/testdata"

	sdk "github.com/KiraCore/cosmos-sdk/types"

	"github.com/stretchr/testify/suite"

	"github.com/KiraCore/cosmos-sdk/client/testutil"
	"github.com/KiraCore/cosmos-sdk/codec"
	codectypes "github.com/KiraCore/cosmos-sdk/codec/types"
	"github.com/KiraCore/cosmos-sdk/std"
)

func TestGenerator(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &testdata.TestMsg{})
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	pubKeyCodec := std.DefaultPublicKeyCodec{}
	signModeHandler := DefaultSignModeHandler()
	suite.Run(t, testutil.NewTxConfigTestSuite(NewTxConfig(marshaler, pubKeyCodec, signModeHandler)))
}
