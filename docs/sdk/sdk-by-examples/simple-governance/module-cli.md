## Command-Line Interface (CLI)

**File: [`x/simple_governance/client/cli/simple_governance.go`](https://github.com/cosmos/cosmos-sdk/blob/fedekunze/module_tutorial/examples/simpleGov/x/simple_governance/client/cli/simple_governance.go)**

Go in the `cli` folder and create a `simple_governance.go` file. This is where we will define the commands for our module.

The CLI builds on top of [Cobra](https://github.com/spf13/cobra). Here is the schema to build a command on top of Cobra:

```go
    // Declare flags
    const(
        Flag = "flag"
        ...
    )

    // Main command function. One function for each command.
    func Command(codec *wire.Codec) *cobra.Command {
        // Create the command to return
        command := &cobra.Command{
            Use: "actual command",
            Short: "Short description",
            Run: func(cmd *cobra.Command, args []string) error {
                // Actual function to run when command is used
            },
        }

        // Add flags to the command
        command.Flags().<Type>(FlagNameConstant, <example_value>, "<Description>")

        return command
    }
```

