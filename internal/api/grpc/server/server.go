package server

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"authmicro/pkg/logger"
)

// AuthServer represents the gRPC server for auth service
type AuthServer struct {
	address      string
	server       *grpc.Server
	logger       logger.Logger
	authService  interface{}
	tokenService interface{}
}

// NewGRPCServer creates a new instance of the gRPC server
func NewGRPCServer(address string, authService interface{}, tokenService interface{}, logger logger.Logger) *AuthServer {
	return &AuthServer{
		address:      address,
		logger:       logger,
		authService:  authService,
		tokenService: tokenService,
	}
}

// Start starts the gRPC server
func (s *AuthServer) Start() error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	// Create a new gRPC server
	server := grpc.NewServer()

	// Enable reflection for grpcurl and other tools
	reflection.Register(server)

	// Store the server instance
	s.server = server

	// Register gRPC services here
	// proto.RegisterAuthServiceServer(server, NewAuthServiceGRPC(s.authService, s.logger))

	s.logger.Infof("gRPC server is listening on %s", s.address)
	return server.Serve(lis)
}

// Stop gracefully stops the gRPC server
func (s *AuthServer) Stop() {
	if s.server != nil {
		s.logger.Info("Gracefully stopping gRPC server")
		s.server.GracefulStop()
	}
}

// RegisterService registers a gRPC service with the server
func (s *AuthServer) RegisterService(sd *grpc.ServiceDesc, ss interface{}) {
	s.server.RegisterService(sd, ss)
}

// Shutdown is an alias for Stop method to maintain interface consistency
func (s *AuthServer) Shutdown(ctx context.Context) error {
	s.Stop()
	return nil
}
