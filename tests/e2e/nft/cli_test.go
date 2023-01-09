package nft

import (
	"fmt"
	"testing"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	"gotest.tools/v3/assert"
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

type fixture struct {
	cfg     network.Config
	network *network.Network
	owner   sdk.AccAddress
}

func initFixture(t *testing.T) (*fixture, func()) {
	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 1

	genesisState := cfg.GenesisState
	nftGenesis := nft.GenesisState{
		Classes: []*nft.Class{&ExpClass},
		Entries: []*nft.Entry{{
			Owner: Owner,
			Nfts:  []*nft.NFT{&ExpNFT},
		}},
	}

	nftDataBz, err := cfg.Codec.MarshalJSON(&nftGenesis)
	assert.NilError(t, err)
	genesisState[nft.ModuleName] = nftDataBz
	cfg.GenesisState = genesisState
	network, err := network.New(t, t.TempDir(), cfg)
	assert.NilError(t, err)
	assert.NilError(t, network.WaitForNextBlock())

	val := network.Validators[0]
	ctx := val.ClientCtx
	assert.NilError(t, ctx.Keyring.ImportPrivKey(OwnerName, OwnerArmor, "1234567890"))

	keyinfo, err := ctx.Keyring.Key(OwnerName)
	assert.NilError(t, err)

	args := []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	owner, err := keyinfo.GetAddress()
	assert.NilError(t, err)

	amount := sdk.NewCoins(sdk.NewCoin(cfg.BondDenom, sdk.NewInt(200)))
	_, err = clitestutil.MsgSendExec(ctx, val.Address, owner, amount, args...)
	assert.NilError(t, err)
	assert.NilError(t, network.WaitForNextBlock())

	return &fixture{
			cfg:     cfg,
			network: network,
			owner:   owner,
		}, func() {
			network.Cleanup()
		}
}

func TestCLITxSend(t *testing.T) {
	t.Parallel()
	f, cleanup := initFixture(t)
	defer cleanup()

	val := f.network.Validators[0]
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, OwnerName),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(f.cfg.BondDenom, sdk.NewInt(10))).String()),
	}
	testCases := []struct {
		name         string
		args         []string
		expectedCode uint32
	}{
		{
			"valid transaction",
			[]string{
				testClassID,
				testID,
				val.Address.String(),
			},
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			clientCtx := val.ClientCtx
			args = append(args, tc.args...)
			out, err := ExecSend(val, args)

			var txResp sdk.TxResponse
			assert.NilError(t, err)
			assert.NilError(t, clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
			assert.NilError(t, clitestutil.CheckTxCode(f.network, clientCtx, txResp.TxHash, tc.expectedCode))
		})
	}
}
