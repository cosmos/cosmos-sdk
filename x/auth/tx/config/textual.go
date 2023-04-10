package tx

import (
	"context"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/x/tx/signing/textual"

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
func NewTextualWithGRPCConn(grpcConn grpc.ClientConnInterface) (*textual.SignModeHandler, error) {
	protoFiles, err := gogoproto.MergedRegistry()
	if err != nil {
		return nil, err
	}

	return textual.NewSignModeHandler(textual.SignModeOptions{
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
	})
}

// NewTextualWithBankKeeper creates a new Textual struct using the given
// BankKeeper to retrieve coin metadata.
//
// Note: Once we switch to ADR-033, and keepers become ADR-033 clients to each
// other, this function could probably be deprecated in favor of
// `NewTextualWithGRPCConn`.
func NewTextualWithBankKeeper(bk BankKeeper) (*textual.SignModeHandler, error) {
	protoFiles, err := gogoproto.MergedRegistry()
	if err != nil {
		return nil, err
	}

	return textual.NewSignModeHandler(textual.SignModeOptions{
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
	})
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
