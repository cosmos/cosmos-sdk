package tx

import (
	"context"
	"fmt"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/tx/textual/valuerenderer"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
)

func init() {
	appmodule.Register(&txconfigv1.Config{},
		appmodule.Provide(ProvideModule),
	)
}

//nolint:revive
type TxInputs struct {
	depinject.In

	Config              *txconfigv1.Config
	ProtoCodecMarshaler codec.ProtoCodecMarshaler

	AccountKeeper ante.AccountKeeper `optional:"true"`
	// AuthBankKeeper is the expected bank keeper to be passed to AnteHandlers
	AuthBankKeeper authtypes.BankKeeper `optional:"true"`
	// TxBankKeeper is the expected bank keeper to be passed to Textual
	TxBankKeeper   BankKeeper
	FeeGrantKeeper feegrantkeeper.Keeper `optional:"true"`
}

//nolint:revive
type TxOutputs struct {
	depinject.Out

	TxConfig      client.TxConfig
	BaseAppOption runtime.BaseAppOption
}

func ProvideModule(in TxInputs) TxOutputs {
	textual := NewTextual(in.TxBankKeeper)
	txConfig := tx.NewTxConfigWithTextual(in.ProtoCodecMarshaler, tx.DefaultSignModes, textual)

	baseAppOption := func(app *baseapp.BaseApp) {
		// AnteHandlers
		if !in.Config.SkipAnteHandler {
			anteHandler, err := newAnteHandler(txConfig, in)
			if err != nil {
				panic(err)
			}
			app.SetAnteHandler(anteHandler)
		}

		// PostHandlers
		if !in.Config.SkipPostHandler {
			// In v0.46, the SDK introduces _postHandlers_. PostHandlers are like
			// antehandlers, but are run _after_ the `runMsgs` execution. They are also
			// defined as a chain, and have the same signature as antehandlers.
			//
			// In baseapp, postHandlers are run in the same store branch as `runMsgs`,
			// meaning that both `runMsgs` and `postHandler` state will be committed if
			// both are successful, and both will be reverted if any of the two fails.
			//
			// The SDK exposes a default empty postHandlers chain.
			//
			// Please note that changing any of the anteHandler or postHandler chain is
			// likely to be a state-machine breaking change, which needs a coordinated
			// upgrade.
			postHandler, err := posthandler.NewPostHandler(
				posthandler.HandlerOptions{},
			)
			if err != nil {
				panic(err)
			}
			app.SetPostHandler(postHandler)
		}

		// TxDecoder/TxEncoder
		app.SetTxDecoder(txConfig.TxDecoder())
		app.SetTxEncoder(txConfig.TxEncoder())
	}

	return TxOutputs{TxConfig: txConfig, BaseAppOption: baseAppOption}
}

func newAnteHandler(txConfig client.TxConfig, in TxInputs) (sdk.AnteHandler, error) {
	if in.AuthBankKeeper == nil {
		return nil, fmt.Errorf("both AccountKeeper and BankKeeper are required")
	}

	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   in.AccountKeeper,
			BankKeeper:      in.AuthBankKeeper,
			SignModeHandler: txConfig.SignModeHandler(),
			FeegrantKeeper:  in.FeeGrantKeeper,
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ante handler: %w", err)
	}

	return anteHandler, nil
}

// NewTextual creates a new Textual struct using the given
// BankKeeper to retrieve coin metadata.
func NewTextual(bk BankKeeper) valuerenderer.Textual {
	textual := valuerenderer.NewTextual(func(ctx context.Context, denom string) (*bankv1beta1.Metadata, error) {
		res, err := bk.DenomMetadata(ctx, &types.QueryDenomMetadataRequest{Denom: denom})
		if err != nil {
			return nil, err
		}

		m := &bankv1beta1.Metadata{
			Base:    res.Metadata.Base,
			Display: res.Metadata.Display,
			// fields below this one are not strictly needed by Textual
			// but added here for completeness.
			Description: res.Metadata.Description,
			Name:        res.Metadata.Name,
			Symbol:      res.Metadata.Symbol,
			Uri:         res.Metadata.URI,
			UriHash:     res.Metadata.URIHash,
		}
		m.DenomUnits = make([]*bankv1beta1.DenomUnit, len(res.Metadata.DenomUnits))
		for i, d := range res.Metadata.DenomUnits {
			m.DenomUnits[i] = &bankv1beta1.DenomUnit{
				Denom:    d.Denom,
				Exponent: d.Exponent,
				Aliases:  d.Aliases,
			}
		}

		return m, nil
	})

	return textual
}
