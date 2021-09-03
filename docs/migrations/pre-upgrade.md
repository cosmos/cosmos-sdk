# Pre Upgrade Handling

Cosmovisor supports custom pre-upgrade handling. Any changes to the application configs prior to the upgrade, which may be needed by the newer version of the application, can be implemted in the application.

For example, any changes to `app.toml` settings which might be needed by the newer version can be handled during the pre upgrade. This makes upgradation proccess seamless as the file does not have to be manually updated.

Prior to the application binary being upgraded, Cosmovisor calls a `pre-upgrade` command which can  be implemented by the application. 
It is not mandatory to implement this command. If it is not implemented, the upgrade proceeds as it did previously.


The `pre-upgrade` command does not take in any command line arguments and is expected to terminate with the following exit codes.


| Exit status code | How it is handled in Cosmosvisor                                                                                    |
|------------------|---------------------------------------------------------------------------------------------------------------------|
| `0`              | Assumes `pre-upgrade` command executed successfully and continues the upgrade.                                      |
| `1`              | Default exit code when `pre-upgrade` command has not been implemented.                                              |
| `30`             | `pre-upgrade` command was executed but failed. This fails the entire upgrade.                                       |
| `31`             | `pre-upgrade` command was executed but failed. But the command is retried until exit code `1` or `30` are returned. |


## Sample

Here is a sample structure of how the `pre-upgrade` command should look like.

```go
func preUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pre-upgrade",
		Short: "Pre upgrade command",
        Long: "Pre upgrade command longer desc",
		Run: func(cmd *cobra.Command, args []string) {

			err := HandlePreUpgrade()

			if err != nil {
				os.Exit(30)
			}

			os.Exit(1)

		},
	}

	return cmd
}
```


Ensure that the pre-upgrade command has been registered in the application
```go
rootCmd.AddCommand(
		// ..
		preUpgradeCommand(),
		// ..
	)
```
