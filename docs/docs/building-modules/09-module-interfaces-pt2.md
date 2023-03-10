# Module Interfaces Continued

:::note Synopsis
This document details how to build CLI and REST interfaces for a module. Examples from various Cosmos SDK modules are included.
:::

:::note

### Pre-requisite Readings

* [Building Modules Intro](./01-intro.md)

:::

## AutoCLI

The `autocli` package is a Go library for generating CLI (command line interface) interfaces for Cosmos SDK-based applications. It provides a simple way to add CLI commands to your application by generating them automatically based on your gRPC service definitions. Autocli generates CLI commands and flags directly from your Protocol Buffer messages, including options, input parameters, and output parameters. This means that you can easily add a CLI interface to your application without having to manually create and manage commands.

### Getting Started

Here are the steps to use the `autocli` package:

1.  Define your app's modules that implement the `appmodule.AppModule` interface.
2.  Create an instance of the `autocli.AppOptions` struct that specifies the modules you defined.
3.  Use the `RootCmd()` method provided by `autocli` to generate a root command that includes all the CLI commands for the specified modules.

Here's an example of how to use `autocli`:

```
// Define your app's modules
testModules := map[string]appmodule.AppModule{
    "testModule": &TestModule{},
}

// Define the autocli AppOptions
autoCliOpts := autocli.AppOptions{
    Modules: testModules,
}

// Generate the root command with autocli
rootCmd, err := autoCliOpts.RootCmd()
if err != nil {
    fmt.Println(err)
}

// Run the root command
if err := rootCmd.Execute(); err != nil {
    fmt.Println(err)
}
```

By following these steps, you can quickly and easily add a CLI interface to your application using `autocli`.

### Flags

`autocli` generates flags for each field in a Protocol Buffer message. By default, the names of the flags are generated based on the names of the fields in the message. You can customise the flag names using the `namingOptions` parameter of the `Builder.AddMessageFlags()` method.

To define flags for a message, you can use the `Builder.AddMessageFlags()` method. This method takes the `cobra.Command` instance and the message type as input, and generates flags for each field in the message.

```
binder, err := b.AddMessageFlags(cmd.Context(), cmd.Flags(), inputType, options)
if err != nil {
    return nil, err
}

cmd.Args = binder.CobraArgs

```

The `binder` variable returned by the `AddMessageFlags()` method is used to bind the command-line arguments to the fields in the message.

You can also customise the behavior of the flags using the `namingOptions` parameter of the `Builder.AddMessageFlags()` method. This parameter allows you to specify a custom prefix for the flags, and to specify whether to generate flags for repeated fields and whether to generate flags for fields with default values.

### Commands and Queries

The `autocli` package generates CLI commands and flags for each method defined in your gRPC service. By default, it generates commands for each RPC method that does not return a stream of messages. The commands are named based on the name of the service method.

For example, given the following protobuf definition for a service:
```
service MyService {
  rpc MyMethod(MyRequest) returns (MyResponse) {}
}

```

`autocli` will generate a command named `my-method` for the `MyMethod` method. The command will have flags for each field in the `MyRequest` message.

If you want to customise the behavior of a command, you can define a custom command by implementing the `autocli.Command` interface. You can then register the command with the `autocli.Builder` instance for your application.

Similarly, you can define a custom query by implementing the `autocli.Query` interface. You can then register the query with the `autocli.Builder` instance for your application.

To add a custom command or query, you can use the `Builder.AddCustomCommand` or `Builder.AddCustomQuery` methods, respectively. These methods take a `cobra.Command` or `cobra.Command` instance, respectively, which can be used to define the behavior of the command or query.

### Advanced Usage

#### Specifying Subcommands

By default, `autocli` generates a command for each method in your gRPC service. However, you can specify subcommands to group related commands together. To specify subcommands, you can use the `autocliv1.ServiceCommandDescriptor` struct.

For example, suppose you have a service with two methods: `GetTest` and `GetTest2`. You can group these methods together under a `get` subcommand using the following code

```
autoCliOpts := &autocliv1.ModuleOptions{
    Tx: &autocliv1.ServiceCommandDescriptor{
        Service: myServiceDesc.ServiceName,
        SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{
            "get": {
                Service: myServiceDesc.ServiceName,
                SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{
                    "test": {Service: myServiceDesc.ServiceName, Method: "GetTest"},
                    "test2": {Service: myServiceDesc.ServiceName, Method: "GetTest2"},
                },
            },
        },
    },
}


```

With this configuration, you can invoke the `GetTest` method by running `./app tx myservice get test`, and you can invoke the `GetTest2` method by running `./app tx myservice get test2`.

#### Adding Custom Commands

Although `autocli` can automatically generate CLI commands and flags based on your Protocol Buffer messages, you may want to add custom commands to your CLI application. To add a custom command, you can use the `AddCommand()` method provided by `cobra`. For instance, to add a custom `test` command to your CLI application, you can use the following code:

```
rootCmd := autocli.RootCmd(appOptions)
rootCmd.AddCommand(&cobra.Command{
    Use:   "test",
    Short: "custom test command",
    RunE: func(cmd *cobra.Command, args []string) error {
        // custom command logic here
        return nil
    },
})

```

This will add a new `test` command to your CLI application that can be invoked by running `./app test`.

### Customising Flag Names

By default, `autocli` generates flag names based on the names of the fields in your Protocol Buffer message. However, you can customise the flag names by providing a `namingOptions` parameter to the `Builder.AddMessageFlags()` method. This parameter allows you to specify custom names for flags based on the names of the message fields. For example, if you have a message with the fields `test` and `test1`, you can use the following naming options to customise the flags
```
options := autocliv1.RpcCommandOptions{ 
	FlagOptions: map[string]*autocliv1.FlagOptions{ 
		"test": { Name: "custom_name", }, 
		"test1": { Name: "other_name", }, 
		}, 
	}
```

With these naming options, the generated flags for the `test` and `test1` fields will have custom names `--custom-name` and `--other-name`, respectively.

### Conclusion

`autocli` is a powerful tool for adding CLI interfaces to your Cosmos SDK-based applications. It allows you to easily generate CLI commands and flags from your Protocol Buffer messages, and provides many options for customising the behavior of your CLI application. With `autocli`, you can quickly create powerful CLI interfaces for your applications with minimal effort.