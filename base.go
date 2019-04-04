// Package base provides a simple, unregistered gRPC server
// with functional options not included in the base library.
// Primarily useful to synchornize configurations across
// development teams.
package base

import (
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server for gRPC without any registered services.
type Server struct {
	S *grpc.Server
}

// Option provide a functional interface for constructing grpc.ServerOptions.
type Option func() ([]grpc.ServerOption, error)

// NewServer returns a unregistered gRPC Server with
// sensible defaults for configuration and middlewares.
func NewServer(opts ...Option) (*Server, error) {
	var oo []grpc.ServerOption
	for _, opt := range opts {
		o, err := opt()
		if err != nil {
			return nil, err
		}
		oo = append(oo, o...)
	}
	return &Server{
		S: grpc.NewServer(oo...),
	}, nil
}

// ListenAndServe from the Server Config's port.
func (s *Server) ListenAndServe(addr string) error {
	if len(s.S.GetServiceInfo()) < 1 {
		return errors.New("no services registered to this server")
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrapf(err, "unable to listen on %s", addr)
	}
	return s.S.Serve(lis)
}

// Bundle standard gRPC options into our custom Option.
func Bundle(opts ...grpc.ServerOption) Option {
	var oo []grpc.ServerOption
	for _, opt := range opts {
		oo = append(oo, opt)
	}
	return func() ([]grpc.ServerOption, error) { return oo, nil }
}

// TLS parses certs for a valid TLS config.
func TLS(certPath, keyPath string) Option {
	return func() ([]grpc.ServerOption, error) {
		creds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
		if err != nil {
			return nil, err
		}
		opts := []grpc.ServerOption{grpc.Creds(creds)}
		return opts, err
	}
}

// Log for server and other basic, non-intrusive interceptors.
func Log(l *zap.Logger) Option {
	return func() ([]grpc.ServerOption, error) {
		if l == nil {
			l = zap.NewExample()
		}
		opts := []grpc.ServerOption{
			grpc_middleware.WithUnaryServerChain(
				grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_zap.UnaryServerInterceptor(l),
				grpc_recovery.UnaryServerInterceptor(),
			),
			grpc_middleware.WithStreamServerChain(
				grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_zap.StreamServerInterceptor(l),
				grpc_recovery.StreamServerInterceptor(),
			)}
		return opts, nil
	}
}
