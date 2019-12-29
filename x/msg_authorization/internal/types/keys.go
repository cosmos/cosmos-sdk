package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the module name constant used in many places
	ModuleName = "msg_authorization"

	// StoreKey is the store key string for msg_authorization
	StoreKey = ModuleName

	// RouterKey is the message route for msg_authorization
	RouterKey = ModuleName

	// QuerierRoute is the querier route for msg_authorization
	QuerierRoute = ModuleName
)

func GetActorAuthorizationKey(grantee sdk.AccAddress, granter sdk.AccAddress, msg sdk.Msg) []byte {
	return []byte(fmt.Sprintf("c/%x/%x/%s/%s", grantee, granter, msg.Route(), msg.Type()))
}
