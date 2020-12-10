<!--
order: 5
-->

# Callbacks

Application modules implementing the IBC module must implement the following callbacks as found in [05-port](../05-port/types/module.go).
More information on how to implement these callbacks can be found in the [implementation guide](../../../../docs/ibc/custom.md).

```go
// IBCModule defines an interface that implements all the callbacks
// that modules must define as specified in ICS-26
type IBCModule interface {
	OnChanOpenInit(
		ctx sdk.Context,
		order channeltypes.Order,
		connectionHops []string,
		portId string,
		channelId string,
		channelCap *capability.Capability,
		counterparty channeltypes.Counterparty,
		version string,
	) error

	OnChanOpenTry(
		ctx sdk.Context,
		order channeltypes.Order,
		connectionHops []string,
		portId,
		channelId string,
		channelCap *capability.Capability,
		counterparty channeltypes.Counterparty,
		version,
		counterpartyVersion string,
	) error

	OnChanOpenAck(
		ctx sdk.Context,
		portId,
		channelId string,
		counterpartyVersion string,
	) error

	OnChanOpenConfirm(
		ctx sdk.Context,
		portId,
		channelId string,
	) error

	OnChanCloseInit(
		ctx sdk.Context,
		portId,
		channelId string,
	) error

	OnChanCloseConfirm(
		ctx sdk.Context,
		portId,
		channelId string,
	) error

	// OnRecvPacket must return the acknowledgement bytes
	// In the case of an asynchronous acknowledgement, nil should be returned.
	OnRecvPacket(
		ctx sdk.Context,
		packet channeltypes.Packet,
	) (*sdk.Result, []byte, error)

	OnAcknowledgementPacket(
		ctx sdk.Context,
		packet channeltypes.Packet,
		acknowledgement []byte,
	) (*sdk.Result, error)

	OnTimeoutPacket(
		ctx sdk.Context,
		packet channeltypes.Packet,
	) (*sdk.Result, error)
}
```
