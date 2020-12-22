<!--
order: 3
-->

# Messages

Messages (Msg) trigger state transitions. Msgs are wrapped in transactions (Txs) that clients submit to the network. The Cosmos SDK wraps and unwraps messages from transactions.

All nameservice messages require a corresponding handler that performs validation logic.

## MsgSetName

This message includes the attributes that set the value for a name.

``` protobuf
// MsgSetName defines a SetName message
type MsgSetName struct {
	Name  string         `json:"name"`
	Value string         `json:"value"`
	Owner sdk.AccAddress `json:"owner"`
}
```

- name - The name to set
- value - The value that the name resolves to
- owner - The current owner of the name


### MsgSetName Handler

The handler checks to see if the Msg sender is the owner of the name:

    - yes - Set the name
    - no - Throw an error and return `"Name and/or Value cannot be empty"` message

## MsgBuyName

This message includes the required attributes to buy a name.

``` protobuf
// MsgBuyName defines the BuyName message
type MsgBuyName struct {
	Name  string         `json:"name"`
	Bid   sdk.Coins      `json:"bid"`
	Buyer sdk.AccAddress `json:"buyer"`
}
```

- name - The name to buy
- bid - The value that the name resolves to
- buyer - The buyer of the name


### MsgBuyName Handler

The handler checks price and owner.

First, verify that the bid is higher than the price paid by the current owner:
    - yes - Proceed.
    - no - Throw an error and return `"Bid not high enough"` message.

Next, verify if the name already has an owner:
    - yes - The former owner receives the money from the `Buyer`
    - no - Send coins from the buyer to an unrecoverable address

If successful, the handler sets the buyer to the new owner, sets the new price to the current bid, and deducts the bid amount from the buyer.

If the SubtractCoins or SendCoins transaction returns a non-nil error, the handler throws an error and reverts the state transition.

## MsgDeleteName

This message includes the required attributes to delete a name.

``` protobuf
// This Msg deletes a name.
type MsgDeleteName struct {
	Name  string         `json:"name" yaml:"name"`
	Owner sdk.AccAddress `json:"owner" yaml:"owner"`
}
```

- name - The name to delete
- owner - The current owner of the name


### MsgDeleteName Handler

The handler performs the state transitions triggered by the message.

Verify that the owner is the current owner:
    - yes - Delete the name.
    - no - Throw an error and return `"Incorrect Owner"` message.
