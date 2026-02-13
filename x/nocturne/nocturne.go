
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

func PerovskiteOrder() float64 {
	return float64(C.nocturne_perovskite_order())
}

func VitaPulse(currentTime float64) float64 {
	return float64(C.nocturne_vita_pulse(C.double(currentTime)))
}

func PublishManifesto() string {
	cStr := C.nocturne_publish_manifesto()
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
	fmt.Printf("Perovskite Interface Order: %.2f\n", PerovskiteOrder())
	fmt.Printf("VITA Time: %.3fs\n", VitaPulse(1.0))
	fmt.Println(NeuralinkSync(0.5))
	fmt.Println(PublishManifesto())
}
