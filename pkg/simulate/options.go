package simulate

import (
	"sync"

	"github.com/kiga-hub/arc-consumer/pkg/goss"
	"github.com/kiga-hub/arc-consumer/pkg/grpc"
	"github.com/kiga-hub/arc/logging"
)

// Option is a function that will set up option.
type Option func(opts *Server)

func loadOptions(options ...Option) *Server {
	opts := &Server{
		frameChans: new(sync.Map),
		sensors:    new(sync.Map),
	}
	for _, option := range options {
		option(opts)
	}
	if opts.logger == nil {
		opts.logger = new(logging.NoopLogger)
	}
	if opts.config == nil {
		opts.config = GetConfig()
	}
	return opts
}

// WithGrpc -
func WithGrpc(g grpc.Handler) Option {
	return func(opts *Server) {
		opts.grpc = g
	}
}

// WithKVCache -
func WithKVCache(g goss.Handler) Option {
	return func(opts *Server) {
		opts.kvCache = g
	}
}

// WithLogger -
func WithLogger(logger logging.ILogger) Option {
	return func(opts *Server) {
		opts.logger = logger
	}
}

// WithConfig -
func WithConfig(c *Config) Option {
	return func(opts *Server) {
		opts.config = c
	}
}
