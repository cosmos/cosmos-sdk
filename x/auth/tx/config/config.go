package tx

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoregistry"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	txconfigv1 "cosmossdk.io/api/cosmos/tx/config/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
	"cosmossdk.io/x/tx/signing/direct"
	"cosmossdk.io/x/tx/signing/directaux"
	"cosmossdk.io/x/tx/signing/textual"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/registry"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func init() {
	appmodule.Register(&txconfigv1.Config{},
		appmodule.Provide(ProvideModule),
		appmodule.Provide(ProvideSignModeOptions),
	)
}

type ModuleInputs struct {
	depinject.In
	Config              *txconfigv1.Config
	ProtoCodecMarshaler codec.ProtoCodecMarshaler
	SignModeOptions     tx.SignModeOptions
	// BankKeeper is the expected bank keeper to be passed to AnteHandlers
	BankKeeper             authtypes.BankKeeper               `optional:"true"`
	AccountKeeper          ante.AccountKeeper                 `optional:"true"`
	FeeGrantKeeper         ante.FeegrantKeeper                `optional:"true"`
	CustomSignModeHandlers func() []txsigning.SignModeHandler `optional:"true"`
}

type ModuleOutputs struct {
	depinject.Out

	TxConfig      client.TxConfig
	BaseAppOption runtime.BaseAppOption
}

// ProvideSignModeOptions provides the default x/tx SignModeOptions for the SDK.
func ProvideSignModeOptions(bk BankKeeper) tx.SignModeOptions {
	opts, err := NewSignModeOptionsWithMetadataQueryFn(NewBankKeeperCoinMetadataQueryFn(bk))
	if err != nil {
		panic(err)
	}
	return opts
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	var txConfig client.TxConfig
	var customSignModeHandlers []txsigning.SignModeHandler
	if in.CustomSignModeHandlers != nil {
		customSignModeHandlers = in.CustomSignModeHandlers()
	}
	txConfig = tx.NewTxConfigWithOptions(in.ProtoCodecMarshaler, in.SignModeOptions, customSignModeHandlers...)

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

	return ModuleOutputs{TxConfig: txConfig, BaseAppOption: baseAppOption}
}

func newAnteHandler(txConfig client.TxConfig, in ModuleInputs) (sdk.AnteHandler, error) {
	if in.BankKeeper == nil {
		return nil, fmt.Errorf("both AccountKeeper and BankKeeper are required")
	}

	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   in.AccountKeeper,
			BankKeeper:      in.BankKeeper,
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

// NewBankKeeperCoinMetadataQueryFn creates a new Textual struct using the given
// BankKeeper to retrieve coin metadata.
//
// Note: Once we switch to ADR-033, and keepers become ADR-033 clients to each
// other, this function could probably be deprecated in favor of
// `NewTextualWithGRPCConn`.
func NewBankKeeperCoinMetadataQueryFn(bk BankKeeper) textual.CoinMetadataQueryFn {
	return func(ctx context.Context, denom string) (*bankv1beta1.Metadata, error) {
		res, err := bk.DenomMetadata(ctx, &types.QueryDenomMetadataRequest{Denom: denom})
		if err != nil {
			return nil, metadataExists(err)
		}

		m := &bankv1beta1.Metadata{
			Base:    res.Metadata.Base,
			Display: res.Metadata.Display,
			// fields below are not strictly needed by Textual
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
	}
}

// NewGRPCCoinMetadataQueryFn returns a new Textual instance where the metadata
// queries are done via gRPC using the provided GRPC client connection. In the
// SDK, you can pass a client.Context as the GRPC connection.
//
// Example:
//
//	clientCtx := client.GetClientContextFromCmd(cmd)
//	txt := tx.NewTextualWithGRPCConn(clientCtxx)
func NewGRPCCoinMetadataQueryFn(grpcConn grpc.ClientConnInterface) textual.CoinMetadataQueryFn {
	return func(ctx context.Context, denom string) (*bankv1beta1.Metadata, error) {
		bankQueryClient := bankv1beta1.NewQueryClient(grpcConn)
		res, err := bankQueryClient.DenomMetadata(ctx, &bankv1beta1.QueryDenomMetadataRequest{
			Denom: denom,
		})
		if err != nil {
			return nil, metadataExists(err)
		}

		return res.Metadata, nil
	}
}

// NewSignModeOptionsWithMetadataQueryFn creates a new SignModeOptions instance
func NewSignModeOptionsWithMetadataQueryFn(fn textual.CoinMetadataQueryFn) (tx.SignModeOptions, error) {
	protoFiles := registry.MergedProtoRegistry()
	typeResolver := protoregistry.GlobalTypes
	signersContext, err := txsigning.NewGetSignersContext(txsigning.GetSignersOptions{ProtoFiles: protoFiles})
	if err != nil {
		return tx.SignModeOptions{}, err
	}

	aminoJSONEncoder := aminojson.NewAminoJSON()
	signModeOptions := tx.SignModeOptions{
		Direct: &direct.SignModeHandler{},
		DirectAux: &directaux.SignModeHandlerOptions{
			FileResolver:   protoFiles,
			TypeResolver:   typeResolver,
			SignersContext: signersContext,
		},
		AminoJSON: &aminojson.SignModeHandlerOptions{
			FileResolver: protoFiles,
			TypeResolver: typeResolver,
			Encoder:      &aminoJSONEncoder,
		},
		Textual: &textual.SignModeOptions{
			CoinMetadataQuerier: fn,
			FileResolver:        protoFiles,
			TypeResolver:        typeResolver,
		},
	}

	return signModeOptions, nil
}

// metadataExists parses the error, and only propagates the error if it's
// different than a "not found" error.
func metadataExists(err error) error {
	status, ok := grpcstatus.FromError(err)
	if !ok {
		return err
	}

	// This means we didn't find any metadata for this denom. Returning
	// empty metadata.
	if status.Code() == codes.NotFound {
		return nil
	}

	return err
}
