1. Download ledger live (make sure you are on ledger.com)

2. Update your ledger to the latest firmware 

2. Install the "Cosmos" application on your ledger device (enable dev mode)

This can take a while 

Note: Not possible to add account direclty from ledger live rn, have to use the CLI 

3. Navigate to the Cosmos app on your ledger device


Installing the binaries 
Install `gaiacli` and `gaiad` 
Check version


Restoring from fundraiser

Using a ledger

1. Brand new ledger
2. Select a PIN
3. Chose restore configuration
4. Choose 12 words
5. Input the 12 words mnemonic

> If you input the wrong 12 words, or if you want to reset an already used ledger, go to settings, device and reset all

Then, instructions on how to install ledger

Then, `gaiacli keys add <yourKeyName> --ledger`
Note: sometimes this show an error, try again until it works. 

Using only a computer

High security -> Offline dedicated computer
Low security -> Online computer 

then `gaiacli keys add <yourKeyName> --recover 


Interracting with the network 

1. You need to be connected to a node

Option 1: Run your own full node - Maximum security

Option 2: Connect to the full-node of a trusted party, like a validator - Less security
How is this less secure than running your own full node?
What are the tradeoffs?
Use --node = 

2. gaiacli

`gaiacli` is the tool you will use to interract with the `gaiad` full-node from step 1. 
It enables you to do two main things: query the state, and send transactions. 

2.1 querying the state

When you query the state, you actually query the full-node you are connected to. If you are not running your own full-node, know that you have to trust the entity that runs it. You can double check by running the query with multiple full-nodes or using third-party block explorer like Hubble or Stargazer.

Here is the list of important queries pretransfers:

`gaiacli query account <yourAddress>`
`gaiacli query 

2.2 Bonding Atoms and withdrawing rewards 

Using a ledger

Key stored on ledger device 

To bond:

`gaiacli tx deleg.. --ledger`

Key stored in an offline computer

Generate -> sign -> broadcast

Key stored in an online computer 

