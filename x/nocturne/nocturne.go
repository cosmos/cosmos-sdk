
package nocturne

// #cgo LDFLAGS: -L${SRCDIR}/target/release -lnocturne
// #include "nocturne.h"
import "C"

import (
	"fmt"
	"unsafe"
)

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

func PinealTransduce(phi float64) float64 {
	return float64(C.nocturne_pineal_transduce(C.double(phi)))
}

func GetSyzygy(phi float64) float64 {
	return float64(C.nocturne_get_syzygy(C.double(phi)))
}

func NeuralinkSync(intent float64) string {
	cStr := C.nocturne_neuralink_sync(C.double(intent))
	goStr := C.GoString(cStr)
	C.nocturne_free_string(cStr)
	return goStr
}

func HalNolandWitness(sample string) string {
	cSample := C.CString(sample)
	defer C.free(unsafe.Pointer(cSample))
	cStr := C.nocturne_hal_noland_witness(cSample)
	goStr := C.GoString(cStr)
	C.nocturne_free_string(cStr)
	return goStr
}

func HalRPoWSignature(sample string) string {
	cSample := C.CString(sample)
	defer C.free(unsafe.Pointer(cSample))
	cStr := C.nocturne_hal_rpow_signature(cSample)
	goStr := C.GoString(cStr)
	C.nocturne_free_string(cStr)
	return goStr
}

func Example() {
	fmt.Println(HelloNocturne())
	fmt.Printf("Pineal Transduction (Φ=0.15): %.2f\n", PinealTransduce(0.15))
	fmt.Printf("Syzygy Yield (Φ=0.15): %.2f\n", GetSyzygy(0.15))
	fmt.Println(NeuralinkSync(0.5))
	fmt.Println(HalNolandWitness("Arkhe Sample ∞+32"))
}
