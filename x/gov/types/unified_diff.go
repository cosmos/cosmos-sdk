package types

import (
	"fmt"
	"strconv"
	"strings"
)

// Hunk represents a single change hunk in a unified diff.
type Hunk struct {
	SrcLine int
	SrcSpan int
	DstLine int
	DstSpan int
	Lines   []string
}

// ParseUnifiedDiff parses the unified diff string into hunks with validation.
func ParseUnifiedDiff(diffStr string) ([]Hunk, error) {
	var (
		hunks []Hunk
		hunk  *Hunk
	)

	diffLines := strings.Split(diffStr, "\n")
	// remove any trailing empty lines
	for len(diffLines) > 0 && diffLines[len(diffLines)-1] == "" {
		diffLines = diffLines[:len(diffLines)-1]
	}

	for i := 0; i < len(diffLines); i++ {
		line := diffLines[i]

		if strings.HasPrefix(line, "@@") {
			// Start of a new hunk
			hunkHeader := line
			h, err := parseHunkHeader(hunkHeader)
			if err != nil {
				return nil, fmt.Errorf("invalid hunk header at line %d: %v", i+1, err)
			}
			if hunk != nil {
				// Validate the previous hunk before starting a new one
				if err := validateHunk(hunk); err != nil {
					return nil, fmt.Errorf("invalid hunk ending at line %d: %v", i, err)
				}
				hunks = append(hunks, *hunk)
			}
			hunk = h
			continue
		}

		if hunk != nil {
			if len(line) > 0 {
				prefix := line[0]
				if prefix != ' ' && prefix != '-' && prefix != '+' {
					return nil, fmt.Errorf("invalid line prefix '%c' at line %d", prefix, i+1)
				}
				hunk.Lines = append(hunk.Lines, line)
			} else {
				// Empty line (could be a valid diff line with no content)
				hunk.Lines = append(hunk.Lines, line)
			}
		} else if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "---") && !strings.HasPrefix(line, "+++") {
			// Non-empty line outside of a hunk and not a file header
			return nil, fmt.Errorf("unexpected content outside of hunks at line %d", i+1)
		}
	}

	// Validate and append the last hunk
	if hunk != nil {
		if err := validateHunk(hunk); err != nil {
			return nil, fmt.Errorf("invalid hunk at end of diff: %v", err)
		}
		hunks = append(hunks, *hunk)
	}

	if len(hunks) == 0 {
		return nil, fmt.Errorf("no valid hunks found in the diff")
	}

	return hunks, nil
}

// parseHunkHeader parses the hunk header line and returns a Hunk.
func parseHunkHeader(header string) (*Hunk, error) {
	// Remove the leading and trailing '@' symbols and any surrounding spaces
	header = strings.TrimPrefix(header, "@@")
	header = strings.TrimSuffix(header, "@@")
	header = strings.TrimSpace(header)

	// Split the header into source and destination parts
	parts := strings.Fields(header)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid hunk header format")
	}

	srcPart := parts[0]
	dstPart := parts[1]

	h := &Hunk{}

	// Parse source part
	srcLine, srcSpan, err := parseHunkRange(srcPart)
	if err != nil {
		return nil, fmt.Errorf("invalid source range: %v", err)
	}
	h.SrcLine = srcLine
	h.SrcSpan = srcSpan

	// Parse destination part
	dstLine, dstSpan, err := parseHunkRange(dstPart)
	if err != nil {
		return nil, fmt.Errorf("invalid destination range: %v", err)
	}
	h.DstLine = dstLine
	h.DstSpan = dstSpan

	return h, nil
}

// parseHunkRange parses a range string like "-srcLine,srcSpan" or "+dstLine"
func parseHunkRange(rangeStr string) (line int, span int, err error) {
	if len(rangeStr) == 0 {
		return 0, 0, fmt.Errorf("empty range string")
	}

	// The range string starts with '-' or '+'
	prefix := rangeStr[0]
	if prefix != '-' && prefix != '+' {
		return 0, 0, fmt.Errorf("invalid range prefix '%c'", prefix)
	}

	rangeStr = rangeStr[1:] // Remove the prefix

	// Split the range into line and span if ',' is present
	if strings.Contains(rangeStr, ",") {
		parts := strings.Split(rangeStr, ",")
		if len(parts) != 2 {
			return 0, 0, fmt.Errorf("invalid range format")
		}
		line, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid line number: %v", err)
		}
		span, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid span number: %v", err)
		}
	} else {
		// No span provided, default to span = 1
		line, err = strconv.Atoi(rangeStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid line number: %v", err)
		}
		span = 1
	}

	if span < 0 {
		return 0, 0, fmt.Errorf("negative span")
	}

	line-- // Adjust line number to 0-based index

	return line, span, nil
}

