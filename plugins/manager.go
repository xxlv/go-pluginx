package plugins

import (
	"encoding/json"
	"fmt"

	"log/slog"
	"net/http"
	"plugin"
	"sync"
	"time"

	msgchannel "github.com/xxlv/go-pluginx/build"
	"github.com/xxlv/go-pluginx/storage"
)

var GlobalTenantPluginManager = NewTenantPluginManager()

// Tenant plugin manager
// each tenant has its own plugin manager
type TenantPluginManager struct {
	ManagerMap map[string]*PluginManager
	mu         sync.Mutex
	Logger     *slog.Logger
}

func NewTenantPluginManager() *TenantPluginManager {
	tpm := &TenantPluginManager{
		ManagerMap: make(map[string]*PluginManager),
		Logger:     slog.New(slog.Default().Handler()), // default(for test), change later if set
	}
	return tpm
}

func (tpm *TenantPluginManager) SetLogger(logger *slog.Logger) {
	tpm.Logger = logger
}

func (tpm *TenantPluginManager) GetPluginManager(tenant string) *PluginManager {
	tpm.mu.Lock()
	defer tpm.mu.Unlock()
	manager, ok := tpm.ManagerMap[tenant]
	if !ok {
		return nil
	}
	return manager
}

func (tpm *TenantPluginManager) RegPluginManager(tenant string) *PluginManager {
	tpm.mu.Lock()
	defer tpm.mu.Unlock()
	manager, ok := tpm.ManagerMap[tenant]
	if !ok {
		manager = &PluginManager{
			Tenant: tenant,
			// Storage: storages.NewDefaultRedisxStorage(), // default
			Logger: tpm.Logger,
		}
		tpm.ManagerMap[tenant] = manager
	}
	return manager
}

// plugin manager control all plugins
// plugin save to Plugins map and each *PluginInfo contains `Symbol“ or `Plugin` key
type PluginManager struct {
	sync.Mutex
	Logger  *slog.Logger
	Tenant  string
	Storage storage.Storage
	Plugins map[string]*PluginInfo
}

// TODO: check test
func (pm *PluginManager) CheckIfLoad(id string) bool {
	if _, ok := pm.Plugins[id]; ok {
		return true
	}
	return false
}

func (pm *PluginManager) SetStorage(s storage.Storage) {
	if s != nil {
		pm.Logger.Warn("new storage has been set!")
		pm.Storage = s
	}
}

// RegisterPlugin.
// Installed , customized config value for checking your plugin is installed or not
func (pm *PluginManager) RegisterPlugin(bp *BuildPayload, symbol plugin.Symbol, plugin *plugin.Plugin, err error, installed bool, onFinish func(pi *PluginInfo, err error)) {
	id := bp.Id
	filepath := bp.PluginFile
	name := bp.Title
	code := bp.Code

	if err == nil {
		if pm.Plugins == nil {
			pm.Plugins = make(map[string]*PluginInfo)
		}
		pm.Logger.Info("RegisterPlugin", "id", id)
		pi := &PluginInfo{
			Tenant:         pm.Tenant,
			Id:             id,
			SoFilepath:     filepath,
			Symbol:         symbol,
			Plugin:         plugin,
			RegisteredTime: time.Now(),
			Name:           name,
			Code:           code,
			Error:          err,
			Installed:      installed,
			Ext:            bp.Ext,
		}
		pm.Plugins[id] = pi
		onFinish(pi, nil)
	} else {
		onFinish(nil, err)
	}
}

func (tm *PluginManager) Sse(w http.ResponseWriter, r *http.Request) {
	buildId := r.URL.Query().Get("buildId")
	mc, ok := msgchannel.GetChannel(buildId)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	if ok {
		for {
			msg := <-mc
			if msg != "" {
				fmt.Fprintf(w, "data: %s\n\n", msg)
				w.(http.Flusher).Flush()
			} else {
				time.Sleep(1 * time.Second)
			}

		}
	} else {
		fmt.Fprintf(w, "data: Build %s not found\n\n", buildId)
		w.(http.Flusher).Flush()
	}

}

func (pm *PluginManager) Handle() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(pm.Plugins)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type PluginInfo struct {
	Tenant         string
	Id             string
	Name           string
	SoFilepath     string
	Code           string
	Plugin         *plugin.Plugin `json:"-"` // Plugin obj
	Symbol         plugin.Symbol  `json:"-"` // if has
	Installed      bool
	Accessed       bool
	RegisteredTime time.Time
	Error          error `json:"-"` // some error
	Ext            any   // any config data
}

func (p *PluginInfo) AsJson() string {
	b, err := json.Marshal(p)
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return string(b)
}
