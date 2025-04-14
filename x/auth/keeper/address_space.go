package keeper

import (
	"context"
	"encoding/binary"
	"errors"

	"cosmossdk.io/collections"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// AddressSpaceManager is an interface for managing foreign address spaces.
type AddressSpaceManager interface {
	// Name is the name of the address space.
	Name() string

	// DeriveAddress derives an address from the given account ID and public key.
	DeriveAddress(id AccountID, pubKey cryptotypes.PubKey) Address
}

var (
	ErrUnknownAddressSpace = errors.New("unknown address space")
)

// DefineAddressSpace defines a new address space with the given prefix and manager.
// In the first upgrade where a new address space is defined, the Migrator.AddAdressSpace migration
// must be performed for this new address space.
func (ak AccountKeeper) DefineAddressSpace(prefix AddressSpacePrefix, manager AddressSpaceManager) {
	if _, ok := ak.addressSpaceManagers[prefix]; ok {
		panic("address space already defined")
	}
	ak.addressSpaceManagers[prefix] = manager

	if _, ok := ak.addressPrefixByName[manager.Name()]; ok {
		panic("address space already defined")
	}
	ak.addressPrefixByName[manager.Name()] = prefix
}

// ResolveAddress resolves an address from the given account ID and address space.
func (ak AccountKeeper) ResolveAddress(ctx context.Context, addressSpace string, id AccountID) (Address, error) {
	if addressSpace == "" {
		val, err := ak.Accounts.Indexes.Number.MatchExact(ctx, accountIdToNum(id))
		if err != nil {
			return nil, err
		}
		return val, nil
	}

	addrTyp, ok := ak.addressPrefixByName[addressSpace]
	if !ok {
		return nil, ErrUnknownAddressSpace
	}

	return ak.AddressByAccountID.Get(ctx, collections.Join(id, addrTyp))
}

// ResolveAccountID resolves an account ID from the given address and address space.
func (ak AccountKeeper) ResolveAccountID(ctx context.Context, addressSpace string, address Address) (AccountID, error) {
	if addressSpace == "" {
		val, err := ak.Accounts.Get(ctx, address)
		if err != nil {
			return nil, err
		}
		acctNum := val.GetAccountNumber()
		return accountNumToId(acctNum), nil
	}

	addrTyp, ok := ak.addressPrefixByName[addressSpace]
	if !ok {
		return nil, ErrUnknownAddressSpace
	}

	return ak.AccountIDByAddress.Get(ctx, collections.Join(addrTyp, address))
}

func accountNumToId(acctNum uint64) AccountID {
	id := make([]byte, 8)
	binary.LittleEndian.PutUint64(id, acctNum)
	return id
}

func accountIdToNum(id AccountID) uint64 {
	if len(id) != 8 {
		panic("account ID must currently be 8 bytes")
	}
	return binary.LittleEndian.Uint64(id)
}
