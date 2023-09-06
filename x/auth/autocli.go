package auth

import (
	"fmt"

	authv1beta1 "cosmossdk.io/api/cosmos/auth/v1beta1"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	_ "cosmossdk.io/api/cosmos/crypto/secp256k1" // register to that it shows up in protoregistry.GlobalTypes
	_ "cosmossdk.io/api/cosmos/crypto/secp256r1" // register to that it shows up in protoregistry.GlobalTypes

	"github.com/cosmos/cosmos-sdk/version"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: authv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Accounts",
					Use:       "accounts",
					Short:     "Query all the accounts",
				},
				{
					RpcMethod:      "Account",
					Use:            "account [address]",
					Short:          "Query account by address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				{
					RpcMethod:      "AccountInfo",
					Use:            "account-info [address]",
					Short:          "Query account info which is common to all account types.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address"}},
				},
				{
					RpcMethod:      "AccountAddressByID",
					Use:            "address-by-acc-num [acc-num]",
					Short:          "Query account address by account number",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "id"}},
				},
				{
					RpcMethod: "ModuleAccounts",
					Use:       "module-accounts",
					Short:     "Query all module accounts",
				},
				{
					RpcMethod:      "ModuleAccountByName",
					Use:            "module-account [module-name]",
					Short:          "Query module account info by module name",
					Example:        fmt.Sprintf("%s q auth module-account gov", version.AppName),
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "name"}},
				},
				{
					RpcMethod:      "AddressBytesToString",
					Use:            "address-bytes-to-string [address-bytes]",
					Short:          "Transform an address bytes to string",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address_bytes"}},
				},
				{
					RpcMethod:      "AddressStringToBytes",
					Use:            "address-string-to-bytes [address-string]",
					Short:          "Transform an address string to bytes",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "address_string"}},
				},
				{
					RpcMethod: "Bech32Prefix",
					Use:       "bech32-prefix",
					Short:     "Query the chain bech32 prefix (if applicable)",
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current auth parameters",
				},
			},
		},
		// Tx is purposely left empty, as the only tx is MsgUpdateParams which is gov gated.
	}
}
