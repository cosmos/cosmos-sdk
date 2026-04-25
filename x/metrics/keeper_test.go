package metrics

import (
	"strings"
	"testing"
	"time"
)

func TestNewKeeper(t *testing.T) {
	cfg := DefaultMetricsConfig()
	k := NewKeeper(cfg)
	if k == nil {
		t.Fatal("expected non-nil keeper")
	}
	if k.config.Namespace != "cosmos" {
		t.Errorf("expected namespace cosmos, got %s", k.config.Namespace)
	}
	if k.registry == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestRecordTransaction(t *testing.T) {
	k := NewKeeper(DefaultMetricsConfig())
	k.RecordTransaction("send", 50000)
	k.RecordTransaction("send", 30000)
	k.RecordTransaction("delegate", 70000)

	if k.txCounts["send"] != 2 {
		t.Errorf("expected 2 send txs, got %v", k.txCounts["send"])
	}
	if k.gasUsed["send"] != 80000 {
		t.Errorf("expected 80000 gas for send, got %v", k.gasUsed["send"])
	}
	if k.txCounts["delegate"] != 1 {
		t.Errorf("expected 1 delegate tx, got %v", k.txCounts["delegate"])
	}
}

func TestRecordBlockTime(t *testing.T) {
	k := NewKeeper(DefaultMetricsConfig())
	k.RecordBlockTime(2 * time.Second)

	if k.blockTime != 2.0 {
		t.Errorf("expected block time 2.0, got %v", k.blockTime)
	}
}

func TestRecordModuleCall(t *testing.T) {
	k := NewKeeper(DefaultMetricsConfig())
	k.RecordModuleCall("bank", "Send", 100*time.Millisecond)
	k.RecordModuleCall("bank", "Send", 200*time.Millisecond)

	key := "bank.Send"
	if k.moduleCalls[key] != 2 {
		t.Errorf("expected 2 calls, got %v", k.moduleCalls[key])
	}
	if k.moduleDurations[key] != 0.3 {
		t.Errorf("expected 0.3s duration, got %v", k.moduleDurations[key])
	}
}

func TestGetMetrics(t *testing.T) {
	k := NewKeeper(DefaultMetricsConfig())
	k.RecordTransaction("send", 1000)
	k.RecordBlockTime(time.Second)

	metrics := k.GetMetrics()
	if len(metrics) < 3 {
		t.Errorf("expected at least 3 metrics, got %d", len(metrics))
	}
}

func TestFormatPrometheus(t *testing.T) {
	k := NewKeeper(DefaultMetricsConfig())
	k.RecordTransaction("send", 5000)
	k.RecordBlockTime(time.Second)

	output := k.FormatPrometheus()
	if !strings.Contains(output, "cosmos_sdk_tx_count") {
		t.Errorf("expected prometheus output to contain cosmos_sdk_tx_count, got:\n%s", output)
	}
	if !strings.Contains(output, "# TYPE") {
		t.Errorf("expected prometheus TYPE annotations, got:\n%s", output)
	}
}

// staticCollector is a test helper implementing MetricCollector.
type staticCollector struct{ m []Metric }

func (sc *staticCollector) Collect() []Metric { return sc.m }

func TestMetricRegistry(t *testing.T) {
	reg := NewMetricRegistry()

	sc := &staticCollector{m: []Metric{{Name: "custom", Type: Gauge, Value: 42}}}
	reg.Register("static", sc)

	all := reg.CollectAll()
	if len(all) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(all))
	}
	if all[0].Value != 42 {
		t.Errorf("expected value 42, got %v", all[0].Value)
	}

	reg.Unregister("static")
	all = reg.CollectAll()
	if len(all) != 0 {
		t.Errorf("expected 0 metrics after unregister, got %d", len(all))
	}
}

func TestReset(t *testing.T) {
	k := NewKeeper(DefaultMetricsConfig())
	k.RecordTransaction("send", 1000)
	k.RecordBlockTime(time.Second)
	k.RecordModuleCall("bank", "Send", time.Millisecond)

	k.Reset()

	if len(k.txCounts) != 0 {
		t.Errorf("expected txCounts to be empty after reset")
	}
	if len(k.gasUsed) != 0 {
		t.Errorf("expected gasUsed to be empty after reset")
	}
	if k.blockTime != 0 {
		t.Errorf("expected blockTime to be 0 after reset")
	}
	if len(k.moduleCalls) != 0 {
		t.Errorf("expected moduleCalls to be empty after reset")
	}
}
