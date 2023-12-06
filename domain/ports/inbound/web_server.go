package inbound

type RouteHandlerFunc func(ctx interface{}) error

type WebServerConfig struct {
	Port int
}

type WebServerPort interface {
	Start(cfg WebServerConfig) error
	AddRoute(method string, path string, handler RouteHandlerFunc)
}
