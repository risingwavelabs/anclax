package wire

import (
	"github.com/gofiber/fiber/v2"
	"github.com/risingwavelabs/anclax/pkg/app"
	"github.com/risingwavelabs/anclax/pkg/config"
	"github.com/risingwavelabs/anclax/pkg/server"
)

// Option configures optional application behavior during initialization.
type Option = server.Option

// WithHTTPMiddleware adds middleware to the Anclax HTTP pipeline before any
// built-in or consumer routes are registered.
func WithHTTPMiddleware(middlewares ...fiber.Handler) Option {
	return server.WithHTTPMiddleware(middlewares...)
}

// InitializeApplication constructs an Anclax application. Optional runtime
// behavior is supplied at this composition root instead of being mixed into
// the application's static LibConfig.
func InitializeApplication(
	cfg *config.Config,
	libCfg *config.LibConfig,
	options ...Option,
) (*app.Application, error) {
	return initializeApplication(cfg, libCfg, server.NewOptions(options...))
}
