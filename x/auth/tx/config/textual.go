package tx

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoregistry"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"
	"cosmossdk.io/x/tx/signing/directaux"
	"cosmossdk.io/x/tx/signing/textual"
	types2 "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// NewTextualWithGRPCConn returns a new Textual instance where the metadata
// queries are done via gRPC using the provided GRPC client connection. In the
// SDK, you can pass a client.Context as the GRPC connection.
//
// Example:
//
//	clientCtx := client.GetClientContextFromCmd(cmd)
//	txt := tx.NewTextualWithGRPCConn(clientCtxx)
//
// TODO: rename
func NewTextualWithGRPCConn(grpcConn grpc.ClientConnInterface) (tx.SignModeOptions, error) {
	protoFiles := types2.MergedProtoRegistry()
	typeResolver := protoregistry.GlobalTypes
	signersContext, err := txsigning.NewGetSignersContext(txsigning.GetSignersOptions{ProtoFiles: protoFiles})
	if err != nil {
		return tx.SignModeOptions{}, err
	}

	aminoJSONEncoder := aminojson.NewAminoJSON()
	signModeOptions := tx.SignModeOptions{
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
			CoinMetadataQuerier: func(ctx context.Context, denom string) (*bankv1beta1.Metadata, error) {
				bankQueryClient := bankv1beta1.NewQueryClient(grpcConn)
				res, err := bankQueryClient.DenomMetadata(ctx, &bankv1beta1.QueryDenomMetadataRequest{
					Denom: denom,
				})
				if err != nil {
					return nil, metadataExists(err)
				}

				return res.Metadata, nil
			},
			FileResolver: protoFiles,
			TypeResolver: typeResolver,
		},
	}

	return signModeOptions, nil
}

// NewTextualWithBankKeeper creates a new Textual struct using the given
// BankKeeper to retrieve coin metadata.
//
// Note: Once we switch to ADR-033, and keepers become ADR-033 clients to each
// other, this function could probably be deprecated in favor of
// `NewTextualWithGRPCConn`.
// TODO: rename
func NewTextualWithBankKeeper(bk BankKeeper) (tx.SignModeOptions, error) {
	protoFiles := types2.MergedProtoRegistry()
	typeResolver := protoregistry.GlobalTypes
	signersContext, err := txsigning.NewGetSignersContext(txsigning.GetSignersOptions{ProtoFiles: protoFiles})
	if err != nil {
		return tx.SignModeOptions{}, err
	}

	txtOpts := textual.SignModeOptions{
		CoinMetadataQuerier: func(ctx context.Context, denom string) (*bankv1beta1.Metadata, error) {
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
		},
		FileResolver: protoFiles,
	}

	aminoJSONEncoder := aminojson.NewAminoJSON()
	return tx.SignModeOptions{
		Textual: &txtOpts,
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
	}, nil
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
