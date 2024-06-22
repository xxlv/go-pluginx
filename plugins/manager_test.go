package plugins

import (
	"log/slog"
	"testing"
	"time"
)

func TestTenantPluginManager(t *testing.T) {
	tpm := NewTenantPluginManager()

	t.Run("GetPluginManager", func(t *testing.T) {
		manager := tpm.GetPluginManager("nonexistent")
		if manager != nil {
			t.Errorf("Expected manager to be nil, got %v", manager)
		}
		manager = tpm.RegPluginManager("existing")
		if manager == nil {
			t.Errorf("Expected manager to be not nil")
		}

		manager2 := tpm.GetPluginManager("existing")
		if manager != manager2 {
			t.Errorf("Expected manager and manager2 to be equal")
		}
	})

	t.Run("RegPluginManager", func(t *testing.T) {
		manager := tpm.RegPluginManager("new")
		if manager == nil {
			t.Errorf("Expected manager to be not nil")
		}

		manager2 := tpm.RegPluginManager("new")
		if manager != manager2 {
			t.Errorf("Expected manager and manager2 to be equal")
		}
	})

}

func TestPluginManager_RegisterPlugin(t *testing.T) {
	pm := &PluginManager{
		Tenant:  "test",
		Storage: nil,
		Plugins: nil,
		Logger:  slog.Default(),
	}
	bp := &BuildPayload{
		Id:         "plugin1",
		Title:      "Plugin 1",
		PluginFile: "/path/to/plugin1.so",
		Code:       "plugin1_code",
	}
	pm.RegisterPlugin(bp, nil, nil, nil, false, func(pi *PluginInfo, err error) {})

	if !pm.CheckIfLoad("plugin1") {
		t.Error("need load but not load")
	}

	expectedPlugin := &PluginInfo{
		Tenant:         "test",
		Id:             "plugin1",
		SoFilepath:     "/path/to/plugin1.so",
		Symbol:         nil,
		Plugin:         nil,
		RegisteredTime: time.Now(),
		Name:           "Plugin 1",
		Code:           "plugin1_code",
	}
	actualPlugin, ok := pm.Plugins["plugin1"]
	if !ok {
		t.Errorf("Expected plugin1 to be registered")
	} else {
		if actualPlugin.Tenant != expectedPlugin.Tenant ||
			actualPlugin.Id != expectedPlugin.Id ||
			actualPlugin.SoFilepath != expectedPlugin.SoFilepath ||
			actualPlugin.Symbol != expectedPlugin.Symbol ||
			actualPlugin.Plugin != expectedPlugin.Plugin ||
			actualPlugin.Name != expectedPlugin.Name ||
			actualPlugin.Code != expectedPlugin.Code {
			t.Errorf("Expected plugin1 details to be %v, got %v", expectedPlugin, actualPlugin)
		}
		// reg is not load
		// load need write info to storage
		// so this installed must be false
		if actualPlugin.Installed == true {
			t.Errorf("Expected plugin1 installed is false ,but got true")
		}
	}
}
