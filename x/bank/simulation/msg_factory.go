package simulation

// weight  (in context of the module)
// msg

// need:
// * account factory: existing account
// * banlance.SpendableCoins

// on failure: simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "invalid transfers")
// on success: simtypes.NewOperationMsg(msg, true, "")

// helper
// random coins
// random fees
