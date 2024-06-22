package loader

import (
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	msgchannel "github.com/xxlv/go-pluginx/build"
	plugins "github.com/xxlv/go-pluginx/plugins"
	storage "github.com/xxlv/go-pluginx/storage"
	utils "github.com/xxlv/go-pluginx/utils"
)

type Loader struct {
	Options Options
	Logger  *slog.Logger
}

type TenantOptions struct {
	TenantID string
}

type Options struct {
	Tenants []TenantOptions
	Logger  *slog.Logger
}

func NewLoader(options Options) *Loader {
	if options.Logger == nil {
		options.Logger = slog.New(slog.Default().Handler())
	}
	return &Loader{
		Options: options,
		Logger:  options.Logger,
	}
}

func (ld *Loader) Run(onFinish func(pi *plugins.PluginInfo, err error)) {
	for _, v := range ld.Options.Tenants {
		// init every plugin manager by RegPluginManager
		plugins.GlobalTenantPluginManager.RegPluginManager(v.TenantID)
	}
	ld.load(onFinish)
}

// Start a timer
// Load plugins from metadata
// Detect new plugins and load them
// Write the plugins to metadata
func (ld *Loader) load(onFinish func(pi *plugins.PluginInfo, err error)) {
	for k, p := range plugins.GlobalTenantPluginManager.ManagerMap {
		ld.Logger.Debug("Plugin Init", "tenant", k)
		pm := p
		ld.loadOthers(pm, true)
		ticker := time.NewTicker(3 * time.Second)
		loadCurrentMachine := func() {
			anyinstalled := false
			for _, v := range pm.Plugins {
				vv := v
				if !vv.Installed {
					anyinstalled = ld.checkAndLoad(pm, vv, onFinish)
				}
			}
			if anyinstalled {
				ld.Logger.Info("some plugins loaded", "tenant", k)
			}
		}
		go func(pm *plugins.PluginManager) {
			// prepare wait
			time.Sleep(5 * time.Second)

			for range ticker.C {
				ld.loadOthers(pm, false)
				loadCurrentMachine()
			}
		}(pm)
	}

}

// Install new plugin to tm storage
func (ld *Loader) checkAndLoad(pm *plugins.PluginManager, v *plugins.PluginInfo, onFinish func(pi *plugins.PluginInfo, err error)) bool {
	v.Installed = true

	ld.Logger.Info("load plugin", "id", v.Id)
	id := strings.ReplaceAll(v.Id, "\n", "-")
	id = strings.ReplaceAll(id, " ", "_")
	if id == "" {
		ld.Logger.Error("miss plugin id", "id", v.Id)
		return false
	}
	_ = pm.Storage.Store(storage.Kv{Key: id, Value: v.AsJson()})
	exists, _ := pm.Storage.Get(plugins.MetadataPliuginsKeyFmt.Key(pm.Tenant))
	var existIds []string = make([]string, 0)
	if exists != "" {
		err := json.Unmarshal([]byte(exists), &existIds)
		if err != nil {
			ld.Logger.Error("unmarshal json data from driver fail", "error", err)
		}
	}
	existIds = append(existIds, id)
	existIds = utils.RemoveDuplicates(existIds)
	data, _ := json.Marshal(existIds)
	listdata := string(data)

	ld.Logger.Info("write to storage with kv key:", "key", plugins.MetadataPliuginsKeyFmt, "value", listdata)
	_ = pm.Storage.Store(storage.Kv{Key: plugins.MetadataPliuginsKeyFmt.Key(pm.Tenant), Value: listdata})
	bp := asBuildPayload(v)

	err := plugins.LoadSoAndRegPlugin(pm, bp, msgchannel.NilMessageChannel, v.Installed, onFinish)
	if err != nil {
		ld.Logger.Error("failed load so file with error", "error", err)
		return false
	}

	return true
}

func asBuildPayload(v *plugins.PluginInfo) *plugins.BuildPayload {
	bp := &plugins.BuildPayload{
		Tenant:     v.Tenant,
		Id:         v.Id,
		PluginFile: v.SoFilepath,
		Code:       v.Code,
		Title:      v.Name,
		Ext:        v.Ext,
	}
	return bp
}

func (ld *Loader) loadOthers(pm *plugins.PluginManager, forceInstall bool) {
	defer func() {
		err := recover()
		if err != nil {
			ld.Logger.Error("faild load plugin from local with error", "tenant", pm.Tenant, "error", err)
		}
	}()
	localKeys, err := pm.Storage.Get(plugins.MetadataPliuginsKeyFmt.Key(pm.Tenant))
	if err != nil && err.Error() != "not found" {
		ld.Logger.Error("local from storage field with error", "error", err, "tenant", pm.Tenant)
	}
	if localKeys != "" && len(strings.Split(localKeys, ",")) < len(pm.Plugins) {
		ld.Logger.Info("local from storage with keys", "localKeys", localKeys, "tenant", pm.Tenant)
	}
	var result []string = make([]string, 0)
	err = json.Unmarshal([]byte(localKeys), &result)

	if err != nil && localKeys != "" {
		ld.Logger.Error("can not process local plugins,err", "tenant", pm.Tenant, "error", err)
	}

	for _, v := range result {
		r, err := pm.Storage.Get(v)
		if err != nil {
			if err.Error() != "not found" {
				ld.Logger.Error("fail to build plugin with err", "tenant", pm.Tenant, "error", err)
			}
			continue
		} else {
			var pinfo plugins.PluginInfo
			err := json.Unmarshal([]byte(r), &pinfo)
			if err != nil {
				ld.Logger.Error("failed unmarshal plugin info with err", "tenant", pm.Tenant, "error", err)
			} else {
				if forceInstall || !pm.CheckIfLoad(pinfo.Id) {
					bp := asBuildPayload(&pinfo)

					ld.Logger.Info("Load from storage:: load so file", "tenant", pm.Tenant, "sofile", pinfo.SoFilepath, "code", pinfo.Code)
					plugins.LoadSoAndRegPlugin(pm, bp, msgchannel.NilMessageChannel, true, func(pi *plugins.PluginInfo, err error) {})
				}
			}
		}
	}

}
