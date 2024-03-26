package simulation

import (
	"math/rand"
	"time"

	v1 "cosmossdk.io/api/cosmos/gov/v1"
	"cosmossdk.io/core/address"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/authz"
	banktypes "cosmossdk.io/x/bank/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// genGrant returns a slice of authorization grants.
func genGrant(r *rand.Rand, accounts []simtypes.Account, genT time.Time, cdc address.Codec) []authz.GrantAuthorization {
	authorizations := make([]authz.GrantAuthorization, len(accounts)-1)
	for i := 0; i < len(accounts)-1; i++ {
		granter := accounts[i]
		grantee := accounts[i+1]
		var expiration *time.Time
		if i%3 != 0 { // generate some grants with no expire time
			e := genT.AddDate(1, 0, 0)
			expiration = &e
		}
		granterAddr, _ := cdc.BytesToString(granter.Address)
		granteeAddr, _ := cdc.BytesToString(grantee.Address)
		authorizations[i] = authz.GrantAuthorization{
			Granter:       granterAddr,
			Grantee:       granteeAddr,
			Authorization: generateRandomGrant(r),
			Expiration:    expiration,
		}
	}

	return authorizations
}

func generateRandomGrant(r *rand.Rand) *codectypes.Any {
	authorizations := make([]*codectypes.Any, 2)
	sendAuthz := banktypes.NewSendAuthorization(sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1000))), nil)
	authorizations[0] = newAnyAuthorization(sendAuthz)
	authorizations[1] = newAnyAuthorization(authz.NewGenericAuthorization(sdk.MsgTypeURL(&v1.MsgSubmitProposal{})))

	return authorizations[r.Intn(len(authorizations))]
}

func newAnyAuthorization(a authz.Authorization) *codectypes.Any {
	any, err := codectypes.NewAnyWithValue(a)
	if err != nil {
		panic(err)
	}

	return any
}

// RandomizedGenState generates a random GenesisState for authz.
func RandomizedGenState(simState *module.SimulationState) {
	var grants []authz.GrantAuthorization
	simState.AppParams.GetOrGenerate("authz", &grants, simState.Rand, func(r *rand.Rand) {
		grants = genGrant(r, simState.Accounts, simState.GenTimestamp, simState.AddressCodec)
	})

	authzGrantsGenesis := authz.NewGenesisState(grants)

	simState.GenState[authz.ModuleName] = simState.Cdc.MustMarshalJSON(authzGrantsGenesis)
}
