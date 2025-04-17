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

	pubKeyAlgorithms map[string]PublicKeyAlgorithm
	vms              map[string]VM

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

	// AccountSequence maps account IDs to their sequence numbers.
	AccountSequence collections.Map[AccountID, uint64]

	UnorderedNonces collections.KeySet[collections.Pair[int64, []byte]]

	// TODO legacy vesting & other data
}

func (a AccountKeeper) DefinePublicKeyAlgorithm(algorithm PublicKeyAlgorithm) {}

func (a AccountKeeper) DefineVM(vm VM) {}

func (a AccountKeeper) CreateOrResolveAccountID(addressType AddressType, address Address) (AccountID, error) {
	panic("not implemented")
}

func (a AccountKeeper) ResolveAddress(addressType AddressType, accountID AccountID) (Address, error) {
	panic("not implemented")
}
