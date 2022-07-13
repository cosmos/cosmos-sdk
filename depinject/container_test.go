package depinject_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"

	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/internal/testgen"
)

func TestScenario(t *testing.T) {
	var (
		handlers map[string]testgen.Handler
		commands []testgen.Command
		a        testgen.KeeperA
		b        testgen.KeeperB
	)
	require.NoError(t,
		depinject.Inject(
			depinject.Configs(
				testgen.ScenarioConfig,
				depinject.Supply(testgen.ModuleA{}, testgen.ModuleB{}),
			),
			&handlers,
			&commands,
			&a,
			&b,
		))

	require.Len(t, handlers, 2)
	require.Equal(t, testgen.Handler{}, handlers["a"])
	require.Equal(t, testgen.Handler{}, handlers["b"])
	require.Len(t, commands, 3)
	require.Equal(t, testgen.KeeperA{
		Key:  testgen.KVStoreKey{Name: "a"},
		Name: "a",
	}, a)
	require.Equal(t, testgen.KeeperB{
		Key: testgen.KVStoreKey{Name: "b"},
		MsgClientA: testgen.MsgClientA{
			Key: "b",
		},
	}, b)
}

func TestResolveError(t *testing.T) {
	var x string
	require.Error(t, depinject.Inject(
		depinject.Provide(
			func(x float64) string { return fmt.Sprintf("%f", x) },
			func(x int) float64 { return float64(x) },
			func(x float32) int { return int(x) },
		),
		&x,
	))
}

func TestCyclic(t *testing.T) {
	var x string
	require.Error(t, depinject.Inject(
		depinject.Provide(
			func(x int) float64 { return float64(x) },
			func(x float64) (int, string) { return int(x), "hi" },
		),
		&x,
	))
}

func TestErrorOption(t *testing.T) {
	err := depinject.Inject(depinject.Error(fmt.Errorf("an error")))
	require.Error(t, err)
}

func TestBadCtr(t *testing.T) {
	_, err := depinject.ExtractProviderDescriptor(testgen.KeeperA{})
	require.Error(t, err)
}

func TestTrivial(t *testing.T) {
	require.NoError(t, depinject.Inject(depinject.Configs()))
}

func TestErrorFunc(t *testing.T) {
	_, err := depinject.ExtractProviderDescriptor(
		func() (error, int) { return nil, 0 },
	)
	require.Error(t, err)

	_, err = depinject.ExtractProviderDescriptor(
		func() (int, error) { return 0, nil },
	)
	require.NoError(t, err)

	var x int
	require.Error(t,
		depinject.Inject(
			depinject.Provide(func() (int, error) {
				return 0, fmt.Errorf("the error")
			}),
			&x,
		))
}

func TestSimple(t *testing.T) {
	var x int
	require.NoError(t,
		depinject.Inject(
			depinject.Provide(
				func() int { return 1 },
			),
			&x,
		),
	)

	require.Error(t,
		depinject.Inject(
			depinject.Provide(
				func() int { return 0 },
				func() int { return 1 },
			),
			&x,
		),
	)
}

func TestModuleScoped(t *testing.T) {
	var x int
	require.Error(t,
		depinject.Inject(
			depinject.Provide(
				func(depinject.ModuleKey) int { return 0 },
			),
			&x,
		),
	)

	var y float64
	require.Error(t,
		depinject.Inject(
			depinject.Configs(
				depinject.Provide(
					func(depinject.ModuleKey) int { return 0 },
					func() int { return 1 },
				),
				depinject.ProvideInModule("a",
					func(x int) float64 { return float64(x) },
				),
			),
			&y,
		),
	)

	require.Error(t,
		depinject.Inject(
			depinject.Configs(
				depinject.Provide(
					func() int { return 0 },
					func(depinject.ModuleKey) int { return 1 },
				),
				depinject.ProvideInModule("a",
					func(x int) float64 { return float64(x) },
				),
			),
			&y,
		),
	)

	require.Error(t,
		depinject.Inject(
			depinject.Configs(
				depinject.Provide(
					func(depinject.ModuleKey) int { return 0 },
					func(depinject.ModuleKey) int { return 1 },
				),
				depinject.ProvideInModule("a",
					func(x int) float64 { return float64(x) },
				),
			),
			&y,
		),
	)

	require.NoError(t,
		depinject.Inject(
			depinject.Configs(
				depinject.Provide(
					func(depinject.ModuleKey) int { return 0 },
				),
				depinject.ProvideInModule("a",
					func(x int) float64 { return float64(x) },
				),
			),
			&y,
		),
	)

	require.Error(t,
		depinject.Inject(
			depinject.Configs(
				depinject.Provide(
					func(depinject.ModuleKey) int { return 0 },
				),
				depinject.ProvideInModule("",
					func(x int) float64 { return float64(x) },
				),
			),
			&y,
		),
	)

	var z float32
	require.NoError(t,
		depinject.Inject(
			depinject.Configs(
				depinject.Provide(
					func(depinject.ModuleKey) int { return 0 },
				),
				depinject.ProvideInModule("a",
					func(x int) float64 { return float64(x) },
					func(x int) float32 { return float32(x) },
				),
			),
			&y, &z,
		),
		"use module dep twice",
	)
}

