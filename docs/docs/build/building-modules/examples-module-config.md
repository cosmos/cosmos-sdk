# Module Configuration Examples

This document provides examples of how to use the new module configuration feature that allows specifying arbitrary configuration blobs by key and type.

## Example 1: Custom Fee Configuration

Let's say you want to create a module that can accept different fee configurations.

### 1. Define the proto messages

```protobuf
// proto/example/fee/v1/fee.proto
syntax = "proto3";
package example.fee.v1;

message FeeConfig {
  string fee_collector = 1;
  repeated FeeRate fee_rates = 2;
}

message FeeRate {
  string operation = 1;
  string amount = 2; // e.g. "100stake"
}
```

### 2. Define the module config

```protobuf
// proto/example/module/v1/module.proto
syntax = "proto3";
package example.module.v1;

import "google/protobuf/any.proto";
import "cosmos/app/v1alpha1/module.proto";

option go_package = "example.com/x/example/types";

message Module {
  option (cosmos.app.v1alpha1.module) = {
    go_import: "example.com/x/example"
  };

  // Regular module configuration
  bool enable_feature = 1;
  
  // Arbitrary configuration entries
  repeated ConfigEntry config_entries = 2;
}

message ConfigEntry {
  string key = 1;
  google.protobuf.Any value = 2;
}
```

### 3. Module implementation to read the configuration

```go
// x/example/module.go
package example

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"example.com/x/example/keeper"
	"example.com/x/example/types"
	feev1 "example.com/x/example/types/fee/v1"
)

func init() {
	appmodule.Register(
		&types.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config *types.Module
	// Other dependencies...
}

type ModuleOutputs struct {
	depinject.Out

	Keeper *keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) (ModuleOutputs, error) {
	// Extract the fee configuration if it exists
	var feeConfig *feev1.FeeConfig
	for _, entry := range in.Config.ConfigEntries {
		if entry.Key == "fee_config" {
			var cfg feev1.FeeConfig
			err := anypb.UnmarshalTo(entry.Value, &cfg, proto.UnmarshalOptions{})
			if err != nil {
				return ModuleOutputs{}, err
			}
			feeConfig = &cfg
			break
		}
	}

	// Use default if not provided
	if feeConfig == nil {
		feeConfig = &feev1.FeeConfig{
			FeeCollector: "fee_collector",
			FeeRates: []*feev1.FeeRate{
				{
					Operation: "default",
					Amount:    "10stake",
				},
			},
		}
	}

	// Create the keeper with the configuration
	k := keeper.NewKeeper(feeConfig)
	m := NewAppModule(k)

	return ModuleOutputs{
		Keeper: k,
		Module: m,
	}, nil
}
```

### 4. App configuration (YAML)

```yaml
modules:
  - name: example
    config:
      "@type": "/example.module.v1.Module"
      enable_feature: true
      config_entries:
        - key: "fee_config"
          value:
            "@type": "/example.fee.v1.FeeConfig"
            fee_collector: "fee_module"
            fee_rates:
              - operation: "transfer"
                amount: "50stake"
              - operation: "delegate"
                amount: "100stake"
```

### 5. App configuration (Go)

```go
// app_config.go
{
	Name: "example",
	Config: appconfig.WrapAny(&examplemodulev1.Module{
		EnableFeature: true,
		ConfigEntries: []*examplemodulev1.ConfigEntry{
			{
				Key: "fee_config",
				Value: appconfig.WrapAny(&feev1.FeeConfig{
					FeeCollector: "fee_module",
					FeeRates: []*feev1.FeeRate{
						{
							Operation: "transfer",
							Amount:    "50stake",
						},
						{
							Operation: "delegate",
							Amount:    "100stake",
						},
					},
				}),
			},
		},
	}),
},
```

## Example 2: Multiple Configuration Types

You can include multiple configuration types in a single module:

```yaml
modules:
  - name: multiconfig
    config:
      "@type": "/multiconfig.module.v1.Module"
      config_entries:
        - key: "network_config"
          value:
            "@type": "/multiconfig.network.v1.NetworkConfig"
            max_connections: 100
            timeout_seconds: 30
        
        - key: "storage_config"
          value:
            "@type": "/multiconfig.storage.v1.StorageConfig"
            max_size_bytes: 104857600  # 100 MB
            retention_days: 14
        
        - key: "notification_config"
          value:
            "@type": "/multiconfig.notification.v1.NotificationConfig"
            enabled: true
            providers:
              - name: "email"
                endpoint: "smtp://example.com:587"
              - name: "webhook"
                endpoint: "https://api.example.com/notify"
```

This approach allows for highly modular and extensible configuration without requiring code changes when adding new configuration options. 
