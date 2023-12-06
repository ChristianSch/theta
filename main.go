package main

import (
	"github.com/ChristianSch/Theta/adapters/inbound"
	"github.com/ChristianSch/Theta/adapters/outbound"
	outboundPorts "github.com/ChristianSch/Theta/domain/ports/outbound"
	"github.com/gofiber/fiber/v2"
)

func main() {
	log := outbound.NewZapLogger(outbound.ZapLoggerConfig{Debug: true})
	web := inbound.NewFiberWebServer(8080, inbound.FiberWebServerAdapters{Log: log})

	web.AddRoute("GET", "/", func(ctx interface{}) error {
		log.Debug("handling request", outboundPorts.LogField{Key: "path", Value: "/"})
		fiberCtx := ctx.(*fiber.Ctx)
		return fiberCtx.SendString("hello world")
	})

	if err := web.Start(); err != nil {
		panic(err)
	}
}
