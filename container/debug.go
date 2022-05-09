package container

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/pkg/errors"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

// DebugOption is a functional option for running a container that controls
// debug logging and visualization output.
type DebugOption interface {
	applyConfig(*debugConfig) error
}

// StdoutLogger is a debug option which routes logging output to stdout.
func StdoutLogger() DebugOption {
	return Logger(func(s string) {
		_, _ = fmt.Fprintln(os.Stdout, s)
	})
}

// Visualizer creates an option which provides a visualizer function which
// will receive a rendering of the container in the Graphiz DOT format
// whenever the container finishes building or fails due to an error. The
// graph is color-coded to aid debugging.
func Visualizer(visualizer func(dotGraph string)) DebugOption {
	return debugOption(func(c *debugConfig) error {
		c.addFuncVisualizer(visualizer)
		return nil
	})
}

// LogVisualizer is a debug option which dumps a graphviz DOT rendering of
// the container to the log.
func LogVisualizer() DebugOption {
	return debugOption(func(c *debugConfig) error {
		c.enableLogVisualizer()
		return nil
	})
}

// FileVisualizer is a debug option which dumps a graphviz rendering of
// the container to the specified file with the specified format. Currently
// supported formats are: dot, svg, png and jpg.
func FileVisualizer(filename, format string) DebugOption {
	return debugOption(func(c *debugConfig) error {
		c.addFileVisualizer(filename, format)
		return nil
	})
}

// Logger creates an option which provides a logger function which will
// receive all log messages from the container.
func Logger(logger func(string)) DebugOption {
	return debugOption(func(c *debugConfig) error {
		logger("Initializing logger")
		c.loggers = append(c.loggers, logger)

		// send conditional log messages batched for onError/onSuccess cases
		if c.logBuf != nil {
			for _, s := range *c.logBuf {
				logger(s)
			}
		}

		return nil
	})
}

const debugContainerSvg = "debug_container.svg"
const debugContainerDot = "debug_container.dot"

// Debug is a default debug option which sends log output to stdout, dumps
// the container in the graphviz DOT and SVG formats to debug_container.dot
// and debug_container.svg respectively.
func Debug() DebugOption {
	return DebugOptions(
		StdoutLogger(),
		FileVisualizer(debugContainerSvg, "svg"),
		FileVisualizer(debugContainerDot, "dot"),
	)
}

func (d *debugConfig) initLogBuf() {
	if d.logBuf == nil {
		d.logBuf = &[]string{}
		d.loggers = append(d.loggers, func(s string) {
			*d.logBuf = append(*d.logBuf, s)
		})
	}
}

// OnError is a debug option that allows setting debug options that are
// conditional on an error happening. Any loggers added error will
// receive the full dump of logs since the start of container processing.
func OnError(option DebugOption) DebugOption {
	return debugOption(func(config *debugConfig) error {
		config.initLogBuf()
		config.onError = option
		return nil
	})
}

// OnSuccess is a debug option that allows setting debug options that are
// conditional on successful container resolution. Any loggers added on success
// will receive the full dump of logs since the start of container processing.
func OnSuccess(option DebugOption) DebugOption {
	return debugOption(func(config *debugConfig) error {
		config.initLogBuf()
		config.onSuccess = option
		return nil
	})
}

// DebugCleanup specifies a clean-up function to be called at the end of
// processing to clean up any resources that may be used during debugging.
func DebugCleanup(cleanup func()) DebugOption {
	return debugOption(func(config *debugConfig) error {
		config.cleanup = append(config.cleanup, cleanup)
		return nil
	})
}

// AutoDebug does the same thing as Debug when there is an error and deletes
// the debug_container.dot and debug_container.dot if they exist when there
// is no error. This is the default debug mode of Run.
func AutoDebug() DebugOption {
	return DebugOptions(
		OnError(Debug()),
		OnSuccess(DebugCleanup(func() {
			deleteIfExists(debugContainerSvg)
			deleteIfExists(debugContainerDot)
		})),
	)
}

func deleteIfExists(filename string) {
	if _, err := os.Stat(filename); err == nil {
		_ = os.Remove(filename)
	}
}

