package main

import (
	"butterfly.orx.me/core"
	"butterfly.orx.me/core/app"
	"github.com/orvice/openapi-proxy/internal/handler"
)

func main() {
	handler.Init()
	proxyAPP := core.New(&app.Config{
		Service: "openai-proxy",
		Router:  handler.Router,
	})
	proxyAPP.Run()
}
