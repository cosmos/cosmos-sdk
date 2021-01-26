# Cosmos SDK v0.41.0 "Stargate" Release Notes

This release includes a breaking change, and a few minor bugfixes.

See the [Cosmos SDK v0.41.0 milestone](https://github.com/cosmos/cosmos-sdk/milestone/37?closed=1) on our issue tracker for details.

### Support Amino JSON for IBC MsgTransfer

This change **breaks state backward compatibility**.

At the moment hardware wallets are [unable to sign messages using `SIGN_MODE_DIRECT` because the cosmos ledger app does not support proto encoding and`SIGN_MODE_TEXTUAL` is not available yet](https://https://github.com/cosmos/cosmos-sdk/issues/8266).

In order to enable hardware wallets users to interact with IBC, amino JSON support was added to `MsgTransfer` only.

### Counterparty.ChannelID not available in OnChanOpenAck callback implementation.

This change **breaks state backward compatibility**.

In a previous version the `Counterparty.ChannelID` was available for an `OnChanOpenAck` callback implementation (read via `channelKeeper.GetChannel()`. Due to a regression, the channelID is currently empty.

The issue has been fixed by reordering IBC `ChanOpenAck` and `ChanOpenConfirm` to execute the core handlers logic first, followed by application callbacks.

### Bug Fixes

Now `x/bank` correctly verifies balances and metadata at init genesis stage.

`simapp` correctly adds the coins of genesis accounts to supply.

### IBC improvements

IBC channel data is persisted in a way that callbacks implementations can access it.
