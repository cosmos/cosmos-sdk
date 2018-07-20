## Errors (`errors.go`)

The `error.go` file allows us to define custom error messages for our module.  Declaring errors should be relatively similar in all modules. You can look in the [error.go](./error.go) file of our simple governance module for a concrete example. The code is self-explanatory.

Note that the errors of our module inherit from the `sdk.Error` interface and therefore possess the method `Result()`. This method is useful when there is an error in the `handler` and an error has to be returned in place of an actual result.