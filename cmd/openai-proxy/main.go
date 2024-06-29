package main

import (
	"butterfly.orx.me/core"
	"butterfly.orx.me/core/app"
	"github.com/orvice/openapi-proxy/internal/handler"
)

func main() {
	handler.Init()
	//err := os.Setenv("PORT", "18788")
	//if err != nil {
	//	panic(err)
	//}
	app := core.New(&app.Config{
		Service: "openai-proxy",
		Router:  handler.Router,
	})
	app.Run()
}
