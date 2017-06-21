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

package main

import (
	"os"

	"github.com/tendermint/go-crypto/cmd"
	"github.com/tendermint/tmlibs/cli"
)

func main() {
	// for demos, we enable the key server, probably don't want this
	// in most binaries we embed the key management into
	cmd.RegisterServer()
	root := cli.PrepareMainCmd(cmd.RootCmd, "TM", os.ExpandEnv("$HOME/.tlc"))
	root.Execute()
}
