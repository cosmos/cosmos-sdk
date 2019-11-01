package types

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

const (
	// SubModuleName defines the IBC transfer name
	SubModuleName = "transfer"

	// StoreKey is the store key string for IBC transfer
	StoreKey = SubModuleName

	// RouterKey is the message route for IBC transfer
	RouterKey = SubModuleName

	// QuerierRoute is the querier route for IBC transfer
	QuerierRoute = SubModuleName

	// BoundPortID defines the name of the capability key
	BoundPortID = "bankbankbank"
)

// GetEscrowAddress returns the escrow address for the specified channel
//
// CONTRACT: this assumes that there's only one bank bridge module that owns the
// port associated with the channel ID so that the address created is actually
// unique.
func GetEscrowAddress(portID, channelID string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(portID + channelID)))
}

// GetDenomPrefix returns the receiving denomination prefix
func GetDenomPrefix(portID, channelID string) string {
	return fmt.Sprintf("%s/%s/", portID, channelID)
}

// GetModuleAccountName returns the IBC transfer module account name for supply
func GetModuleAccountName() string {
	return fmt.Sprintf("%s/%s", ibctypes.ModuleName, SubModuleName)
}
