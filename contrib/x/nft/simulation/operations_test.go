package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"

	sdkapp "github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	nft "github.com/cosmos/cosmos-sdk/contrib/x/nft"
	nftkeeper "github.com/cosmos/cosmos-sdk/contrib/x/nft/keeper"
	nftmodule "github.com/cosmos/cosmos-sdk/contrib/x/nft/module"
	"github.com/cosmos/cosmos-sdk/contrib/x/nft/simulation"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// nftAppModule wraps nftmodule.AppModule to satisfy the sdkapp.Module interface.
type nftAppModule struct {
	nftmodule.AppModule
	storeKey *storetypes.KVStoreKey
}

func (m nftAppModule) StoreKeys() map[string]*storetypes.KVStoreKey {
	return map[string]*storetypes.KVStoreKey{m.storeKey.Name(): m.storeKey}
}

func (nftAppModule) ModuleAccountPermissions() map[string][]string {
	return map[string][]string{nft.ModuleName: nil}
}

var _ sdkapp.Module = nftAppModule{}

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context

	app               *sdkapp.SDKApp
	codec             codec.Codec
	interfaceRegistry codectypes.InterfaceRegistry
	txConfig          client.TxConfig
	accountKeeper     authkeeper.AccountKeeper
	bankKeeper        bankkeeper.Keeper
	stakingKeeper     *stakingkeeper.Keeper
	nftKeeper         nftkeeper.Keeper
}

// setupWithNFT constructs an SDKApp with the NFT module added and returns the
// initialized app plus the nft keeper.
func setupWithNFT(t *testing.T) (*sdkapp.SDKApp, nftkeeper.Keeper) {
	t.Helper()

	opts := simtestutil.AppOptionsMap{
		flags.FlagHome:    t.TempDir(),
		flags.FlagChainID: "test-chain",
	}
	cfg := sdkapp.DefaultSDKAppConfig("app", opts)
	ta := sdkapp.NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)

	// Pre-register NFT module account permission so NewKeeper can validate it
	// before AddModules+LoadModules run. AddModules will add it to moduleAccountPerms
	// and LoadModules will persist it via LoadMaccPerms.
	ta.AccountKeeper.AddModuleAccountPerm(nft.ModuleName, nil)

	nftStoreKey := storetypes.NewKVStoreKey(nft.StoreKey)
	nftSvc := sdk.NewKVStoreService(nftStoreKey)
	k := nftkeeper.NewKeeper(nftSvc, ta.AppCodec(), ta.AccountKeeper, ta.BankKeeper)
	nftMod := nftAppModule{
		AppModule: nftmodule.NewAppModule(ta.AppCodec(), k, ta.AccountKeeper, ta.BankKeeper, ta.InterfaceRegistry()),
		storeKey:  nftStoreKey,
	}
	if err := ta.AddModules(nftMod); err != nil {
		t.Fatalf("failed to add nft module: %v", err)
	}

	ta.LoadModules()

	if err := ta.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}

	// Build a minimal genesis with one validator and one funded account.
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	if err != nil {
		t.Fatalf("failed to get pub key: %v", err)
	}
	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	priv := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(priv.PubKey().Address().Bytes(), priv.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000000000))),
	}

	genesisState, err := buildGenesis(ta.AppCodec(), ta.DefaultGenesis(), valSet, []authtypes.GenesisAccount{acc}, balance)
	if err != nil {
		t.Fatalf("failed to build genesis: %v", err)
	}

	stateBytes, err := cmtjson.MarshalIndent(genesisState, "", " ")
	if err != nil {
		t.Fatalf("failed to marshal genesis: %v", err)
	}

	if _, err := ta.InitChain(&abci.RequestInitChain{
		ChainId:         "test-chain",
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: simtestutil.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	}); err != nil {
		t.Fatalf("failed to init chain: %v", err)
	}

	if _, err := ta.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             ta.LastBlockHeight() + 1,
		NextValidatorsHash: valSet.Hash(),
	}); err != nil {
		t.Fatalf("failed to finalize block: %v", err)
	}

	return ta, k
}

