package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeperI is the interface contract that x/auth's keeper implements.
type AccountKeeperI interface {
	address.Codec

	// Return a new account with the next account number and the specified address. Does not save the new account to the store.
	NewAccountWithAddress(context.Context, sdk.AccAddress) sdk.AccountI

	// Return a new account with the next account number. Does not save the new account to the store.
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI

	// Check if an account exists in the store.
	HasAccount(context.Context, sdk.AccAddress) bool

	// Retrieve an account from the store.
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI

	// Set an account in the store.
	SetAccount(context.Context, sdk.AccountI)

	// Remove an account from the store.
	RemoveAccount(context.Context, sdk.AccountI)

	// Iterate over all accounts, calling the provided function. Stop iteration when it returns true.
	IterateAccounts(context.Context, func(sdk.AccountI) bool)

	// Fetch the public key of an account at a specified address
	GetPubKey(context.Context, sdk.AccAddress) (cryptotypes.PubKey, error)

	// Fetch the sequence of an account at a specified address.
	GetSequence(context.Context, sdk.AccAddress) (uint64, error)

	// Fetch the next account number, and increment the internal counter.
	NextAccountNumber(context.Context) uint64

	// GetModulePermissions fetches per-module account permissions
	GetModulePermissions() map[string]types.PermissionsForAddress
}

// AccountKeeper encodes/decodes accounts using the go-amino (binary)
// encoding/decoding library.
type AccountKeeper struct {
	address.Codec

	storeService store.KVStoreService
	cdc          codec.BinaryCodec
	permAddrs    map[string]types.PermissionsForAddress
	bech32Prefix string

	// The prototypical AccountI constructor.
	proto func() sdk.AccountI

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	// State
	ParamsState   collections.Item[types.Params] // NOTE: name is this because it conflicts with the Params gRPC method impl
	AccountNumber collections.Sequence
}

var _ AccountKeeperI = &AccountKeeper{}

// NewAccountKeeper returns a new AccountKeeperI that uses go-amino to
// (binary) encode and decode concrete sdk.Accounts.
// `maccPerms` is a map that takes accounts' addresses as keys, and their respective permissions as values. This map is used to construct
// types.PermissionsForAddress and is used in keeper.ValidatePermissions. Permissions are plain strings,
// and don't have to fit into any predefined structure. This auth module does not use account permissions internally, though other modules
// may use auth.Keeper to access the accounts permissions map.
func NewAccountKeeper(
	cdc codec.BinaryCodec, storeService store.KVStoreService, proto func() sdk.AccountI,
	maccPerms map[string][]string, bech32Prefix, authority string,
) AccountKeeper {
	permAddrs := make(map[string]types.PermissionsForAddress)
	for name, perms := range maccPerms {
		permAddrs[name] = types.NewPermissionsForAddress(name, perms)
	}

	sb := collections.NewSchemaBuilder(storeService)

	return AccountKeeper{
		Codec:         authcodec.NewBech32Codec(bech32Prefix),
		bech32Prefix:  bech32Prefix,
		storeService:  storeService,
		proto:         proto,
		cdc:           cdc,
		permAddrs:     permAddrs,
		authority:     authority,
		ParamsState:   collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		AccountNumber: collections.NewSequence(sb, types.GlobalAccountNumberKey, "account_number"),
	}
}

// GetAuthority returns the x/auth module's authority.
func (ak AccountKeeper) GetAuthority() string {
	return ak.authority
}

// GetAddressCodec returns the x/auth module's address.
// x/auth is tied to bech32 encoded user accounts
func (ak AccountKeeper) GetAddressCodec() address.Codec {
	return ak.Codec
}

// Logger returns a module-specific logger.
func (ak AccountKeeper) Logger(ctx context.Context) log.Logger {
	return sdk.UnwrapSDKContext(ctx).Logger().With("module", "x/"+types.ModuleName)
}

