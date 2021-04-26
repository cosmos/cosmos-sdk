package main

import (
	"fmt"
	"os"
	"time"

	amino "github.com/tendermint/go-amino"
)

func main() {
	times := []time.Time{
		time.Unix(0, 0),
		time.Now(),
		time.Date(1979, time.January, 02, 10, 11, 12, 7999192, time.UTC),
		time.Now().Add(10000 * time.Hour),
		time.Now().Add(100 * time.Hour),
		time.Now().Add(-100 * time.Hour),
		time.Now().Add(-10000 * time.Hour),
	}

	for i, t := range times {
		func() {
			f, err := os.Create(fmt.Sprintf("%d.txt", i+1))
			if err != nil {
				panic(err)
			}
			defer f.Close()

			if err := amino.EncodeTime(f, t); err != nil {
				panic(err)
			}
		}()
	}
}
