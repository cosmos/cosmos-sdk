package keeper

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/circuit/types"
	"github.com/stretchr/testify/require"
)

func Test_msgServer_AuthorizeCircuitBreaker(t *testing.T) {

	ft := SetupFixture(t)

	srv := msgServer{
		Keeper: ft.Keeper,
	}

	// add a new super admin
	msg := &types.MsgAuthorizeCircuitBreaker{Granter: addresses[0], Grantee: addresses[1], Permissions: &types.Permissions{Level: types.Permissions_LEVEL_SUPER_ADMIN, LimitTypeUrls: []string{}}}
	_, err := srv.AuthorizeCircuitBreaker(ft.Ctx, msg)
	require.NoError(t, err)

	perms, err := ft.Keeper.GetPermissions(ft.Ctx, addresses[1])
	require.NoError(t, err)

	fmt.Println(perms)
}

// func Test_msgServer_TripCircuitBreaker(t *testing.T) {

// }

// func Test_msgServer_ResetCircuitBreaker(t *testing.T) {

// }
