package dto

import (
	"github.com/victus-est-deus/shipment/internal/domain/entity"
	pb "github.com/victus-est-deus/shipment/proto/shipment"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ShipmentToProto(s *entity.Shipment) *pb.Shipment {
	return &pb.Shipment{
		Id:                    s.ID.String(),
		ReferenceNumber:       s.ReferenceNumber,
		Origin:                s.Origin,
		Destination:           s.Destination,
		Status:                s.Status.String(),
		DriverName:            s.DriverName,
		DriverPhone:           s.DriverPhone,
		UnitNumber:            s.UnitNumber,
		ShipmentAmount:        s.ShipmentAmount.Amount(),
		ShipmentCurrency:      s.ShipmentAmount.Currency(),
		DriverRevenue:         s.DriverRevenue.Amount(),
		DriverRevenueCurrency: s.DriverRevenue.Currency(),
		CreatedAt:             timestamppb.New(s.CreatedAt),
		UpdatedAt:             timestamppb.New(s.UpdatedAt),
	}
}

func StatusEventToProto(e *entity.StatusEvent) *pb.StatusEvent {
	return &pb.StatusEvent{
		Id:         e.ID.String(),
		ShipmentId: e.ShipmentID.String(),
		Status:     e.Status.String(),
		Location:   e.Location,
		Notes:      e.Notes,
		CreatedAt:  timestamppb.New(e.CreatedAt),
	}
}

func StatusEventsToProto(events []*entity.StatusEvent) []*pb.StatusEvent {
	result := make([]*pb.StatusEvent, len(events))
	for i, e := range events {
		result[i] = StatusEventToProto(e)
	}
	return result
}
