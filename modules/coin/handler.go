package coin

import (
	"fmt"

	"github.com/tendermint/go-wire/data"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/modules/ibc"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

const (
	//NameCoin - name space of the coin module
	NameCoin = "coin"
	// CostSend is GasAllocation per input/output
	CostSend = int64(10)
	// CostCredit is GasAllocation of a credit allocation
	CostCredit = int64(20)
)

// Handler includes an accountant
type Handler struct {
	stack.PassInitValidate
}

var _ stack.Dispatchable = Handler{}

// NewHandler - new accountant handler for the coin module
func NewHandler() Handler {
	return Handler{}
}

// Name - return name space
func (Handler) Name() string {
	return NameCoin
}

// AssertDispatcher - to fulfill Dispatchable interface
func (Handler) AssertDispatcher() {}

// CheckTx checks if there is enough money in the account
func (h Handler) CheckTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, _ sdk.Checker) (res sdk.CheckResult, err error) {

	err = tx.ValidateBasic()
	if err != nil {
		return res, err
	}

	switch t := tx.Unwrap().(type) {
	case SendTx:
		// price based on inputs and outputs
		used := int64(len(t.Inputs) + len(t.Outputs))
		return sdk.NewCheck(used*CostSend, ""), h.checkSendTx(ctx, store, t)
	case CreditTx:
		// default price of 20, constant work
		return sdk.NewCheck(CostCredit, ""), h.creditTx(ctx, store, t)
	}
	return res, errors.ErrUnknownTxType(tx.Unwrap())
}

// DeliverTx moves the money
func (h Handler) DeliverTx(ctx sdk.Context, store state.SimpleDB,
	tx sdk.Tx, cb sdk.Deliver) (res sdk.DeliverResult, err error) {

	err = tx.ValidateBasic()
	if err != nil {
		return res, err
	}

	switch t := tx.Unwrap().(type) {
	case SendTx:
		return h.sendTx(ctx, store, t, cb)
	case CreditTx:
		return res, h.creditTx(ctx, store, t)
	}
	return res, errors.ErrUnknownTxType(tx.Unwrap())
}

// InitState - sets the genesis account balance
func (h Handler) InitState(l log.Logger, store state.SimpleDB,
	module, key, value string, cb sdk.InitStater) (log string, err error) {
	if module != NameCoin {
		return "", errors.ErrUnknownModule(module)
	}
	switch key {
	case "account":
		return setAccount(store, value)
	case "issuer":
		return setIssuer(store, value)
	}
	return "", errors.ErrUnknownKey(key)
}

func (h Handler) sendTx(ctx sdk.Context, store state.SimpleDB,
	send SendTx, cb sdk.Deliver) (res sdk.DeliverResult, err error) {

	err = checkTx(ctx, send)
	if err != nil {
		return res, err
	}

	// deduct from all input accounts
	senders := sdk.Actors{}
	for _, in := range send.Inputs {
		_, err = ChangeCoins(store, in.Address, in.Coins.Negative())
		if err != nil {
			return res, err
		}
		senders = append(senders, in.Address)
	}

	// add to all output accounts
	for _, out := range send.Outputs {
		// TODO: cleaner way, this makes sure we don't consider
		// incoming ibc packets with our chain to be remote packets
		if out.Address.ChainID == ctx.ChainID() {
			out.Address.ChainID = ""
		}

		_, err = ChangeCoins(store, out.Address, out.Coins)
		if err != nil {
			return res, err
		}
		// now send ibc packet if needed...
		if out.Address.ChainID != "" {
			// FIXME: if there are many outputs, we need to adjust inputs
			// so the amounts in and out match.  how?
			inputs := make([]TxInput, len(send.Inputs))
			for i := range send.Inputs {
				inputs[i] = send.Inputs[i]
				inputs[i].Address = inputs[i].Address.WithChain(ctx.ChainID())
			}

			outTx := NewSendTx(inputs, []TxOutput{out})
			packet := ibc.CreatePacketTx{
				DestChain:   out.Address.ChainID,
				Permissions: senders,
				Tx:          outTx,
			}
			ibcCtx := ctx.WithPermissions(ibc.AllowIBC(NameCoin))
			_, err := cb.DeliverTx(ibcCtx, store, packet.Wrap())
			if err != nil {
				return res, err
			}
		}
	}

	// now we build the tags
	tags := make([]cmn.KVPair, 0, 1+len(send.Inputs)+len(send.Outputs))

	tags = append(tags, cmn.KVPair{[]byte("height"), []byte(fmt.Sprintf("%d", int64(ctx.BlockHeight())))})

	for _, in := range send.Inputs {
		addr := []byte(in.Address.Address)
		tags = append(tags, cmn.KVPair{[]byte("coin.sender"), addr})
	}

	for _, out := range send.Outputs {
		addr := []byte(out.Address.Address)
		tags = append(tags, cmn.KVPair{[]byte("coin.receiver"), addr})
	}

	// a-ok!
	return sdk.DeliverResult{Tags: tags}, nil
}

func (h Handler) creditTx(ctx sdk.Context, store state.SimpleDB,
	credit CreditTx) error {

	// first check permissions!!
	info, err := loadHandlerInfo(store)
	if err != nil {
		return err
	}
	if info.Issuer.Empty() || !ctx.HasPermission(info.Issuer) {
		return errors.ErrUnauthorized()
	}

	// load up the account
	addr := ChainAddr(credit.Debitor)
	acct, err := GetAccount(store, addr)
	if err != nil {
		return err
	}

	// make and check changes
	acct.Coins = acct.Coins.Plus(credit.Credit)
	if !acct.Coins.IsNonnegative() {
		return ErrInsufficientFunds()
	}
	acct.Credit = acct.Credit.Plus(credit.Credit)
	if !acct.Credit.IsNonnegative() {
		return ErrInsufficientCredit()
	}

	err = storeAccount(store, addr.Bytes(), acct)
	return err
}

func checkTx(ctx sdk.Context, send SendTx) error {
	// check if all inputs have permission
	for _, in := range send.Inputs {
		if !ctx.HasPermission(in.Address) {
			return errors.ErrUnauthorized()
		}
	}
	return nil
}

func (Handler) checkSendTx(ctx sdk.Context, store state.SimpleDB, send SendTx) error {
	err := checkTx(ctx, send)
	if err != nil {
		return err
	}
	// now make sure there is money
	for _, in := range send.Inputs {
		_, err := CheckCoins(store, in.Address, in.Coins.Negative())
		if err != nil {
			return err
		}
	}
	return nil
}

func setAccount(store state.SimpleDB, value string) (log string, err error) {
	var acc GenesisAccount
	err = data.FromJSON([]byte(value), &acc)
	if err != nil {
		return "", err
	}
	acc.Balance.Sort()
	addr, err := acc.GetAddr()
	if err != nil {
		return "", ErrInvalidAddress()
	}
	// this sets the permission for a public key signature, use that app
	actor := auth.SigPerm(addr)
	err = storeAccount(store, actor.Bytes(), acc.ToAccount())
	if err != nil {
		return "", err
	}
	return "Success", nil
}

// setIssuer sets a permission for some super-powerful account to
// mint money
func setIssuer(store state.SimpleDB, value string) (log string, err error) {
	var issuer sdk.Actor
	err = data.FromJSON([]byte(value), &issuer)
	if err != nil {
		return "", err
	}
	err = storeIssuer(store, issuer)
	if err != nil {
		return "", err
	}
	return "Success", nil
}
