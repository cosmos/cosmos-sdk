package errors

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestStackTrace(t *testing.T) {
	cases := map[string]struct {
		err       error
		wantError string
	}{
		"New gives us a stacktrace": {
			err:       Extend(ErrNoSignatures, "name"),
			wantError: "no signatures supplied: name",
		},
		"Wrapping stderr gives us a stacktrace": {
			err:       Extend(fmt.Errorf("foo"), "standard"),
			wantError: "foo: standard",
		},
		"Wrapping pkg/errors gives us clean stacktrace": {
			err:       Extend(errors.New("bar"), "pkg"),
			wantError: "bar: pkg",
		},
		"Wrapping inside another function is still clean": {
			err:       Extend(fmt.Errorf("indirect"), "do the do"),
			wantError: "indirect: do the do",
		},
	}

	// Wrapping code is unwanted in the errors stack trace.
	unwantedSrc := []string{
		"github.com/cosmos/cosmos-sdk/types/errors.Extend\n",
		"github.com/cosmos/cosmos-sdk/types/errors.Extendf\n",
		"runtime.goexit\n",
	}
	const thisTestSrc = "types/errors/stacktrace_test.go"

	for testName, tc := range cases {
		tc := tc
		t.Run(testName, func(t *testing.T) {
			if !reflect.DeepEqual(tc.err.Error(), tc.wantError) {
				t.Fatalf("errors not equal, got '%s', want '%s'", tc.err.Error(), tc.wantError)
			}

			if stackTrace(tc.err) == nil {
				t.Fatal("expected a stack trace to be present")
			}

			fullStack := fmt.Sprintf("%+v", tc.err)
			if !strings.Contains(fullStack, thisTestSrc) {
				t.Logf("Stack trace below\n----%s\n----", fullStack)
				t.Error("full stack trace should contain this test source code information")
			}
			if !strings.Contains(fullStack, tc.wantError) {
				t.Logf("Stack trace below\n----%s\n----", fullStack)
				t.Error("full stack trace should contain the error description")
			}
			for _, src := range unwantedSrc {
				if strings.Contains(fullStack, src) {
					t.Logf("Stack trace below\n----%s\n----", fullStack)
					t.Logf("full stack contains unwanted source file path: %q", src)
				}
			}

			tinyStack := fmt.Sprintf("%v", tc.err)
			if !strings.HasPrefix(tinyStack, tc.wantError) {
				t.Fatalf("prefix mimssing: %s", tinyStack)
			}
			if strings.Contains(tinyStack, "\n") {
				t.Fatal("only one stack line is expected")
			}
			// contains a link to where it was created, which must
			// be here, not the Extend() function
			if !strings.Contains(tinyStack, thisTestSrc) {
				t.Fatalf("this file missing in stack info:\n %s", tinyStack)
			}
		})
	}
}
