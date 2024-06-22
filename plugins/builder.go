package plugins

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	build "github.com/xxlv/go-pluginx/build"
	storage "github.com/xxlv/go-pluginx/storage"
	"github.com/xxlv/go-pluginx/storage/storages"
	utils "github.com/xxlv/go-pluginx/utils"
)

const (
	Success  BuildResultStatus = "SUCCESS"
	Running  BuildResultStatus = "RUNNING"
	Fail     BuildResultStatus = "FAIL"
	NotFound BuildResultStatus = "NOT_FOUND"
)

type BuildResultStatus string

type BuildPayload struct {
	Tenant     string `json:"tenant"`
	Id         string `json:"id"`
	Title      string `json:"title" title:"Your tool title"`
	Code       string `json:"code" title:"Pure go code"`
	PluginFile string `json:"plugin_file"`
	Ext        any    `json:"ext"`
}

type BuildResult struct {
	BuildId string `json:"build_id"`
}

type BuildResultInfo struct {
	Status    BuildResultStatus
	MachineIp string
	BuildId   string
	OpTime    time.Time
}

type PluginBuilder struct {
	Logger   *slog.Logger
	Storage  storage.Storage
	BuildLog build.BuildLog
}

func NewDefaultPluginBuilder() *PluginBuilder {
	return &PluginBuilder{
		Storage:  &storages.MemoryStorage{},
		Logger:   slog.Default(),
		BuildLog: &build.BuildQueue{},
	}
}

func (pb *PluginBuilder) Build(ctx context.Context, bp *BuildPayload, onFinish func(pi *PluginInfo, err error)) (*BuildResult, error) {
	buildId := fmt.Sprintf("%d-%d", time.Now().Unix(), rand.Intn(1000))
	mc := build.NewChannel(buildId)
	// first choice:  use so file if provide
	if bp.PluginFile != "" {
		return pb.BuildFromFile(ctx, buildId, bp, mc, onFinish)
	} else if bp.Code != "" {
		// if so file does not provice, compile from code
		return pb.BuildFromCode(ctx, buildId, bp, mc, onFinish)
	}

	return &BuildResult{BuildId: ""}, fmt.Errorf("not find code or so file")
}

// TODO:(xxlv) fix onFinish
func (pb *PluginBuilder) BuildFromFile(ctx context.Context, buildId string, bp *BuildPayload, mc build.MessageChannel, onFinish func(pi *PluginInfo, err error)) (build *BuildResult, err error) {
	base64Code := bp.Code
	decoded, _ := base64.StdEncoding.DecodeString(base64Code)
	code := string(decoded)
	pb.setBuildStatus(ctx, buildId, Running)
	bp.Code = code
	go func() {
		defer func() {
			panicErr := recover()
			if panicErr != nil && panicErr.(error) != nil {
				err = panicErr.(error)
			}
		}()
		err := pb.ProcessSofile(ctx, buildId, bp, mc, onFinish)
		if err != nil {
			pb.Logger.ErrorContext(ctx, "can not process so file", "so", bp.PluginFile, "err", err)
		}
	}()
	return &BuildResult{BuildId: buildId}, nil
}

func (pb *PluginBuilder) BuildFromCode(ctx context.Context, buildId string, bp *BuildPayload, mc build.MessageChannel, onFinish func(pi *PluginInfo, err error)) (*BuildResult, error) {
	code := string(bp.Code)
	pb.setBuildStatus(ctx, buildId, Running)
	bp.Code = code

	go func() {
		err := pb.ProcessPureFunc(ctx, buildId, bp, mc, onFinish)
		if err != nil {
			pb.Logger.ErrorContext(ctx, "can not ProcessPureFunc", "error", err, "buildId", buildId)
		}
	}()

	return &BuildResult{BuildId: buildId}, nil
}

func (pb *PluginBuilder) ProcessPureFunc(ctx context.Context, buildId string, bp *BuildPayload, mc build.MessageChannel, onFinish func(pi *PluginInfo, err error)) error {
	soFilePath, err := CompileToSO(bp.Code, mc)
	if err != nil {
		pb.setBuildStatus(ctx, buildId, Fail)
		return err
	}
	bp.PluginFile = soFilePath
	return pb.ProcessSofile(ctx, buildId, bp, mc, onFinish)
}

func (pb *PluginBuilder) ProcessSofile(ctx context.Context, buildId string, bp *BuildPayload, mc build.MessageChannel, onFinish func(pi *PluginInfo, err error)) error {
	pm := GlobalTenantPluginManager.GetPluginManager(bp.Tenant)
	if pm == nil {
		return fmt.Errorf("plugin manager is nil, tenant is %s", bp.Tenant)
	}

	soFilePath := bp.PluginFile
	mc <- "prepare load so file  " + bp.PluginFile + " id#" + bp.Id

	err := LoadSoAndRegPlugin(pm, bp, mc, false, onFinish)
	if err != nil {
		mc <- "load plugin error " + err.Error()
		mc <- " 😭😭 finish this job, and got error,please check your code or so file"
		pb.setBuildStatus(ctx, buildId, Fail)
	} else {
		mc <- "system load this [" + soFilePath + "] " + "success"
		mc <- " 🎉🎉 clear all data 🎉🎉"
		pb.setBuildStatus(ctx, buildId, Success)
	}
	build.DeleteChannel(buildId)

	return nil
}

