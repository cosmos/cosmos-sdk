package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
)

// define constants used for testing
const (
	validPort        = "testportid"
	invalidPort      = "(invalidport1)"
	invalidShortPort = "p"
	invalidLongPort  = "invalidlongportinvalidlongportinvalidlongportinvalidlongportinvalid"

	validChannel        = "testchannel"
	invalidChannel      = "(invalidchannel1)"
	invalidShortChannel = "invalid"
	invalidLongChannel  = "invalidlongchannelinvalidlongchannelinvalidlongchannelinvalidlongchannel"
)

var (
	addr1     = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addr2     = sdk.AccAddress("testaddr2").String()
	emptyAddr sdk.AccAddress

	coin             = sdk.NewCoin("atom", sdk.NewInt(100))
	ibcCoin          = sdk.NewCoin("ibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2", sdk.NewInt(100))
	invalidIBCCoin   = sdk.NewCoin("notibc/7F1D3FCF4AE79E1554D670D1AD949A9BA4E4A3C76C63093E17E446A46061A7A2", sdk.NewInt(100))
	invalidDenomCoin = sdk.Coin{Denom: "0atom", Amount: sdk.NewInt(100)}
	zeroCoin         = sdk.Coin{Denom: "atoms", Amount: sdk.NewInt(0)}

	timeoutHeight = clienttypes.NewHeight(0, 10)
)

// TestMsgTransferRoute tests Route for MsgTransfer
func TestMsgTransferRoute(t *testing.T) {
	msg := NewMsgTransfer(validPort, validChannel, coin, addr1, addr2, timeoutHeight, 0)

	require.Equal(t, RouterKey, msg.Route())
}

// TestMsgTransferType tests Type for MsgTransfer
func TestMsgTransferType(t *testing.T) {
	msg := NewMsgTransfer(validPort, validChannel, coin, addr1, addr2, timeoutHeight, 0)

	require.Equal(t, "transfer", msg.Type())
}

func TestMsgTransferGetSignBytes(t *testing.T) {
	msg := NewMsgTransfer(validPort, validChannel, coin, addr1, addr2, timeoutHeight, 0)
	expected := fmt.Sprintf(`{"type":"cosmos-sdk/MsgTransfer","value":{"receiver":"%s","sender":"%s","source_channel":"testchannel","source_port":"testportid","timeout_height":{"revision_height":"10"},"token":{"amount":"100","denom":"atom"}}}`, addr2, addr1)
	require.NotPanics(t, func() {
		res := msg.GetSignBytes()
		require.Equal(t, expected, string(res))
	})
}

// TestMsgTransferValidation tests ValidateBasic for MsgTransfer
func TestMsgTransferValidation(t *testing.T) {
	testCases := []struct {
		name    string
		msg     *MsgTransfer
		expPass bool
	}{
		{"valid msg with base denom", NewMsgTransfer(validPort, validChannel, coin, addr1, addr2, timeoutHeight, 0), true},
		{"valid msg with trace hash", NewMsgTransfer(validPort, validChannel, ibcCoin, addr1, addr2, timeoutHeight, 0), true},
		{"invalid ibc denom", NewMsgTransfer(validPort, validChannel, invalidIBCCoin, addr1, addr2, timeoutHeight, 0), false},
		{"too short port id", NewMsgTransfer(invalidShortPort, validChannel, coin, addr1, addr2, timeoutHeight, 0), false},
		{"too long port id", NewMsgTransfer(invalidLongPort, validChannel, coin, addr1, addr2, timeoutHeight, 0), false},
		{"port id contains non-alpha", NewMsgTransfer(invalidPort, validChannel, coin, addr1, addr2, timeoutHeight, 0), false},
		{"too short channel id", NewMsgTransfer(validPort, invalidShortChannel, coin, addr1, addr2, timeoutHeight, 0), false},
		{"too long channel id", NewMsgTransfer(validPort, invalidLongChannel, coin, addr1, addr2, timeoutHeight, 0), false},
		{"channel id contains non-alpha", NewMsgTransfer(validPort, invalidChannel, coin, addr1, addr2, timeoutHeight, 0), false},
		{"invalid denom", NewMsgTransfer(validPort, validChannel, invalidDenomCoin, addr1, addr2, timeoutHeight, 0), false},
		{"zero coin", NewMsgTransfer(validPort, validChannel, zeroCoin, addr1, addr2, timeoutHeight, 0), false},
		{"missing sender address", NewMsgTransfer(validPort, validChannel, coin, emptyAddr, addr2, timeoutHeight, 0), false},
		{"missing recipient address", NewMsgTransfer(validPort, validChannel, coin, addr1, "", timeoutHeight, 0), false},
		{"empty coin", NewMsgTransfer(validPort, validChannel, sdk.Coin{}, addr1, addr2, timeoutHeight, 0), false},
	}

	for i, tc := range testCases {
		err := tc.msg.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %s", i, tc.name)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

// TestMsgTransferGetSigners tests GetSigners for MsgTransfer
func TestMsgTransferGetSigners(t *testing.T) {
	msg := NewMsgTransfer(validPort, validChannel, coin, addr1, addr2, timeoutHeight, 0)
	res := msg.GetSigners()

	require.Equal(t, []sdk.AccAddress{addr1}, res)
}
