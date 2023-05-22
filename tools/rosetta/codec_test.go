package rosetta_test

import (
	"testing"

	"cosmossdk.io/tools/rosetta"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/stretchr/testify/suite"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

type CodecTestSuite struct {
	suite.Suite

	ir  codectypes.InterfaceRegistry
	cdc *codec.ProtoCodec

	requiredInterfaces []string
}

func (s *CodecTestSuite) SetupTest() {
	s.ir = codectypes.NewInterfaceRegistry()
	s.cdc = codec.NewProtoCodec(s.ir)
	s.requiredInterfaces = []string{
		"cosmos.base.v1beta1.Msg",
		"cosmos.tx.v1beta1.Tx",
		"cosmos.crypto.PubKey",
		"cosmos.crypto.PrivKey",
		"ibc.core.client.v1.ClientState",
		"ibc.core.client.v1.Height",
		"cosmos.tx.v1beta1.MsgResponse",
		"ibc.core.client.v1.Header",
	}
}

func (s *CodecTestSuite) TestInterfaceRegistry() {
	s.Run("Required interfaces", func() {
		rosetta.RegisterInterfaces(s.ir)
		interfaceList := s.ir.ListAllInterfaces()

		interfaceListMap := make(map[string]bool)
		for _, interfaceTypeUrl := range interfaceList {
			interfaceListMap[interfaceTypeUrl] = true
		}

		for _, requiredInterfaceTypeUrl := range s.requiredInterfaces {
			s.Require().True(interfaceListMap[requiredInterfaceTypeUrl])
		}
	})
}

func TestCodecTestSuite(t *testing.T) {
	suite.Run(t, new(CodecTestSuite))
}
