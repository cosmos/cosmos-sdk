IBC
===

Types
-----

**IBCPacket**
^^^^^^^^^^^^^

- **SrcAddr** (``sdk.Address``) -
- **DestAddr** (``sdk.Address``) -
- **Coins** (``sdk.Coins``) -
- **SrcChain** (``string``) -
- **DestChain** (``string``) -

IBCPacket defines a piece of data that can be send between two separate blockchains.

Methods
"""""""

``NewIBCPacket(srcAddr sdk.Address, destAddr sdk.Address, coins sdk.Coins, srcChain string, destChain string)``
***************************************************************************************************************

  Returns a new ``IBCPacket``.

``ibcp.ValidateBasic()``
************************

  Returns the ``sdk.Address`` of the keeper.

  Validates the IBC packet.


**IBCTransferMsg**
^^^^^^^^^^^^^^^^^^

- **IBCPacket** (``IBCPacket``) -

IBCTransferMsg defines how another module can send an IBCPacket.


Methods
"""""""

``msg.Type()``
**************

  Sets the an Address for keeper. Returns ``error`` if fails.

``msg.GetSigners()``
********************

  Returns the signers' addresses of the message.

``msg.GetSignBytes()``
**********************

  Get the sign bytes for IBC transfer message.

``msg.ValidateBasic()``
***********************

  Validates IBC transfer message.

**IBCReceiveMsg**
^^^^^^^^^^^^^^^^^

- **IBCPacket** (``IBCPacket``) -
- **Relayer** (``sdk.Address``) -
- **Sequence** (``int64``) -

IBCReceiveMsg defines the message that a relayer uses to post an IBCPacket to the destination chain.

Methods
"""""""

``msg.Type()``
**************

  Sets the an Address for keeper. Returns ``error`` if fails.

``msg.ValidateBasic()``
***********************

  Validates IBC transfer message.

``msg.GetSigners()``
********************

  Returns the signers' addresses of the message.

``msg.GetSignBytes()``
**********************

  Get the sign bytes for IBC transfer message.

Handlers
--------

IBC Handlers
^^^^^^^^^^^^

Methods
"""""""

``handleIBCTransferMsg(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, msg IBCTransferMsg)``
******************************************************************************************

  Returns: ``sdk.Result``

  Deducts coins from the account and creates an egress IBC packet.

``handleIBCReceiveMsg(ctx sdk.Context, ibcm Mapper, ck bank.Keeper, msg IBCReceiveMsg)``
****************************************************************************************

  Returns: ``sdk.Result``

  Adds coins to the destination address and creates an ingress IBC packet.
