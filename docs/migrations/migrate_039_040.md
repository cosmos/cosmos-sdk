# Update to Cosmos-SDK 0.40

### Global changes
* Previously, `context.CLIContext` imported from `github.com/cosmos/cosmos-sdk/client/context` is changed to `client.Context`, which we import from` github.com/cosmos/cosmos-sdk/client`
* `RegisterCodec` is renamed to `RegisterLegacyAminoCodec`

### `codec.go`
* `codec.New()` is changed to `codec.NewLegacyAmino()`.
* We declare a new variable `ModuleCdc` which can be either `NewAminoCodec` or `NewProtoCodec` based on module requirements.
* `RegisterCodec` is renamed to `RegisterLegacyAminoCodec` with same logic.
* Added `RegisterInterfaces` method in which we add `RegisterImplementations` and `RegisterInterfaces` based on msgs and interfaces in module.
    ```go
    func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
        registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitProposal{},
	    )
	    registry.RegisterInterface(
		"cosmos.gov.v1beta1.Content",
		(*Content)(nil),
		&TextProposal{},
	    )
    }
    ```
* `init()` changed from:
    ```go
    func init() {
	    RegisterCodec(cdc)
    }
    ```
    to:
    ```go
    func init() {
	    RegisterLegacyAminoCodec(amino)
	    cryptocodec.RegisterCrypto(amino)
    }
    ```
    Above `crptocodec` is imported from `github.com/cosmos/cosmos-sdk/crypto/codec`
    
### `handler.go`
* There are not many changed need to be done in handler. Just need to change msg to pointer, as msg will be converted to proto. For example, if you have `types.MsgSend`, then it will be modified to `*types.MsgSend`. And one more change we have is, `ctx.EventManager().Events()` is changed to `ctx.EventManager().ABCIEvents()`.
### `keeper.go`
* `*codec.Codec` is changed to any of `codec.BinaryMarshaler`,`codec.JSONMarshaler`, `codec.Marshaler` based on your requirements.

### `tx.go`
* `github.com/cosmos/cosmos-sdk/x/auth/client/utils`, `github.com/cosmos/cosmos-sdk/client/context` are removed now.
* `flags.PostCommands` is removed now and you can add flags by adding line `flags.AddTxFlagsToCmd(cmd)` in each cmd logic.
* Below code:
    ```go
    ctx := context.NewCLIContext().WithCodec(cdc)
    bldr := auth.NewTxBuilderFromCLI(os.Stdin).WithTxEncoder(utils.GetTxEncoder(cdc))
    ```
    is changed to:
    ```go
    clientCtx := client.GetClientContextFromCmd(cmd)
    clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
	if err != nil {
	    return err
	}
    ```
* `utils.GenerateOrBroadcastMsgs(ctx, bldr, []sdk.Msg{msg})` is changed to `tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)`. Here `tx` is imported from `github.com/cosmos/cosmos-sdk/client/tx`

### `query.go`
* `github.com/cosmos/cosmos-sdk/client/context` is removed now.
* `flags.GetCommands` is removed now and you can add flags by adding line `flags.AddQueryFlagsToCmd(cmd)` in each cmd logic.
* We will be going to use `NewQueryClient` from generated `query.pb.go` instead of `QueryWithData`. We will be declaring grpc queries seperately same like querier previously and use those methods here. Please check below code for changes:
    ```go
    cliCtx := context.NewCLIContext().WithCodec(cdc)

    valAddr, err := sdk.ValAddressFromBech32(args[1])
    if err != nil {
        return err
    }

    delAddr, err := sdk.AccAddressFromBech32(args[0])
    if err != nil {
        return err
    }

    bz, err := cdc.MarshalJSON(types.NewQueryBondsParams(delAddr, valAddr))
    if err != nil {
        return err
    }

    route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryUnbondingDelegation)
    res, _, err := cliCtx.QueryWithData(route, bz)
    if err != nil {
        return err
    }

    return cliCtx.PrintOutput(types.MustUnmarshalUBD(cdc, res))
    ```
    The above code is changed to:
    ```go
    clientCtx := client.GetClientContextFromCmd(cmd)
	clientCtx, err := client.ReadQueryCommandFlags(clientCtx, cmd.Flags())
	if err != nil {
        return err
	}

	queryClient := types.NewQueryClient(clientCtx)
	valAddr, err := sdk.ValAddressFromBech32(args[1])
    if err != nil {
        return err
    }

    delAddr, err := sdk.AccAddressFromBech32(args[0])
    if err != nil {
        return err
    }

    params := &types.QueryUnbondingDelegationRequest{
        DelegatorAddr: delAddr.String(),
        ValidatorAddr: valAddr.String(),
    }

    res, err := queryClient.UnbondingDelegation(context.Background(), params)
    if err != nil {
        return err
    }

    return clientCtx.PrintOutput(&res.Unbond)
    ```

### `module.go`
* `type AppModuleBasic struct{}` is  updated to:

    ```go
    type AppModuleBasic struct {
        cdc codec.Marshaler
    }
    ```
* `RegisterCodec(cdc *codec.Codec)` method is changed to `RegisterLegacyAminoCodec(cdc *codec.LegacyAmino)`
* Added `RegisterInterfaces` method which implements `AppModuleBasic` which takes one parameter of type `"github.com/cosmos/cosmos-sdk/codec/types".InterfaceRegistry`. This method is used for registering interface types of module.
    ```go
    func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
        // register all interfaces in module. 
        // We will register all interfaces generally in types/codec.go of module .
    }
    ```
* `func (AppModuleBasic) DefaultGenesis() json.RawMessage {}` => ` func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {}`
* `func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) {}` => `func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, config client.TxEncodingConfig, bz json.RawMessage) error {}`. Here `cdc`, which we get from parameters is used for `UnmarshalJSON` instead of using cdc from types in same module.
* `GetQueryCmd(cdc *codec.Codec)`,`GetTxCmd(cdc *codec.Codec)` is changed to `GetQueryCmd()`,`GetTxCmd()` respectively.
* Return type of `Route()` method which implements `AppModule` is changed from `string` to `"github.com/cosmos/cosmos-sdk/types".Route`. We will return a NewRoute which includes `RouterKey` and `NewHandler` as params.
    ```go
    func (am AppModule) Route() sdk.Route {
        return sdk.NewRoute(types.RouterKey, handler.NewHandler(am.keeper))
    }
    ```
* `func (am AppModule) NewQuerierHandler() sdk.Querier {}` => `func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {}`
* `RegisterQueryService(server grpc.Server)` is added which implements `AppModule`, which registers a GRPC query service to respond to the module-specific GRPC queries. Below `RegisterQueryServer` is present in `query.pb.go` which is generated `query.proto`
   
   ```go
    func (am AppModule) RegisterQueryService(server grpc.Server) {
	    querier := keeper.Querier{Keeper: am.keeper}
	    types.RegisterQueryServer(server, querier)
    }
    ```
   
* `func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {}` => `func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {}`
* `func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {}` => `func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {}`