// GetPubKey Returns the PubKey of the account at address
func (ak AccountKeeper) GetPubKey(ctx context.Context, addr sdk.AccAddress) (cryptotypes.PubKey, error) {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	return acc.GetPubKey(), nil
}

// GetSequence Returns the Sequence of the account at address
func (ak AccountKeeper) GetSequence(ctx context.Context, addr sdk.AccAddress) (uint64, error) {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return 0, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	return acc.GetSequence(), nil
}

// NextAccountNumber returns and increments the global account number counter.
// If the global account number is not set, it initializes it with value 0.
func (ak AccountKeeper) NextAccountNumber(ctx context.Context) uint64 {
	n, err := ak.AccountNumber.Next(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

// GetModulePermissions fetches per-module account permissions.
func (ak AccountKeeper) GetModulePermissions() map[string]types.PermissionsForAddress {
	return ak.permAddrs
}

// ValidatePermissions validates that the module account has been granted
// permissions within its set of allowed permissions.
func (ak AccountKeeper) ValidatePermissions(macc sdk.ModuleAccountI) error {
	permAddr := ak.permAddrs[macc.GetName()]
	for _, perm := range macc.GetPermissions() {
		if !permAddr.HasPermission(perm) {
			return fmt.Errorf("invalid module permission %s", perm)
		}
	}

	return nil
}

// GetModuleAddress returns an address based on the module name
func (ak AccountKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	permAddr, ok := ak.permAddrs[moduleName]
	if !ok {
		return nil
	}

	return permAddr.GetAddress()
}

// GetModuleAddressAndPermissions returns an address and permissions based on the module name
func (ak AccountKeeper) GetModuleAddressAndPermissions(moduleName string) (addr sdk.AccAddress, permissions []string) {
	permAddr, ok := ak.permAddrs[moduleName]
	if !ok {
		return addr, permissions
	}

	return permAddr.GetAddress(), permAddr.GetPermissions()
}

// GetModuleAccountAndPermissions gets the module account from the auth account store and its
// registered permissions
func (ak AccountKeeper) GetModuleAccountAndPermissions(ctx context.Context, moduleName string) (sdk.ModuleAccountI, []string) {
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
func (ak AccountKeeper) GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI {
	acc, _ := ak.GetModuleAccountAndPermissions(ctx, moduleName)
	return acc
}

// SetModuleAccount sets the module account to the auth account store
func (ak AccountKeeper) SetModuleAccount(ctx context.Context, macc sdk.ModuleAccountI) {
	ak.SetAccount(ctx, macc)
}

func (ak AccountKeeper) decodeAccount(bz []byte) sdk.AccountI {
	acc, err := ak.UnmarshalAccount(bz)
	if err != nil {
		panic(err)
	}

	return acc
}

// MarshalAccount protobuf serializes an Account interface
func (ak AccountKeeper) MarshalAccount(accountI sdk.AccountI) ([]byte, error) {
	return ak.cdc.MarshalInterface(accountI)
}

// UnmarshalAccount returns an Account interface from raw encoded account
// bytes of a Proto-based Account type
func (ak AccountKeeper) UnmarshalAccount(bz []byte) (sdk.AccountI, error) {
	var acc sdk.AccountI
	return acc, ak.cdc.UnmarshalInterface(bz, &acc)
}

// GetCodec return codec.Codec object used by the keeper
func (ak AccountKeeper) GetCodec() codec.BinaryCodec { return ak.cdc }

// add getter for bech32Prefix
func (ak AccountKeeper) getBech32Prefix() (string, error) {
	return ak.bech32Prefix, nil
}

// SetParams sets the auth module's parameters.
// CONTRACT: This method performs no validation of the parameters.
func (ak AccountKeeper) SetParams(ctx context.Context, params types.Params) error {
	return ak.ParamsState.Set(ctx, params)
}

// GetParams gets the auth module's parameters.
func (ak AccountKeeper) GetParams(ctx context.Context) (params types.Params) {
	params, err := ak.ParamsState.Get(ctx)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		panic(err)
	}
	return params
}
