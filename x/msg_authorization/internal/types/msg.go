package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"
)

// MsgGrant grants the provided capability to the grantee on the granter's
// account with the provided expiration time.
type MsgGrant struct {
	Granter    sdk.AccAddress `json:"granter"`
	Grantee    sdk.AccAddress `json:"grantee"`
	Capability Capability     `json:"capability"`
	// Expiration specifies the expiration time of the grant
	Expiration time.Time `json:"expiration"`
}

// MsgRevoke revokes any capability with the provided sdk.Msg type on the
// granter's account with that has been granted to the grantee.
type MsgRevoke struct {
	Granter sdk.AccAddress `json:"granter"`
	Grantee sdk.AccAddress `json:"grantee"`
	// CapabilityMsgType is the type of sdk.Msg that the revoked Capability refers to.
	// i.e. this is what `Capability.MsgType()` returns
	CapabilityMsgType sdk.Msg `json:"capability_msg_type"`
}

// MsgExecDelegated attempts to execute the provided messages using
// capabilities granted to the grantee. Each message should have only
// one signer corresponding to the granter of the capability.
type MsgExecDelegated struct {
	Grantee sdk.AccAddress `json:"grantee"`
	Msgs    []sdk.Msg      `json:"msg"`
}
