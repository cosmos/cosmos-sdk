package keeper_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/nft/types"
)

func (suite *KeeperTestSuite) TestGRPCQueryNFT() {
	type args struct {
		c   context.Context
		req *types.QueryNFTRequest
	}

	nft := types.NFT{
		Id:    "painting1",
		Owner: suite.addrs[0].String(),
		Data:  nil,
	}
	suite.app.NFTkeeper.SetNFT(suite.ctx, nft)

	tests := []struct {
		name    string
		args    args
		want    *types.QueryNFTResponse
		wantErr bool
	}{
		{
			name: "valid request",
			args: args{
				c: context.Background(),
				req: &types.QueryNFTRequest{
					Id: nft.Id,
				},
			},
			want: &types.QueryNFTResponse{
				NFT: &nft,
			},
			wantErr: false,
		},

		{
			name: "empty nft id",
			args: args{
				c:   context.Background(),
				req: &types.QueryNFTRequest{},
			},
			want: &types.QueryNFTResponse{
				NFT: &nft,
			},
			wantErr: true,
		},
		{
			name: "not exist nft id",
			args: args{
				c:   context.Background(),
				req: &types.QueryNFTRequest{Id: "painting"},
			},
			want: &types.QueryNFTResponse{
				NFT: &nft,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			got, err := suite.queryClient.NFT(tt.args.c, tt.args.req)
			if (err != nil) == tt.wantErr {
				return
			}
			suite.Require().Equal(got, tt.want)
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryNFTs() {
	type args struct {
		c   context.Context
		req *types.QueryNFTsRequest
	}

	nfts := []types.NFT{
		{
			Id:    "painting1",
			Owner: suite.addrs[0].String(),
		},
		{
			Id:    "painting2",
			Owner: suite.addrs[0].String(),
		},
		{
			Id:    "painting3",
			Owner: suite.addrs[1].String(),
		},
	}

	for _, nft := range nfts {
		suite.app.NFTkeeper.SetNFT(suite.ctx, nft)
	}

	tests := []struct {
		name    string
		args    args
		want    *types.QueryNFTsResponse
		wantErr bool
	}{
		{
			name: "valid request,two nft",
			args: args{
				c: context.Background(),
				req: &types.QueryNFTsRequest{
					Owner: suite.addrs[0].String(),
				},
			},
			want: &types.QueryNFTsResponse{
				NFTs: []*types.NFT{&nfts[0], &nfts[1]},
			},
			wantErr: false,
		},

		{
			name: "valid request,one nft",
			args: args{
				c: context.Background(),
				req: &types.QueryNFTsRequest{
					Owner: suite.addrs[1].String(),
				},
			},
			want: &types.QueryNFTsResponse{
				NFTs: []*types.NFT{&nfts[2]},
			},
			wantErr: false,
		},
		{
			name: "valid request,query all nft",
			args: args{
				c:   context.Background(),
				req: &types.QueryNFTsRequest{},
			},
			want: &types.QueryNFTsResponse{
				NFTs: []*types.NFT{&nfts[0], &nfts[1], &nfts[2]},
			},
			wantErr: false,
		},
		{
			name: "not exist owner",
			args: args{
				c: context.Background(),
				req: &types.QueryNFTsRequest{
					Owner: suite.addrs[2].String(),
				},
			},
			want: &types.QueryNFTsResponse{
				NFTs: []*types.NFT{nil},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			got, err := suite.queryClient.NFTs(tt.args.c, tt.args.req)
			if (err != nil) == tt.wantErr {
				return
			}
			suite.Require().Equal(got, tt.want)
		})
	}
}
