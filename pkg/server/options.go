package server

import "github.com/gofiber/fiber/v2"

// Option configures optional HTTP server behavior at application assembly
// time. Implementations are private to this package so the public API can grow
// without exposing the server's internal options representation.
type Option interface {
	apply(*Options)
}

// Options is the resolved set of server options passed through dependency
// injection. Callers should construct it with NewOptions.
type Options struct {
	middlewares []fiber.Handler
}

type httpMiddlewareOption struct {
	middlewares []fiber.Handler
}

func (o httpMiddlewareOption) apply(options *Options) {
	options.middlewares = append(options.middlewares, o.middlewares...)
}

// WithHTTPMiddleware adds middleware after panic recovery and before Anclax's
// built-in middleware and routes. Middleware runs in declaration order and
// should normally call c.Next() so the rest of the HTTP pipeline can execute.
func WithHTTPMiddleware(middlewares ...fiber.Handler) Option {
	return httpMiddlewareOption{middlewares: append([]fiber.Handler(nil), middlewares...)}
}

// NewOptions resolves functional options into the value consumed by NewServer.
func NewOptions(options ...Option) Options {
	var resolved Options
	for _, option := range options {
		if option != nil {
			option.apply(&resolved)
		}
	}
	return resolved
}
