package version

import "strconv"

const Version int64 = 1

func DefaultPrefix() []byte {
	return Prefix(Version)
}

func Prefix(version int64) []byte {
	return []byte(strconv.FormatInt(version, 10) + "/")
}
