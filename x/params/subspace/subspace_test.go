package subspace_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

type SubspaceTestSuite struct {
	suite.Suite

	cdc *codec.Codec
	ctx sdk.Context
	ss  subspace.Subspace
}

func (suite *SubspaceTestSuite) SetupTest() {
	cdc := codec.New()
	db := dbm.NewMemDB()

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkey, sdk.StoreTypeTransient, db)
	suite.NoError(ms.LoadLatestVersion())

	ss := subspace.NewSubspace(cdc, key, tkey, "testsubspace")

	suite.cdc = cdc
	suite.ctx = sdk.NewContext(ms, abci.Header{}, false, log.NewNopLogger())
	suite.ss = ss.WithKeyTable(paramKeyTable())
}

func (suite *SubspaceTestSuite) TestKeyTable() {
	suite.True(suite.ss.HasKeyTable())
	suite.Panics(func() {
		suite.ss.WithKeyTable(paramKeyTable())
	})
	suite.NotPanics(func() {
		ss := subspace.NewSubspace(codec.New(), key, tkey, "testsubspace2")
		ss = ss.WithKeyTable(paramKeyTable())
	})
}

func (suite *SubspaceTestSuite) TestGetSet() {
	var v time.Duration
	t := time.Hour * 48

	suite.Panics(func() {
		suite.ss.Get(suite.ctx, keyUnbondingTime, &v)
	})
	suite.NotEqual(t, v)
	suite.NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})
	suite.NotPanics(func() {
		suite.ss.Get(suite.ctx, keyUnbondingTime, &v)
	})
	suite.Equal(t, v)
}

func (suite *SubspaceTestSuite) TestGetIfExists() {
	var v time.Duration

	suite.NotPanics(func() {
		suite.ss.GetIfExists(suite.ctx, keyUnbondingTime, &v)
	})
	suite.Equal(time.Duration(0), v)
}

func (suite *SubspaceTestSuite) TestGetRaw() {
	t := time.Hour * 48

	suite.NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})
	suite.NotPanics(func() {
		res := suite.ss.GetRaw(suite.ctx, keyUnbondingTime)
		suite.Equal("2231373238303030303030303030303022", fmt.Sprintf("%X", res))
	})
}

func (suite *SubspaceTestSuite) TestHas() {
	t := time.Hour * 48

	suite.False(suite.ss.Has(suite.ctx, keyUnbondingTime))
	suite.NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})
	suite.True(suite.ss.Has(suite.ctx, keyUnbondingTime))
}

func (suite *SubspaceTestSuite) TestModified() {
	t := time.Hour * 48

	suite.False(suite.ss.Modified(suite.ctx, keyUnbondingTime))
	suite.NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})
	suite.True(suite.ss.Modified(suite.ctx, keyUnbondingTime))
}

func (suite *SubspaceTestSuite) TestUpdate() {
	suite.Panics(func() {
		suite.ss.Update(suite.ctx, []byte("invalid_key"), nil)
	})

	t := time.Hour * 48
	suite.NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, t)
	})

	bad := time.Minute * 5

	bz, err := suite.cdc.MarshalJSON(bad)
	suite.NoError(err)
	suite.Error(suite.ss.Update(suite.ctx, keyUnbondingTime, bz))

	good := time.Hour * 360
	bz, err = suite.cdc.MarshalJSON(good)
	suite.NoError(err)
	suite.NoError(suite.ss.Update(suite.ctx, keyUnbondingTime, bz))

	var v time.Duration

	suite.NotPanics(func() {
		suite.ss.Get(suite.ctx, keyUnbondingTime, &v)
	})
	suite.Equal(good, v)
}

func (suite *SubspaceTestSuite) TestGetParamSet() {
	a := params{
		UnbondingTime: time.Hour * 48,
		MaxValidators: 100,
		BondDenom:     "stake",
	}
	suite.NotPanics(func() {
		suite.ss.Set(suite.ctx, keyUnbondingTime, a.UnbondingTime)
		suite.ss.Set(suite.ctx, keyMaxValidators, a.MaxValidators)
		suite.ss.Set(suite.ctx, keyBondDenom, a.BondDenom)
	})

	b := params{}
	suite.NotPanics(func() {
		suite.ss.GetParamSet(suite.ctx, &b)
	})
	suite.Equal(a.UnbondingTime, b.UnbondingTime)
	suite.Equal(a.MaxValidators, b.MaxValidators)
	suite.Equal(a.BondDenom, b.BondDenom)
}

func (suite *SubspaceTestSuite) TestSetParamSet() {
	testCases := []struct {
		name string
		ps   subspace.ParamSet
	}{
		{"invalid unbonding time", &params{time.Hour * 1, 100, "stake"}},
		{"invalid bond denom", &params{time.Hour * 48, 100, ""}},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.Panics(func() {
				suite.ss.SetParamSet(suite.ctx, tc.ps)
			})
		})
	}

	a := params{
		UnbondingTime: time.Hour * 48,
		MaxValidators: 100,
		BondDenom:     "stake",
	}
	suite.NotPanics(func() {
		suite.ss.SetParamSet(suite.ctx, &a)
	})

	b := params{}
	suite.NotPanics(func() {
		suite.ss.GetParamSet(suite.ctx, &b)
	})
	suite.Equal(a.UnbondingTime, b.UnbondingTime)
	suite.Equal(a.MaxValidators, b.MaxValidators)
	suite.Equal(a.BondDenom, b.BondDenom)
}

func (suite *SubspaceTestSuite) TestName() {
	suite.Equal("testsubspace", suite.ss.Name())
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(SubspaceTestSuite))
}
