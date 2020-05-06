package keeper

import (
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// AccountKeeper encodes/decodes accounts using the go-amino (binary)
// encoding/decoding library.
type AccountKeeper struct {
	key           sdk.StoreKey
	cdc           types.Codec
	paramSubspace paramtypes.Subspace
	permAddrs     map[string]types.PermissionsForAddress

	// The prototypical Account constructor.
	proto func() exported.Account
}

// NewAccountKeeper returns a new sdk.AccountKeeper that uses go-amino to
// (binary) encode and decode concrete sdk.Accounts.
func NewAccountKeeper(
	cdc types.Codec, key sdk.StoreKey, paramstore paramtypes.Subspace, proto func() exported.Account,
	maccPerms map[string][]string,
) AccountKeeper {

	// set KeyTable if it has not already been set
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	permAddrs := make(map[string]types.PermissionsForAddress)
	for name, perms := range maccPerms {
		permAddrs[name] = types.NewPermissionsForAddress(name, perms)
	}

	return AccountKeeper{
		key:           key,
		proto:         proto,
		cdc:           cdc,
		paramSubspace: paramstore,
		permAddrs:     permAddrs,
	}
}

// Logger returns a module-specific logger.
func (ak AccountKeeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetPubKey Returns the PubKey of the account at address
func (ak AccountKeeper) GetPubKey(ctx sdk.Context, addr sdk.AccAddress) (crypto.PubKey, error) {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	return acc.GetPubKey(), nil
}

// GetSequence Returns the Sequence of the account at address
func (ak AccountKeeper) GetSequence(ctx sdk.Context, addr sdk.AccAddress) (uint64, error) {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return 0, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "account %s does not exist", addr)
	}

	return acc.GetSequence(), nil
}

// GetNextAccountNumber returns and increments the global account number counter.
// If the global account number is not set, it initializes it with value 0.
func (ak AccountKeeper) GetNextAccountNumber(ctx sdk.Context) uint64 {
	var accNumber uint64
	store := ctx.KVStore(ak.key)

	bz := store.Get(types.GlobalAccountNumberKey)
	if bz == nil {
		// initialize the account numbers
		accNumber = 0
	} else {
		val := gogotypes.UInt64Value{}

		err := ak.cdc.UnmarshalBinaryBare(bz, &val)
		if err != nil {
			panic(err)
		}

		accNumber = val.GetValue()
	}

	bz = ak.cdc.MustMarshalBinaryBare(&gogotypes.UInt64Value{Value: accNumber + 1})
	store.Set(types.GlobalAccountNumberKey, bz)

	return accNumber
}

// ValidatePermissions validates that the module account has been granted
// permissions within its set of allowed permissions.
func (ak AccountKeeper) ValidatePermissions(macc exported.ModuleAccountI) error {
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
func (ak AccountKeeper) GetModuleAccountAndPermissions(ctx sdk.Context, moduleName string) (exported.ModuleAccountI, []string) {
	addr, perms := ak.GetModuleAddressAndPermissions(moduleName)
	if addr == nil {
		return nil, []string{}
	}

	acc := ak.GetAccount(ctx, addr)
	if acc != nil {
		macc, ok := acc.(exported.ModuleAccountI)
		if !ok {
			panic("account is not a module account")
		}
		return macc, perms
	}

	// create a new module account
	macc := types.NewEmptyModuleAccount(moduleName, perms...)
	maccI := (ak.NewAccount(ctx, macc)).(exported.ModuleAccountI) // set the account number
	ak.SetModuleAccount(ctx, maccI)

	return maccI, perms
}

// GetModuleAccount gets the module account from the auth account store, if the account does not
// exist in the AccountKeeper, then it is created.
func (ak AccountKeeper) GetModuleAccount(ctx sdk.Context, moduleName string) exported.ModuleAccountI {
	acc, _ := ak.GetModuleAccountAndPermissions(ctx, moduleName)
	return acc
}

// SetModuleAccount sets the module account to the auth account store
func (ak AccountKeeper) SetModuleAccount(ctx sdk.Context, macc exported.ModuleAccountI) { //nolint:interfacer
	ak.SetAccount(ctx, macc)
}

func (ak AccountKeeper) decodeAccount(bz []byte) exported.Account {
	acc, err := ak.cdc.UnmarshalAccount(bz)
	if err != nil {
		panic(err)
	}

	return acc
}
