package appdata

import "testing"

func TestListenerMux(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		res := ListenerMux(Listener{}, Listener{})
		if res.InitializeModuleData != nil {
			t.Error("expected nil")
		}
		if res.StartBlock != nil {
			t.Error("expected nil")
		}
		if res.OnTx != nil {
			t.Error("expected nil")
		}
		if res.OnEvent != nil {
			t.Error("expected nil")
		}
		if res.OnKVPair != nil {
			t.Error("expected nil")
		}
		if res.OnObjectUpdate != nil {
			t.Error("expected nil")
		}
		if res.Commit != nil {
			t.Error("expected nil")
		}
	})

	t.Run("all called once", func(t *testing.T) {
		var calls []string
		res := ListenerMux(Listener{
			InitializeModuleData: func(ModuleInitializationData) error {
				calls = append(calls, "InitializeModuleData 1")
				return nil
			},
			StartBlock: func(StartBlockData) error {
				calls = append(calls, "StartBlock 1")
				return nil
			},
			OnTx: func(TxData) error {
				calls = append(calls, "OnTx 1")
				return nil
			},
			OnEvent: func(EventData) error {
				calls = append(calls, "OnEvent 1")
				return nil
			},
			OnKVPair: func(KVPairData) error {
				calls = append(calls, "OnKVPair 1")
				return nil
			},
			OnObjectUpdate: func(ObjectUpdateData) error {
				calls = append(calls, "OnObjectUpdate 1")
				return nil
			},
			Commit: func(CommitData) error {
				calls = append(calls, "Commit 1")
				return nil
			},
		}, Listener{
			InitializeModuleData: func(ModuleInitializationData) error {
				calls = append(calls, "InitializeModuleData 2")
				return nil
			},
			StartBlock: func(StartBlockData) error {
				calls = append(calls, "StartBlock 2")
				return nil
			},
			OnTx: func(TxData) error {
				calls = append(calls, "OnTx 2")
				return nil
			},
			OnEvent: func(EventData) error {
				calls = append(calls, "OnEvent 2")
				return nil
			},
			OnKVPair: func(KVPairData) error {
				calls = append(calls, "OnKVPair 2")
				return nil
			},
			OnObjectUpdate: func(ObjectUpdateData) error {
				calls = append(calls, "OnObjectUpdate 2")
				return nil
			},
			Commit: func(CommitData) error {
				calls = append(calls, "Commit 2")
				return nil
			},
		})

		if err := res.InitializeModuleData(ModuleInitializationData{}); err != nil {
			t.Error(err)
		}
		if err := res.StartBlock(StartBlockData{}); err != nil {
			t.Error(err)
		}
		if err := res.OnTx(TxData{}); err != nil {
			t.Error(err)
		}
		if err := res.OnEvent(EventData{}); err != nil {
			t.Error(err)
		}
		if err := res.OnKVPair(KVPairData{}); err != nil {
			t.Error(err)
		}
		if err := res.OnObjectUpdate(ObjectUpdateData{}); err != nil {
			t.Error(err)
		}
		if err := res.Commit(CommitData{}); err != nil {
			t.Error(err)
		}

		expected := []string{
			"InitializeModuleData 1",
			"InitializeModuleData 2",
			"StartBlock 1",
			"StartBlock 2",
			"OnTx 1",
			"OnTx 2",
			"OnEvent 1",
			"OnEvent 2",
			"OnKVPair 1",
			"OnKVPair 2",
			"OnObjectUpdate 1",
			"OnObjectUpdate 2",
			"Commit 1",
			"Commit 2",
		}

		if len(calls) != len(expected) {
			t.Fatalf("expected %d calls, got %d", len(expected), len(calls))
		}

		for i := range calls {
			if calls[i] != expected[i] {
				t.Errorf("expected %q, got %q", expected[i], calls[i])
			}
		}
	})
}