func buildGenesis(
	cdc codec.Codec,
	genesisState map[string]json.RawMessage,
	valSet *cmttypes.ValidatorSet,
	genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) (map[string]json.RawMessage, error) {
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = cdc.MustMarshalJSON(authGenesis)

	bondAmt := sdk.DefaultPowerReduction
	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
		if err != nil {
			return nil, err
		}
		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			return nil, err
		}
		validators = append(validators, stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdkmath.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()),
			MinSelfDelegation: sdkmath.ZeroInt(),
		})
		delegations = append(delegations, stakingtypes.NewDelegation(
			genAccs[0].GetAddress().String(),
			sdk.ValAddress(val.Address).String(),
			sdkmath.LegacyOneDec(),
		))
	}

	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		totalSupply = totalSupply.Add(b.Coins...)
	}
	for range delegations {
		totalSupply = totalSupply.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))
	}
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
	})
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = cdc.MustMarshalJSON(bankGenesis)

	return genesisState, nil
}

func (suite *SimTestSuite) SetupTest() {
	ta, k := setupWithNFT(suite.T())

	suite.app = ta
	suite.codec = ta.AppCodec()
	suite.interfaceRegistry = ta.InterfaceRegistry()
	suite.txConfig = ta.TxConfig()
	suite.accountKeeper = ta.AccountKeeper
	suite.bankKeeper = ta.BankKeeper
	suite.stakingKeeper = ta.StakingKeeper
	suite.nftKeeper = k
	suite.ctx = ta.NewContext(false)
}

func (suite *SimTestSuite) TestWeightedOperations() {
	weightedOps := simulation.WeightedOperations(
		suite.interfaceRegistry,
		make(simtypes.AppParams),
		suite.codec,
		suite.txConfig,
		suite.accountKeeper,
		suite.bankKeeper,
		suite.nftKeeper,
	)

	// setup 3 accounts
	s := rand.NewSource(1)
	r := rand.New(s)
	accs := suite.getTestingAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simulation.WeightSend, nft.ModuleName, simulation.TypeMsgSend},
	}

	for i, w := range weightedOps {
		operationMsg, _, err := w.Op()(r, suite.app.BaseApp, suite.ctx, accs, suite.app.ChainID())
		suite.Require().NoError(err)

		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		suite.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		suite.Require().Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		suite.Require().Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 200000)
	initCoins := sdk.NewCoins(sdk.NewCoin("stake", initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.accountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.accountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(banktestutil.FundAccount(suite.ctx, suite.bankKeeper, account.Address, initCoins))
	}

	return accounts
}

func (suite *SimTestSuite) TestSimulateMsgSend() {
	s := rand.NewSource(1)
	r := rand.New(s)
	accounts := suite.getTestingAccounts(r, 2)
	blockTime := time.Now().UTC()
	ctx := suite.ctx.WithBlockTime(blockTime)

	// begin new block
	_, err := suite.app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: suite.app.LastBlockHeight() + 1,
		Hash:   suite.app.LastCommitID().Hash,
	})
	suite.Require().NoError(err)

	// execute operation
	registry := suite.interfaceRegistry
	op := simulation.SimulateMsgSend(codec.NewProtoCodec(registry), suite.txConfig, suite.accountKeeper, suite.bankKeeper, suite.nftKeeper)
	operationMsg, futureOperations, err := op(r, suite.app.BaseApp, ctx, accounts, suite.app.ChainID())
	suite.Require().NoError(err)

	var msg nft.MsgSend
	_ = suite.codec.UnmarshalJSON(operationMsg.Msg, &msg)
	suite.Require().True(operationMsg.OK)
	suite.Require().Len(futureOperations, 0)
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
