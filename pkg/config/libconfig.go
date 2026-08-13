package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/risingwavelabs/anclax/lib/ws"
)

type PgCfg struct {
	MaxConnections int32
	MinConnections int32
}

type LogCfg struct {
	// (optional) If set, only log entries where the request path starts with this prefix will be logged.
	RequestPathPrefix *string

	// (optional) If set, only error will be logged for the health check path.
	HealthCheckPath *string
}

type LibConfig struct {
	Cors *cors.Config
	Pg   *PgCfg
	Log  LogCfg
	Ws   *ws.WsCfg

	// GlobalMiddlewares are registered after panic recovery and before
	// Anclax's built-in middleware and routes. Libraries embedding Anclax can
	// use this hook for policies that must cover every response, including
	// authentication routes registered by NewServer.
	GlobalMiddlewares []fiber.Handler
}

func DefaultLibConfig() *LibConfig {
	return &LibConfig{
		Pg: &PgCfg{
			MaxConnections: 10,
			MinConnections: 1,
		},
	}
}
