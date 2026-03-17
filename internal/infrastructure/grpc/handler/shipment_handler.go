package handler

import (
	"context"
	"errors"

	"github.com/victus-est-deus/shipment/internal/application/dto"
	"github.com/victus-est-deus/shipment/internal/application/usecase"
	"github.com/victus-est-deus/shipment/internal/domain/service"
	pb "github.com/victus-est-deus/shipment/proto/shipment"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ShipmentHandler struct {
	pb.UnimplementedShipmentServiceServer
	useCase *usecase.ShipmentUseCase
}

func NewShipmentHandler(uc *usecase.ShipmentUseCase) *ShipmentHandler {
	return &ShipmentHandler{useCase: uc}
}

func (h *ShipmentHandler) CreateShipment(ctx context.Context, req *pb.CreateShipmentRequest) (*pb.CreateShipmentResponse, error) {
	shipment, err := h.useCase.CreateShipment(ctx, dto.CreateShipmentRequest{
		ReferenceNumber:       req.GetReferenceNumber(),
		Origin:                req.GetOrigin(),
		Destination:           req.GetDestination(),
		DriverName:            req.GetDriverName(),
		DriverPhone:           req.GetDriverPhone(),
		UnitNumber:            req.GetUnitNumber(),
		ShipmentAmount:        req.GetShipmentAmount(),
		ShipmentCurrency:      req.GetShipmentCurrency(),
		DriverRevenue:         req.GetDriverRevenue(),
		DriverRevenueCurrency: req.GetDriverRevenueCurrency(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.CreateShipmentResponse{
		Shipment: dto.ShipmentToProto(shipment),
	}, nil
}

func (h *ShipmentHandler) GetShipment(ctx context.Context, req *pb.GetShipmentRequest) (*pb.GetShipmentResponse, error) {
	shipment, err := h.useCase.GetShipment(ctx, req.GetId())
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.GetShipmentResponse{
		Shipment: dto.ShipmentToProto(shipment),
	}, nil
}

func (h *ShipmentHandler) UpdateStatus(ctx context.Context, req *pb.UpdateStatusRequest) (*pb.UpdateStatusResponse, error) {
	event, err := h.useCase.UpdateStatus(ctx, dto.UpdateStatusRequest{
		ShipmentID: req.GetShipmentId(),
		Status:     req.GetStatus(),
		Location:   req.GetLocation(),
		Notes:      req.GetNotes(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.UpdateStatusResponse{
		Event: dto.StatusEventToProto(event),
	}, nil
}

func (h *ShipmentHandler) GetEventHistory(ctx context.Context, req *pb.GetEventHistoryRequest) (*pb.GetEventHistoryResponse, error) {
	events, err := h.useCase.GetEventHistory(ctx, req.GetShipmentId())
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.GetEventHistoryResponse{
		Events: dto.StatusEventsToProto(events),
	}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, service.ErrShipmentNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrDuplicateReference):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrInvalidTransition):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, service.ErrShipmentTerminated):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
