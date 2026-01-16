package internal

import (
	"bytes"
	"testing"
)

func TestCmpInlineKeyPrefix(t *testing.T) {
	tests := []struct {
		name         string
		inlinePrefix []byte
		inlineLen    int
		other        []byte
		wantCmp      int
		wantNeedFull bool
	}{
		// ============================================================
		// Empty key cases (inlineLen = 0)
		// ============================================================
		{
			name:         "empty key vs empty other",
			inlinePrefix: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			inlineLen:    0,
			other:        []byte{},
			wantCmp:      0,
			wantNeedFull: false,
		},
		{
			name:         "empty key vs non-empty other",
			inlinePrefix: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			inlineLen:    0,
			other:        []byte{1},
			wantCmp:      -1,
			wantNeedFull: false,
		},
		{
			name:         "empty key vs long other",
			inlinePrefix: []byte{0, 0, 0, 0, 0, 0, 0, 0},
			inlineLen:    0,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
			wantCmp:      -1,
			wantNeedFull: false,
		},

		// ============================================================
		// Short keys (inlineLen 1-7) - full key is inline
		// ============================================================
		{
			name:         "short key equal",
			inlinePrefix: []byte{1, 2, 3, 0, 0, 0, 0, 0},
			inlineLen:    3,
			other:        []byte{1, 2, 3},
			wantCmp:      0,
			wantNeedFull: false,
		},
		{
			name:         "short key less than other (content differs)",
			inlinePrefix: []byte{1, 2, 3, 0, 0, 0, 0, 0},
			inlineLen:    3,
			other:        []byte{1, 2, 4},
			wantCmp:      -1,
			wantNeedFull: false,
		},
		{
			name:         "short key greater than other (content differs)",
			inlinePrefix: []byte{1, 2, 5, 0, 0, 0, 0, 0},
			inlineLen:    3,
			other:        []byte{1, 2, 4},
			wantCmp:      1,
			wantNeedFull: false,
		},
		{
			name:         "short key less than longer other (our key is prefix of other)",
			inlinePrefix: []byte{1, 2, 3, 0, 0, 0, 0, 0},
			inlineLen:    3,
			other:        []byte{1, 2, 3, 4, 5},
			wantCmp:      -1,
			wantNeedFull: false,
		},
		{
			name:         "short key greater than shorter other (other is prefix of our key)",
			inlinePrefix: []byte{1, 2, 3, 0, 0, 0, 0, 0},
			inlineLen:    3,
			other:        []byte{1, 2},
			wantCmp:      1,
			wantNeedFull: false,
		},
		{
			name:         "short key vs long other, content differs early",
			inlinePrefix: []byte{1, 2, 3, 0, 0, 0, 0, 0},
			inlineLen:    3,
			other:        []byte{1, 9, 0, 0, 0, 0, 0, 0, 0, 0},
			wantCmp:      -1,
			wantNeedFull: false,
		},
		{
			name:         "1-byte key equal",
			inlinePrefix: []byte{42, 0, 0, 0, 0, 0, 0, 0},
			inlineLen:    1,
			other:        []byte{42},
			wantCmp:      0,
			wantNeedFull: false,
		},
		{
			name:         "7-byte key equal",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 0},
			inlineLen:    7,
			other:        []byte{1, 2, 3, 4, 5, 6, 7},
			wantCmp:      0,
			wantNeedFull: false,
		},
		{
			name:         "7-byte key vs 8-byte other (prefix matches)",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 0},
			inlineLen:    7,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
			wantCmp:      -1,
			wantNeedFull: false,
		},

		// ============================================================
		// Exactly 8-byte keys (inlineLen = 8) - boundary case
		// ============================================================
		{
			name:         "8-byte key equal to 8-byte other",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    8,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
			wantCmp:      0,
			wantNeedFull: false,
		},
		{
			name:         "8-byte key less than 8-byte other (content)",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    8,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 9},
			wantCmp:      -1,
			wantNeedFull: false,
		},
		{
			name:         "8-byte key greater than 8-byte other (content)",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 9},
			inlineLen:    8,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
			wantCmp:      1,
			wantNeedFull: false,
		},
		{
			name:         "8-byte key less than 9-byte other (prefix matches, other longer)",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    8,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
			wantCmp:      -1,
			wantNeedFull: false,
		},
		{
			name:         "8-byte key greater than 7-byte other (prefix matches, other shorter)",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    8,
			other:        []byte{1, 2, 3, 4, 5, 6, 7},
			wantCmp:      1,
			wantNeedFull: false,
		},
		{
			name:         "8-byte key greater than 5-byte other",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    8,
			other:        []byte{1, 2, 3, 4, 5},
			wantCmp:      1,
			wantNeedFull: false,
		},
		{
			name:         "8-byte key vs long other, content differs in first 8",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    8,
			other:        []byte{1, 2, 3, 4, 5, 6, 9, 0, 0, 0},
			wantCmp:      -1,
			wantNeedFull: false,
		},

		// ============================================================
		// Long keys (inlineLen = 9) - just over boundary
		// ============================================================
		{
			name:         "9-byte key vs 8-byte other (prefix matches) - no read needed",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    9,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
			wantCmp:      1, // our key is longer
			wantNeedFull: false,
		},
		{
			name:         "9-byte key vs 9-byte other (prefix matches) - need full key",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    9,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8, 0},
			wantCmp:      0,
			wantNeedFull: true,
		},
		{
			name:         "9-byte key vs 9-byte other, prefix differs - no read needed",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 9},
			inlineLen:    9,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8, 0},
			wantCmp:      1, // byte 7: 9 > 8
			wantNeedFull: false,
		},
		{
			name:         "9-byte key vs 5-byte other (prefix matches) - no read needed",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    9,
			other:        []byte{1, 2, 3, 4, 5},
			wantCmp:      1,
			wantNeedFull: false,
		},

		// ============================================================
		// Long keys (inlineLen = 12) - typical long key
		// ============================================================
		{
			name:         "12-byte key vs 3-byte other, content differs",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    12,
			other:        []byte{1, 2, 9},
			wantCmp:      -1,
			wantNeedFull: false,
		},
		{
			name:         "12-byte key vs 3-byte other, prefix greater",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    12,
			other:        []byte{1, 2, 0},
			wantCmp:      1,
			wantNeedFull: false,
		},
		{
			name:         "12-byte key vs 5-byte other, prefix matches exactly",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    12,
			other:        []byte{1, 2, 3, 4, 5},
			wantCmp:      1, // our key is longer
			wantNeedFull: false,
		},
		{
			name:         "12-byte key vs 8-byte other, prefix matches - no read needed",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    12,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
			wantCmp:      1, // our key is longer
			wantNeedFull: false,
		},
		{
			name:         "12-byte key vs 10-byte other, prefix matches - need full key",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    12,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8, 0, 0},
			wantCmp:      0,
			wantNeedFull: true,
		},
		{
			name:         "12-byte key vs 10-byte other, prefix differs at byte 7",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    12,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 9, 0, 0},
			wantCmp:      -1, // byte 7: 8 < 9
			wantNeedFull: false,
		},
		{
			name:         "12-byte key vs 12-byte other, prefix differs at byte 0",
			inlinePrefix: []byte{5, 0, 0, 0, 0, 0, 0, 0},
			inlineLen:    12,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			wantCmp:      1, // byte 0: 5 > 1
			wantNeedFull: false,
		},

		// ============================================================
		// Very long keys (boundary at 127, simulating max stored prefix len)
		// ============================================================
		{
			name:         "127-byte key vs 8-byte other, prefix matches - no read",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    127,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
			wantCmp:      1,
			wantNeedFull: false,
		},
		{
			name:         "127-byte key vs 10-byte other, prefix matches - need full key",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    127,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			wantCmp:      0,
			wantNeedFull: true,
		},
		{
			name:         "127-byte key vs 127-byte other, prefix matches - need full key",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    127,
			other:        make([]byte, 127), // all zeros except we set first 8
			wantCmp:      0,
			wantNeedFull: true,
		},

		// ============================================================
		// Keys > 127 bytes (testing that large inlineLen works)
		// ============================================================
		{
			name:         "1000-byte key vs 8-byte other, prefix matches - no read",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    1000,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8},
			wantCmp:      1,
			wantNeedFull: false,
		},
		{
			name:         "1000-byte key vs 500-byte other, prefix matches - need full key",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    1000,
			other:        append([]byte{1, 2, 3, 4, 5, 6, 7, 8}, make([]byte, 492)...),
			wantCmp:      0,
			wantNeedFull: true,
		},
		{
			name:         "65535-byte key (max uint16) vs 9-byte other, prefix matches - need full key",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    65535,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 8, 9},
			wantCmp:      0,
			wantNeedFull: true,
		},

		// ============================================================
		// Content difference at each byte position (0-7)
		// ============================================================
		{
			name:         "difference at byte 0",
			inlinePrefix: []byte{2, 0, 0, 0, 0, 0, 0, 0},
			inlineLen:    8,
			other:        []byte{1, 0, 0, 0, 0, 0, 0, 0},
			wantCmp:      1,
			wantNeedFull: false,
		},
		{
			name:         "difference at byte 1",
			inlinePrefix: []byte{1, 2, 0, 0, 0, 0, 0, 0},
			inlineLen:    8,
			other:        []byte{1, 1, 0, 0, 0, 0, 0, 0},
			wantCmp:      1,
			wantNeedFull: false,
		},
		{
			name:         "difference at byte 7",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    8,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 0},
			wantCmp:      1,
			wantNeedFull: false,
		},
		{
			name:         "long keys, difference at byte 7",
			inlinePrefix: []byte{1, 2, 3, 4, 5, 6, 7, 8},
			inlineLen:    20,
			other:        []byte{1, 2, 3, 4, 5, 6, 7, 0, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9},
			wantCmp:      1,
			wantNeedFull: false,
		},

		// ============================================================
		// Verify garbage bytes in prefix beyond inlineLen are ignored
		// ============================================================
		{
			name:         "garbage in prefix ignored (inlineLen=3)",
			inlinePrefix: []byte{1, 2, 3, 99, 99, 99, 99, 99}, // garbage after byte 2
			inlineLen:    3,
			other:        []byte{1, 2, 3},
			wantCmp:      0,
			wantNeedFull: false,
		},
		{
			name:         "garbage in prefix ignored when other is longer",
			inlinePrefix: []byte{1, 2, 3, 99, 99, 99, 99, 99},
			inlineLen:    3,
			other:        []byte{1, 2, 3, 4, 5}, // should compare only first 3 bytes
			wantCmp:      -1,                    // our key (3 bytes) < other (5 bytes)
			wantNeedFull: false,
		},
	}

	// Set up the 127-byte other for the test case
	for i := range tests {
		if tests[i].name == "127-byte key vs 127-byte other, prefix matches - need full key" {
			tests[i].other[0] = 1
			tests[i].other[1] = 2
			tests[i].other[2] = 3
			tests[i].other[3] = 4
			tests[i].other[4] = 5
			tests[i].other[5] = 6
			tests[i].other[6] = 7
			tests[i].other[7] = 8
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmp, gotNeedFull := cmpInlineKeyPrefix(tt.inlinePrefix, tt.inlineLen, tt.other)
			if gotCmp != tt.wantCmp || gotNeedFull != tt.wantNeedFull {
				t.Errorf("cmpInlineKeyPrefix() = (%d, %v), want (%d, %v)",
					gotCmp, gotNeedFull, tt.wantCmp, tt.wantNeedFull)
			}
		})
	}
}

