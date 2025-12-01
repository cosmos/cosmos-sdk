package utils

import (
	"fmt"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

// GenerateUnifiedDiff generates a unified diff from src and dst strings using gotextdiff.
// This is the only function that uses the gotextdiff library as its primary use is for
// clients.
func GenerateUnifiedDiff(src, dst string) (string, error) {
	// Create spans for the source and destination texts
	srcURI := span.URIFromPath("src")

	if src == "" || src[len(src)-1] != '\n' {
		src += "\n" // Add an EOL to src if it's empty or newline is missing
	}
	if dst == "" || dst[len(dst)-1] != '\n' {
		dst += "\n" // Add an EOL to dst if it's empty or newline is missing
	}

	// Compute the edits using the Myers diff algorithm
	eds := myers.ComputeEdits(srcURI, src, dst)

	// Generate the unified diff string
	diff := gotextdiff.ToUnified("src", "dst", src, eds)

	// Convert the diff to a string
	diffStr := fmt.Sprintf("%v", diff)

	return diffStr, nil
}
