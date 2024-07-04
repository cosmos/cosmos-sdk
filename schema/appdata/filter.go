package appdata

// ModuleFilter returns an updated listener that filters state updates based on the module name.
func ModuleFilter(listener Listener, filter func(moduleName string) bool) Listener {
	if initModData := listener.InitializeModuleData; initModData != nil {
		listener.InitializeModuleData = func(data ModuleInitializationData) error {
			if !filter(data.ModuleName) {
				return nil
			}

			return initModData(data)
		}
	}

	if onKVPair := listener.OnKVPair; onKVPair != nil {
		listener.OnKVPair = func(data KVPairData) error {
			for _, update := range data.Updates {
				if !filter(update.ModuleName) {
					continue
				}

				if err := onKVPair(KVPairData{Updates: []ModuleKVPairUpdate{update}}); err != nil {
					return err
				}
			}

			return nil
		}
	}

	if onObjectUpdate := listener.OnObjectUpdate; onObjectUpdate != nil {
		listener.OnObjectUpdate = func(data ObjectUpdateData) error {
			if !filter(data.ModuleName) {
				return nil
			}

			return onObjectUpdate(data)
		}
	}

	return listener
}
