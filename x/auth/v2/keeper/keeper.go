package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

type AccountKeeper struct {
	addressCodec address.Codec

	storeService store.KVStoreService
	cdc          codec.BinaryCodec
	bech32Prefix string

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	// State
	Schema        collections.Schema
	Params        collections.Item[types.Params]
	AccountNumber collections.Sequence

	// AccountCode represents the public key, abstract account code, or module key representing the account.
	// Public key accounts will normally have no entry here until they are claimed in a signed transaction.
	AccountCode collections.Map[AccountID, CodeData]
	// AccountIDByAddress maps addresses to account IDs.
	AccountIDByAddress collections.Map[collections.Pair[Address, AddressType], AccountID]
	//  AddressByAccountID maps account IDs to addresses.
	AddressByAccountID collections.Map[collections.Pair[AccountID, AddressType], Address]

	UnorderedNonces collections.KeySet[collections.Pair[int64, []byte]]
	// TODO legacy vesting & other data
}

// CodeData represents either a public key or a pointer to an account's implementation code in some VM or module.
type CodeData struct {
	// Type represents the type of the data. This may be a public key algorithm, or a registered VM or module.
	Type string
	Data []byte
}
