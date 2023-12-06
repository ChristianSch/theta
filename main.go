package main

import (
	"github.com/ChristianSch/Theta/adapters/inbound"
	"github.com/ChristianSch/Theta/adapters/outbound"
	outboundPorts "github.com/ChristianSch/Theta/domain/ports/outbound"
	"github.com/gofiber/fiber/v2"
)

func main() {
	log := outbound.NewZapLogger(outbound.ZapLoggerConfig{Debug: true})
	web := inbound.NewFiberWebServer(inbound.FiberWebServerConfig{
		Port:                8080,
		TemplatesPath:       "./infrastructure/views",
		TemplatesExtension:  ".gohtml",
		StaticResourcesPath: "./infrastructure/static",
	}, inbound.FiberWebServerAdapters{Log: log})

	web.AddRoute("GET", "/", func(ctx interface{}) error {
		log.Debug("handling request", outboundPorts.LogField{Key: "path", Value: "/"})
		fiberCtx := ctx.(*fiber.Ctx)
		return fiberCtx.Render("chat", fiber.Map{
			"Title": "Hello, World!",
		}, "layouts/main")
	})

	if err := web.Start(); err != nil {
		panic(err)
	}
}
