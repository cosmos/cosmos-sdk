package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

var granter = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
var grantee = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
var msgType = banktypes.MsgSend{}.Type()

func TestGrantkey(t *testing.T) {
	actor := GetActorAuthorizationKey(grantee, granter, msgType)
	granter1, grantee1 := ExtractAddressesFromGrantKey(actor)
	require.Equal(t, granter, granter1)
	require.Equal(t, grantee, grantee1)
}
