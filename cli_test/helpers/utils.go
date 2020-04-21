package helpers

import "strings"

func addFlags(cmd string, flags []string) string {
	for _, f := range flags {
		cmd += " " + f
	}
	return strings.TrimSpace(cmd)
}
