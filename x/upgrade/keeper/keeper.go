package keeper

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	ibcexported "github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// UpgradeInfoFileName file to store upgrade information
const UpgradeInfoFileName string = "upgrade-info.json"

type Keeper struct {
	homePath           string
	skipUpgradeHeights map[int64]bool
	storeKey           sdk.StoreKey
	cdc                codec.BinaryMarshaler
	upgradeHandlers    map[string]types.UpgradeHandler
}

// NewKeeper constructs an upgrade Keeper
func NewKeeper(skipUpgradeHeights map[int64]bool, storeKey sdk.StoreKey, cdc codec.BinaryMarshaler, homePath string) Keeper {
	return Keeper{
		homePath:           homePath,
		skipUpgradeHeights: skipUpgradeHeights,
		storeKey:           storeKey,
		cdc:                cdc,
		upgradeHandlers:    map[string]types.UpgradeHandler{},
	}
}

// SetUpgradeHandler sets an UpgradeHandler for the upgrade specified by name. This handler will be called when the upgrade
// with this name is applied. In order for an upgrade with the given name to proceed, a handler for this upgrade
// must be set even if it is a no-op function.
func (k Keeper) SetUpgradeHandler(name string, upgradeHandler types.UpgradeHandler) {
	k.upgradeHandlers[name] = upgradeHandler
}

// ScheduleUpgrade schedules an upgrade based on the specified plan.
// If there is another Plan already scheduled, it will overwrite it
// (implicitly cancelling the current plan)
// ScheduleUpgrade will also write the upgraded client to the upgraded client path
// if an upgraded client is specified in the plan
func (k Keeper) ScheduleUpgrade(ctx sdk.Context, plan types.Plan) error {
	if err := plan.ValidateBasic(); err != nil {
		return err
	}

	if !plan.Time.IsZero() {
		if !plan.Time.After(ctx.BlockHeader().Time) {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "upgrade cannot be scheduled in the past")
		}
	} else if plan.Height <= ctx.BlockHeight() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "upgrade cannot be scheduled in the past")
	}

	if k.GetDoneHeight(ctx, plan.Name) != 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "upgrade with name %s has already been completed", plan.Name)
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&plan)
	store.Set(types.PlanKey(), bz)

	if plan.UpgradedClientState == nil {
		// if latest UpgradedClientState is nil, but upgraded client exists in store,
		// then delete client state from store.
		_, height, _ := k.GetUpgradedClient(ctx)
		if height != 0 {
			store.Delete(types.UpgradedClientKey(height))
		}
		return nil
	}

	// Set UpgradedClientState in store
	clientState, err := clienttypes.UnpackClientState(plan.UpgradedClientState)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "could not unpack clientstate: %v", err)
	}
	// deletes any previously stored upgraded client and sets the new upgraded client in
	return k.SetUpgradedClient(ctx, plan.Height, clientState)
}

// SetUpgradedClient sets the expected upgraded client for the next version of this chain
func (k Keeper) SetUpgradedClient(ctx sdk.Context, upgradeHeight int64, cs ibcexported.ClientState) error {
	store := ctx.KVStore(k.storeKey)

	// delete any previously stored upgraded client before setting a new one
	// since there should only ever be one upgraded client in the store at any given time
	_, setHeight, _ := k.GetUpgradedClient(ctx)
	if setHeight != 0 {
		store.Delete(types.UpgradedClientKey(setHeight))
	}

	// zero out any custom fields before setting
	cs = cs.ZeroCustomFields()
	bz, err := clienttypes.MarshalClientState(k.cdc, cs)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "could not marshal clientstate: %v", err)
	}

	store.Set(types.UpgradedClientKey(upgradeHeight), bz)
	return nil
}

// GetUpgradedClient gets the expected upgraded client for the next version of this chain
// along with the planned upgrade height
// Since there is only ever one upgraded client in store, we do not need to know key beforehand
func (k Keeper) GetUpgradedClient(ctx sdk.Context) (ibcexported.ClientState, int64, error) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(types.KeyUpgradedClient))
	var (
		count, height int
		bz            []byte
	)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		count++
		// we must panic if the upgraded clients in store is ever more than one since
		// that would break upgrade functionality and chain must halt and fix issue manually
		if count > 1 {
			panic("more than 1 upgrade client stored in state")
		}

		keySplit := strings.Split(string(iterator.Key()), "/")
		var err error
		height, err = strconv.Atoi(keySplit[len(keySplit)-1])
		if err != nil {
			return nil, 0, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "could not parse upgrade height from key: %s", err)
		}

		bz = iterator.Value()
	}

	if count == 0 {
		return nil, 0, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "upgrade client not found in store")
	}

	clientState, err := clienttypes.UnmarshalClientState(k.cdc, bz)
	if err != nil {
		return nil, 0, err
	}
	return clientState, int64(height), nil
}

