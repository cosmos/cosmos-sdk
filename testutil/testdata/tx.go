package testdata

import (
	"encoding/json"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// AddressGenerator creates and returns a random address generator using rapid.
func AddressGenerator(t *rapid.T) *rapid.Generator[sdk.AccAddress] {
	return rapid.Custom(func(t *rapid.T) sdk.AccAddress {
		pkBz := rapid.SliceOfN(rapid.Byte(), 20, 20).Draw(t, "hex")
		return sdk.AccAddress(pkBz)
	})
}

// PaginationGenerator creates and returns a pagination PageRequest generator
// using rapid.
func PaginationGenerator(t *rapid.T, maxLimit uint64) *rapid.Generator[*query.PageRequest] {
	return rapid.Custom(func(t *rapid.T) *query.PageRequest {
		return &query.PageRequest{
			Offset:     rapid.Uint64Range(0, maxLimit).Draw(t, "offset"),
			Limit:      rapid.Uint64Range(0, maxLimit).Draw(t, "limit"),
			CountTotal: rapid.Bool().Draw(t, "count-total"),
			Reverse:    rapid.Bool().Draw(t, "reverse"),
		}
	})
}

// KeyTestPubAddr generates a new secp256k1 keypair.
func KeyTestPubAddr() (cryptotypes.PrivKey, cryptotypes.PubKey, sdk.AccAddress) {
	key := secp256k1.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

// KeyTestPubAddr generates a new secp256r1 keypair.
func KeyTestPubAddrSecp256R1(require *require.Assertions) (cryptotypes.PrivKey, cryptotypes.PubKey, sdk.AccAddress) {
	key, err := secp256r1.GenPrivKey()
	require.NoError(err)
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

// NewTestFeeAmount is a test fee amount.
func NewTestFeeAmount() sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
}

// NewTestGasLimit is a test fee gas limit.
func NewTestGasLimit() uint64 {
	return 200000
}

// NewTestMsg creates a message for testing with the given signers.
func NewTestMsg(addrs ...sdk.AccAddress) *TestMsg {
	var accAddresses []string

	for _, addr := range addrs {
		accAddresses = append(accAddresses, addr.String())
	}

	return &TestMsg{
		Signers: accAddresses,
	}
}

var _ sdk.Msg = (*TestMsg)(nil)

func (msg *TestMsg) Route() string { return "TestMsg" }
func (msg *TestMsg) Type() string  { return "Test message" }
func (msg *TestMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg.Signers)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(bz)
}

func (msg *TestMsg) GetSigners() []sdk.AccAddress {
	signers := make([]sdk.AccAddress, 0, len(msg.Signers))
	for _, addr := range msg.Signers {
		a, _ := sdk.AccAddressFromBech32(addr)
		signers = append(signers, a)
	}
	return signers
}

func (msg *TestMsg) ValidateBasic() error {
	for _, addr := range msg.Signers {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid signer address: %s", err)
		}
	}
	return nil
}

var _ sdk.Msg = &MsgCreateDog{}

func (msg *MsgCreateDog) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{} }
func (msg *MsgCreateDog) ValidateBasic() error         { return nil }
