package params

import (
	"strings"
)

func GetParamSpaceFromKey(keystr string) string {
	strs := strings.Split(keystr, "/")
	if len(strs) != 2 {
		return ""
	}
	return strs[0]
}

func GetParamKey(keystr string) string {
	strs := strings.Split(keystr, "/")
	if len(strs) != 2 {
		return ""
	}
	return strs[1]
}
