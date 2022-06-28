# Wiring up app modules for use with appconfig

The `appconfig` framework allows Cosmos SDK modules to be composed declaratively using a configuration file without
requiring the app developer to understand the details of inter-module dependencies.

## 1. Create a module config protobuf message

The first step in creating a module that works with `appconfig`, is to create a protobuf message for the module configuration. The best practices for defining the module configuration message are:
* Use a dedicated protobuf package for the module configuration message  instead of placing it in the API protobuf package. For example, the module configuration for bank would go in `cosmos.bank.module.v1` instead of just `cosmos.bank.v1`. This decouples the state machine version from the API version.
* The module configuration message is usually called simply `Module`, ex. `cosmos.bank.module.v1.Module`.
* Create a new protobuf package and configuration message for each state machine breaking version of the module, ex. `cosmos.bank.module.v2.Module`, etc.

The module configuration message should include any parameters which should be initialized at application startup. For example, the auth module needs to know the bech32 prefix of the app and the permissions of module accounts.

In the future, it may be possible to update the app config through a governance proposal at runtime.

All module configuration messages should define a module descriptor, using the `cosmos.app.v1alpha1.module` message option.

Here is an example module configuration message for the `auth` module:

```protobuf
package cosmos.auth.module.v1;

import "cosmos/app/v1alpha1/module.proto";

message Module {
  option (cosmos.app.v1alpha1.module) = {
    go_import: "github.com/cosmos/cosmos-sdk/x/auth"
  };
  string bech32_prefix = 1;
  repeated ModuleAccountPermission module_account_permissions = 2;
}
```

## 2. Register module depinject providers and invokers
Once we have a module config object, we need to register depinject providers and invokers for the module using the `cosmossdk.io/core/appmodule` package.

At the most basic level, we must define an `init` function in the package listed as the `go_import` in the module descriptor. This `init` function must call `appmodule.Register` with an empty instance of the config object and some options for initializing the module, ex:

```go
func init() {
	appmodule.Register(&modulev1.Module{},
    // options
  )
}
```

### Defining providers and invokers

