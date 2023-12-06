package inbound

import (
	"fmt"

	"github.com/ChristianSch/Theta/domain/ports/inbound"
	"github.com/ChristianSch/Theta/domain/ports/outbound"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

type FiberWebServerConfig struct {
	Port                int
	StaticResourcesPath string
	TemplatesPath       string
	TemplatesExtension  string
}

type FiberWebServer struct {
	TemplatesPath string
	Port          int
	Server        *fiber.App
	adapters      FiberWebServerAdapters
}

type FiberWebServerAdapters struct {
	Log outbound.Log
}

func NewFiberWebServer(cfg FiberWebServerConfig, adapters FiberWebServerAdapters) *FiberWebServer {
	server := fiber.New(fiber.Config{
		Views: html.New(cfg.TemplatesPath, cfg.TemplatesExtension),
	})

	server.Static("/static", cfg.StaticResourcesPath)

	return &FiberWebServer{
		Port:     cfg.Port,
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
