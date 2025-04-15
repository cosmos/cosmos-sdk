package appmodule

import (
	depinjectappconfig "cosmossdk.io/depinject/appconfig"
)

// Register registers a module with the global module registry. The provided
// protobuf message is used only to uniquely identify the protobuf module config
// type. The instance of the protobuf message used in the actual configuration
// will be injected into the container and can be requested by a provider
// function. All module initialization should be handled by the provided options.
//
// Protobuf message types used for module configuration should define the
// cosmos.app.v1alpha.module option and must explicitly specify go_package
// to make debugging easier for users.
var Register = depinjectappconfig.RegisterModule
