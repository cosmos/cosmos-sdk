package generator

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client/testutil"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/stretchr/testify/suite"
)

func TestGenerator(t *testing.T) {
	marshaler := codec.NewHybridCodec(codec.New(), codectypes.NewInterfaceRegistry())
	pubKeyCodec := std.DefaultPublicKeyCodec{}
	suite.Run(t, testutil.NewTxGeneratorTestSuite(New(marshaler, pubKeyCodec)))
}