// TestCmpInlineKeyPrefixNeverUnnecessaryRead verifies that needFullKey=true
// is ONLY returned when absolutely necessary (both keys > 8 bytes AND first 8 match).
func TestCmpInlineKeyPrefixNeverUnnecessaryRead(t *testing.T) {
	prefix := []byte{1, 2, 3, 4, 5, 6, 7, 8}

	// Test all combinations of lengths around the boundary
	for inlineLen := 0; inlineLen <= 20; inlineLen++ {
		for otherLen := 0; otherLen <= 20; otherLen++ {
			// Create other with matching prefix
			other := make([]byte, otherLen)
			for i := 0; i < otherLen && i < 8; i++ {
				other[i] = byte(i + 1)
			}

			_, needFull := cmpInlineKeyPrefix(prefix, inlineLen, other)

			// needFullKey should ONLY be true when BOTH > 8 and first 8 match
			shouldNeedFull := inlineLen > 8 && otherLen > 8

			if needFull != shouldNeedFull {
				t.Errorf("inlineLen=%d, otherLen=%d: got needFullKey=%v, want %v",
					inlineLen, otherLen, needFull, shouldNeedFull)
			}
		}
	}
}

// TestCmpInlineKeyPrefixMatchesBytesCompare verifies that when needFullKey=false,
// the returned cmp matches what bytes.Compare would return for the full keys.
func TestCmpInlineKeyPrefixMatchesBytesCompare(t *testing.T) {
	testCases := []struct {
		fullKey []byte
		other   []byte
	}{
		{[]byte{}, []byte{}},
		{[]byte{1}, []byte{}},
		{[]byte{}, []byte{1}},
		{[]byte{1, 2, 3}, []byte{1, 2, 3}},
		{[]byte{1, 2, 3}, []byte{1, 2, 4}},
		{[]byte{1, 2, 3}, []byte{1, 2}},
		{[]byte{1, 2}, []byte{1, 2, 3}},
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8}, []byte{1, 2, 3, 4, 5, 6, 7, 8}},
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8}, []byte{1, 2, 3, 4, 5, 6, 7, 9}},
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8}, []byte{1, 2, 3, 4, 5, 6, 7}},
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8}, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}},
	}

	for _, tc := range testCases {
		// Build the inline prefix (first 8 bytes of fullKey, padded)
		var prefix [8]byte
		copy(prefix[:], tc.fullKey)

		cmp, needFull := cmpInlineKeyPrefix(prefix[:], len(tc.fullKey), tc.other)
		if needFull {
			continue // can't verify without full key
		}

		expected := bytes.Compare(tc.fullKey, tc.other)
		// Normalize to -1, 0, 1
		if expected < 0 {
			expected = -1
		} else if expected > 0 {
			expected = 1
		}

		if cmp != expected {
			t.Errorf("fullKey=%v, other=%v: got cmp=%d, want %d",
				tc.fullKey, tc.other, cmp, expected)
		}
	}
}
