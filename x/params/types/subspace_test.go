package types_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	prototypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/types"
)

type SubspaceTestSuite struct {
	suite.Suite

	cdc   codec.BinaryMarshaler
	amino *codec.LegacyAmino
	ctx   sdk.Context
	ss    types.Subspace
}

func (suite *SubspaceTestSuite) SetupTest() {
	db := dbm.NewMemDB()

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkey, sdk.StoreTypeTransient, db)
	suite.NoError(ms.LoadLatestVersion())

	encCfg := simapp.MakeEncodingConfig()
	ss := types.NewSubspace(encCfg.Marshaler, key, tkey, "testsubspace")

	suite.cdc = encCfg.Marshaler
	suite.amino = encCfg.Amino
	suite.ctx = sdk.NewContext(ms, tmproto.Header{}, false, log.NewNopLogger())
	suite.ss = ss.WithKeyTable(paramKeyTable())
}

func (suite *SubspaceTestSuite) TestKeyTable() {
	suite.Require().True(suite.ss.HasKeyTable())
	suite.Require().Panics(func() {
		suite.ss.WithKeyTable(paramKeyTable())
	})
	suite.Require().NotPanics(func() {
		ss := types.NewSubspace(suite.cdc, key, tkey, "testsubspace2")
		ss = ss.WithKeyTable(paramKeyTable())
	})
}

func (suite *SubspaceTestSuite) TestGetSet() {
	var v prototypes.Duration
	t := prototypes.DurationProto(time.Hour * 48)

	suite.Require().Panics(func() {
		suite.ss.Get(suite.ctx, keyUnbondingTime, &v)
	})
	suite.Require().NotEqual(t, v)
	suite.Require().NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})
	suite.Require().NotPanics(func() {
		suite.ss.Get(suite.ctx, keyUnbondingTime, &v)
	})
	suite.Require().Equal(t, v)
}

func (suite *SubspaceTestSuite) TestGetIfExists() {
	var v prototypes.Duration

	suite.Require().NotPanics(func() {
		suite.ss.GetIfExists(suite.ctx, keyUnbondingTime, &v)
	})
	suite.Require().Equal(time.Duration(0), v)
}

func (suite *SubspaceTestSuite) TestGetRaw() {
	t := prototypes.DurationProto(time.Hour * 48)

	suite.Require().NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})
	suite.Require().NotPanics(func() {
		res := suite.ss.GetRaw(suite.ctx, keyUnbondingTime)
		suite.Require().Equal("2231373238303030303030303030303022", fmt.Sprintf("%X", res))
	})
}

func (suite *SubspaceTestSuite) TestHas() {
	t := prototypes.DurationProto(time.Hour * 48)

	suite.Require().False(suite.ss.Has(suite.ctx, keyUnbondingTime))
	suite.Require().NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})
	suite.Require().True(suite.ss.Has(suite.ctx, keyUnbondingTime))
}

func (suite *SubspaceTestSuite) TestModified() {
	t := prototypes.DurationProto(time.Hour * 48)

	suite.Require().False(suite.ss.Modified(suite.ctx, keyUnbondingTime))
	suite.Require().NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})
	suite.Require().True(suite.ss.Modified(suite.ctx, keyUnbondingTime))
}

func (suite *SubspaceTestSuite) TestUpdate() {
	suite.Require().Panics(func() {
		suite.ss.Update(suite.ctx, []byte("invalid_key"), nil) // nolint:errcheck
	})

	t := prototypes.DurationProto(time.Hour * 48)
	suite.Require().NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})

	bad := prototypes.DurationProto(time.Minute * 5)

	bz, err := suite.cdc.MarshalBinaryBare(bad)
	suite.Require().NoError(err)
	suite.Require().Error(suite.ss.Update(suite.ctx, keyUnbondingTime, bz))

	good := prototypes.DurationProto(time.Hour * 360)
	bz, err = suite.amino.MarshalJSON(good)
	suite.Require().NoError(err)
	suite.Require().NoError(suite.ss.Update(suite.ctx, keyUnbondingTime, bz))

	var v prototypes.Duration

	suite.Require().NotPanics(func() {
		suite.ss.Get(suite.ctx, keyUnbondingTime, &v)
	})
	suite.Require().Equal(good, v)
}

func (suite *SubspaceTestSuite) TestGetParamSet() {
	a := params{
		UnbondingTime: *prototypes.DurationProto(time.Hour * 48),
		MaxValidators: prototypes.UInt64Value{Value: 100},
		BondDenom:     prototypes.StringValue{Value: "stake"},
	}
	suite.Require().NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, &a.UnbondingTime)
		suite.ss.Set(suite.ctx, keyMaxValidators, &a.MaxValidators)
		suite.ss.Set(suite.ctx, keyBondDenom, &a.BondDenom)
	})

	b := params{}
	suite.Require().NotPanics(func() {
		suite.ss.GetParamSet(suite.ctx, &b)
	})
	suite.Require().Equal(a.UnbondingTime, b.UnbondingTime)
	suite.Require().Equal(a.MaxValidators, b.MaxValidators)
	suite.Require().Equal(a.BondDenom, b.BondDenom)
}


func (suite *SubspaceTestSuite) TestSetParamSet() {
	var testCases = []struct {
		name string
		ps   types.ParamSet
	}{
		{"invalid unbonding time", &params{
			*prototypes.DurationProto(time.Hour * 1),
			prototypes.UInt64Value{Value: 100},
			prototypes.StringValue{Value: "stake"}}},
		{"invalid bond denom", &params{
			*prototypes.DurationProto(time.Hour * 48),
			prototypes.UInt64Value{Value: 100},
			prototypes.StringValue{Value: ""}}},
	}
	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.Require().Panics(func() {
				suite.ss.SetParamSet(suite.ctx, tc.ps)
			})
		})
	}

	a := params{
		UnbondingTime: *prototypes.DurationProto(time.Hour * 48),
		MaxValidators: prototypes.UInt64Value{Value: 100},
		BondDenom:     prototypes.StringValue{Value: "stake"},
	}
	suite.Require().NotPanics(func() {
		suite.ss.SetParamSet(suite.ctx, &a)
	})

	b := params{}
	suite.Require().NotPanics(func() {
		suite.ss.GetParamSet(suite.ctx, &b)
	})
	suite.Require().Equal(a.UnbondingTime, b.UnbondingTime)
	suite.Require().Equal(a.MaxValidators, b.MaxValidators)
	suite.Require().Equal(a.BondDenom, b.BondDenom)
}

func (suite *SubspaceTestSuite) TestName() {
	suite.Require().Equal("testsubspace", suite.ss.Name())
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(SubspaceTestSuite))
}