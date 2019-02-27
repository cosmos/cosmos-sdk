package simulation

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

// Builds a function to add logs for this particular block
func addLogMessage(testingmode bool,
	blockLogBuilders []*strings.Builder, height int) func(string) {

	if !testingmode {
		return func(_ string) {}
	}

	blockLogBuilders[height] = &strings.Builder{}
	return func(x string) {
		(*blockLogBuilders[height]).WriteString(x)
		(*blockLogBuilders[height]).WriteString("\n")
	}
}

// Creates a function to print out the logs
func logPrinter(testingmode bool, logs []*strings.Builder) func() {
	if !testingmode {
		return func() {}
	}

	return func() {
		numLoggers := 0
		for i := 0; i < len(logs); i++ {
			// We're passed the last created block
			if logs[i] == nil {
				numLoggers = i
				break
			}
		}

		var f *os.File
		if numLoggers > 10 {
			fileName := fmt.Sprintf("%s.txt", time.Now().Format("2006-01-02_15:04:05"))

			folderPath := os.ExpandEnv("$HOME/.gaiad/simulations")
			filePath := path.Join(folderPath, fileName)
			fmt.Printf("Too many logs to display, instead writing to %s\n", filePath)

			err := os.MkdirAll(folderPath, os.ModePerm)
			if err != nil {
				panic(err)
			}
			f, _ = os.Create(filePath)
		}

		for i := 0; i < numLoggers; i++ {
			if f == nil {
				fmt.Printf("Begin block %d\n", i+1)
				fmt.Println((*logs[i]).String())
				continue
			}

			_, err := f.WriteString(fmt.Sprintf("Begin block %d\n", i+1))
			if err != nil {
				panic("Failed to write logs to file")
			}

			_, err = f.WriteString((*logs[i]).String())
			if err != nil {
				panic("Failed to write logs to file")
			}
		}
	}
}
