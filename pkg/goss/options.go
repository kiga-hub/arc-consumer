package goss

import (
	"github.com/kiga-hub/arc/logging"
	microComponent "github.com/kiga-hub/arc/micro/component"
)

// Option is a function that will set up option.
type Option func(opts *Server)

func loadOptions(options ...Option) *Server {
	opts := &Server{}
	for _, option := range options {
		option(opts)
	}
	if opts.logger == nil {
		opts.logger = new(logging.NoopLogger)
	}
	return opts
}

// WithLogger -
func WithLogger(logger logging.ILogger) Option {
	return func(opts *Server) {
		opts.logger = logger
	}
}

// WithKVCache -
func WithKVCache(kvc *microComponent.GossipKVCacheComponent) Option {
	return func(opts *Server) {
		opts.gossipKVCache = kvc
	}
}
