package module

import (
	"embed"

	"google.golang.org/protobuf/proto"
)

func Register(configType proto.Message, pinnedProtoImageFS embed.FS, options ...Option) {
}
