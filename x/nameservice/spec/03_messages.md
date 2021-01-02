<!--
order: 3
-->

# Messages

Messages are contained in transactions. They trigger state transitions. Each module defines a list of messages and how to handle them. Here are the messages you need to implement the desired functionality for the nameservice application:

- `MsgSetName`: This message allows name owners to set a value for a given name.
- `MsgBuyName`: This message allows accounts to buy a name and become its owner.
    - When someone buys a name, they are required to pay the previous owner of the name a price higher than the price the previous owner paid for it. If a name does not have a previous owner yet, they must burn a `MinPrice` amount.
- `MsgDeleteName`: This message allows name owners to delete a given name.

When a transaction (included in a block) reaches a Tendermint node, it is passed to the application through the ABCI (opens new window)and decoded to get the message. The message is then routed to the appropriate module and handled there according to the logic defined in the `Handler`. If the state needs to be updated, the `Handler` calls the `Keeper` to perform the update.
