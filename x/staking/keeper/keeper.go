package keeper

import (
	"fmt"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/collections/indexes"
	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/math"
	"cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements ValidatorSet interface
var _ types.ValidatorSet = Keeper{}

// Implements DelegationSet interface
var _ types.DelegationSet = Keeper{}

type rotationHistoryIndexes struct {
	Block *indexes.Multi[uint64, collections.Pair[[]byte, uint64], types.ConsPubKeyRotationHistory]
}

func (a rotationHistoryIndexes) IndexesList() []collections.Index[collections.Pair[[]byte, uint64], types.ConsPubKeyRotationHistory] {
	return []collections.Index[collections.Pair[[]byte, uint64], types.ConsPubKeyRotationHistory]{
		a.Block,
	}
}

func NewRotationHistoryIndexes(sb *collections.SchemaBuilder) rotationHistoryIndexes {
	return rotationHistoryIndexes{
		Block: indexes.NewMulti(
			sb,
			types.BlockConsPubKeyRotationHistoryKey,
			"cons_pubkey_history_by_block",
			collections.Uint64Key,
			collections.PairKeyCodec(collections.BytesKey, collections.Uint64Key),
			func(key collections.Pair[[]byte, uint64], v types.ConsPubKeyRotationHistory) (uint64, error) {
				return v.Height, nil
			},
		),
	}
}

// Keeper of the x/staking store
type Keeper struct {
	appmodule.Environment

	cdc                   codec.BinaryCodec
	authKeeper            types.AccountKeeper
	bankKeeper            types.BankKeeper
	consensusKeeper       types.ConsensusKeeper
	hooks                 types.StakingHooks
	authority             string
	validatorAddressCodec addresscodec.Codec
	consensusAddressCodec addresscodec.Codec
	cometInfoService      comet.Service

	Schema collections.Schema

	// LastTotalPower value: LastTotalPower
	LastTotalPower collections.Item[math.Int]
	// DelegationsByValidator key: valAddr+delAddr | value: none used (index key for delegations by validator index)
	DelegationsByValidator collections.Map[collections.Pair[sdk.ValAddress, sdk.AccAddress], []byte]
	// ValidatorByConsensusAddress key: consAddr | value: valAddr
	ValidatorByConsensusAddress collections.Map[sdk.ConsAddress, sdk.ValAddress]
	// Redelegations key: AccAddr+SrcValAddr+DstValAddr | value: Redelegation
	Redelegations collections.Map[collections.Triple[[]byte, []byte, []byte], types.Redelegation]
	// Delegations key: AccAddr+valAddr | value: Delegation
	Delegations collections.Map[collections.Pair[sdk.AccAddress, sdk.ValAddress], types.Delegation]
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
	// ValidatorConsensusKeyRotationRecordIndexKey: this key is used to restrict the validator next rotation within waiting (unbonding) period
	ValidatorConsensusKeyRotationRecordIndexKey collections.KeySet[collections.Pair[[]byte, time.Time]]
	// ValidatorConsensusKeyRotationRecordQueue: this key is used to set the unbonding period time on each rotation
	ValidatorConsensusKeyRotationRecordQueue collections.Map[time.Time, types.ValAddrsOfRotatedConsKeys]
	// ConsAddrToValidatorIdentifierMap: maps the new cons addr to the initial cons addr
	ConsAddrToValidatorIdentifierMap collections.Map[[]byte, []byte]
	// OldToNewConsAddrMap: maps the old cons addr to the new cons addr
	OldToNewConsAddrMap collections.Map[[]byte, []byte]
	// ValidatorConsPubKeyRotationHistory: consPubkey rotation history by validator
	// A index is being added with key `BlockConsPubKeyRotationHistory`: consPubkey rotation history by height
	RotationHistory *collections.IndexedMap[collections.Pair[[]byte, uint64], types.ConsPubKeyRotationHistory, rotationHistoryIndexes]
}

// NewKeeper creates a new staking Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	env appmodule.Environment,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	ck types.ConsensusKeeper,
	authority string,
	validatorAddressCodec addresscodec.Codec,
	consensusAddressCodec addresscodec.Codec,
	cometInfoService comet.Service,
) *Keeper {
	sb := collections.NewSchemaBuilder(env.KVStoreService)
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
		Environment:           env,
		cdc:                   cdc,
		authKeeper:            ak,
		bankKeeper:            bk,
		consensusKeeper:       ck,
		hooks:                 nil,
		authority:             authority,
		validatorAddressCodec: validatorAddressCodec,
		consensusAddressCodec: consensusAddressCodec,
		cometInfoService:      cometInfoService,
		LastTotalPower:        collections.NewItem(sb, types.LastTotalPowerKey, "last_total_power", sdk.IntValue),
		Delegations: collections.NewMap(
			sb, types.DelegationKey, "delegations",
			collections.NamedPairKeyCodec(
				"delegator",
				sdk.LengthPrefixedAddressKey(sdk.AccAddressKey), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
				"validator_address_key",
				sdk.LengthPrefixedAddressKey(sdk.ValAddressKey), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			),
			codec.CollValue[types.Delegation](cdc),
		),
		DelegationsByValidator: collections.NewMap(
			sb, types.DelegationByValIndexKey,
			"delegations_by_validator",
			collections.NamedPairKeyCodec(
				"validator_address",
				sdk.LengthPrefixedAddressKey(sdk.ValAddressKey), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
				"delegator",
				sdk.AccAddressKey,
			),
			collections.BytesValue,
		),
		ValidatorByConsensusAddress: collections.NewMap(
			sb, types.ValidatorsByConsAddrKey,
			"validator_by_cons_addr",
			sdk.LengthPrefixedAddressKey(sdk.ConsAddressKey).WithName("cons_address"), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			collcodec.KeyToValueCodec(sdk.ValAddressKey),
		),
		// key format is: 52 | lengthPrefixedBytes(AccAddr) | lengthPrefixedBytes(SrcValAddr) | lengthPrefixedBytes(DstValAddr)
		Redelegations: collections.NewMap(
			sb, types.RedelegationKey,
			"redelegations",
			collections.NamedTripleKeyCodec(
				"delegator",
				collections.BytesKey,
				"src_validator",
				collections.BytesKey,
				"dst_validator",
				sdk.LengthPrefixedBytesKey, // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
			),
			codec.CollValue[types.Redelegation](cdc),
		),
		UnbondingDelegationByValIndex: collections.NewMap(
			sb, types.UnbondingDelegationByValIndexKey,
			"unbonding_delegation_by_val_index",
			collections.NamedPairKeyCodec(
				"validator_address",
				sdk.LengthPrefixedBytesKey,
				"delegator",
				sdk.LengthPrefixedBytesKey,
			), // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
			collections.BytesValue,
		),
		UnbondingQueue: collections.NewMap(sb, types.UnbondingQueueKey, "unbonidng_queue", sdk.TimeKey.WithName("unbonding_time"), codec.CollValue[types.DVPairs](cdc)),
		// key format is: 53 | lengthPrefixedBytes(SrcValAddr) | lengthPrefixedBytes(AccAddr) | lengthPrefixedBytes(DstValAddr)
		RedelegationsByValSrc: collections.NewMap(
			sb, types.RedelegationByValSrcIndexKey,
			"redelegations_by_val_src",
			collections.NamedTripleKeyCodec(
				"src_validator",
				collections.BytesKey,
				"delegator",
				collections.BytesKey,
				"dst_validator",
				sdk.LengthPrefixedBytesKey, // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
			),
			collections.BytesValue,
		),
		// key format is: 17 | lengthPrefixedBytes(valAddr) | power
		LastValidatorPower: collections.NewMap(sb, types.LastValidatorPowerKey, "last_validator_power", sdk.LengthPrefixedBytesKey.WithName("validator_address"), codec.CollValue[gogotypes.Int64Value](cdc)), // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
		// key format is: 54 | lengthPrefixedBytes(DstValAddr) | lengthPrefixedBytes(AccAddr) | lengthPrefixedBytes(SrcValAddr)
		RedelegationsByValDst: collections.NewMap(
			sb, types.RedelegationByValDstIndexKey,
			"redelegations_by_val_dst",
			collections.NamedTripleKeyCodec(
				"dst_validator",
				collections.BytesKey,
				"delegator",
				collections.BytesKey,
				"src_validator",
				sdk.LengthPrefixedBytesKey, // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
			),
			collections.BytesValue,
		),
		RedelegationQueue: collections.NewMap(sb, types.RedelegationQueueKey, "redelegation_queue", sdk.TimeKey.WithName("completion_time"), codec.CollValue[types.DVVTriplets](cdc)),
		Validators: collections.NewMap(
			sb,
			types.ValidatorsKey,
			"validators",
			sdk.LengthPrefixedBytesKey.WithName("validator_address"), // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
			codec.CollValue[types.Validator](cdc),
		),
		UnbondingDelegations: collections.NewMap(
			sb, types.UnbondingDelegationKey,
			"unbonding_delegation",
			collections.NamedPairKeyCodec(
				"delegator",
				collections.BytesKey,
				"validator",
				sdk.LengthPrefixedBytesKey, // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
			),
			codec.CollValue[types.UnbondingDelegation](cdc),
		),
		// key format is: 67 | length(timestamp Bytes) | timestamp | height
		// Note: We use 3 keys here because we prefixed time bytes with its length previously and to retain state compatibility we remain to use the same
		ValidatorQueue: collections.NewMap(
			sb, types.ValidatorQueueKey,
			"validator_queue",
			collections.NamedTripleKeyCodec(
				"ts_length",
				collections.Uint64Key,
				"timestamp",
				sdk.TimeKey,
				"height",
				collections.Uint64Key,
			),
			codec.CollValue[types.ValAddresses](cdc),
		),
		// key is: 113 (it's a direct prefix)
		Params: collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),

		// key format is: 104 | valAddr | time
		ValidatorConsensusKeyRotationRecordIndexKey: collections.NewKeySet(
			sb, types.ValidatorConsensusKeyRotationRecordIndexKey,
			"cons_pub_rotation_index",
			collections.NamedPairKeyCodec(
				"validator_address",
				collections.BytesKey,
				"time",
				sdk.TimeKey),
		),

		// key format is: 103 | time
		ValidatorConsensusKeyRotationRecordQueue: collections.NewMap(
			sb, types.ValidatorConsensusKeyRotationRecordQueueKey,
			"cons_pub_rotation_queue",
			sdk.TimeKey,
			codec.CollValue[types.ValAddrsOfRotatedConsKeys](cdc),
		),

		// key format is: 105 | consAddr
		ConsAddrToValidatorIdentifierMap: collections.NewMap(
			sb, types.ConsAddrToValidatorIdentifierMapPrefix,
			"new_to_old_cons_addr_map",
			collections.BytesKey,
			collections.BytesValue,
		),

		// key format is: 106 | consAddr
		OldToNewConsAddrMap: collections.NewMap(
			sb, types.OldToNewConsAddrMap,
			"old_to_new_cons_key_map",
			collections.BytesKey,
			collections.BytesValue,
		),

		// key format is : 101 | rotation history
		// index is : 102 | rotation history
		RotationHistory: collections.NewIndexedMap(
			sb,
			types.ValidatorConsPubKeyRotationHistoryKey,
			"cons_pub_rotation_history",
			collections.NamedPairKeyCodec(
				"validator_address",
				collections.BytesKey,
				"height_key",
				collections.Uint64Key),
			codec.CollValue[types.ConsPubKeyRotationHistory](cdc),
			NewRotationHistoryIndexes(sb),
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
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
