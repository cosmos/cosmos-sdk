package appconfig

import (
	depinjectappconfig "cosmossdk.io/depinject/appconfig"
)

// LoadJSON loads an app config in JSON format.
var LoadJSON = depinjectappconfig.LoadJSON

// LoadYAML loads an app config in YAML format.
var LoadYAML = depinjectappconfig.LoadYAML

// WrapAny marshals a proto message into a proto Any instance
var WrapAny = depinjectappconfig.WrapAny

// Compose composes a v1alpha1 app config into a container option by resolving
// the required modules and composing their options.
var Compose = depinjectappconfig.Compose
