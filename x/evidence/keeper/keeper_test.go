package keeper_test

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/std"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

var (
	pubkeys = []crypto.PubKey{
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51"),
		newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52"),
	}

	valAddresses = []sdk.ValAddress{
		sdk.ValAddress(pubkeys[0].Address()),
		sdk.ValAddress(pubkeys[1].Address()),
		sdk.ValAddress(pubkeys[2].Address()),
	}

	initAmt   = sdk.TokensFromConsensusPower(200)
	initCoins = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))
)

func newPubKey(pk string) (res crypto.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}

	var pubkey ed25519.PubKeyEd25519
	copy(pubkey[:], pkBytes)

	return pubkey
}

func testEquivocationHandler(k interface{}) types.Handler {
	return func(ctx sdk.Context, e exported.Evidence) error {
		if err := e.ValidateBasic(); err != nil {
			return err
		}

		ee, ok := e.(types.Equivocation)
		if !ok {
			return fmt.Errorf("unexpected evidence type: %T", e)
		}
		if ee.Height%2 == 0 {
			return fmt.Errorf("unexpected even evidence height: %d", ee.Height)
		}

		return nil
	}
}

type KeeperTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	querier sdk.Querier
	app     *simapp.SimApp
}

func (suite *KeeperTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)

	// recreate keeper in order to use custom testing types
	evidenceKeeper := evidence.NewKeeper(
		std.NewAppCodec(app.Codec()), app.GetKey(evidence.StoreKey),
		app.StakingKeeper, app.SlashingKeeper,
	)
	router := evidence.NewRouter()
	router = router.AddRoute(types.RouteEquivocation, testEquivocationHandler(*evidenceKeeper))
	evidenceKeeper.SetRouter(router)

	app.EvidenceKeeper = *evidenceKeeper

	suite.ctx = app.BaseApp.NewContext(checkTx, abci.Header{Height: 1})
	suite.querier = keeper.NewQuerier(*evidenceKeeper)
	suite.app = app

	for i, addr := range valAddresses {
		addr := sdk.AccAddress(addr)
		app.AccountKeeper.SetAccount(suite.ctx, auth.NewBaseAccount(addr, pubkeys[i], uint64(i), 0))
	}
}

func (suite *KeeperTestSuite) populateEvidence(ctx sdk.Context, numEvidence int) []exported.Evidence {
	evidence := make([]exported.Evidence, numEvidence)

	for i := 0; i < numEvidence; i++ {
		pk := ed25519.GenPrivKey()

		evidence[i] = types.Equivocation{
			Height:           11,
			Power:            100,
			Time:             time.Now().UTC(),
			ConsensusAddress: sdk.ConsAddress(pk.PubKey().Address().Bytes()),
		}

		suite.Nil(suite.app.EvidenceKeeper.SubmitEvidence(ctx, evidence[i]))
	}

	return evidence
}

func (suite *KeeperTestSuite) populateValidators(ctx sdk.Context) {
	// add accounts and set total supply
	totalSupplyAmt := initAmt.MulRaw(int64(len(valAddresses)))
	totalSupply := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, totalSupplyAmt))
	suite.app.BankKeeper.SetSupply(ctx, bank.NewSupply(totalSupply))

	for _, addr := range valAddresses {
		_, err := suite.app.BankKeeper.AddCoins(ctx, sdk.AccAddress(addr), initCoins)
		suite.NoError(err)
	}
}

func (suite *KeeperTestSuite) TestSubmitValidEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	pk := ed25519.GenPrivKey()

	e := types.Equivocation{
		Height:           1,
		Power:            100,
		Time:             time.Now().UTC(),
		ConsensusAddress: sdk.ConsAddress(pk.PubKey().Address().Bytes()),
	}

	suite.Nil(suite.app.EvidenceKeeper.SubmitEvidence(ctx, e))

	res, ok := suite.app.EvidenceKeeper.GetEvidence(ctx, e.Hash())
	suite.True(ok)
	suite.Equal(&e, res)
}

func (suite *KeeperTestSuite) TestSubmitValidEvidence_Duplicate() {
	ctx := suite.ctx.WithIsCheckTx(false)
	pk := ed25519.GenPrivKey()

	e := types.Equivocation{
		Height:           1,
		Power:            100,
		Time:             time.Now().UTC(),
		ConsensusAddress: sdk.ConsAddress(pk.PubKey().Address().Bytes()),
	}

	suite.Nil(suite.app.EvidenceKeeper.SubmitEvidence(ctx, e))
	suite.Error(suite.app.EvidenceKeeper.SubmitEvidence(ctx, e))

	res, ok := suite.app.EvidenceKeeper.GetEvidence(ctx, e.Hash())
	suite.True(ok)
	suite.Equal(&e, res)
}

func (suite *KeeperTestSuite) TestSubmitInvalidEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	pk := ed25519.GenPrivKey()
	e := types.Equivocation{
		Height:           0,
		Power:            100,
		Time:             time.Now().UTC(),
		ConsensusAddress: sdk.ConsAddress(pk.PubKey().Address().Bytes()),
	}

	suite.Error(suite.app.EvidenceKeeper.SubmitEvidence(ctx, e))

	res, ok := suite.app.EvidenceKeeper.GetEvidence(ctx, e.Hash())
	suite.False(ok)
	suite.Nil(res)
}

func (suite *KeeperTestSuite) TestIterateEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	numEvidence := 100
	suite.populateEvidence(ctx, numEvidence)

	evidence := suite.app.EvidenceKeeper.GetAllEvidence(ctx)
	suite.Len(evidence, numEvidence)
}

func (suite *KeeperTestSuite) TestGetEvidenceHandler() {
	handler, err := suite.app.EvidenceKeeper.GetEvidenceHandler(types.Equivocation{}.Route())
	suite.NoError(err)
	suite.NotNil(handler)

	handler, err = suite.app.EvidenceKeeper.GetEvidenceHandler("invalidHandler")
	suite.Error(err)
	suite.Nil(handler)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
