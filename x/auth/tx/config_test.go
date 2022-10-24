package tx

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/pointnetwork/cosmos-point-sdk/codec"
	codectypes "github.com/pointnetwork/cosmos-point-sdk/codec/types"
	"github.com/pointnetwork/cosmos-point-sdk/std"
	"github.com/pointnetwork/cosmos-point-sdk/testutil/testdata"
	sdk "github.com/pointnetwork/cosmos-point-sdk/types"
	"github.com/pointnetwork/cosmos-point-sdk/x/auth/testutil"
)

func TestGenerator(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &testdata.TestMsg{})
	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	suite.Run(t, testutil.NewTxConfigTestSuite(NewTxConfig(protoCodec, DefaultSignModes)))
}
