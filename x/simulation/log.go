package simulation

import (
	"fmt"
	"os"
	"path"
	"time"
)

// log writter
type LogWriter interface {
	AddEntry(OperationEntry)
	PrintLogs()
}

// LogWriter - return a dummy or standard log writer given the testingmode
func NewLogWriter(testingmode bool) LogWriter {
	if !testingmode {
		return &DummyLogWriter{}
	}
	return &StandardLogWriter{}
}

// log writter
type StandardLogWriter struct {
	OpEntries []OperationEntry `json:"op_entries"`
}

// add an entry to the log writter
func (lw *StandardLogWriter) AddEntry(opEntry OperationEntry) {
	lw.OpEntries = append(lw.OpEntries, opEntry)
}

// PrintLogs - print the logs to a simulation file
func (lw *StandardLogWriter) PrintLogs() {
	f := createLogFile()
	for i := 0; i < len(lw.OpEntries); i++ {
		writeEntry := fmt.Sprintf("%s\n", (lw.OpEntries[i]).MustMarshal())
		_, err := f.WriteString(writeEntry)
		if err != nil {
			panic("Failed to write logs to file")
		}
	}
}

//// Builds a function to append logs
//func getLogWriter(testingmode bool, opEntries []*OperationEntry) func(OperationEntry) {

//if !testingmode {
//return func(_ OperationEntry) {}
//}

//return func(opEntry OperationEntry) {
//opEntries = append(opEntries, opEntry)
//}
//}

//// Creates a function to print out the logs
//func logPrinter(testingmode bool, opEntries []*OperationEntry) func() {
//if !testingmode {
//return func() {}
//}

//return func() {
//f := createLogFile()
//_, _ = f.WriteString(fmt.Sprintf("debug opEntries: %v\n", opEntries))
//for i := 0; i < len(opEntries); i++ {
//writeEntry := fmt.Sprintf("%s\n", (*opEntries[i]).MustMarshal())
//_, err := f.WriteString(writeEntry)
//if err != nil {
//panic("Failed to write logs to file")
//}
//}
//}
//}

func createLogFile() *os.File {
	var f *os.File
	fileName := fmt.Sprintf("%s.log", time.Now().Format("2006-01-02_15:04:05"))

	folderPath := os.ExpandEnv("$HOME/.gaiad/simulations")
	filePath := path.Join(folderPath, fileName)

	err := os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
	f, _ = os.Create(filePath)
	fmt.Printf("Logs to writing to %s\n", filePath)
	return f
}

//_____________________
// dummy log writter
type DummyLogWriter struct{}

// do nothing
func (lw *DummyLogWriter) AddEntry(_ OperationEntry) {}

// do nothing
func (lw *DummyLogWriter) PrintLogs() {}
