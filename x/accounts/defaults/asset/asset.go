package asset

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	assettypes "cosmossdk.io/x/accounts/defaults/asset/v1"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	DenomPrefix   = collections.NewPrefix(0)
	BalancePrefix = collections.NewPrefix(1)
	SupplyPrefix  = collections.NewPrefix(2)
	OwnerPrefix   = collections.NewPrefix(3)
	Type          = "asset-account"
)

var (
	CONTINUOUS_LOCKING_ACCOUNT = "continuous-locking-account"
	DELAYED_LOCKING_ACCOUNT    = "delayed-locking-account"
	PERIODIC_LOCKING_ACCOUNT   = "periodic-locking-account"
	PERMANENT_LOCKING_ACCOUNT  = "permanent-locking-account"
)

type getLockedCoinsFunc = func(ctx context.Context, time time.Time, denoms ...string) (sdk.Coins, error)

// newBaseLockup creates a new BaseLockup object.
func NewAssetAccount(d accountstd.Dependencies) (*AssetAccount, error) {
	fmt.Println("go here, addr codec", d.AddressCodec)
	AssetAccount := &AssetAccount{
		Owner:   collections.NewItem(d.SchemaBuilder, OwnerPrefix, "owner", collections.BytesValue),
		Denom:   collections.NewItem(d.SchemaBuilder, DenomPrefix, "denom", collections.StringValue),
		Balance: collections.NewMap(d.SchemaBuilder, BalancePrefix, "balance", collections.BytesKey, sdk.IntValue),
		Supply:  collections.NewItem(d.SchemaBuilder, SupplyPrefix, "supply", sdk.IntValue),

		addressCodec:  d.AddressCodec,
		headerService: d.Environment.HeaderService,
	}
	return AssetAccount, nil
}

type AssetAccount struct {
	// Owner is the address of the account owner.
	Owner         collections.Item[[]byte]
	Denom         collections.Item[string]
	Balance       collections.Map[[]byte, math.Int]
	Supply        collections.Item[math.Int]
	addressCodec  address.Codec
	headerService header.Service
	transferFunc  func(ctx context.Context, from, to []byte, amount math.Int) error
}

func (aa *AssetAccount) Init(ctx context.Context, msg *assettypes.MsgInitAssetAccountWrapper) (
	*assettypes.MsgInitAssetAccountResponse, error,
) {
	fmt.Println("msg", msg, msg.Owner)
	fmt.Println("transfer func", msg.TransferFunc)
	fmt.Println("address codec", aa.addressCodec)
	owner, err := aa.addressCodec.StringToBytes(msg.Owner)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'owner' address: %s", err)
	}
	err = aa.Owner.Set(ctx, owner)
	if err != nil {
		return nil, err
	}

	err = aa.Denom.Set(ctx, msg.Denom)
	if err != nil {
		return nil, err
	}

	totalSupply := math.ZeroInt()
	for _, balance := range msg.InitBalance {
		totalSupply = totalSupply.Add(balance.Amount)
		fmt.Println("Init addr", balance.Addr)
		err = aa.Balance.Set(ctx, balance.Addr, balance.Amount)
		if err != nil {
			return nil, err
		}
	}

	err = aa.Supply.Set(ctx, totalSupply)
	if err != nil {
		return nil, err
	}

	aa.transferFunc = msg.TransferFunc(aa)

	return &assettypes.MsgInitAssetAccountResponse{}, nil
}

func (aa *AssetAccount) GetDenom(ctx context.Context) (
	string, error,
) {
	denom, err := aa.Denom.Get(ctx)
	if err != nil {
		return "", err
	}
	return denom, nil
}

func (aa *AssetAccount) GetBalance(ctx context.Context, addr []byte) math.Int {
	balance, err := aa.Balance.Get(ctx, addr)
	if err != nil {
		return math.ZeroInt()
	}
	return balance
}

func (aa *AssetAccount) SetBalance(ctx context.Context, addr []byte, amt math.Int) error {
	return aa.Balance.Set(ctx, addr, amt)
}

func (aa *AssetAccount) GetSupply(ctx context.Context) (
	math.Int, error,
) {
	supply, err := aa.Supply.Get(ctx)
	if err != nil {
		return math.ZeroInt(), err
	}
	return supply, nil
}

func (aa *AssetAccount) Transfer(ctx context.Context, msg *assettypes.MsgTransfer) (*assettypes.MsgTransferResponse, error) {
	if msg == nil {
		return nil, errors.New("empty msg")
	}
	err := aa.transferFunc(ctx, msg.From, msg.To, msg.Amount)
	if err != nil {
		return nil, err
	}
	fromBalance := aa.GetBalance(ctx, msg.From)
	toBalance := aa.GetBalance(ctx, msg.To)
	
	return &assettypes.MsgTransferResponse{
		FromBalance: assettypes.Balance{
			Addr:   msg.From,
			Amount: fromBalance,
		},
		ToBalance: assettypes.Balance{
			Addr:   msg.To,
			Amount: toBalance,
		},
	}, nil
}

func (aa *AssetAccount) SubUnlockedCoins(ctx context.Context, addr []byte, amt math.Int) error {
	denom, err := aa.GetDenom(ctx)
	if err != nil {
		return err
	}

	fmt.Println("From addr", addr)
	balance := aa.GetBalance(ctx, addr)
	fmt.Println("GetBalance", balance, err)

	_, err = balance.SafeSub(amt)
	if err != nil {
		return errorsmod.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"%s spendable balance %s is smaller than %s",
			denom, balance, amt,
		)
	}

	newBalance := balance.Sub(amt)

	return aa.SetBalance(ctx, addr, newBalance)

}

func (aa *AssetAccount) AddCoins(ctx context.Context, addr []byte, amt math.Int) error {
	balance := aa.GetBalance(ctx, addr)

	newBalance := balance.Add(amt)

	return aa.SetBalance(ctx, addr, newBalance)
}

// RegisterInitHandler implements implementation.Account.
func (a *AssetAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, a.Init)
}

func (aa *AssetAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, aa.Transfer)
}

// RegisterQueryHandlers implements implementation.Account.
func (a *AssetAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
}
