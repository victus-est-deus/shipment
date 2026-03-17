package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/victus-est-deus/shipment/internal/application/usecase"
	"github.com/victus-est-deus/shipment/internal/domain/repository"
	"github.com/victus-est-deus/shipment/internal/domain/service"
	"github.com/victus-est-deus/shipment/internal/infrastructure/config"
	"github.com/victus-est-deus/shipment/internal/infrastructure/grpc"
	"github.com/victus-est-deus/shipment/internal/infrastructure/grpc/handler"
	"github.com/victus-est-deus/shipment/internal/infrastructure/persistence/jsonfile"
	"github.com/victus-est-deus/shipment/internal/infrastructure/persistence/postgres"
)

func main() {
	cfg, err := config.Load("config.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var shipmentRepo repository.ShipmentRepository
	var statusEventRepo repository.StatusEventRepository
	var logRepo repository.LogRepository

	if cfg.Storage == config.StoragePostgres {
		log.Println("Using Postgres for storage")
		db, err := postgres.NewConnection(cfg.Database)
		if err != nil {
			log.Fatalf("Failed to connect to postgres: %v", err)
		}
		defer db.Close()

		shipmentRepo = postgres.NewShipmentRepository(db)
		statusEventRepo = postgres.NewStatusEventRepository(db)
		logRepo = postgres.NewLogRepository(db)
	} else {
		log.Println("Using JSON files for storage")
		store, err := jsonfile.NewStoreFromConfig()
		if err != nil {
			log.Fatalf("Failed to initialize JSON store: %v", err)
		}

		shipmentRepo = jsonfile.NewShipmentRepository(store)
		statusEventRepo = jsonfile.NewStatusEventRepository(store)
		logRepo = jsonfile.NewLogRepository(store)
	}

	shipmentService := service.NewShipmentService(shipmentRepo, statusEventRepo, logRepo)
	shipmentUseCase := usecase.NewShipmentUseCase(shipmentService)
	shipmentHandler := handler.NewShipmentHandler(shipmentUseCase)

	server := grpc.NewServer(cfg.GRPC.Port, shipmentHandler)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("gRPC server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Stop()
	log.Println("Server gracefully stopped")
}