// validateHunk validates the hunk's content against its header.
func validateHunk(hunk *Hunk) error {
	srcCount := 0
	dstCount := 0

	for _, line := range hunk.Lines {
		if len(line) == 0 {
			// Empty line, does not contribute to counts
			continue
		}
		switch line[0] {
		case ' ':
			srcCount++
			dstCount++
		case '-':
			srcCount++
		case '+':
			dstCount++
		default:
			return fmt.Errorf("invalid line prefix '%c'", line[0])
		}
	}

	if srcCount != hunk.SrcSpan {
		return fmt.Errorf("source line count (%d) does not match SrcSpan (%d)", srcCount, hunk.SrcSpan)
	}
	if dstCount != hunk.DstSpan {
		return fmt.Errorf("destination line count (%d) does not match DstSpan (%d)", dstCount, hunk.DstSpan)
	}

	return nil
}

// applyHunks applies the parsed hunks to the source lines.
func applyHunks(srcStr string, hunks []Hunk) ([]string, error) {
	srcLines := strings.Split(srcStr, "\n")
	result := make([]string, 0)
	srcIndex := 0

	for _, hunk := range hunks {
		// Validate hunk.SrcLine is within bounds
		if hunk.SrcLine > len(srcLines) {
			return nil, fmt.Errorf("hunk starts at line %d but source only has %d lines", hunk.SrcLine+1, len(srcLines))
		}

		// Add unchanged lines before the hunk
		for srcIndex < hunk.SrcLine {
			if srcIndex >= len(srcLines) {
				return nil, fmt.Errorf("source index %d exceeds source length %d", srcIndex, len(srcLines))
			}
			result = append(result, srcLines[srcIndex])
			srcIndex++
		}

		// Apply hunk lines
		for _, line := range hunk.Lines {
			if len(line) == 0 {
				continue
			}
			prefix := line[0]
			content := line[1:]

			switch prefix {
			case ' ':
				// Context line, should match source
				if srcIndex >= len(srcLines) {
					return nil, fmt.Errorf("context line at hunk position exceeds source length (srcIndex: %d, srcLines: %d)", srcIndex, len(srcLines))
				}
				if srcLines[srcIndex] != content {
					return nil, fmt.Errorf("context line mismatch at line %d", srcIndex+1)
				}
				result = append(result, content)
				srcIndex++
			case '-':
				// Deletion, skip source line
				if srcIndex >= len(srcLines) {
					return nil, fmt.Errorf("deletion line at hunk position exceeds source length (srcIndex: %d, srcLines: %d)", srcIndex, len(srcLines))
				}
				if srcLines[srcIndex] != content {
					return nil, fmt.Errorf("deletion line mismatch at line %d", srcIndex+1)
				}
				srcIndex++
			case '+':
				// Insertion, add to result
				result = append(result, content)
			default:
				return nil, fmt.Errorf("invalid diff line: %s", line)
			}
		}
	}

	// Add any remaining lines
	for srcIndex < len(srcLines) {
		result = append(result, srcLines[srcIndex])
		srcIndex++
	}

	return result, nil
}

// ApplyUnifiedDiff applies a unified diff patch to the src string and returns the result.
// Does not make use of any external libraries to ensure deterministic behavior.
func ApplyUnifiedDiff(src, diffStr string) (string, error) {
	// Parse the unified diff into hunks
	hunks, err := ParseUnifiedDiff(diffStr)
	if err != nil {
		return "", err
	}

	// Apply the hunks to the source lines
	resultLines, err := applyHunks(src, hunks)
	if err != nil {
		return "", err
	}

	return strings.Join(resultLines, "\n"), nil
}
