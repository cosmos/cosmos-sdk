// Copyright © 2017 Ethan Frey
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get details of one key",
	Long:  `Return public details of one local key.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 || len(args[0]) == 0 {
			fmt.Println("You must provide a name for the key")
			return
		}
		name := args[0]

		info, err := GetKeyManager().Get(name)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		printInfo(info)
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
}
