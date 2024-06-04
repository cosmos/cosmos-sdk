package keeper

import (
	"fmt"
	"time"

	gogotypes "github.com/cosmos/gogoproto/types"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	addresscodec "cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/testutil/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements ValidatorSet interface
// var _ types.ValidatorSet = Keeper{}

// Keeper of the x/staking store
type Keeper struct {
	appmodule.Environment

	cdc                   codec.BinaryCodec
	authKeeper            types.AccountKeeper
	bankKeeper            types.BankKeeper
	hooks                 types.StakingHooks
	authority             string
	validatorAddressCodec addresscodec.Codec
	consensusAddressCodec addresscodec.Codec
	cometInfoService      comet.Service

	Schema collections.Schema

	// LastTotalPower value: LastTotalPower
	LastTotalPower collections.Item[math.Int]
	// ValidatorByConsensusAddress key: consAddr | value: valAddr
	ValidatorByConsensusAddress collections.Map[sdk.ConsAddress, sdk.ValAddress]
	// Delegations key: AccAddr+valAddr | value: Delegation
	Delegations collections.Map[collections.Pair[sdk.AccAddress, sdk.ValAddress], types.Delegation]
	// Validators key: valAddr | value: Validator
	Validators collections.Map[[]byte, types.Validator]
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
	env appmodule.Environment,
	ak types.AccountKeeper,
	bk types.BankKeeper,
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
		authority:             authority,
		validatorAddressCodec: validatorAddressCodec,
		consensusAddressCodec: consensusAddressCodec,
		cometInfoService:      cometInfoService,
		LastTotalPower:        collections.NewItem(sb, types.LastTotalPowerKey, "last_total_power", sdk.IntValue),
		Delegations: collections.NewMap(
			sb, types.DelegationKey, "delegations",
			collections.PairKeyCodec(
				sdk.LengthPrefixedAddressKey(sdk.AccAddressKey), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
				sdk.LengthPrefixedAddressKey(sdk.ValAddressKey), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			),
			codec.CollValue[types.Delegation](cdc),
		),
		ValidatorByConsensusAddress: collections.NewMap(
			sb, types.ValidatorsByConsAddrKey,
			"validator_by_cons_addr",
			sdk.LengthPrefixedAddressKey(sdk.ConsAddressKey), //nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			collcodec.KeyToValueCodec(sdk.ValAddressKey),
		),
		// key format is: 17 | lengthPrefixedBytes(valAddr) | power
		LastValidatorPower: collections.NewMap(sb, types.LastValidatorPowerKey, "last_validator_power", sdk.LengthPrefixedBytesKey, codec.CollValue[gogotypes.Int64Value](cdc)), // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
		Validators:         collections.NewMap(sb, types.ValidatorsKey, "validators", sdk.LengthPrefixedBytesKey, codec.CollValue[types.Validator](cdc)),                        // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
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
