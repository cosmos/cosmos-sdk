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

  Returns: ``IBCPacket``

  Creates an IBC packet to relay a massage to another chain.

``ibcp.ValidateBasic()``
************************

  Returns: ``sdk.Error``

  Validates the IBC packet. Returns error if fails.


**IBCTransferMsg**
^^^^^^^^^^^^^^^^^^

- **IBCPacket** (``IBCPacket``) -

IBCTransferMsg defines how another module can send an IBCPacket.


Methods
"""""""

``msg.Type()``
**************

  Returns: ``string``

  Returns the type of the message.

``msg.GetSigners()``
********************

  Returns: ``[]sdk.Address``

  Returns the signers' addresses of the message.

``msg.GetSignBytes()``
**********************

  Returns: ``[]byte``

  Get the signature bytes of the message.

``msg.ValidateBasic()``
***********************

  Returns: ``sdk.Error``

  Basic validation of the message. Returns error if fails.

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

  Returns: ``string``

  Returns the type of the message.

``msg.GetSigners()``
********************

  Returns: ``[]sdk.Address``

  Returns the signers' addresses of the message.

``msg.GetSignBytes()``
**********************

  Returns: ``[]byte``

  Get the signature bytes of the message.

``msg.ValidateBasic()``
***********************

  Returns: ``sdk.Error``

  Basic validation of the message. Returns error if fails.


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
