package server

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	"github.com/risingwavelabs/anclax/pkg/config"
)

func TestRegisterMiddlewareAppliesHTTPMiddlewareInPipelineOrder(t *testing.T) {
	app := fiber.New()
	var calls []string
	s := &Server{
		app:    app,
		libCfg: &config.LibConfig{},
		options: NewOptions(
			WithHTTPMiddleware(
				func(c *fiber.Ctx) error {
					calls = append(calls, "outer-request")
					c.Set("X-Library-Middleware", "enabled")
					err := c.Next()
					calls = append(calls, "outer-response")
					return err
				},
				func(c *fiber.Ctx) error {
					calls = append(calls, "inner-request")
					err := c.Next()
					calls = append(calls, "inner-response")
					return err
				},
			),
		),
		skipLogRequest:  func(*fiber.Ctx) bool { return true },
		skipLogResponse: func(*fiber.Ctx) bool { return true },
	}

	s.registerMiddleware()
	app.Get("/auth/probe", func(c *fiber.Ctx) error {
		calls = append(calls, "route")
		return c.SendStatus(fiber.StatusNoContent)
	})

	response, err := app.Test(httptest.NewRequest("GET", "/auth/probe", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNoContent, response.StatusCode)
	require.Equal(t, "enabled", response.Header.Get("X-Library-Middleware"))
	require.Equal(t, []string{
		"outer-request",
		"inner-request",
		"route",
		"inner-response",
		"outer-response",
	}, calls)
}

func TestRegisterMiddlewareRecoveryWrapsHTTPMiddleware(t *testing.T) {
	app := fiber.New()
	s := &Server{
		app:    app,
		libCfg: &config.LibConfig{},
		options: NewOptions(WithHTTPMiddleware(func(c *fiber.Ctx) error {
			c.Set("X-Library-Middleware", "enabled")
			return c.Next()
		})),
		skipLogRequest:  func(*fiber.Ctx) bool { return true },
		skipLogResponse: func(*fiber.Ctx) bool { return true },
	}

	s.registerMiddleware()
	app.Get("/panic", func(*fiber.Ctx) error {
		panic("test panic")
	})

	response, err := app.Test(httptest.NewRequest("GET", "/panic", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, response.StatusCode)
	require.Equal(t, "enabled", response.Header.Get("X-Library-Middleware"))
}
