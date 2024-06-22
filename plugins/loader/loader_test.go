package loader

import (
	"encoding/json"

	"log/slog"
	"os"
	"testing"

	"github.com/xxlv/go-pluginx/plugins"
	"github.com/xxlv/go-pluginx/storage/storages"
)

func TestLoader_Run(t *testing.T) {
	options := Options{
		Tenants: []TenantOptions{
			{
				TenantID: "tenant1",
			},
			{
				TenantID: "tenant2",
			},
		},
	}
	loader := NewLoader(options)
	loader.Run(func(pi *plugins.PluginInfo, err error) {})
	if len(plugins.GlobalTenantPluginManager.ManagerMap) != 2 {
		t.Error("Expected 2 registered plugin managers, but got a different count")
	}

}

func Test_checkAndLoad(t *testing.T) {
	options := Options{
		Logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
		Tenants: []TenantOptions{
			{TenantID: "tenanta"},
			{TenantID: "tenantb"},
			{TenantID: "tenantc"},
		},
	}
	ld := NewLoader(options)
	pm := plugins.GlobalTenantPluginManager.RegPluginManager("tenant1")
	pm.Storage = &storages.MemoryStorage{}

	v := &plugins.PluginInfo{
		Installed: false,
		Id:        "1",
		Name:      "Plugin 1",
		Code: `package main
		
		import "fmt"
		
		func main() {
			fmt.Println("Hello, World!")
		}`,
	}

	result := ld.checkAndLoad(pm, v, func(pi *plugins.PluginInfo, err error) {})

	if !result {
		t.Errorf("Expected checkAndLoad to return true, but got false")
	}

	if !v.Installed {
		t.Errorf("Expected Installed field to be true, but got false")
	}

	var err error
	kv, _ := pm.Storage.Get("1")
	var storedPluginInfo plugins.PluginInfo
	err = json.Unmarshal([]byte(kv), &storedPluginInfo)
	if err != nil {
		t.Error(err)
	}
	if storedPluginInfo != *v {
		t.Errorf("Expected plugin info to be stored correctly, but got different values")
	}

	kv, _ = pm.Storage.Get(plugins.MetadataPliuginsKeyFmt.Key(pm.Tenant))
	var storedPluginIDs []string
	err = json.Unmarshal([]byte(kv), &storedPluginIDs)
	if err != nil {
		t.Error(err)
	}
	found := false
	for _, id := range storedPluginIDs {
		if id == "1" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected plugin ID to be added to stored plugin IDs, but not found")
	}

	v = &plugins.PluginInfo{
		Installed: false,
		Id:        "1",
		Name:      "Plugin 1",
		Code: `bad code package main
		
		import "fmt"
		
		func main() {
			fmt.Println("Hello, World!)
		}`,
	}
	result = ld.checkAndLoad(pm, v, func(pi *plugins.PluginInfo, err error) {})

	if result == true {
		t.Error("bad code success load")
	}

	if !v.Installed {
		t.Error("bad code also need mark as installed")
	}

}
