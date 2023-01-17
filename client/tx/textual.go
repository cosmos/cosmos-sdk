package tx

import (
	"context"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	"cosmossdk.io/x/tx/textual"
	"github.com/cosmos/cosmos-sdk/client"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// NewTextualWithClientCtx returns a new Textual instance where the metadata
// queries are done via gRPC using the provided client.Context.
func NewTextualWithClientCtx(clientCtx client.Context) textual.Textual {
	return textual.NewTextual(func(ctx context.Context, denom string) (*bankv1beta1.Metadata, error) {
		bankQueryClient := bankv1beta1.NewQueryClient(clientCtx)
		res, err := bankQueryClient.DenomMetadata(ctx, &bankv1beta1.QueryDenomMetadataRequest{
			Denom: denom,
		})

		status, ok := grpcstatus.FromError(err)
		if !ok {
			return nil, err
		}

		// This means we didn't find any metadata for this denom. Returning
		// empty metadata.
		if status.Code() == codes.NotFound {
			return nil, nil
		}

		return res.Metadata, nil
	})
}
