package nft

import (
	"fmt"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

const (
	OwnerName  = "owner"
	Owner      = "cosmos1kznrznww4pd6gx0zwrpthjk68fdmqypjpkj5hp"
	OwnerArmor = `-----BEGIN TENDERMINT PRIVATE KEY-----
salt: C3586B75587D2824187D2CDA22B6AFB6
type: secp256k1
kdf: bcrypt

1+15OrCKgjnwym1zO3cjo/SGe3PPqAYChQ5wMHjdUbTZM7mWsH3/ueL6swgjzI3b
DDzEQAPXBQflzNW6wbne9IfT651zCSm+j1MWaGk=
=wEHs
-----END TENDERMINT PRIVATE KEY-----`

	testClassID          = "kitty"
	testClassName        = "Crypto Kitty"
	testClassSymbol      = "kitty"
	testClassDescription = "Crypto Kitty"
	testClassURI         = "class uri"
	testID               = "kitty1"
	testURI              = "kitty uri"
)

var (
	ExpClass = nft.Class{
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
	}

	ExpNFT = nft.NFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     testURI,
	}
)

type E2ETestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
	owner   sdk.AccAddress
}

func NewE2ETestSuite(cfg network.Config) *E2ETestSuite {
	return &E2ETestSuite{cfg: cfg}
}

func (s *E2ETestSuite) SetupSuite() {
	s.T().Log("setting up e2e test suite")

	genesisState := s.cfg.GenesisState
	nftGenesis := nft.GenesisState{
		Classes: []*nft.Class{&ExpClass},
		Entries: []*nft.Entry{{
			Owner: Owner,
			Nfts:  []*nft.NFT{&ExpNFT},
		}},
	}

	nftDataBz, err := s.cfg.Codec.MarshalJSON(&nftGenesis)
	s.Require().NoError(err)
	genesisState[nft.ModuleName] = nftDataBz
	s.cfg.GenesisState = genesisState
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)
	s.Require().NoError(s.network.WaitForNextBlock())

	s.initAccount()
	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *E2ETestSuite) TearDownSuite() {
	s.T().Log("tearing down e2e test suite")
	s.network.Cleanup()
}

func (s *E2ETestSuite) TestCLITxSend() {
	val := s.network.Validators[0]
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, OwnerName),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}
	testCases := []struct {
		name         string
		args         []string
		expectedCode uint32
		expectErr    bool
	}{
		{
			"valid transaction",
			[]string{
				testClassID,
				testID,
				val.Address.String(),
			},
			0,
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx
			args = append(args, tc.args...)
			out, err := ExecSend(
				val,
				args,
			)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var txResp sdk.TxResponse
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
				s.Require().NoError(clitestutil.CheckTxCode(s.network, clientCtx, txResp.TxHash, tc.expectedCode))
			}
		})
	}
}

func (s *E2ETestSuite) initAccount() {
	val := s.network.Validators[0]
	ctx := val.ClientCtx
	err := ctx.Keyring.ImportPrivKey(OwnerName, OwnerArmor, "1234567890")
	s.Require().NoError(err)

	keyinfo, err := ctx.Keyring.Key(OwnerName)
	s.Require().NoError(err)

	args := []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	s.owner, err = keyinfo.GetAddress()
	s.Require().NoError(err)

	amount := sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(200)))
	_, err = clitestutil.MsgSendExec(ctx, val.Address, s.owner, amount, args...)
	s.Require().NoError(err)
}
