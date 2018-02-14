package rendezvous

import (
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

type options struct {
	log         *zap.Logger
	credentials credentials.TransportCredentials
}

func newOptions() *options {
	return &options{
		log:         zap.NewNop(),
		credentials: nil,
	}
}

type Option func(options *options)

func WithLogger(log *zap.Logger) Option {
	return func(options *options) {
		options.log = log
	}
}

func WithCredentials(credentials credentials.TransportCredentials) Option {
	return func(options *options) {
		options.credentials = credentials
	}
}
