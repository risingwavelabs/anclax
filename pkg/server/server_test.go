package server

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"github.com/risingwavelabs/anclax/pkg/config"
)

func TestRegisterMiddlewareAppliesLibraryMiddlewareBeforeRoutes(t *testing.T) {
	app := fiber.New()
	s := &Server{
		app: app,
		libCfg: &config.LibConfig{
			GlobalMiddlewares: []fiber.Handler{
				func(c *fiber.Ctx) error {
					c.Set("X-Library-Middleware", "enabled")
					return c.Next()
				},
			},
		},
		skipLogRequest:  func(*fiber.Ctx) bool { return true },
		skipLogResponse: func(*fiber.Ctx) bool { return true },
	}

	s.registerMiddleware()
	app.Get("/auth/probe", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	response, err := app.Test(httptest.NewRequest("GET", "/auth/probe", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNoContent, response.StatusCode)
	require.Equal(t, "enabled", response.Header.Get("X-Library-Middleware"))
}
