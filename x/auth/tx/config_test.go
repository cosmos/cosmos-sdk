package tx

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/cosmos-sdk/codec"
	codectypes "github.com/Stride-Labs/cosmos-sdk/codec/types"
	"github.com/Stride-Labs/cosmos-sdk/std"
	"github.com/Stride-Labs/cosmos-sdk/testutil/testdata"
	sdk "github.com/Stride-Labs/cosmos-sdk/types"
	"github.com/Stride-Labs/cosmos-sdk/x/auth/testutil"
)

func TestGenerator(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &testdata.TestMsg{})
	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	suite.Run(t, testutil.NewTxConfigTestSuite(NewTxConfig(protoCodec, DefaultSignModes)))
}
