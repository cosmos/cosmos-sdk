IBC
===

**IBCPacket**

- **SrcAddr** (``sdk.Address``) -
- **DestAddr** (``sdk.Address``) -
- **Coins** (``sdk.Coins``) -
- **SrcChain** (``string``) -
- **DestChain** (``string``) -

**IBCTransferMsg**

- **IBCPacket** (``IBCPacket``) -

**IBCReceiveMsg**

- **IBCPacket** (``IBCPacket``) -
- **Relayer** (``sdk.Address``) -
- **Sequence** (``int64``) -
