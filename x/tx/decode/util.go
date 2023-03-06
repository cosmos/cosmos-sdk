package decode

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func MessageNameFromTypeURL(url string) protoreflect.FullName {
	messagename := protoreflect.FullName(url)
	if i := strings.LastIndexByte(url, '/'); i >= 0 {
		messagename = messagename[i+len("/"):]
	}
	return messagename
}
