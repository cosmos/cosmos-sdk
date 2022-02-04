package moduleimpl

import (
	"embed"

	modulev1alpha1 "github.com/cosmos/cosmos-sdk/api/cosmos/app/module/v1alpha1"
	"github.com/cosmos/cosmos-sdk/app/module"
)

// go:embed proto_image.bin.gz
var pinnedProtoImage embed.FS

func init() {
	module.Register(&modulev1alpha1.Module{}, pinnedProtoImage)
}
