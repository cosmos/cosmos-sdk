package common

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func TestSimulationReporterTransitions(t *testing.T) {
	specs := map[string]struct {
		setup     func(r SimulationReporter)
		expStatus ReporterStatus
		expPanic  bool
	}{
		"undefined->skipped": {
			setup: func(r SimulationReporter) {
				r.Skip("testing")
			},
			expStatus: ReporterStatusSkipped,
		},
		"skipped->skipped": {
			setup: func(r SimulationReporter) {
				r.Skip("testing1")
				r.Skip("testing2")
			},
			expStatus: ReporterStatusSkipped,
		},
		"skipped->completed": {
			setup: func(r SimulationReporter) {
				r.Skip("testing1")
				r.Success(nil, "testing2")
			},
			expStatus: ReporterStatusCompleted,
		},
		"completed->completed": {
			setup: func(r SimulationReporter) {
				r.Success(nil, "testing1")
				r.Fail(nil, "testing2")
			},
			expStatus: ReporterStatusCompleted,
		},
		"completed->completed2": {
			setup: func(r SimulationReporter) {
				r.Fail(nil, "testing1")
				r.Success(nil, "testing2")
			},
			expStatus: ReporterStatusCompleted,
		},
		"completed->skipped: rejected": {
			setup: func(r SimulationReporter) {
				r.Success(nil, "testing1")
				r.Skip("testing2")
			},
			expPanic: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			r := NewBasicSimulationReporter()
			if !spec.expPanic {
				spec.setup(r)
				assert.Equal(t, uint32(spec.expStatus), r.status.Load())
				return
			}
			require.Panics(t, func() {
				spec.setup(r)
			})
		})
	}
}

func TestSkipHook(t *testing.T) {
	myHook := func() (SkipHookFn, *bool) {
		var hookCalled bool
		return func(args ...any) {
			hookCalled = true
		}, &hookCalled
	}
	fn, myHookCalled := myHook()
	r := NewBasicSimulationReporter(fn)
	r.Skip("testing")
	require.True(t, *myHookCalled)

	// and with nested reporter
	fn, myHookCalled = myHook()
	r = NewBasicSimulationReporter(fn)
	fn2, myOtherHookCalled := myHook()
	r2 := r.WithScope(testdata.NewTestMsg([]byte{1}), fn2)
	r2.Skipf("testing %d", 2)
	assert.True(t, *myHookCalled)
	assert.True(t, *myOtherHookCalled)
}

func TestReporterSummary(t *testing.T) {
	specs := map[string]struct {
		do         func(t *testing.T, r SimulationReporter)
		expSummary map[string]int
		expReasons map[string]map[string]int
	}{
		"skipped": {
			do: func(t *testing.T, r SimulationReporter) { //nolint:thelper // not a helper
				r2 := r.WithScope(testdata.NewTestMsg([]byte{1}))
				r2.Skip("testing")
				require.NoError(t, r2.Close())
			},
			expSummary: map[string]int{"TestMsg_skipped": 1},
			expReasons: map[string]map[string]int{"/testpb.TestMsg": {"testing": 1}},
		},
		"success result": {
			do: func(t *testing.T, r SimulationReporter) { //nolint:thelper // not a helper
				msg := testdata.NewTestMsg([]byte{1})
				r2 := r.WithScope(msg)
				r2.Success(msg)
				require.NoError(t, r2.Close())
			},
			expSummary: map[string]int{"TestMsg_completed": 1},
			expReasons: map[string]map[string]int{},
		},
		"error result": {
			do: func(t *testing.T, r SimulationReporter) { //nolint:thelper // not a helper
				msg := testdata.NewTestMsg([]byte{1})
				r2 := r.WithScope(msg)
				r2.Fail(errors.New("testing"))
				require.Error(t, r2.Close())
			},
			expSummary: map[string]int{"TestMsg_completed": 1},
			expReasons: map[string]map[string]int{},
		},
		"multiple skipped": {
			do: func(t *testing.T, r SimulationReporter) { //nolint:thelper // not a helper
				r2 := r.WithScope(testdata.NewTestMsg([]byte{1}))
				r2.Skip("testing1")
				require.NoError(t, r2.Close())
				r3 := r.WithScope(testdata.NewTestMsg([]byte{2}))
				r3.Skip("testing2")
				require.NoError(t, r3.Close())
			},
			expSummary: map[string]int{"TestMsg_skipped": 2},
			expReasons: map[string]map[string]int{
				"/testpb.TestMsg": {"testing1": 1, "testing2": 1},
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			r := NewBasicSimulationReporter()
			// when
			spec.do(t, r)
			gotSummary := r.Summary()
			// then
			require.Equal(t, spec.expSummary, gotSummary.counts)
			require.Equal(t, spec.expReasons, gotSummary.skipReasons)
		})
	}
}
