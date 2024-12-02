package asset

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	assettypes "cosmossdk.io/x/accounts/defaults/asset/v1"

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

// newBaseLockup creates a new BaseLockup object.
func NewAssetAccount(d accountstd.Dependencies) (*AssetAccount, error) {
	AssetAccount := &AssetAccount{
		Owner:        collections.NewItem(d.SchemaBuilder, OwnerPrefix, "owner", collections.BytesValue),
		Denom:        collections.NewItem(d.SchemaBuilder, DenomPrefix, "denom", collections.StringValue),
		Balance:      collections.NewMap(d.SchemaBuilder, BalancePrefix, "balance", collections.BytesKey, sdk.IntValue),
		Supply:       collections.NewItem(d.SchemaBuilder, SupplyPrefix, "supply", sdk.IntValue),
		transferFunc: make(map[string]func(ctx context.Context, from, to []byte, amount math.Int) ([][]byte, error)),
		mintFunc:     make(map[string]func(ctx context.Context, to []byte, amount math.Int) ([][]byte, error)),
		burnFunc:     make(map[string]func(ctx context.Context, from []byte, amount math.Int) ([][]byte, error)),

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
	transferFunc  map[string]func(ctx context.Context, from, to []byte, amount math.Int) ([][]byte, error)
	mintFunc      map[string]func(ctx context.Context, to []byte, amount math.Int) ([][]byte, error)
	burnFunc      map[string]func(ctx context.Context, from []byte, amount math.Int) ([][]byte, error)
}

func (aa *AssetAccount) Init(ctx context.Context, msg *assettypes.MsgInitAssetAccountWrapper) (
	*assettypes.MsgInitAssetAccountResponse, error,
) {
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
		err = aa.Balance.Set(ctx, balance.Addr, balance.Amount)
		if err != nil {
			return nil, err
		}
	}

	err = aa.Supply.Set(ctx, totalSupply)
	if err != nil {
		return nil, err
	}

	aa.transferFunc[msg.Denom] = msg.TransferFunc(aa)
	aa.mintFunc[msg.Denom] = msg.MintFunc(aa)
	aa.burnFunc[msg.Denom] = msg.BurnFunc(aa)

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

func (aa *AssetAccount) GetOwner(ctx context.Context) (
	[]byte, error,
) {
	owner, err := aa.Owner.Get(ctx)
	if err != nil {
		return []byte{}, err
	}
	return owner, nil
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

func (aa *AssetAccount) GetSupply(ctx context.Context) math.Int {
	supply, err := aa.Supply.Get(ctx)
	if err != nil {
		return math.ZeroInt()
	}
	return supply
}

func (aa *AssetAccount) SetSupply(ctx context.Context, supply math.Int) error {
	return aa.Supply.Set(ctx, supply)
}

func (aa *AssetAccount) Transfer(ctx context.Context, msg *assettypes.MsgTransfer) (*assettypes.MsgTransferResponse, error) {
	if msg == nil {
		return nil, errors.New("empty msg")
	}
	denom, err := aa.GetDenom(ctx)
	if err != nil {
		return nil, err
	}
	changeAddr, err := aa.transferFunc[denom](ctx, msg.From, msg.To, msg.Amount)
	if err != nil {
		return nil, err
	}
	resp := &assettypes.MsgTransferResponse{
		Supply: aa.GetSupply(ctx),
	}

	for _, addr := range changeAddr {
		balance := aa.GetBalance(ctx, addr)
		resp.Balances = append(resp.Balances, assettypes.Balance{Addr: addr, Amount: balance})
	}

	return resp, nil
}

func (aa *AssetAccount) Mint(ctx context.Context, msg *assettypes.MsgMint) (*assettypes.MsgMintResponse, error) {
	if msg == nil {
		return nil, errors.New("empty msg")
	}
	denom, err := aa.GetDenom(ctx)
	if err != nil {
		return nil, err
	}
	changeAddr, err := aa.mintFunc[denom](ctx, msg.To, msg.Amount)
	if err != nil {
		return nil, err
	}
	resp := &assettypes.MsgMintResponse{
		Supply: aa.GetSupply(ctx),
	}

	for _, addr := range changeAddr {
		balance := aa.GetBalance(ctx, addr)
		resp.Balances = append(resp.Balances, assettypes.Balance{Addr: addr, Amount: balance})
	}

	return resp, nil
}

func (aa *AssetAccount) Burn(ctx context.Context, msg *assettypes.MsgBurn) (*assettypes.MsgBurnResponse, error) {
	if msg == nil {
		return nil, errors.New("empty msg")
	}

	denom, err := aa.GetDenom(ctx)
	if err != nil {
		return nil, err
	}

	changeAddr, err := aa.burnFunc[denom](ctx, msg.From, msg.Amount)
	if err != nil {
		return nil, err
	}
	resp := &assettypes.MsgBurnResponse{
		Supply: aa.GetSupply(ctx),
	}

	for _, addr := range changeAddr {
		balance := aa.GetBalance(ctx, addr)
		resp.Balances = append(resp.Balances, assettypes.Balance{Addr: addr, Amount: balance})
	}

	return resp, nil
}

func (aa *AssetAccount) QueryOwner(ctx context.Context, msg *assettypes.QueryOwnerRequest) (*assettypes.QueryOwnerResponse, error) {
	if msg == nil {
		return nil, errors.New("empty msg")
	}
	owner, err := aa.GetOwner(ctx)
	if err != nil {
		return nil, err
	}

	return &assettypes.QueryOwnerResponse{
		Owner: owner,
	}, nil
}

func (aa *AssetAccount) SubUnlockedCoins(ctx context.Context, addr []byte, amt math.Int) error {
	denom, err := aa.GetDenom(ctx)
	if err != nil {
		return err
	}

	balance := aa.GetBalance(ctx, addr)
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
	accountstd.RegisterExecuteHandler(builder, aa.Mint)
	accountstd.RegisterExecuteHandler(builder, aa.Burn)
}

// RegisterQueryHandlers implements implementation.Account.
func (a *AssetAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
}
