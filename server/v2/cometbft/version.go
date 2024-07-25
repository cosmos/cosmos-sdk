package cometbft

import "runtime/debug"

var Version = ""

func getCometBFTServerVersion() string {
	deps, ok := debug.ReadBuildInfo()
	if !ok {
		return Version
	}

	var serverVersion string
	for _, dep := range deps.Deps {
		if dep.Path == "cosmossdk.io/server/v2/cometbft" {
			if dep.Replace != nil && dep.Replace.Version != "(devel)" {
				serverVersion = dep.Replace.Version
			} else {
				serverVersion = dep.Version
			}
		}
	}

	Version = serverVersion
	return serverVersion
}
