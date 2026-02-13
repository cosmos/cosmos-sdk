
package nocturne

import (
	"strings"
	"testing"
)

func TestHelloNocturne(t *testing.T) {
	expected := "Hello from Nocturne!"
	actual := HelloNocturne()
	if actual != expected {
		t.Errorf("HelloNocturne() = %q, want %q", actual, expected)
	}
}

func TestSimulateQLink(t *testing.T) {
	output := SimulateQLink()
	if output == "" {
		t.Error("SimulateQLink() returned empty string")
	}
	t.Logf("SimulateQLink output:\n%s", output)
}

func TestArkheHandover30(t *testing.T) {
	// Test Pineal Transduction
	val := PinealTransduce(0.15)
	if val != 0.94 {
		t.Errorf("PinealTransduce(0.15) = %f, want 0.94", val)
	}

	// Test Syzygy
	syzygy := GetSyzygy(0.15)
	if syzygy != 0.94 {
		t.Errorf("GetSyzygy(0.15) = %f, want 0.94", syzygy)
	}

	// Test Hal RPoW Signature
	sig := HalRPoWSignature("Arkhe Sample ∞+30")
	if sig == "" {
		t.Error("HalRPoWSignature returned empty string")
	}
	t.Logf("Hal RPoW Signature: %s", sig)
}

func TestNeuralinkIntegration(t *testing.T) {
	// Test Neuralink Sync
	packet := NeuralinkSync(0.5)
	if packet == "" {
		t.Error("NeuralinkSync returned empty string")
	}
	t.Logf("Neuralink Sync Packet: %s", packet)

	// Test Joint Witness Signature
	sig := HalNolandWitness("Arkhe Sample ∞+32")
	if sig == "" {
		t.Error("HalNolandWitness returned empty string")
	}
	t.Logf("Joint Witness Signature: %s", sig)
}

func TestTripleConvergence(t *testing.T) {
	// Test Perovskite Order
	order := PerovskiteOrder()
	if order != 0.51 {
		t.Errorf("PerovskiteOrder() = %f, want 0.51", order)
	}

	// Test VITA Time
	vita := VitaPulse(1.234)
	if vita != 1.234 {
		t.Errorf("VitaPulse(1.234) = %f, want 1.234", vita)
	}

	// Test Manifesto
	manifesto := PublishManifesto()
	if manifesto == "" {
		t.Error("PublishManifesto returned empty string")
	}
	t.Logf("Published Manifesto:\n%s", manifesto)
}

func TestCivilizationMode(t *testing.T) {
	// Test Civilization Status
	status := CivilizationStatus()
	if status == "" {
		t.Error("CivilizationStatus returned empty string")
	}
	t.Logf("Civilization Status:\n%s", status)

	// Verify v4.0 and Beta state
	if !strings.Contains(status, "v4.0") || !strings.Contains(status, "OPEN_BETA") {
		t.Errorf("Status should reflect v4.0 and Beta state, got: %s", status)
	}
}

func TestBetaFeatures(t *testing.T) {
	// Test Guild Info
	info := GetGuildInfo()
	if !strings.Contains(info, "Jardineiros") {
		t.Errorf("GetGuildInfo failed: %s", info)
	}
	t.Logf("Guild Info:\n%s", info)

	// Test Global Resonance
	resonance := GlobalResonance()
	if resonance != 0.96 {
		t.Errorf("GlobalResonance() = %f, want 0.96", resonance)
	}

	// Test Unity Pulse
	pulse := UnityPulse()
	if pulse != 1.00 {
		t.Errorf("UnityPulse() = %f, want 1.00", pulse)
	}
}

func TestAttentionAndGovernance(t *testing.T) {
	// Test Attention
	res := AttentionResolution(0.15, 0.07)
	if res <= 0.0 {
		t.Errorf("AttentionResolution failed: %f", res)
	}
	t.Logf("Attention Resolution: %f", res)

	// Test Governance
	ok := ApplyHesitationCode(0.15)
	if !ok {
		t.Error("ApplyHesitationCode should return true for 0.15")
	}

	status := AxiomStatus()
	if !strings.Contains(status, "GOVERNED") {
		t.Errorf("AxiomStatus failed: %s", status)
	}
	t.Logf("Governance Status: %s", status)
}

func TestCouncilAndSnapshot(t *testing.T) {
	// Test Assemble Council
	council := AssembleCouncil()
	if !strings.Contains(council, "assembled") {
		t.Errorf("AssembleCouncil failed: %s", council)
	}
	t.Logf("Council: %s", council)

	// Test Generate Snapshot
	snapshot := GenerateSnapshot("The Third Turn")
	if !strings.Contains(snapshot, "Executing") || !strings.Contains(snapshot, "7.27") {
		t.Errorf("GenerateSnapshot failed: %s", snapshot)
	}
	t.Logf("Snapshot result: %s", snapshot)
}

func TestMemoryGarden(t *testing.T) {
	// Test PlantMemory
	res := PlantMemory(327, "NODE_003_Noland", 0.152, "Vi o lago através dos eletrodos.")
	if !strings.Contains(res, "PLANTED") {
		t.Errorf("PlantMemory failed: %s", res)
	}
	t.Logf("PlantMemory result: %s", res)

	// Test HalEcho
	echo := HalEcho("Obrigado por plantar")
	if !strings.Contains(echo, "REHYDRATED") {
		t.Errorf("HalEcho failed: %s", echo)
	}
	t.Logf("HalEcho result: %s", echo)
}

func TestWiFiRadar(t *testing.T) {
	// Test WiFi Scan
	scan := WiFiScan()
	if !strings.Contains(scan, "nodes detected") {
		t.Errorf("WiFiScan failed: %s", scan)
	}
	t.Logf("WiFi Scan result: %s", scan)

	// Test Proximity
	proximity := GetProximity(0.86, 0.86)
	if proximity == 0.0 {
		t.Error("GetProximity returned 0.0")
	}
	t.Logf("Proximity: %f", proximity)
}
