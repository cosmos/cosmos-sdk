## Types 

**File: [`x/simple_governance/types.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/x/simple_governance/types.go)**

In this file, we define the custom types for our module. This includes the types from the [State](app-design.md#State) section and the custom message types defined in the [Messages](app-design#Messages) section.

For each new type that is not a message, it is possible to add methods that make sense in the context of the application. In our case, we will implement an `updateTally` function to easily update the tally of a given proposal as vote messages come in.

Messages are a bit different. They implement the `Message` interface defined in the SDK's `types` folder. Here are the methods you need to implement when you define a custom message type:

- `Type()`: This function returns the name of our module's route. When messages are processed by the application, they are routed using the string returned by the `Type()` method.
- `GetSignBytes()`: Returns the byte representation of the message. It is used to sign the message.
- `GetSigners()`: Returns address(es) of the signer(s).
- `ValidateBasic()`: This function is used to discard obviously invalid messages. It is called at the beginning of `runTx()` in the baseapp file. If `ValidateBasic()` does not return `nil`, the app stops running the transaction.
- `Get()`: A basic getter, returns some property of the message.
- `String()`: Returns a human-readable version of the message

For our simple governance messages, this means:

- `Type()` will return `"simpleGov"`
- For `SubmitProposalMsg`, we need to make sure that the attributes are not empty and that the deposit is both valid and positive. Note that this is only basic validation, we will therefore not check in this method that the sender has sufficient funds to pay for the deposit
- For `VoteMsg`, we check that the address and option are valid and that the proposalID is not negative.
- As for other methods, less customization is required. You can check the code to see a standard way of implementing these.