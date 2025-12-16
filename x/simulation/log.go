package simulation

import (
	"fmt"
	"os"
	"path"
	"sync"
	"time"
)

type LogWriter interface {
	AddEntry(OperationEntry)
	PrintLogs()
}

// NewLogWriter returns a dummy or standard log writer given the testingmode
func NewLogWriter(testingmode bool) LogWriter {
	if !testingmode {
		return &DummyLogWriter{}
	}

	return &StandardLogWriter{}
}

// StandardLogWriter is a standard log writer that writes the logs to a file.
type StandardLogWriter struct {
	Seed int64

	OpEntries []OperationEntry `json:"op_entries" yaml:"op_entries"`
	wMtx      sync.Mutex
	written   bool
}

// AddEntry adds an entry to the log writer
func (lw *StandardLogWriter) AddEntry(opEntry OperationEntry) {
	lw.OpEntries = append(lw.OpEntries, opEntry)
}

// PrintLogs - print the logs to a simulation file
func (lw *StandardLogWriter) PrintLogs() {
	lw.wMtx.Lock()
	defer lw.wMtx.Unlock()
	if lw.written { // print once only
		return
	}
	f := createLogFile(lw.Seed)
	defer f.Close()

	for i := range lw.OpEntries {
		writeEntry := fmt.Sprintf("%s\n", (lw.OpEntries[i]).MustMarshal())
		_, err := f.WriteString(writeEntry)
		if err != nil {
			panic("Failed to write logs to file")
		}
	}
	lw.written = true
}

func createLogFile(seed int64) *os.File {
	var f *os.File
	var prefix string
	if seed != 0 {
		prefix = fmt.Sprintf("seed_%10d", seed)
	}
	fileName := fmt.Sprintf("%s--%d.log", prefix, time.Now().UnixNano())
	folderPath := path.Join(os.ExpandEnv("$HOME"), ".simapp", "simulations")
	filePath := path.Join(folderPath, fileName)

	err := os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		panic(err)
	}

	f, err = os.Create(filePath)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Logs to writing to %s\n", filePath)

	return f
}

type DummyLogWriter struct{}

func (lw *DummyLogWriter) AddEntry(_ OperationEntry) {}

func (lw *DummyLogWriter) PrintLogs() {}
