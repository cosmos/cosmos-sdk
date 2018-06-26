// Copyright 2017 Tendermint. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crypto_test

import (
	"fmt"

	"github.com/tendermint/go-crypto"
)

func ExampleSha256() {
	sum := crypto.Sha256([]byte("This is Tendermint"))
	fmt.Printf("%x\n", sum)
	// Output:
	// f91afb642f3d1c87c17eb01aae5cb65c242dfdbe7cf1066cc260f4ce5d33b94e
}

func ExampleRipemd160() {
	sum := crypto.Ripemd160([]byte("This is Tendermint"))
	fmt.Printf("%x\n", sum)
	// Output:
	// 051e22663e8f0fd2f2302f1210f954adff009005
}
