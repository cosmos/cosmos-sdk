
package nocturne

// #cgo LDFLAGS: -L${SRCDIR}/target/release -lnocturne
// #include "nocturne.h"
import "C"

import "fmt"

func HelloNocturne() string {
	cStr := C.hello_nocturne()
	goStr := C.GoString(cStr)
	C.nocturne_free_string(cStr)
	return goStr
}

func SimulateQLink() string {
	cStr := C.simulate_qlink()
	goStr := C.GoString(cStr)
	C.nocturne_free_string(cStr)
	return goStr
}

func Example() {
	fmt.Println(HelloNocturne())
}