// GetDoneHeight returns the height at which the given upgrade was executed
func (k Keeper) GetDoneHeight(ctx sdk.Context, name string) int64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{types.DoneByte})
	bz := store.Get([]byte(name))
	if len(bz) == 0 {
		return 0
	}

	return int64(binary.BigEndian.Uint64(bz))
}

// ClearUpgradePlan clears any schedule upgrade
func (k Keeper) ClearUpgradePlan(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.PlanKey())
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetUpgradePlan returns the currently scheduled Plan if any, setting havePlan to true if there is a scheduled
// upgrade or false if there is none
func (k Keeper) GetUpgradePlan(ctx sdk.Context) (plan types.Plan, havePlan bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.PlanKey())
	if bz == nil {
		return plan, false
	}

	k.cdc.MustUnmarshalBinaryBare(bz, &plan)
	return plan, true
}

// setDone marks this upgrade name as being done so the name can't be reused accidentally
func (k Keeper) setDone(ctx sdk.Context, name string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{types.DoneByte})
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, uint64(ctx.BlockHeight()))
	store.Set([]byte(name), bz)
}

// HasHandler returns true iff there is a handler registered for this name
func (k Keeper) HasHandler(name string) bool {
	_, ok := k.upgradeHandlers[name]
	return ok
}

// ApplyUpgrade will execute the handler associated with the Plan and mark the plan as done.
func (k Keeper) ApplyUpgrade(ctx sdk.Context, plan types.Plan) {
	handler := k.upgradeHandlers[plan.Name]
	if handler == nil {
		panic("ApplyUpgrade should never be called without first checking HasHandler")
	}

	handler(ctx, plan)

	k.ClearUpgradePlan(ctx)
	k.setDone(ctx, plan.Name)
}

// IsSkipHeight checks if the given height is part of skipUpgradeHeights
func (k Keeper) IsSkipHeight(height int64) bool {
	return k.skipUpgradeHeights[height]
}

// DumpUpgradeInfoToDisk writes upgrade information to UpgradeInfoFileName.
func (k Keeper) DumpUpgradeInfoToDisk(height int64, name string) error {
	upgradeInfoFilePath, err := k.GetUpgradeInfoPath()
	if err != nil {
		return err
	}

	upgradeInfo := store.UpgradeInfo{
		Name:   name,
		Height: height,
	}
	info, err := json.Marshal(upgradeInfo)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(upgradeInfoFilePath, info, 0600)
}

// GetUpgradeInfoPath returns the upgrade info file path
func (k Keeper) GetUpgradeInfoPath() (string, error) {
	upgradeInfoFileDir := path.Join(k.getHomeDir(), "data")
	err := tmos.EnsureDir(upgradeInfoFileDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	return filepath.Join(upgradeInfoFileDir, UpgradeInfoFileName), nil
}

// getHomeDir returns the height at which the given upgrade was executed
func (k Keeper) getHomeDir() string {
	return k.homePath
}

// ReadUpgradeInfoFromDisk returns the name and height of the upgrade which is
// written to disk by the old binary when panicking. An error is returned if
// the upgrade path directory cannot be created or if the file exists and
// cannot be read or if the upgrade info fails to unmarshal.
func (k Keeper) ReadUpgradeInfoFromDisk() (store.UpgradeInfo, error) {
	var upgradeInfo store.UpgradeInfo

	upgradeInfoPath, err := k.GetUpgradeInfoPath()
	if err != nil {
		return upgradeInfo, err
	}

	data, err := ioutil.ReadFile(upgradeInfoPath)
	if err != nil {
		// if file does not exist, assume there are no upgrades
		if os.IsNotExist(err) {
			return upgradeInfo, nil
		}

		return upgradeInfo, err
	}

	if err := json.Unmarshal(data, &upgradeInfo); err != nil {
		return upgradeInfo, err
	}

	return upgradeInfo, nil
}
