package inbound

import (
	"fmt"

	"github.com/ChristianSch/Theta/domain/ports/inbound"
	"github.com/ChristianSch/Theta/domain/ports/outbound"
	"github.com/gofiber/fiber/v2"
)

type FiberWebServer struct {
	Port     int
	Server   *fiber.App
	adapters FiberWebServerAdapters
}

type FiberWebServerAdapters struct {
	Log outbound.Log
}

func NewFiberWebServer(port int, adapters FiberWebServerAdapters) *FiberWebServer {
	server := fiber.New()

	return &FiberWebServer{
		Port:     port,
		Server:   server,
		adapters: adapters,
	}
}

func (f *FiberWebServer) Start() error {
	return f.Server.Listen(fmt.Sprintf(":%d", f.Port))
}

func (f *FiberWebServer) AddRoute(method string, path string, handler inbound.RouteHandlerFunc) {
	adaptedHandler := func(c *fiber.Ctx) error {
		return handler(c)
	}

	switch method {
	case "GET":
		f.Server.Get(path, adaptedHandler)
	case "POST":
		f.Server.Post(path, adaptedHandler)
	case "PUT":
		f.Server.Put(path, adaptedHandler)
	case "DELETE":
		f.Server.Delete(path, adaptedHandler)
	default:
		panic(fmt.Sprintf("unsupported method %s", method))
	}

	f.adapters.Log.Debug("added route", outbound.LogField{Key: "method", Value: method}, outbound.LogField{Key: "path", Value: path})
}
