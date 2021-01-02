<!--
order: 2
-->

# State

The state represents the application at a given moment. It tells how much token each account possesses, who are the owners and the price of each name, and to what value each name resolves to.

The state of tokens and accounts is defined by the `auth`, and `bank` modules. You need to define the part of the state that relates specifically to the `nameservice` module.

In the SDK, everything is stored in one store called the `multistore`. Any number of key/value stores (called KVStores (opens new window)in the Cosmos SDK) can be created in this multistore. For example, you can use one store to map `name`s to its respective whois, a struct that holds the value, owner, and price of the name.

```protobuf
 // Whois is a struct that contains all the metadata of a name
type Whois struct {
	Value string         `json:"value" yaml:"value"`
	Owner sdk.AccAddress `json:"owner" yaml:"owner"`
	Price sdk.Coins      `json:"price" yaml:"price"`
}

```