type OnePerModuleInt int

func (OnePerModuleInt) IsOnePerModuleType() {}

func TestOnePerModule(t *testing.T) {
	var x OnePerModuleInt
	require.Error(t,
		depinject.Inject(depinject.Configs(), &x),
		"bad input type",
	)

	var y map[string]OnePerModuleInt
	var z string
	require.NoError(t,
		depinject.Inject(
			depinject.Configs(
				depinject.ProvideInModule("a",
					func() OnePerModuleInt { return 3 },
				),
				depinject.ProvideInModule("b",
					func() OnePerModuleInt { return 4 },
				),
				depinject.Provide(func(x map[string]OnePerModuleInt) string {
					sum := 0
					for _, v := range x {
						sum += int(v)
					}
					return fmt.Sprintf("%d", sum)
				}),
			),
			&y,
			&z,
		),
	)

	require.Equal(t, map[string]OnePerModuleInt{
		"a": 3,
		"b": 4,
	}, y)
	require.Equal(t, "7", z)

	var m map[string]OnePerModuleInt
	require.Error(t,
		depinject.Inject(
			depinject.ProvideInModule("a",
				func() OnePerModuleInt { return 0 },
				func() OnePerModuleInt { return 0 },
			),
			&m,
		),
		"duplicate",
	)

	require.Error(t,
		depinject.Inject(
			depinject.Provide(
				func() OnePerModuleInt { return 0 },
			),
			&m,
		),
		"out of scope",
	)

	require.Error(t,
		depinject.Inject(
			depinject.Provide(
				func() map[string]OnePerModuleInt { return nil },
			),
			&m,
		),
		"bad return type",
	)

	require.NoError(t,
		depinject.Inject(
			depinject.Configs(),
			&m,
		),
		"no providers",
	)
}

type ManyPerContainerInt int

func (ManyPerContainerInt) IsManyPerContainerType() {}

func TestManyPerContainer(t *testing.T) {
	var xs []ManyPerContainerInt
	var sum string
	require.NoError(t,
		depinject.Inject(
			depinject.Provide(
				func() ManyPerContainerInt { return 4 },
				func() ManyPerContainerInt { return 9 },
				func(xs []ManyPerContainerInt) string {
					sum := 0
					for _, x := range xs {
						sum += int(x)
					}
					return fmt.Sprintf("%d", sum)
				},
			),
			&xs,
			&sum,
		),
	)
	require.Len(t, xs, 2)
	require.Contains(t, xs, ManyPerContainerInt(4))
	require.Contains(t, xs, ManyPerContainerInt(9))
	require.Equal(t, "13", sum)

	var z ManyPerContainerInt
	require.Error(t,
		depinject.Inject(
			depinject.Provide(
				func() ManyPerContainerInt { return 0 },
			),
			&z,
		),
		"bad input type",
	)

	require.NoError(t,
		depinject.Inject(
			depinject.Configs(),
			&xs,
		),
		"no providers",
	)
}

func TestSupply(t *testing.T) {
	var x int
	require.NoError(t,
		depinject.Inject(
			depinject.Supply(3),
			&x,
		),
	)
	require.Equal(t, 3, x)

	require.Error(t,
		depinject.Inject(
			depinject.Configs(
				depinject.Supply(3),
				depinject.Provide(func() int { return 4 }),
			),
			&x,
		),
		"can't supply then provide",
	)

	require.Error(t,
		depinject.Inject(
			depinject.Configs(
				depinject.Supply(3),
				depinject.Provide(func() int { return 4 }),
			),
			&x,
		),
		"can't provide then supply",
	)

	require.Error(t,
		depinject.Inject(
			depinject.Supply(3, 4),
			&x,
		),
		"can't supply twice",
	)
}

