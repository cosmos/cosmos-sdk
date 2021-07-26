package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper is the interface contract that x/auth's keeper implements.
type AccountKeeper interface {
	// Return a new account with the next account number and the specified address. Does not save the new account to the store.
	NewAccountWithAddress(context.Context, sdk.AccAddress) sdk.AccountI

	// Return a new account with the next account number. Does not save the new account to the store.
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI

	// Check if an account exists in the store.
	HasAccount(context.Context, sdk.AccAddress) bool

	// Retrieve an account from the store.
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI

	// GetAllAccounts returns all accounts in the accountKeeper.
	GetAllAccounts(sdk.Context) []types.AccountI

	// Set an account in the store.
	SetAccount(context.Context, sdk.AccountI)

	// Remove an account from the store.
	RemoveAccount(context.Context, sdk.AccountI)

	types.QueryServer

	// Logger returns a module-specific logger.
	Logger(ctx sdk.Context) log.Logger

	// Fetch the public key of an account at a specified address
	GetPubKey(context.Context, sdk.AccAddress) (cryptotypes.PubKey, error)

	// Fetch the sequence of an account at a specified address.
	GetSequence(context.Context, sdk.AccAddress) (uint64, error)

	// Fetch the next account number, and increment the internal counter.
	GetNextAccountNumber(sdk.Context) uint64

	// ValidatePermissions validates that the module account has been granted
	// permissions within its set of allowed permissions.
	ValidatePermissions(types.ModuleAccountI) error

	// GetModuleAddress returns an address based on the module name
	GetModuleAddress(string) sdk.AccAddress

	// GetModuleAddressAndPermissions returns an address and permissions based on the module name
	GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string)

	// GetModuleAccountAndPermissions gets the module account from the auth account store and its
	// registered permissions
	GetModuleAccountAndPermissions(ctx sdk.Context, moduleName string) (types.ModuleAccountI, []string)

	// GetModuleAccount gets the module account from the auth account store, if the account does not
	// exist in the AccountKeeper, then it is created.
	GetModuleAccount(sdk.Context, string) types.ModuleAccountI

	// SetModuleAccount sets the module account to the auth account store
	SetModuleAccount(sdk.Context, types.ModuleAccountI)

	// MarshalAccount protobuf serializes an Account interface
	MarshalAccount(types.AccountI) ([]byte, error)

	// UnmarshalAccount returns an Account interface from raw encoded account
	// bytes of a Proto-based Account type
	UnmarshalAccount([]byte) (types.AccountI, error)

	// GetCodec return codec.Codec object used by the keeper
	GetCodec() codec.BinaryCodec

	// GetParams gets the auth module's parameters.
	GetParams(sdk.Context) types.Params

	// SetParams sets the auth module's parameters.
	SetParams(sdk.Context, types.Params)
}

// accountKeeper encodes/decodes accounts using the go-amino (binary)
// encoding/decoding library.
type accountKeeper struct {
	key           sdk.StoreKey
	cdc           codec.BinaryCodec
	paramSubspace paramtypes.Subspace
	permAddrs     map[string]types.PermissionsForAddress

	// The prototypical AccountI constructor.
	proto func() sdk.AccountI

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	// State
	Schema collections.Schema
	Params collections.Item[types.Params]

	// only use for upgrade handler
	//
	// Deprecated: move to accounts module accountNumber
	accountNumber collections.Sequence
	// Accounts key: AccAddr | value: AccountI | index: AccountsIndex
	Accounts *collections.IndexedMap[sdk.AccAddress, sdk.AccountI, AccountsIndexes]
}

var _ AccountKeeper = &accountKeeper{}

// NewAccountKeeper returns a new AccountKeeperI that uses go-amino to
// (binary) encode and decode concrete sdk.Accounts.
// `maccPerms` is a map that takes accounts' addresses as keys, and their respective permissions as values. This map is used to construct
// types.PermissionsForAddress and is used in keeper.ValidatePermissions. Permissions are plain strings,
// and don't have to fit into any predefined structure. This auth module does not use account permissions internally, though other modules
// may use auth.Keeper to access the accounts permissions map.
func NewAccountKeeper(
	env appmodule.Environment, cdc codec.BinaryCodec, proto func() sdk.AccountI, accountsModKeeper types.AccountsModKeeper,
	maccPerms map[string][]string, ac address.Codec, bech32Prefix, authority string,
) AccountKeeper {
	permAddrs := make(map[string]types.PermissionsForAddress)
	for name, perms := range maccPerms {
		permAddrs[name] = types.NewPermissionsForAddress(name, perms)
	}

	return accountKeeper{
		key:           key,
		proto:         proto,
		cdc:           cdc,
		paramSubspace: paramstore,
		permAddrs:     permAddrs,
	}
	ak.Schema = schema
	return ak
}

