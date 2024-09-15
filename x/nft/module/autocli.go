package module

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	nftv1beta1 "cosmossdk.io/api/cosmos/nft/v1beta1"
	"cosmossdk.io/x/nft"

	"github.com/cosmos/cosmos-sdk/version"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: nftv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Balance",
					Use:       "balance <owner> <class-id>",
					Short:     "Query the number of NFTs of a given class owned by the owner.",
					Example:   fmt.Sprintf(`%s query %s balance <owner> <class-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "owner"},
						{ProtoField: "class_id"},
					},
				},
				{
					RpcMethod: "Owner",
					Use:       "owner <class-id> <nft-id>",
					Short:     "Query the owner of the NFT based on its class and id.",
					Example:   fmt.Sprintf(`%s query %s owner <class-id> <nft-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "Supply",
					Use:       "supply <class-id>",
					Short:     "Query the number of nft based on the class.",
					Example:   fmt.Sprintf(`%s query %s supply <class-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
					},
				},
				{
					RpcMethod: "NFTs",
					Use:       "nfts <class-id>",
					Short:     "Query all NFTs of a given class or owner address.",
					Example:   fmt.Sprintf(`%s query %s nfts <class-id> --owner=<owner>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
					},
				},
				{
					RpcMethod: "NFT",
					Use:       "nft <class-id> <nft-id>",
					Short:     "Query an NFT based on its class and id.",
					Example:   fmt.Sprintf(`%s query %s nft <class-id> <nft-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "Class",
					Use:       "class <class-id>",
					Short:     "Query an NFT class based on its id",
					Example:   fmt.Sprintf(`%s query %s class <class-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
					},
				},
				{
					RpcMethod: "Royalties",
					Use:       "royalties <class-id> <nft-id>",
					Short:     "Query the royalties of an NFT",
					Long:      "Query the accumulated royalties for a specific NFT based on its class and id.",
					Example:   fmt.Sprintf(`%s query %s royalties my-music-nfts song-001`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "Classes",
					Use:       "classes",
					Short:     "Query all NFT classes.",
					Example:   fmt.Sprintf(`%s query %s classes`, version.AppName, nft.ModuleName),
				},
				{
					RpcMethod: "TotalPlays",
					Use:       "total-plays <class-id> <nft-id>",
					Short:     "Query the total number of plays for an NFT",
					Long:      "Query the total number of plays for a specific NFT based on its class and id.",
					Example:   fmt.Sprintf(`%s query %s total-plays my-music-nfts song-001`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "TotalRoyalties",
					Use:       "total-royalties <class-id> <nft-id>",
					Short:     "Query the total royalties generated for an NFT",
					Long:      "Query the total royalties generated for a specific NFT based on its class and id.",
					Example:   fmt.Sprintf(`%s query %s total-royalties my-music-nfts song-001`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "ListedNFTs",
					Use:       "listed-nfts",
					Short:     "Query all NFTs listed on the marketplace",
					Long:      "Query all NFTs that are currently listed for sale on the marketplace",
				},
				{
					RpcMethod: "ListedNFT",
					Use:       "listed-nft [class-id] [nft-id]",
					Short:     "Query a single listed NFT on the marketplace",
					Long:      "Query details of a specific NFT listed for sale on the marketplace",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: nftv1beta1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Send",
					Use:       "send <class-id> <nft-id> <receiver> --from <sender>",
					Short:     "Transfer ownership of an NFT.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
						{ProtoField: "receiver"},
					},
				},
				{
					RpcMethod: "MintNFT",
					Use:       "mint [class-id] [nft-id] [uri] [uri-hash] --from [sender]",
					Short:     "Mint a new NFT",
					Long:      "Mint a new NFT, automatically creating the class if it doesn't exist. Class name, symbol, and description are optional.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
						{ProtoField: "uri"},
						{ProtoField: "uri_hash"},
					},
				},
				{
					RpcMethod: "BurnNFT",
					Use:       "burn <class-id> <nft-id> --from <sender>",
					Short:     "Burn an NFT.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "StakeNFT",
					Use:       "stake <class-id> <nft-id> <stake-duration> --from <sender>",
					Short:     "Stake an NFT for a specified duration.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
						{ProtoField: "stake_duration"},
					},
				},

				{
					RpcMethod: "StreamNFT",
					Use:       "stream <class-id> <nft-id> <payment> --from <sender>",
					Short:     "Stream an NFT and pay royalties.",
					Long:      "Stream an NFT by paying the specified amount. Royalties will be automatically distributed among creator, platform, and owner.",
					Example:   fmt.Sprintf(`%s tx %s stream my-music-nfts song-001 10token --from alice`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
						{ProtoField: "payment"},
					},
				},
				{
					RpcMethod: "WithdrawRoyalties",
					Use:       "withdraw-royalties <class-id> <nft-id> <role> --from <recipient>",
					Short:     "Withdraw accumulated royalties for a specific role.",
					Long:      "Withdraw accumulated royalties for a specific role (creator, platform, or owner) from an NFT.",
					Example:   fmt.Sprintf(`%s tx %s withdraw-royalties my-music-nfts song-001 creator --from alice`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
						{ProtoField: "role"},
					},
				},
				{
					RpcMethod: "ListNFT",
					Use:       "list-nft [class-id] [nft-id] [price]",
					Short:     "List an NFT for sale on the marketplace",
					Long:      "List an NFT for sale on the marketplace with a specified price",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
						{ProtoField: "price"},
					},
				},
				{
					RpcMethod: "BuyNFT",
					Use:       "buy-nft [class-id] [nft-id]",
					Short:     "Buy an NFT from the marketplace",
					Long:      "Purchase an NFT that is listed for sale on the marketplace",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "DelistNFT",
					Use:       "delist-nft [class-id] [nft-id]",
					Short:     "Delist an NFT from the marketplace",
					Long:      "Remove an NFT from the marketplace listing",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
					},
				},
			},
		},
	}
}