// DebugOptions creates a debug option which bundles together other debug options.
func DebugOptions(options ...DebugOption) DebugOption {
	return debugOption(func(c *debugConfig) error {
		for _, opt := range options {
			err := opt.applyConfig(c)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

type debugConfig struct {
	// logging
	loggers   []func(string)
	indentStr string
	logBuf    *[]string // a log buffer for onError/onSuccess processing

	// graphing
	graphviz      *graphviz.Graphviz
	graph         *cgraph.Graph
	visualizers   []func(string)
	logVisualizer bool

	// extra processing
	onError   DebugOption
	onSuccess DebugOption
	cleanup   []func()
}

type debugOption func(*debugConfig) error

func (c debugOption) applyConfig(ctr *debugConfig) error {
	return c(ctr)
}

var _ DebugOption = (*debugOption)(nil)

func newDebugConfig() (*debugConfig, error) {
	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		return nil, errors.Wrap(err, "error initializing graph")
	}

	return &debugConfig{
		graphviz: g,
		graph:    graph,
	}, nil
}

func (c *debugConfig) indentLogger() {
	c.indentStr = c.indentStr + " "
}

func (c *debugConfig) dedentLogger() {
	if len(c.indentStr) > 0 {
		c.indentStr = c.indentStr[1:]
	}
}

func (c debugConfig) logf(format string, args ...interface{}) {
	s := fmt.Sprintf(c.indentStr+format, args...)
	for _, logger := range c.loggers {
		logger(s)
	}
}

func (c *debugConfig) generateGraph() {
	buf := &bytes.Buffer{}
	err := c.graphviz.Render(c.graph, graphviz.XDOT, buf)
	if err != nil {
		c.logf("Error rendering DOT graph: %+v", err)
		return
	}

	dot := buf.String()
	if c.logVisualizer {
		c.logf("DOT Graph: %s", dot)
	}

	for _, v := range c.visualizers {
		v(dot)
	}

	err = c.graph.Close()
	if err != nil {
		c.logf("Error closing graph: %+v", err)
		return
	}

	err = c.graphviz.Close()
	if err != nil {
		c.logf("Error closing graphviz: %+v", err)
	}
}

func (c *debugConfig) addFuncVisualizer(f func(string)) {
	c.visualizers = append(c.visualizers, func(dot string) {
		f(dot)
	})
}

func (c *debugConfig) enableLogVisualizer() {
	c.logVisualizer = true
}

func (c *debugConfig) addFileVisualizer(filename string, format string) {
	c.visualizers = append(c.visualizers, func(_ string) {
		err := c.graphviz.RenderFilename(c.graph, graphviz.Format(format), filename)
		if err != nil {
			c.logf("Error saving graphviz file %s with format %s: %+v", filename, format, err)
		} else {
			path, err := filepath.Abs(filename)
			if err == nil {
				c.logf("Saved graph of container to %s", path)
			}
		}
	})
}

func (c *debugConfig) locationGraphNode(location Location, key *moduleKey) (*cgraph.Node, error) {
	graph := c.moduleSubGraph(key)
	node, found, err := c.findOrCreateGraphNode(graph, location.Name())
	if err != nil {
		return nil, err
	}

	if found {
		return node, nil
	}

	node.SetShape(cgraph.BoxShape)
	node.SetColor("lightgrey")
	return node, nil
}

func (c *debugConfig) typeGraphNode(typ reflect.Type) (*cgraph.Node, error) {
	node, found, err := c.findOrCreateGraphNode(c.graph, moreUsefulTypeString(typ))
	if err != nil {
		return nil, err
	}

	if found {
		return node, nil
	}

	node.SetColor("lightgrey")
	return node, err
}

// moreUsefulTypeString is more useful than reflect.Type.String()
func moreUsefulTypeString(ty reflect.Type) string {
	switch ty.Kind() {
	case reflect.Struct, reflect.Interface:
		return fmt.Sprintf("%s.%s", ty.PkgPath(), ty.Name())
	case reflect.Pointer:
		return fmt.Sprintf("*%s", moreUsefulTypeString(ty.Elem()))
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", moreUsefulTypeString(ty.Key()), moreUsefulTypeString(ty.Elem()))
	case reflect.Slice:
		return fmt.Sprintf("[]%s", moreUsefulTypeString(ty.Elem()))
	default:
		return ty.String()
	}
}

func (c *debugConfig) findOrCreateGraphNode(subGraph *cgraph.Graph, name string) (node *cgraph.Node, found bool, err error) {
	node, err = c.graph.Node(name)
	if err != nil {
		return nil, false, errors.Wrapf(err, "error finding graph node %s", name)
	}

	if node != nil {
		return node, true, nil
	}

	node, err = subGraph.CreateNode(name)
	if err != nil {
		return nil, false, errors.Wrapf(err, "error creating graph node %s", name)
	}

	return node, false, nil
}

func (c *debugConfig) moduleSubGraph(key *moduleKey) *cgraph.Graph {
	graph := c.graph
	if key != nil {
		gname := fmt.Sprintf("cluster_%s", key.name)
		graph = c.graph.SubGraph(gname, 1)
		graph.SetLabel(fmt.Sprintf("Module: %s", key.name))
	}
	return graph
}

func (c *debugConfig) addGraphEdge(from *cgraph.Node, to *cgraph.Node) {
	_, err := c.graph.CreateEdge("", from, to)
	if err != nil {
		c.logf("error creating graph edge")
	}
}
