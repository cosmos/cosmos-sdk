# Testing the Oracle Module

We will guide you through the process of testing the Oracle module in your application. The Oracle module uses vote extensions to provide current price data. If you would like to see the complete working oracle module please see [here](https://github.com/cosmos/sdk-tutorials/blob/master/tutorials/oracle/base/x/oracle).

## Step 1: Compile and Install the Application

First, we need to compile and install the application. Please ensure you are in the `tutorials/oracle/base` directory. Run the following command in your terminal:

```shell
make install
```

This command compiles the application and moves the resulting binary to a location in your system's PATH.

## Step 2: Initialise the Application

Next, we need to initialise the application. Run the following command in your terminal:

```shell
make init
```

This command runs the script `tutorials/oracle/base/scripts/init.sh`, which sets up the necessary configuration for your application to run. This includes creating the `app.toml` configuration file and initialising the blockchain with a genesis block.

## Step 3: Start the Application

Now, we can start the application. Run the following command in your terminal:

```shell
exampled start
```

This command starts your application, begins the blockchain node, and starts processing transactions.

## Step 4: Query the Oracle Prices

Finally, we can query the current prices from the Oracle module. Run the following command in your terminal:

```shell
exampled q oracle prices
```

This command queries the current prices from the Oracle module. The expected output shows that the vote extensions were successfully included in the block and the Oracle module was able to retrieve the price data.

## Understanding Vote Extensions in Oracle

In the Oracle module, the `ExtendVoteHandler` function is responsible for creating the vote extensions. This function fetches the current prices from the provider, creates a `OracleVoteExtension` struct with these prices, and then marshals this struct into bytes. These bytes are then set as the vote extension.

In the context of testing, the Oracle module uses a mock provider to simulate the behavior of a real price provider. This mock provider is defined in the mockprovider package and is used to return predefined prices for specific currency pairs.

## Conclusion

In this tutorial, we've delved into the concept of Oracle's in blockchain technology, focusing on their role in providing external data to a blockchain network. We've explored vote extensions, a powerful feature of ABCI++, and integrated them into a Cosmos SDK application to create a price oracle module.

Through hands-on exercises, you've implemented vote extensions, and tested their effectiveness in providing timely and accurate asset price information. You've gained practical insights by setting up a mock provider for testing and analysing the process of extending votes, verifying vote extensions, and preparing and processing proposals.

Keep experimenting with these concepts, engage with the community, and stay updated on new advancements. The knowledge you've acquired here is crucial for developing robust and reliable blockchain applications that can interact with real-world data.
