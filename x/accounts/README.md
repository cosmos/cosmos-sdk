# x/accounts

The x/accounts module provides module and facilities for writing smart cosmos-sdk accounts.

# Genesis

## Creating accounts on genesis

In order to create accounts at genesis, the `x/accounts` module allows developers to provide
a list of genesis `MsgInit` messages that will be executed in the `x/accounts` genesis flow.

This follows the same initialization flow and rules that would happen if the chain is running. 
The only concrete difference is that this is happening at the genesis block.

For example, given the following `genesis.json` file:

```json
{
  "app_state": {
    "accounts": {
      "init_account_msgs": [
        {
          "sender": "account_creator_address",
          "account_type": "lockup",
          "message": {
            "@type": "cosmos.accounts.defaults.lockup.MsgInitLockupAccount",
            "owner": "some_owner",
            "end_time": "..",
            "start_time": ".."
          },
          "funds": [
            {
              "denom": "stake",
              "amount": "1000"
            }
          ]
        }
      ]
    }
  }
}
```

The accounts module will run the lockup account initialization message.