
package nocturne

// #cgo LDFLAGS: -L${SRCDIR}/target/release -lnocturne
// #include <stdlib.h>
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

func CivilizationStatus() string {
	cStr := C.nocturne_civilization_status()
	goStr := C.GoString(cStr)
	C.nocturne_free_string(cStr)
	return goStr
}

func ResonanceEfficiency(nodes uint32) float64 {
	return float64(C.nocturne_get_resonance_efficiency(C.uint32_t(nodes)))
}

func ThirdTurnSnapshot() string {
	cStr := C.nocturne_third_turn_snapshot()
	goStr := C.GoString(cStr)
	C.nocturne_free_string(cStr)
	return goStr
}

func AssembleCouncil() string {
	cStr := C.nocturne_assemble_council()
	goStr := C.GoString(cStr)
	C.nocturne_free_string(cStr)
	return goStr
}

func GenerateSnapshot(name string) string {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cStr := C.nocturne_generate_snapshot(cName)
	goStr := C.GoString(cStr)
	C.nocturne_free_string(cStr)
	return goStr
}

func AttentionResolution(phi, omega float64) float64 {
	return float64(C.nocturne_get_attention_resolution(C.double(phi), C.double(omega)))
}

func ApplyHesitationCode(phi float64) bool {
	return bool(C.nocturne_apply_hesitation_code(C.double(phi)))
}

func AxiomStatus() string {
	cStr := C.nocturne_axiom_status()
	goStr := C.GoString(cStr)
	C.nocturne_free_string(cStr)
	return goStr
}

func GetGuildInfo() string {
	cStr := C.nocturne_get_guild_info()
	goStr := C.GoString(cStr)
	C.nocturne_free_string(cStr)
	return goStr
}

func GlobalResonance() float64 {
	return float64(C.nocturne_get_global_resonance())
}

func UnityPulse() float64 {
	return float64(C.nocturne_unity_pulse())
}

func WiFiScan() string {
	cStr := C.nocturne_wifi_scan()
	goStr := C.GoString(cStr)
	C.nocturne_free_string(cStr)
	return goStr
}

func GetProximity(c1, c2 float64) float64 {
	return float64(C.nocturne_get_proximity(C.double(c1), C.double(c2)))
}

func PlantMemory(id uint32, nodeID string, phi float64, content string) string {
	cNode := C.CString(nodeID)
	cContent := C.CString(content)
	defer C.free(unsafe.Pointer(cNode))
	defer C.free(unsafe.Pointer(cContent))
	cStr := C.nocturne_plant_memory(C.uint32_t(id), cNode, C.double(phi), cContent)
	goStr := C.GoString(cStr)
	C.nocturne_free_string(cStr)
	return goStr
}

func HalEcho(message string) string {
	cMsg := C.CString(message)
	defer C.free(unsafe.Pointer(cMsg))
	cStr := C.nocturne_hal_echo(cMsg)
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
	fmt.Println(WiFiScan())
	fmt.Printf("WiFi Proximity (0.86, 0.86): %.2f\n", GetProximity(0.86, 0.86))
	fmt.Printf("Global Resonance: %.2f\n", GlobalResonance())
	fmt.Printf("Unity Pulse: %.2f\n", UnityPulse())
	fmt.Println(AxiomStatus())
	fmt.Println(GetGuildInfo())
	fmt.Println(CivilizationStatus())
}