type TestInput struct {
	depinject.In

	X int `optional:"true"`
	Y float64
}

type TestOutput struct {
	depinject.Out

	X string
	Y int64
}

func TestStructArgs(t *testing.T) {
	var input TestInput
	require.Error(t, depinject.Inject(depinject.Configs(), &input))

	require.NoError(t, depinject.Inject(
		depinject.Supply(1.3),
		&input,
	))
	require.Equal(t, 0, input.X)
	require.Equal(t, 1.3, input.Y)

	require.NoError(t, depinject.Inject(
		depinject.Supply(1.3, 1),
		&input,
	))
	require.Equal(t, 1, input.X)
	require.Equal(t, 1.3, input.Y)

	var x string
	var y int64
	require.NoError(t, depinject.Inject(
		depinject.Provide(func() (TestOutput, error) {
			return TestOutput{X: "A", Y: -10}, nil
		}),
		&x, &y,
	))
	require.Equal(t, "A", x)
	require.Equal(t, int64(-10), y)

	require.Error(t, depinject.Inject(
		depinject.Provide(func() (TestOutput, error) {
			return TestOutput{}, fmt.Errorf("error")
		}),
		&x,
	))
}

func TestDebugOptions(t *testing.T) {
	var logOut string
	var dotGraph string

	outfile, err := os.CreateTemp("", "out")
	require.NoError(t, err)
	stdout := os.Stdout
	os.Stdout = outfile
	defer func() { os.Stdout = stdout }()
	defer os.Remove(outfile.Name())

	graphfile, err := os.CreateTemp("", "graph")
	require.NoError(t, err)
	defer os.Remove(graphfile.Name())

	require.NoError(t, depinject.InjectDebug(
		depinject.DebugOptions(
			depinject.Logger(func(s string) {
				logOut += s
			}),
			depinject.Visualizer(func(g string) {
				dotGraph = g
			}),
			depinject.LogVisualizer(),
			depinject.FileVisualizer(graphfile.Name()),
			depinject.StdoutLogger(),
		),
		depinject.Configs(),
	))

	require.Contains(t, logOut, "digraph")
	require.Contains(t, dotGraph, "digraph")

	outfileContents, err := os.ReadFile(outfile.Name())
	require.NoError(t, err)
	require.Contains(t, string(outfileContents), "digraph")

	graphfileContents, err := os.ReadFile(graphfile.Name())
	require.NoError(t, err)
	require.Contains(t, string(graphfileContents), "digraph")
}

func TestGraphAndLogOutput(t *testing.T) {
	var graphOut string
	var b testgen.KeeperB
	debugOpts := depinject.DebugOptions(
		depinject.Visualizer(func(dotGraph string) {
			graphOut = dotGraph
		}))
	require.NoError(t, depinject.InjectDebug(debugOpts, testgen.ScenarioConfig, &b))
	golden.Assert(t, graphOut, "example.dot")

	badConfig := depinject.Configs(testgen.ScenarioConfig)
	require.Error(t, depinject.InjectDebug(debugOpts, badConfig, &b))
	golden.Assert(t, graphOut, "example_error.dot")
}

func TestConditionalDebugging(t *testing.T) {
	logs := ""
	success := false
	conditionalDebugOpt := depinject.DebugOptions(
		depinject.OnError(depinject.Logger(func(s string) {
			logs += s + "\n"
		})),
		depinject.OnSuccess(depinject.DebugCleanup(func() {
			success = true
		})))

	var input TestInput
	require.Error(t, depinject.InjectDebug(
		conditionalDebugOpt,
		depinject.Configs(),
		&input,
	))
	require.Contains(t, logs, `Initializing logger`)
	require.Contains(t, logs, `Registering providers`)
	require.Contains(t, logs, `Registering outputs`)
	require.False(t, success)

	logs = ""
	success = false
	require.NoError(t, depinject.InjectDebug(
		conditionalDebugOpt,
		depinject.Configs(),
	))
	require.Empty(t, logs)
	require.True(t, success)
}
