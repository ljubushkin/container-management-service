package grpctransport

import (
	"context"

	"github.com/ljubushkin/container-management-service/internal/domain"
	"github.com/ljubushkin/container-management-service/internal/usecase"
	containerv1 "github.com/ljubushkin/container-management-service/pkg/api/container/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	containerv1.UnimplementedContainerServiceServer

	service *usecase.Service
}

func NewServer(service *usecase.Service) *Server {
	return &Server{
		service: service,
	}
}

func (s *Server) GetContainer(
	ctx context.Context,
	req *containerv1.GetContainerRequest,
) (*containerv1.GetContainerResponse, error) {

	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	container, err := s.service.GetByID(req.GetId())
	if err != nil {
		return nil, mapError(err)
	}

	return &containerv1.GetContainerResponse{
		Container: toProtoContainer(container),
	}, nil
}

func (s *Server) CreateContainer(
	ctx context.Context, req *containerv1.CreateContainerRequest) (*containerv1.CreateContainerResponse, error) {

	if req.GetType() == "" {
		return nil, status.Error(codes.InvalidArgument, "type is required")
	}

	container, err := s.service.CreateContainer(req.GetType())
	if err != nil {
		return nil, mapError(err)
	}

	return &containerv1.CreateContainerResponse{
		Container: toProtoContainer(container)}, nil
}

func (s *Server) ListContainers(
	ctx context.Context,
	req *containerv1.ListContainersRequest,
) (*containerv1.ListContainersResponse, error) {
	filter := domain.ContainerFilter{}

	if req.GetType() != "" {
		v := req.GetType()
		filter.TypeCode = &v
	}

	if req.GetWarehouseId() != "" {
		v := req.GetWarehouseId()
		filter.WarehouseID = &v
	}

	if req.GetStatus() != "" {
		status := domain.Status(req.GetStatus())
		filter.Status = &status
	}

	filter.Limit = int(req.GetLimit())
	filter.Offset = int(req.GetOffset())

	containers, err := s.service.List(filter)
	if err != nil {
		return nil, mapError(err)
	}

	return &containerv1.ListContainersResponse{
		Containers: toProtoContainers(containers),
	}, nil
}

func (s *Server) CreateContainerBatch(
	ctx context.Context, req *containerv1.CreateContainerBatchRequest) (*containerv1.CreateContainerBatchResponse, error) {

	if req.GetType() == "" {
		return nil, status.Error(codes.InvalidArgument, "type is required")
	}

	if req.GetCount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "count must be greater than 0")
	}

	containers, err := s.service.CreateBatch(req.GetType(), int(req.GetCount()))
	if err != nil {
		return nil, mapError(err)
	}

	return &containerv1.CreateContainerBatchResponse{
		Containers: toProtoContainers(containers)}, nil
}

func (s *Server) AssignWarehouse(
	ctx context.Context,
	req *containerv1.AssignWarehouseRequest,
) (*containerv1.AssignWarehouseResponse, error) {

	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if req.GetWarehouseId() == "" {
		return nil, status.Error(codes.InvalidArgument, "warehouse id is required")
	}

	err := s.service.AssignWarehouse(req.GetId(), req.GetWarehouseId())
	if err != nil {
		return nil, mapError(err)
	}

	return &containerv1.AssignWarehouseResponse{}, nil
}

func (s *Server) ListWarehouses(
	ctx context.Context,
	req *containerv1.ListWarehousesRequest,
) (*containerv1.ListWarehousesResponse, error) {
	warehouses, err := s.service.ListWarehouses()
	if err != nil {
		return nil, mapError(err)
	}

	resp := make([]*containerv1.Warehouse, 0, len(warehouses))
	for _, wh := range warehouses {
		resp = append(resp, &containerv1.Warehouse{
			WarehouseId: wh.ID,
			Name:        wh.Name,
		})
	}
	return &containerv1.ListWarehousesResponse{
		Warehouses: resp,
	}, nil
}

func (s *Server) ListContainerTypes(
	ctx context.Context,
	req *containerv1.ListContainerTypesRequest,
) (*containerv1.ListContainerTypesResponse, error) {
	types, err := s.service.ListContainerTypes()
	if err != nil {
		return nil, mapError(err)
	}

	resp := make([]*containerv1.ContainerType, 0, len(types))
	for _, t := range types {
		resp = append(resp, &containerv1.ContainerType{
			Code: t.Code,
			Name: t.Name,
		})
	}
	return &containerv1.ListContainerTypesResponse{
		ContainerTypes: resp,
	}, nil
}

func (s *Server) UpdateStatus(
	ctx context.Context,
	req *containerv1.UpdateStatusRequest,
) (*containerv1.UpdateStatusResponse, error) {

	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if req.GetStatus() == "" {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	statusValue := domain.Status(req.GetStatus())

	if !domain.IsValidStatus(statusValue) {
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}

	err := s.service.UpdateStatus(req.GetId(), statusValue)
	if err != nil {
		return nil, mapError(err)
	}

	return &containerv1.UpdateStatusResponse{}, nil
}
