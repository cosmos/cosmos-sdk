// nolint
package keeper // noalias

// DONTCOVER

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
)

// dummy addresses used for testing
var (
	delPk1   = newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB51")
	delPk2   = newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB50")
	delPk3   = newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB52")
	delAddr1 = sdk.AccAddress(delPk1.Address())
	delAddr2 = sdk.AccAddress(delPk2.Address())
	delAddr3 = sdk.AccAddress(delPk3.Address())

	valOpPk1    = newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB53")
	valOpPk2    = newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB54")
	valOpPk3    = newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB55")
	valOpAddr1  = sdk.ValAddress(valOpPk1.Address())
	valOpAddr2  = sdk.ValAddress(valOpPk2.Address())
	valOpAddr3  = sdk.ValAddress(valOpPk3.Address())
	valAccAddr1 = sdk.AccAddress(valOpPk1.Address())
	valAccAddr2 = sdk.AccAddress(valOpPk2.Address())
	valAccAddr3 = sdk.AccAddress(valOpPk3.Address())

	TestAddrs = []sdk.AccAddress{
		delAddr1, delAddr2, delAddr3,
		valAccAddr1, valAccAddr2, valAccAddr3,
	}

	emptyDelAddr sdk.AccAddress
	emptyValAddr sdk.ValAddress
	emptyPubkey  crypto.PubKey
)

// TODO: remove dependency with staking
var (
	TestProposal        = types.NewTextProposal("Test", "description")
	TestDescription     = staking.NewDescription("T", "E", "S", "T", "Z")
	TestCommissionRates = staking.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
)

// TODO move to common testing framework
func newPubKey(pk string) (res crypto.PubKey) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		panic(err)
	}
	var pkEd ed25519.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	return pkEd
}

func makeTestCodec() *codec.Codec {
	var cdc = codec.New()
	auth.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	supply.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}

func createTestInput(t *testing.T, isCheckTx bool, initPower int64) (sdk.Context, auth.AccountKeeper, Keeper, staking.Keeper, types.SupplyKeeper) {

	initTokens := sdk.TokensFromConsensusPower(initPower)

	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyGov := sdk.NewKVStoreKey(types.StoreKey)
	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	keySupply := sdk.NewKVStoreKey(supply.StoreKey)
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keySupply, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyGov, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	require.Nil(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "gov-chain"}, isCheckTx, log.NewNopLogger())
	ctx = ctx.WithConsensusParams(
		&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519},
			},
		},
	)
	cdc := makeTestCodec()

	maccPerms := map[string][]string{
		auth.FeeCollectorName:     nil,
		types.ModuleName:          nil,
		staking.NotBondedPoolName: {supply.Burner, supply.Staking},
		staking.BondedPoolName:    {supply.Burner, supply.Staking},
	}

	// create module accounts
	feeCollectorAcc := supply.NewEmptyModuleAccount(auth.FeeCollectorName)
	govAcc := supply.NewEmptyModuleAccount(types.ModuleName, supply.Burner)
	notBondedPool := supply.NewEmptyModuleAccount(staking.NotBondedPoolName, supply.Burner, supply.Staking)
	bondPool := supply.NewEmptyModuleAccount(staking.BondedPoolName, supply.Burner, supply.Staking)

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[feeCollectorAcc.GetAddress().String()] = true
	blacklistedAddrs[govAcc.GetAddress().String()] = true
	blacklistedAddrs[notBondedPool.GetAddress().String()] = true
	blacklistedAddrs[bondPool.GetAddress().String()] = true

	pk := params.NewKeeper(cdc, keyParams, tkeyParams)
	accountKeeper := auth.NewAccountKeeper(cdc, keyAcc, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	bankKeeper := bank.NewBaseKeeper(accountKeeper, pk.Subspace(bank.DefaultParamspace), blacklistedAddrs)
	supplyKeeper := supply.NewKeeper(cdc, keySupply, accountKeeper, bankKeeper, maccPerms)

	sk := staking.NewKeeper(cdc, keyStaking, supplyKeeper, pk.Subspace(staking.DefaultParamspace))
	sk.SetParams(ctx, staking.DefaultParams())

	rtr := types.NewRouter().
		AddRoute(types.RouterKey, types.ProposalHandler)

	keeper := NewKeeper(
		cdc, keyGov, pk.Subspace(types.DefaultParamspace).WithKeyTable(types.ParamKeyTable()), supplyKeeper, sk, rtr,
	)

	keeper.SetProposalID(ctx, types.DefaultStartingProposalID)
	keeper.SetDepositParams(ctx, types.DefaultDepositParams())
	keeper.SetVotingParams(ctx, types.DefaultVotingParams())
	keeper.SetTallyParams(ctx, types.DefaultTallyParams())

	initCoins := sdk.NewCoins(sdk.NewCoin(sk.BondDenom(ctx), initTokens))
	totalSupply := sdk.NewCoins(sdk.NewCoin(sk.BondDenom(ctx), initTokens.MulRaw(int64(len(TestAddrs)))))
	supplyKeeper.SetSupply(ctx, supply.NewSupply(totalSupply))

	for _, addr := range TestAddrs {
		_, err := bankKeeper.AddCoins(ctx, addr, initCoins)
		require.Nil(t, err)
	}

	keeper.supplyKeeper.SetModuleAccount(ctx, feeCollectorAcc)
	keeper.supplyKeeper.SetModuleAccount(ctx, govAcc)
	keeper.supplyKeeper.SetModuleAccount(ctx, bondPool)
	keeper.supplyKeeper.SetModuleAccount(ctx, notBondedPool)

	return ctx, accountKeeper, keeper, sk, supplyKeeper
}

// ProposalEqual checks if two proposals are equal (note: slow, for tests only)
func ProposalEqual(proposalA types.Proposal, proposalB types.Proposal) bool {
	return bytes.Equal(types.ModuleCdc.MustMarshalBinaryBare(proposalA),
		types.ModuleCdc.MustMarshalBinaryBare(proposalB))
}

func createValidators(ctx sdk.Context, sk staking.Keeper, powers []int64) {
	val1 := staking.NewValidator(valOpAddr1, valOpPk1, staking.Description{})
	val2 := staking.NewValidator(valOpAddr2, valOpPk2, staking.Description{})
	val3 := staking.NewValidator(valOpAddr3, valOpPk3, staking.Description{})

	sk.SetValidator(ctx, val1)
	sk.SetValidator(ctx, val2)
	sk.SetValidator(ctx, val3)
	sk.SetValidatorByConsAddr(ctx, val1)
	sk.SetValidatorByConsAddr(ctx, val2)
	sk.SetValidatorByConsAddr(ctx, val3)
	sk.SetNewValidatorByPowerIndex(ctx, val1)
	sk.SetNewValidatorByPowerIndex(ctx, val2)
	sk.SetNewValidatorByPowerIndex(ctx, val3)

	_, _ = sk.Delegate(ctx, valAccAddr1, sdk.TokensFromConsensusPower(powers[0]), sdk.Unbonded, val1, true)
	_, _ = sk.Delegate(ctx, valAccAddr2, sdk.TokensFromConsensusPower(powers[1]), sdk.Unbonded, val2, true)
	_, _ = sk.Delegate(ctx, valAccAddr3, sdk.TokensFromConsensusPower(powers[2]), sdk.Unbonded, val3, true)

	_ = staking.EndBlocker(ctx, sk)
}
