package client

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	sentinel "github.com/cosmos/cosmos-sdk/examples/sut"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

const (
	flagvpnaddr         = "vpn_addr"
	flaguseraddr        = "user_addr"
	flagamount          = "amount"
	flagreceivedbytes   = "receivedBytes"
	flagsessionduration = "sessionduration"
	flagtimestamp       = "timestamp"
	flagsessionid       = "sessionid"
	flagip              = "vpn_ip"
	flagnetspeed        = "netspeed"
	flagppgb            = "price_per_gb"
	flagqueryaddress    = "address"
	flagmspubkey        = "pubkey"
	flaglocation = "location"
) 
func RegisterVpnServiceCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registervpn",
		Short: "Register for sentinel vpn service",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))
			sender, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}
			ip := viper.GetString(flagip)
			if ip == "" {
				return fmt.Errorf("Ip flag is not m entioned")

			}
			ppgb := viper.GetInt64(flagppgb)
			if ppgb > 0 {
				return fmt.Errorf("price per gb not mentioned")
			}
			netspeed := viper.GetInt64(flagnetspeed)
			if netspeed > 0 {
				return fmt.Errorf("net speed not mentioned")
			}
			address, err := sdk.GetAccAddressBech32(ctx.FromAddressName)
			if err != nil {
				return err
			}
			location := viper.GetString(flaglocation)
			if location == "" {
				return fmt.Errorf("location flag is not m entioned")

			}

			msg := sentinel.MsgRegisterVpnService{sender, ip,netspeed,ppgb,location}
			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			fmt.Printf("Vpn serivicer registered with address: %s\n", sender)
			return nil
		},
	}
	cmd.Flags().String(flagvpnaddr, "", "ddress")
	cmd.Flags().String(flagip, "", "ip")
	cmd.Flags().Int64(flagnetspeed, "", "net speed")
	cmd.Flags().Int64(flagppgb, "", "price per gb")
	cmd.Flags().String(flaglocation),"","location of vpn service provider")
	return cmd
}

//
//
//

func RegisterMasterNodeCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register_mn_node",
		Short: "Register Master node",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))
			sender, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}
			msg := sentinel.MsgRegisterMasterNode{sender}
			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			fmt.Printf("Masternode registered with address: %s\n", sender)
			return nil
		},
	}
	//cmd.Flags().String(flagmspubkey, "", "register master node")
	return cmd
}

func QueryVpnServiceCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query_vpn [address]",
		Short: "query for sentinel vpn service",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper()
			// sender, err := ctx.GetFromAddress()
			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				 return err
			}
			res,err := ctx.QueryStore(key, storeName)
			if err != nil {
				return err
			} else if len(res) == 0 {
				return fmt.Errorf("No vpn service provider found with address %s", args[0])
			}
			vpn_service := types.MustUnmarshalValidator(cdc, addr, res)
			switch viper.Get(cli.OutputFlag) {
			case "text":
				human, err := validator.HumanReadableString()
				if err != nil {
					return err
				}
				fmt.Println(human)

			case "json":
				// parse out the validator
				output, err := wire.MarshalJSONIndent(cdc, validator)
				if err != nil {
					return err
				}
				fmt.Println(string(output))
			}
			// TODO output with proofs / machine parseable etc.
			return cmd
		},
	}

	return cmd
}

			// msg := sentinel.MsgQueryRegisteredVpnService{address}
			// res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			//fmt.Printf("Vpn serivicer registered with address: %s\n", sender)
			return nil
		},
	}
	cmd.Flags().String(flagqueryaddress, "", "to query vpn service")
	return cmd
}
func QueryMasterNodeCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query_master [address]",
		Short: "query to master node",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper()
			// sender, err := ctx.GetFromAddress()
			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				 return err
			}
			res,err := ctx.QueryStore(key, storeName)
			if err != nil {
				return err
			} else if len(res) == 0 {
				return fmt.Errorf("No vpn service provider found with address %s", args[0])
			}
			vpn_service := types.MustUnmarshalValidator(cdc, addr, res)
			switch viper.Get(cli.OutputFlag) {
			case "text":
				human, err := validator.HumanReadableString()
				if err != nil {
					return err
				}
				fmt.Println(human)

			case "json":
				// parse out the validator
				output, err := wire.MarshalJSONIndent(cdc, validator)
				if err != nil {
					return err
				}
				fmt.Println(string(output))
			}
			// TODO output with proofs / machine parseable etc.
			return cmd
		},
	}
	return cmd
}

func UnRegisterMasterNodeCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unreg_msnode",
		Short: "Unregister Master node",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))
			sender, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}
			addr := viper.GetString(flagqueryaddress)
			if addr == "" {
				return fmt.Errorf("address is not mentioned")

			}
			key, err := sdk.GetAccAddressBech32(addr)
			if err != nil {
				return err
			}
			msg := sentinel.MsgDeleteMasterNode{key}
			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			//fmt.Printf("Masternode registered with address: %s\n", sender)
			return nil
		},
	}
	cmd.Flags().String(flagmspubkey, "", "register master node")
	return cmd
}
func UnRegisterVpnServiceCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unreg_vpn",
		Short: "Unregister vpn service",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))
			sender, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}
			addr := viper.GetString(flagqueryaddress)
			if addr == "" {
				return fmt.Errorf("address is not mentioned")

			}
			key, err := sdk.GetAccAddressBech32(addr)
			if err != nil {
				return err
			}

			msg := sentinel.MsgDeleteVpnUser{key}
			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			//fmt.Printf("Masternode registered with address: %s\n", sender)
			return nil
		},
	}
	cmd.Flags().String(flagqueryaddress, "", "Unregister vpn node")
	return cmd
}
