package group

import (
	"bytes"
	"fmt"
	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/gherkin"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/regen-network/regen-ledger/util"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	"testing"
)

var cdc *codec.Codec
var ctx sdk.Context
var keeper Keeper

func setupTestInput() {
	db := dbm.NewMemDB()

	cdc = codec.New()
	auth.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	RegisterCodec(cdc)

	paramsKey := sdk.NewKVStoreKey("params")
	tparamsKey := sdk.NewTransientStoreKey("tparams")
	accKey := sdk.NewKVStoreKey("acc")
	groupKey := sdk.NewKVStoreKey("groupKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(accKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(paramsKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(groupKey, sdk.StoreTypeIAVL, db)
	_ = ms.LoadLatestVersion()

	paramsKeeper := params.NewKeeper(cdc, paramsKey, tparamsKey, params.DefaultCodespace)
	accKeeper := auth.NewAccountKeeper(cdc, accKey, paramsKeeper.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	keeper = NewKeeper(groupKey, cdc, accKeeper, nil)
	ctx = sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())
}

var privKey secp256k1.PrivKeySecp256k1

var pubKey crypto.PubKey

var myAddr sdk.AccAddress

var group Group

var groupId sdk.AccAddress

func aPublicKeyAddress() error {
	privKey = secp256k1.GenPrivKey()
	pubKey = privKey.PubKey()
	myAddr = sdk.AccAddress(pubKey.Address())
	return nil
}

func aUserCreatesAGroupWithThatAddressAndADecisionThresholdOfOne() error {
	mem := Member{
		Address: myAddr,
		Weight:  sdk.NewInt(1),
	}
	group = Group{
		Members:           []Member{mem},
		DecisionThreshold: sdk.NewInt(1),
	}
	var err error
	groupId, err = keeper.CreateGroup(ctx, group)
	return err
}

func theyShouldGetANewGroupAddressBack() error {
	if groupId == nil || len(groupId) <= 0 {
		return fmt.Errorf("group ID was empty")
	}
	return nil
}

func beAbleToRetrieveTheGroupDetailsWithThatAddress() error {
	groupRetrieved, err := keeper.GetGroupInfo(ctx, groupId)
	if err != nil {
		return fmt.Errorf("error retrieving group info %+v", err)
	}
	if !group.DecisionThreshold.Equal(groupRetrieved.DecisionThreshold) {
		return fmt.Errorf("got wrong DecisionThreshold")
	}
	if len(group.Members) != len(groupRetrieved.Members) {
		return fmt.Errorf("wrong number of members")
	}
	for i, mem := range group.Members {
		memRetrieved := groupRetrieved.Members[i]
		if !bytes.Equal(mem.Address, memRetrieved.Address) {
			return fmt.Errorf("wrong member GeoAddress")
		}
		if !mem.Weight.Equal(memRetrieved.Weight) {
			return fmt.Errorf("wrong member Weight")
		}
	}
	return nil
}

func authorizationShouldSucceedWithOnlyThereVote() error {
	if !keeper.Authorize(ctx, groupId, []sdk.AccAddress{myAddr}) {
		return fmt.Errorf("auth failed")
	}
	return nil
}

func authorizationShouldFailWithNoVotes() error {
	if keeper.Authorize(ctx, groupId, []sdk.AccAddress{}) {
		return fmt.Errorf("auth succeeded, but should fail")
	}
	return nil
}

func authorizationShouldFailWithAnyOtherVotes() error {
	otherAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	if keeper.Authorize(ctx, groupId, []sdk.AccAddress{otherAddr}) {
		return fmt.Errorf("auth succeeded, but should fail")
	}
	return nil
}

func TestMain(m *testing.M) {
	util.GodogMain(m, "group", FeatureContext)
}

func FeatureContext(s *godog.Suite) {
	s.BeforeFeature(func(*gherkin.Feature) {
		setupTestInput()
	})
	s.Step(`^a public key address$`, aPublicKeyAddress)
	s.Step(`^they should get a new group address back$`, theyShouldGetANewGroupAddressBack)
	s.Step(`^a user creates a group with that address and a decision threshold of 1$`, aUserCreatesAGroupWithThatAddressAndADecisionThresholdOfOne)
	s.Step(`^be able to retrieve the group details with that address$`, beAbleToRetrieveTheGroupDetailsWithThatAddress)
	s.Step(`^authorization should succeed with only there vote$`, authorizationShouldSucceedWithOnlyThereVote)
	s.Step(`^authorization should fail with no votes$`, authorizationShouldFailWithNoVotes)
	s.Step(`^authorization should fail with any other votes$`, authorizationShouldFailWithAnyOtherVotes)
}
