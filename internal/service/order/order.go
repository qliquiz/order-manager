package orderservice

import (
	pb "349877-artemkagor05-course-1478/api/gen"
	"349877-artemkagor05-course-1478/internal/storage/order"
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Order interface {
	CreateOrder(ctx context.Context, item string, quantity int32) (string, error)
	GetOrder(ctx context.Context, id string) (*pb.Order, error)
	UpdateOrder(ctx context.Context, id string, item string, quantity int32) (*pb.Order, error)
	DeleteOrder(ctx context.Context, id string) error
	ListOrders(ctx context.Context) []*pb.Order
}

type OrderService struct {
	pb.UnimplementedOrderServiceServer
	order Order
}

func Register(gRPC *grpc.Server, order Order) {
	pb.RegisterOrderServiceServer(gRPC, New(order))
}

func New(order Order) *OrderService {
	return &OrderService{
		order: order,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	if req.GetItem() == "" {
		return nil, status.Error(codes.InvalidArgument, "item name cannot be empty")
	}
	if req.GetQuantity() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be positive")
	}

	id, err := s.order.CreateOrder(ctx, req.GetItem(), req.GetQuantity())
	if err != nil {
		if errors.Is(err, orderstore.ErrOrderAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "order with item '%s' already exists", req.GetItem())
		}
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}

	return &pb.CreateOrderResponse{Id: id}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id cannot be empty")
	}

	order, err := s.order.GetOrder(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, orderstore.ErrOrderNotFound) {
			return nil, status.Errorf(codes.NotFound, "order with id '%s' does not exists", req.GetId())
		}
		return nil, status.Errorf(codes.Internal, "failed to get order: %v", err)
	}

	return &pb.GetOrderResponse{Order: order}, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.UpdateOrderResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id cannot be empty")
	}
	if req.GetQuantity() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be positive")
	}

	order, err := s.order.UpdateOrder(ctx, req.GetId(), req.GetItem(), req.GetQuantity())
	if err != nil {
		if errors.Is(err, orderstore.ErrOrderNotFound) {
			return nil, status.Errorf(codes.NotFound, "order with id '%s' does not exists", req.GetId())
		}
		return nil, status.Errorf(codes.Internal, "failed to update order: %v", err)
	}

	return &pb.UpdateOrderResponse{Order: order}, nil
}

func (s *OrderService) DeleteOrder(ctx context.Context, req *pb.DeleteOrderRequest) (*pb.DeleteOrderResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id cannot be empty")
	}

	err := s.order.DeleteOrder(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, orderstore.ErrOrderNotFound) {
			return nil, status.Errorf(codes.NotFound, "order with id '%s' does not exists", req.GetId())
		}
		return nil, status.Errorf(codes.Internal, "failed to delete order: %v", err)
	}

	return &pb.DeleteOrderResponse{Success: true}, nil
}

func (s *OrderService) ListOrders(ctx context.Context, req *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	orders := s.order.ListOrders(ctx)

	return &pb.ListOrdersResponse{Orders: orders}, nil
}