func (pb *PluginBuilder) setBuildStatus(ctx context.Context, buildId string, status BuildResultStatus) {
	clientIp := utils.GetLocalIP()
	buildInfo := BuildResultInfo{
		Status:    status,
		BuildId:   buildId,
		MachineIp: clientIp,
		OpTime:    time.Now(),
	}
	result := &[]*BuildResultInfo{}
	found := false
	v, err := pb.Storage.Get(BuildInfoPrefix + buildId)
	if err != nil {
		pb.Logger.ErrorContext(ctx, "can not get build info", "error", err)
	} else {
		err = json.Unmarshal([]byte(v), result)
		if err == nil {
			for _, v := range *result {
				if v.MachineIp == clientIp {
					pb.Logger.InfoContext(ctx, "set build status found old client", "clientIp", clientIp)
					v.Status = status
					v.OpTime = time.Now()
					found = true
				}
			}
		} else {
			pb.Logger.ErrorContext(ctx, "json Unmarshal build result", "error", err)
		}
	}
	if !found {
		*result = append(*result, &buildInfo)
	}
	r, err := json.Marshal(result)
	if err != nil {
		pb.Logger.ErrorContext(ctx, "can not serialized build info", "error", err)
	} else {
		buildInfo := string(r)
		pb.Logger.InfoContext(ctx, "write build info", "buildInfo", buildInfo, "buildId", buildId)
		err = pb.Storage.Store(storage.Kv{
			Key:   BuildInfoPrefix + buildId,
			Value: buildInfo,
		})
		if err != nil {
			pb.Logger.ErrorContext(ctx, "store plugin build info failed", "err", err)
		}
	}
}

func (pb *PluginBuilder) GetResult(ctx context.Context, buildId string) ([]BuildResultInfo, error) {
	v, err := pb.Storage.Get(BuildInfoPrefix + buildId)
	if err != nil {
		return nil, err
	}
	buildInfo := &[]BuildResultInfo{}
	err = json.Unmarshal([]byte(v), buildInfo)
	if err != nil {
		return nil, err
	}
	return *buildInfo, err
}

func (pb *PluginBuilder) SseHandle(w http.ResponseWriter, r *http.Request) {

	buildId := r.URL.Query().Get("buildId")
	mc, ok := build.GetChannel(buildId)
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

func checkSoOrRecompile(soFilePath, code string, mc build.MessageChannel, logger *slog.Logger) string {
	// if file is local and does not exists,
	// this program will build new so file from code by invoke `CompileToSo`
	if filepath.IsAbs(soFilePath) {
		_, err := os.Stat(soFilePath)
		if os.IsNotExist(err) {
			logger.Info("prepare recompile code", "code", code)
			soFilePath, err = CompileToSO(code, mc)
			if err != nil {
				logger.Error("can not compile from code", "error", err)
				return ""
			}
		}
	} else if soFilePath == "" && code != "" {
		logger.Info("soFilePath is empty and find code, starting compile", "code", code)
		var err error
		soFilePath, err = CompileToSO(code, mc)
		if err != nil {
			logger.Error("can not compile directly from code", "error", err)
			return ""
		}
	}
	logger.Info("success get sofilepath", "soFilePath", soFilePath)

	return soFilePath
}

func loadFileFromUrl(soFilePath string) (string, error) {
	if strings.HasPrefix(soFilePath, "http://") || strings.HasPrefix(soFilePath, "https://") {
		resp, err := http.Get(soFilePath)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		tmpFile, err := os.CreateTemp(os.TempDir(), fmt.Sprintf("sofile-%s", time.Now().String()))
		if err != nil {
			return "", err
		}
		_, err = io.Copy(tmpFile, resp.Body)
		if err != nil {
			return "", err
		}
		if err := tmpFile.Close(); err != nil {
			return "", err
		}
		soFilePath = tmpFile.Name()
	}
	return soFilePath, nil
}

func LoadSoAndRegPlugin(pm *PluginManager, bp *BuildPayload, mc build.MessageChannel, install bool, onFinish func(pi *PluginInfo, err error)) error {
	soFilePath := bp.PluginFile
	code := bp.Code
	id := bp.Id
	name := bp.Title

	soFilePathLocal, err := loadFileFromUrl(soFilePath)
	if err != nil {
		return err
	}
	if !utils.CheckFileExists(soFilePathLocal) {
		soFilePathLocal = "" //remove this file
		pm.Logger.Info("so file does not exists", "soFilePath", soFilePath)
	}
	// check or re-compile code to sofile
	soFilePathLocal = checkSoOrRecompile(soFilePathLocal, code, mc, pm.Logger)

	// TODO:(xxlv) dont remove
	// defer os.Remove(soFilePathLocal)

	_, err = os.Stat(soFilePathLocal)
	if os.IsNotExist(err) {
		// File doesn't exist
		pm.Logger.Error("so file does not exists", "error", err)
		mc <- err.Error()
		return err
	} else {
		pm.Logger.Info("prepare open plugin", "id", id, "name", name, "sofile", soFilePathLocal, "install", install)
		// File exists, open it
		p, err := plugin.Open(soFilePathLocal)
		if err != nil {
			// Error occurred while opening the file
			// Handle the error or take appropriate action
			pm.Logger.Error("faild load plugin with error", "error", err)
			mc <- err.Error()
			mc <- "please check so file "
			pm.RegisterPlugin(bp, nil, p, err, install, onFinish)
			return err
		} else {
			pm.Logger.Info("reg plugin success with sofile", "sofile", soFilePath)
			mc <- fmt.Sprintf("reg plugin success with sofile %s", soFilePath)
			pm.RegisterPlugin(bp, nil, p, nil, install, onFinish)
		}
	}

	return nil
}
