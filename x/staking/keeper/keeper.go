package keeper

import (
	"context"
	"fmt"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	addresscodec "cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Implements ValidatorSet interface
var _ types.ValidatorSet = Keeper{}

// Implements DelegationSet interface
var _ types.DelegationSet = Keeper{}

func HistoricalInfoCodec(cdc codec.BinaryCodec) collcodec.ValueCodec[types.HistoricalRecord] {
	return collcodec.NewAltValueCodec(codec.CollValue[types.HistoricalRecord](cdc), func(b []byte) (types.HistoricalRecord, error) {
		historicalinfo := types.HistoricalInfo{} //nolint: staticcheck // HistoricalInfo is deprecated
		err := historicalinfo.Unmarshal(b)
		if err != nil {
			return types.HistoricalRecord{}, err
		}

		return types.HistoricalRecord{
			Apphash:        historicalinfo.Header.AppHash,
			Time:           &historicalinfo.Header.Time,
			ValidatorsHash: historicalinfo.Header.NextValidatorsHash,
		}, nil
	})
}

// Keeper of the x/staking store
type Keeper struct {
	storeService          storetypes.KVStoreService
	cdc                   codec.BinaryCodec
	authKeeper            types.AccountKeeper
	bankKeeper            types.BankKeeper
	hooks                 types.StakingHooks
	authority             string
	validatorAddressCodec addresscodec.Codec
	consensusAddressCodec addresscodec.Codec

	Schema collections.Schema

	// HistoricalInfo key: Height | value: HistoricalInfo
	HistoricalInfo collections.Map[uint64, types.HistoricalRecord]
	// LastTotalPower value: LastTotalPower
	LastTotalPower collections.Item[math.Int]
	// ValidatorUpdates value: ValidatorUpdates
	ValidatorUpdates collections.Item[types.ValidatorUpdates]
	// DelegationsByValidator key: valAddr+delAddr | value: none used (index key for delegations by validator index)
	DelegationsByValidator collections.Map[collections.Pair[sdk.ValAddress, sdk.AccAddress], []byte]
	UnbondingID            collections.Sequence
	// ValidatorByConsensusAddress key: consAddr | value: valAddr
	ValidatorByConsensusAddress collections.Map[sdk.ConsAddress, sdk.ValAddress]
	// UnbondingType key: unbondingID | value: index of UnbondingType
	UnbondingType collections.Map[uint64, uint64]
	// Redelegations key: AccAddr+SrcValAddr+DstValAddr | value: Redelegation
	Redelegations collections.Map[collections.Triple[[]byte, []byte, []byte], types.Redelegation]
	// Delegations key: AccAddr+valAddr | value: Delegation
	Delegations collections.Map[collections.Pair[sdk.AccAddress, sdk.ValAddress], types.Delegation]
	// UnbondingIndex key:UnbondingID | value: ubdKey (ubdKey = [UnbondingDelegationKey(Prefix)+len(delAddr)+delAddr+len(valAddr)+valAddr])
	UnbondingIndex collections.Map[uint64, []byte]
	// UnbondingQueue key: Timestamp | value: DVPairs [delAddr+valAddr]
	UnbondingQueue collections.Map[time.Time, types.DVPairs]
	// Validators key: valAddr | value: Validator
	Validators collections.Map[[]byte, types.Validator]
	// UnbondingDelegations key: delAddr+valAddr | value: UnbondingDelegation
	UnbondingDelegations collections.Map[collections.Pair[[]byte, []byte], types.UnbondingDelegation]
	// RedelegationsByValDst key: DstValAddr+DelAccAddr+SrcValAddr | value: none used (index key for Redelegations stored by DstVal index)
	RedelegationsByValDst collections.Map[collections.Triple[[]byte, []byte, []byte], []byte]
	// RedelegationsByValSrc key: SrcValAddr+DelAccAddr+DstValAddr |  value: none used (index key for Redelegations stored by SrcVal index)
	RedelegationsByValSrc collections.Map[collections.Triple[[]byte, []byte, []byte], []byte]
	// UnbondingDelegationByValIndex key: valAddr+delAddr | value: none used (index key for UnbondingDelegations stored by validator index)
	UnbondingDelegationByValIndex collections.Map[collections.Pair[[]byte, []byte], []byte]
	// RedelegationQueue key: Timestamp | value: DVVTriplets [delAddr+valSrcAddr+valDstAddr]
	RedelegationQueue collections.Map[time.Time, types.DVVTriplets]
	// ValidatorQueue key: len(timestamp bytes)+timestamp+height | value: ValAddresses
	ValidatorQueue collections.Map[collections.Triple[uint64, time.Time, uint64], types.ValAddresses]
	// LastValidatorPower key: valAddr | value: power(gogotypes.Int64Value())
	LastValidatorPower collections.Map[[]byte, gogotypes.Int64Value]
	// Params key: ParamsKeyPrefix | value: Params
	Params collections.Item[types.Params]
}

// NewKeeper creates a new staking Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	authority string,
	validatorAddressCodec addresscodec.Codec,
	consensusAddressCodec addresscodec.Codec,
) *Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	// ensure bonded and not bonded module accounts are set
	if addr := ak.GetModuleAddress(types.BondedPoolName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.BondedPoolName))
	}

	if addr := ak.GetModuleAddress(types.NotBondedPoolName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.NotBondedPoolName))
	}

	// ensure that authority is a valid AccAddress
	if _, err := ak.AddressCodec().StringToBytes(authority); err != nil {
		panic("authority is not a valid acc address")
	}

	if validatorAddressCodec == nil || consensusAddressCodec == nil {
		panic("validator and/or consensus address codec are nil")
	}

	k := &Keeper{
		storeService:          storeService,
		cdc:                   cdc,
		authKeeper:            ak,
		bankKeeper:            bk,
		hooks:                 nil,
		authority:             authority,
		validatorAddressCodec: validatorAddressCodec,
		consensusAddressCodec: consensusAddressCodec,
		LastTotalPower:        collections.NewItem(sb, types.LastTotalPowerKey, "last_total_power", sdk.IntValue),
		HistoricalInfo:        collections.NewMap(sb, types.HistoricalInfoKey, "historical_info", collections.Uint64Key, HistoricalInfoCodec(cdc)),
		ValidatorUpdates:      collections.NewItem(sb, types.ValidatorUpdatesKey, "validator_updates", codec.CollValue[types.ValidatorUpdates](cdc)),
		Delegations: collections.NewMap(
			sb, types.DelegationKey, "delegations",
			collections.PairKeyCodec(
				sdk.LengthPrefixedAddressKey(sdk.AccAddressKey), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
				sdk.LengthPrefixedAddressKey(sdk.ValAddressKey), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			),
			codec.CollValue[types.Delegation](cdc),
		),
		DelegationsByValidator: collections.NewMap(
			sb, types.DelegationByValIndexKey,
			"delegations_by_validator",
			collections.PairKeyCodec(sdk.LengthPrefixedAddressKey(sdk.ValAddressKey), sdk.AccAddressKey), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			collections.BytesValue,
		),
		UnbondingID: collections.NewSequence(sb, types.UnbondingIDKey, "unbonding_id"),
		ValidatorByConsensusAddress: collections.NewMap(
			sb, types.ValidatorsByConsAddrKey,
			"validator_by_cons_addr",
			sdk.LengthPrefixedAddressKey(sdk.ConsAddressKey), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			collcodec.KeyToValueCodec(sdk.ValAddressKey),
		),
		UnbondingType: collections.NewMap(sb, types.UnbondingTypeKey, "unbonding_type", collections.Uint64Key, collections.Uint64Value),
		// key format is: 52 | lengthPrefixedBytes(AccAddr) | lengthPrefixedBytes(SrcValAddr) | lengthPrefixedBytes(DstValAddr)
		Redelegations: collections.NewMap(
			sb, types.RedelegationKey,
			"redelegations",
			collections.TripleKeyCodec(
				collections.BytesKey,
				collections.BytesKey,
				sdk.LengthPrefixedBytesKey, // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
			),
			codec.CollValue[types.Redelegation](cdc),
		),
		UnbondingIndex: collections.NewMap(sb, types.UnbondingIndexKey, "unbonding_index", collections.Uint64Key, collections.BytesValue),
		UnbondingDelegationByValIndex: collections.NewMap(
			sb, types.UnbondingDelegationByValIndexKey,
			"unbonding_delegation_by_val_index",
			collections.PairKeyCodec(sdk.LengthPrefixedBytesKey, sdk.LengthPrefixedBytesKey), // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
			collections.BytesValue,
		),
		UnbondingQueue: collections.NewMap(sb, types.UnbondingQueueKey, "unbonidng_queue", sdk.TimeKey, codec.CollValue[types.DVPairs](cdc)),
		// key format is: 53 | lengthPrefixedBytes(SrcValAddr) | lengthPrefixedBytes(AccAddr) | lengthPrefixedBytes(DstValAddr)
		RedelegationsByValSrc: collections.NewMap(
			sb, types.RedelegationByValSrcIndexKey,
			"redelegations_by_val_src",
			collections.TripleKeyCodec(
				collections.BytesKey,
				collections.BytesKey,
				sdk.LengthPrefixedBytesKey, // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
			),
			collections.BytesValue,
		),
		// key format is: 17 | lengthPrefixedBytes(valAddr) | power
		LastValidatorPower: collections.NewMap(sb, types.LastValidatorPowerKey, "last_validator_power", sdk.LengthPrefixedBytesKey, codec.CollValue[gogotypes.Int64Value](cdc)), // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
		// key format is: 54 | lengthPrefixedBytes(DstValAddr) | lengthPrefixedBytes(AccAddr) | lengthPrefixedBytes(SrcValAddr)
		RedelegationsByValDst: collections.NewMap(
			sb, types.RedelegationByValDstIndexKey,
			"redelegations_by_val_dst",
			collections.TripleKeyCodec(
				collections.BytesKey,
				collections.BytesKey,
				sdk.LengthPrefixedBytesKey, // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
			),
			collections.BytesValue,
		),
		RedelegationQueue: collections.NewMap(sb, types.RedelegationQueueKey, "redelegation_queue", sdk.TimeKey, codec.CollValue[types.DVVTriplets](cdc)),
		Validators:        collections.NewMap(sb, types.ValidatorsKey, "validators", sdk.LengthPrefixedBytesKey, codec.CollValue[types.Validator](cdc)), // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
		UnbondingDelegations: collections.NewMap(
			sb, types.UnbondingDelegationKey,
			"unbonding_delegation",
			collections.PairKeyCodec(
				collections.BytesKey,
				sdk.LengthPrefixedBytesKey, // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
			),
			codec.CollValue[types.UnbondingDelegation](cdc),
		),
		// key format is: 67 | length(timestamp Bytes) | timestamp | height
		// Note: We use 3 keys here because we prefixed time bytes with its length previously and to retain state compatibility we remain to use the same
		ValidatorQueue: collections.NewMap(
			sb, types.ValidatorQueueKey,
			"validator_queue",
			collections.TripleKeyCodec(
				collections.Uint64Key,
				sdk.TimeKey,
				collections.Uint64Key,
			),
			codec.CollValue[types.ValAddresses](cdc),
		),
		// key is: 113 (it's a direct prefix)
		Params: collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

// Hooks gets the hooks for staking *Keeper {
func (k *Keeper) Hooks() types.StakingHooks {
	if k.hooks == nil {
		// return a no-op implementation if no hooks are set
		return types.MultiStakingHooks{}
	}

	return k.hooks
}

// SetHooks sets the validator hooks.  In contrast to other receivers, this method must take a pointer due to nature
// of the hooks interface and SDK start up sequence.
func (k *Keeper) SetHooks(sh types.StakingHooks) {
	if k.hooks != nil {
		panic("cannot set validator hooks twice")
	}

	k.hooks = sh
}

// GetAuthority returns the x/staking module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// ValidatorAddressCodec returns the app validator address codec.
func (k Keeper) ValidatorAddressCodec() addresscodec.Codec {
	return k.validatorAddressCodec
}

// ConsensusAddressCodec returns the app consensus address codec.
func (k Keeper) ConsensusAddressCodec() addresscodec.Codec {
	return k.consensusAddressCodec
}
