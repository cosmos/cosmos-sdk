package main

import (
	"context"

	"google.golang.org/grpc"

	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GetAccSeqNumber returns the account number and sequence number for the given address
func GetAccSeqNumber(grpcConn *grpc.ClientConn, address string) (uint64, uint64, error) {
	info, err := auth.NewQueryClient(grpcConn).AccountInfo(context.Background(), &auth.QueryAccountInfoRequest{Address: address})
	if err != nil {
		return 0, 0, err
	}
	return info.Info.GetAccountNumber(), info.Info.GetSequence(), nil
}
