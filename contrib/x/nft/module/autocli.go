package module

import (
	"fmt"

	autocli "cosmossdk.io/core/autocli"

	nftv1beta1 "github.com/cosmos/cosmos-sdk/contrib/api/cosmos/nft/v1beta1"
	"github.com/cosmos/cosmos-sdk/contrib/x/nft"
	"github.com/cosmos/cosmos-sdk/version"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocli.ModuleOptions {
	return &autocli.ModuleOptions{
		Query: &autocli.ServiceCommandDescriptor{
			Service: nftv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "Balance",
					Use:       "balance [owner] [class-id]",
					Short:     "Query the number of NFTs of a given class owned by the owner.",
					Example:   fmt.Sprintf(`%s query %s balance <owner> <class-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "owner"},
						{ProtoField: "class_id"},
					},
				},
				{
					RpcMethod: "Owner",
					Use:       "owner [class-id] [nft-id]",
					Short:     "Query the owner of the NFT based on its class and id.",
					Example:   fmt.Sprintf(`%s query %s owner <class-id> <nft-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "Supply",
					Use:       "supply [class-id]",
					Short:     "Query the number of nft based on the class.",
					Example:   fmt.Sprintf(`%s query %s supply <class-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "class_id"},
					},
				},
				{
					RpcMethod: "NFTs",
					Use:       "nfts [class-id]",
					Short:     "Query all NFTs of a given class or owner address.",
					Example:   fmt.Sprintf(`%s query %s nfts <class-id> --owner=<owner>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "class_id"},
					},
				},
				{
					RpcMethod: "NFT",
					Use:       "nft [class-id] [nft-id]",
					Short:     "Query an NFT based on its class and id.",
					Example:   fmt.Sprintf(`%s query %s nft <class-id> <nft-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "Class",
					Use:       "class [class-id]",
					Short:     "Query an NFT class based on its id",
					Example:   fmt.Sprintf(`%s query %s class <class-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "class_id"},
					},
				},
				{
					RpcMethod: "Classes",
					Use:       "classes",
					Short:     "Query all NFT classes.",
					Example:   fmt.Sprintf(`%s query %s classes`, version.AppName, nft.ModuleName),
				},
			},
		},
		Tx: &autocli.ServiceCommandDescriptor{
			Service: nftv1beta1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocli.RpcCommandOptions{
				{
					RpcMethod: "Send",
					Use:       "send [class-id] [nft-id] [receiver] --from [sender]",
					Short:     "Transfer ownership of NFT",
					PositionalArgs: []*autocli.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
						{ProtoField: "receiver"},
					},
					// Sender is the signer of the transaction and is automatically added as from flag by AutoCLI.
				},
			},
		},
	}
}
