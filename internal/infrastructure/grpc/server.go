package grpc

import (
	"fmt"
	"log"
	"net"

	"github.com/victus-est-deus/shipment/internal/infrastructure/grpc/handler"
	pb "github.com/victus-est-deus/shipment/proto/shipment"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	port       int
}

func NewServer(port int, shipmentHandler *handler.ShipmentHandler) *Server {
	grpcServer := grpc.NewServer()
	pb.RegisterShipmentServiceServer(grpcServer, shipmentHandler)

	return &Server{
		grpcServer: grpcServer,
		port:       port,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	log.Printf("gRPC server listening on port %d", s.port)
	return s.grpcServer.Serve(listener)
}

func (s *Server) Stop() {
	log.Println("stopping gRPC server...")
	s.grpcServer.GracefulStop()
}
