package server

import (
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "authmicro/internal/api/grpc/proto"
	"authmicro/internal/api/grpc/service"
	"authmicro/pkg/logger"
)

type GRPCServer struct {
	addr       string
	grpcServer *grpc.Server
	logger     logger.Logger
}

func NewGRPCServer(addr string, authService authService, tokenService tokenService, logger logger.Logger) *GRPCServer {
	// Create gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	authGRPCService := service.NewAuthGRPCService(authService, tokenService, logger)
	pb.RegisterAuthServiceServer(grpcServer, authGRPCService)

	// Enable reflection for development tools
	reflection.Register(grpcServer)

	return &GRPCServer{
		addr:       addr,
		grpcServer: grpcServer,
		logger:     logger,
	}
}

// Start starts the gRPC server
func (s *GRPCServer) Start() error {
	// Create listener
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.logger.Errorf("Failed to listen: %v", err)
		return err
	}

	// Start server
	s.logger.Infof("gRPC server listening on %s", s.addr)
	return s.grpcServer.Serve(lis)
}

// Stop stops the gRPC server
func (s *GRPCServer) Stop() {
	s.logger.Info("Stopping gRPC server...")
	s.grpcServer.GracefulStop()
	s.logger.Info("gRPC server stopped")
}
