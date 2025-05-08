// Deprecated: This package is deprecated and will be removed in the next major release. The `x/nft` module will be moved to a separate repo `github.com/cosmos/cosmos-sdk-legacy`.
package module

import (
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	nftv1beta1 "cosmossdk.io/api/cosmos/nft/v1beta1"

	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/nft" //nolint:staticcheck // deprecated and to be removed
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: nftv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Balance",
					Use:       "balance [owner] [class-id]",
					Short:     "Query the number of NFTs of a given class owned by the owner.",
					Example:   fmt.Sprintf(`%s query %s balance <owner> <class-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "owner"},
						{ProtoField: "class_id"},
					},
				},
				{
					RpcMethod: "Owner",
					Use:       "owner [class-id] [nft-id]",
					Short:     "Query the owner of the NFT based on its class and id.",
					Example:   fmt.Sprintf(`%s query %s owner <class-id> <nft-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "Supply",
					Use:       "supply [class-id]",
					Short:     "Query the number of nft based on the class.",
					Example:   fmt.Sprintf(`%s query %s supply <class-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
					},
				},
				{
					RpcMethod: "NFTs",
					Use:       "nfts [class-id]",
					Short:     "Query all NFTs of a given class or owner address.",
					Example:   fmt.Sprintf(`%s query %s nfts <class-id> --owner=<owner>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
					},
				},
				{
					RpcMethod: "NFT",
					Use:       "nft [class-id] [nft-id]",
					Short:     "Query an NFT based on its class and id.",
					Example:   fmt.Sprintf(`%s query %s nft <class-id> <nft-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "class_id"},
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "Class",
					Use:       "class [class-id]",
					Short:     "Query an NFT class based on its id",
					Example:   fmt.Sprintf(`%s query %s class <class-id>`, version.AppName, nft.ModuleName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
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
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: nftv1beta1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Send",
					Use:       "send [class-id] [nft-id] [receiver] --from [sender]",
					Short:     "Transfer ownership of NFT",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
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
