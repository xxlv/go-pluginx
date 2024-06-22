# How it works ?

## Builder

1. Create a Plugin Builder

```go
builder = &plugins.PluginBuilder{
	Logger: slog.Default(),
}

```

2. Build the Go Plugin

```go
pluginBuilder.Build(context.TODO,payload, func(pi plugins.PluginInfo) {
    // when finish
})
```

The `payload` structure is defined as follows. If code is provided, it takes precedence.

```go
type BuildPayload struct {
	Tenant string `json:"tenant"`
	Id     string `json:"id"`
	Title  string `json:"title" title:"Your tool title"`
	Code   string `json:"code" title:"Upload pure go code which in main package!"`
	SoFile string `json:"soFile"`
	Ext    any    `json:"ext"`
}
```

## Multi-Tenant Plugin Loader

```go
    options := engine.Options{
            Tenants: []engine.TenantOptions{
                {TenantID: "tenantA"},
            },
    }
	engine.NewLoader(options).Run()

```
