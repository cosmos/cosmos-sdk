package tx

import (
	"context"
	"fmt"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	txconfigv1 "cosmossdk.io/api/cosmos/tx/config/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/x/auth/ante"
	"cosmossdk.io/x/auth/posthandler"
	"cosmossdk.io/x/auth/tx"
	authtypes "cosmossdk.io/x/auth/types"
	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/textual"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

func init() {
	appconfig.RegisterModule(&txconfigv1.Config{},
		appconfig.Provide(ProvideModule),
		appconfig.Provide(ProvideProtoRegistry),
	)
}

type ModuleInputs struct {
	depinject.In

	Config                *txconfigv1.Config
	AddressCodec          address.Codec
	ValidatorAddressCodec address.ValidatorAddressCodec
	Codec                 codec.Codec
	ProtoFileResolver     txsigning.ProtoFileResolver
	// BankKeeper is the expected bank keeper to be passed to AnteHandlers
	BankKeeper             authtypes.BankKeeper               `optional:"true"`
	MetadataBankKeeper     BankKeeper                         `optional:"true"`
	AccountKeeper          ante.AccountKeeper                 `optional:"true"`
	FeeGrantKeeper         ante.FeegrantKeeper                `optional:"true"`
	CustomSignModeHandlers func() []txsigning.SignModeHandler `optional:"true"`
	CustomGetSigners       []txsigning.CustomGetSigner        `optional:"true"`
}

type ModuleOutputs struct {
	depinject.Out

	TxConfig        client.TxConfig
	TxConfigOptions tx.ConfigOptions
	BaseAppOption   runtime.BaseAppOption
}

func ProvideProtoRegistry() txsigning.ProtoFileResolver {
	return gogoproto.HybridResolver
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	var customSignModeHandlers []txsigning.SignModeHandler
	if in.CustomSignModeHandlers != nil {
		customSignModeHandlers = in.CustomSignModeHandlers()
	}

	txConfigOptions := tx.ConfigOptions{
		EnabledSignModes: tx.DefaultSignModes,
		SigningOptions: &txsigning.Options{
			FileResolver:          in.ProtoFileResolver,
			AddressCodec:          in.AddressCodec,
			ValidatorAddressCodec: in.ValidatorAddressCodec,
			CustomGetSigners:      make(map[protoreflect.FullName]txsigning.GetSignersFunc),
		},
		CustomSignModes: customSignModeHandlers,
	}

	for _, mode := range in.CustomGetSigners {
		txConfigOptions.SigningOptions.CustomGetSigners[mode.MsgType] = mode.Fn
	}

	// enable SIGN_MODE_TEXTUAL only if bank keeper is available
	if in.MetadataBankKeeper != nil {
		txConfigOptions.EnabledSignModes = append(txConfigOptions.EnabledSignModes, signingtypes.SignMode_SIGN_MODE_TEXTUAL)
		txConfigOptions.TextualCoinMetadataQueryFn = NewBankKeeperCoinMetadataQueryFn(in.MetadataBankKeeper)
	}

	txConfig, err := tx.NewTxConfigWithOptions(in.Codec, txConfigOptions)
	if err != nil {
		panic(err)
	}

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

	return ModuleOutputs{TxConfig: txConfig, TxConfigOptions: txConfigOptions, BaseAppOption: baseAppOption}
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
// This function should be used in the server (app.go) and is already injected thanks to app wiring for app_di.
func NewBankKeeperCoinMetadataQueryFn(bk BankKeeper) textual.CoinMetadataQueryFn {
	return func(ctx context.Context, denom string) (*bankv1beta1.Metadata, error) {
		res, err := bk.DenomMetadataV2(ctx, &bankv1beta1.QueryDenomMetadataRequest{Denom: denom})
		if err != nil {
			return nil, metadataExists(err)
		}

		return res.Metadata, nil
	}
}

// NewGRPCCoinMetadataQueryFn returns a new Textual instance where the metadata
// queries are done via gRPC using the provided GRPC client connection. In the
// SDK, you can pass a client.Context as the GRPC connection.
//
// Example:
//
//	clientCtx := client.GetClientContextFromCmd(cmd)
//	txt := tx.NewTextualWithGRPCConn(clientCtx)
//
// This should be used in the client (root.go) of an application.
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
