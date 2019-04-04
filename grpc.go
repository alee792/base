// Pacakge grpc provides a simple, unregistered gRPC server with sensible defaults.
package grpc

import (
	"net"

	"google.golang.org/grpc/credentials"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Server for gRPC without any registered services.
type Server struct {
	svr    *grpc.Server
	Logger *zap.Logger
	Config Config
}

// Config for Server that allows additional options to be injected.
type Config struct {
	Addr       string
	CertPath   string
	KeyPath    string
	ServerOpts []grpc.ServerOption
}

// NewServer returns a unregistered gRPC Server with
// sensible defaults for configuration and middlewares.
func NewServer(logger *zap.Logger, cfg Config) (*Server, error) {
	if cfg.Addr == "" {
		cfg.Addr = ":8443"
	}

	// Set up options and middleware.
	var oo []grpc.ServerOption
	if cfg.CertPath != "" && cfg.KeyPath != "" {
		creds, err := resolveCreds(cfg.CertPath, cfg.KeyPath)
		if err != nil {
			return nil, errors.Wrap(err, "could not resolve TLS credentials")
		}
		oo = append(oo, creds...)
	}

	if logger != nil {
		oo = append(oo,
			grpc_middleware.WithUnaryServerChain(
				grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_zap.UnaryServerInterceptor(logger),
				grpc_recovery.UnaryServerInterceptor(),
			),
			grpc_middleware.WithStreamServerChain(
				grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_zap.StreamServerInterceptor(logger),
				grpc_recovery.StreamServerInterceptor(),
			),
		)
	}

	gs := grpc.NewServer(append(oo, cfg.ServerOpts...)...)
	s := &Server{
		svr:    gs,
		Config: cfg,
	}

	return s, nil
}

// ListenAndServe from the Server Config's port.
func (s *Server) ListenAndServe() error {
	if len(s.svr.GetServiceInfo()) < 1 {
		return errors.New("no services registered to this server")
	}
	lis, err := net.Listen("tcp", s.Config.Addr)
	if err != nil {
		return errors.Wrapf(err, "unable to listen on %s", s.Config.Addr)
	}
	return s.svr.Serve(lis)
}

// resolveCreds parses certs for a valid TLS config.
func resolveCreds(certPath, keyPath string) ([]grpc.ServerOption, error) {
	creds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	opts := []grpc.ServerOption{grpc.Creds(creds)}
	return opts, err
}
