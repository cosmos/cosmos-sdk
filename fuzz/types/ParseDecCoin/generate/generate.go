package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	seeds := []struct {
		name string
		coin string
	}{
		{"empty", ""},
		{"valid", "1.0stake"},
		{"space_between_amount_denom", "1.0 stake"},
		{"leading_space", "  1.0stake"},
		{"trailing_space", "1.0stake  "},
		{"plus_sign", "+1.0stake"},
		{"minus_sign", "-1.0stake"},
		{"no_denom", "1.0"},
		{"less_than_3char_denom", "1.0s"},
		{"greater_than_128char_denom", "1.0" + strings.Repeat("s", 129)},
	}

	for _, seed := range seeds {
		func() {
			f, err := os.Create(fmt.Sprintf("corpus/%s.seed", seed.name))
			if err != nil {
				return
			}
			if _, err := f.Write([]byte(seed.coin)); err != nil {
				log.Fatal(err)
			}
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}()
	}
}
