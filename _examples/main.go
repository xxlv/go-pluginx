package main

import (
	"context"
	"fmt"
	"log/slog"

	msgchannel "github.com/xxlv/go-pluginx/build"
	"github.com/xxlv/go-pluginx/plugins"
	"github.com/xxlv/go-pluginx/storage/storages"
)

func main() {

	plugins.GlobalTenantPluginManager.RegPluginManager("tenantA")

	buildStatus := make(chan byte)
	builder := &plugins.PluginBuilder{
		Storage: &storages.MemoryStorage{},
		Logger:  slog.Default(),
	}
	code := `package main
	
		type Hello struct{}

		var PluginSymbol=Hello{}
	`
	r, err := builder.BuildFromCode(context.Background(), "buildId", &plugins.BuildPayload{
		Tenant: "tenantA",
		Id:     "uuid",
		Title:  "this is your plugin title",
		Code:   code,
	}, msgchannel.NilMessageChannel, func(pi *plugins.PluginInfo, err error) {
		println(pi.Id + " installed")
		buildStatus <- 0
	})
	if err != nil {
		panic(err)
	}

	<-buildStatus

	println("new build started" + r.BuildId)
	for _, v := range plugins.GlobalTenantPluginManager.GetPluginManager("tenantA").Plugins {
		fmt.Printf("%+v", v)
	}
}
