package cli

import (
	"github.com/cosmos/cosmos-sdk/container"
)

func Run(options ...container.Option) {
	err := container.Run(runRoot,
		container.Provide(
			ProvideQueryCommand,
			ProvideTxCommand,
		),
		container.Options(options...),
	)
	if err != nil {
		panic(err)
	}
}