// removeLegacyAccountNumberUnsafe is used for migration purpose only. It deletes the sequence in the DB
// and returns the last value used on success.
// Deprecated
func (ak AccountKeeper) removeLegacyAccountNumberUnsafe(ctx context.Context) (uint64, error) {
	accNum, err := ak.accountNumber.Peek(ctx)
	if err != nil {
		return 0, err
	}

	// Delete DB entry for legacy account number
	store := ak.KVStoreService.OpenKVStore(ctx)
	err = store.Delete(types.GlobalAccountNumberKey.Bytes())

	return accNum, err
}

// Logger returns a module-specific logger.
func (ak accountKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetPubKey Returns the PubKey of the account at address
func (ak accountKeeper) GetPubKey(ctx sdk.Context, addr sdk.AccAddress) (cryptotypes.PubKey, error) {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	return acc.GetPubKey(), nil
}

// GetSequence Returns the Sequence of the account at address
func (ak accountKeeper) GetSequence(ctx sdk.Context, addr sdk.AccAddress) (uint64, error) {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return 0, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	return acc.GetSequence(), nil
}

// NextAccountNumber returns and increments the global account number counter.
// If the global account number is not set, it initializes it with value 0.
func (ak accountKeeper) GetNextAccountNumber(ctx sdk.Context) uint64 {
	var accNumber uint64
	store := ctx.KVStore(ak.key)

	bz := store.Get(types.GlobalAccountNumberKey)
	if bz == nil {
		// initialize the account numbers
		accNumber = 0
	} else {
		val := gogotypes.UInt64Value{}

		err := ak.cdc.Unmarshal(bz, &val)
		if err != nil {
			panic(err)
		}

		accNumber = val.GetValue()
	}
	return n
}

// GetModulePermissions fetches per-module account permissions.
func (ak AccountKeeper) GetModulePermissions() map[string]types.PermissionsForAddress {
	return ak.permAddrs
}

// ValidatePermissions validates that the module account has been granted
// permissions within its set of allowed permissions.
func (ak accountKeeper) ValidatePermissions(macc types.ModuleAccountI) error {
	permAddr := ak.permAddrs[macc.GetName()]
	for _, perm := range macc.GetPermissions() {
		if !permAddr.HasPermission(perm) {
			return fmt.Errorf("invalid module permission %s", perm)
		}
	}

	return nil
}

// GetModuleAddress returns an address based on the module name
func (ak accountKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	permAddr, ok := ak.permAddrs[moduleName]
	if !ok {
		return nil
	}

	return permAddr.GetAddress()
}

// GetModuleAddressAndPermissions returns an address and permissions based on the module name
func (ak accountKeeper) GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string) {
	permAddr, ok := ak.permAddrs[moduleName]
	if !ok {
		return addr, permissions
	}

	return permAddr.GetAddress(), permAddr.GetPermissions()
}

// GetModuleAccountAndPermissions gets the module account from the auth account store and its
// registered permissions
func (ak accountKeeper) GetModuleAccountAndPermissions(ctx sdk.Context, moduleName string) (types.ModuleAccountI, []string) {
	addr, perms := ak.GetModuleAddressAndPermissions(moduleName)
	if addr == nil {
		return nil, []string{}
	}

	acc := ak.GetAccount(ctx, addr)
	if acc != nil {
		macc, ok := acc.(sdk.ModuleAccountI)
		if !ok {
			panic("account is not a module account")
		}
		return macc, perms
	}

	// create a new module account
	macc := types.NewEmptyModuleAccount(moduleName, perms...)
	maccI := (ak.NewAccount(ctx, macc)).(sdk.ModuleAccountI) // set the account number
	ak.SetModuleAccount(ctx, maccI)

	return maccI, perms
}

// GetModuleAccount gets the module account from the auth account store, if the account does not
// exist in the AccountKeeper, then it is created.
func (ak accountKeeper) GetModuleAccount(ctx sdk.Context, moduleName string) types.ModuleAccountI {
	acc, _ := ak.GetModuleAccountAndPermissions(ctx, moduleName)
	return acc
}

// SetModuleAccount sets the module account to the auth account store
func (ak accountKeeper) SetModuleAccount(ctx sdk.Context, macc types.ModuleAccountI) {
	ak.SetAccount(ctx, macc)
}

func (ak accountKeeper) decodeAccount(bz []byte) types.AccountI {
	acc, err := ak.UnmarshalAccount(bz)
	if err != nil {
		panic(err)
	}
	return params
}

// MarshalAccount protobuf serializes an Account interface
func (ak accountKeeper) MarshalAccount(accountI types.AccountI) ([]byte, error) { // nolint:interfacer
	return ak.cdc.MarshalInterface(accountI)
}

// UnmarshalAccount returns an Account interface from raw encoded account
// bytes of a Proto-based Account type
func (ak accountKeeper) UnmarshalAccount(bz []byte) (types.AccountI, error) {
	var acc types.AccountI
	return acc, ak.cdc.UnmarshalInterface(bz, &acc)
}

// GetCodec return codec.Codec object used by the keeper
func (ak accountKeeper) GetCodec() codec.BinaryCodec { return ak.cdc }
