package orderservice

import (
	pb "349877-artemkagor05-course-1478/gen/api"
	"349877-artemkagor05-course-1478/internal/cache"
	orderrepo "349877-artemkagor05-course-1478/internal/repository/order"
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"time"
)

type OrderServer interface {
	CreateOrder(ctx context.Context, item string, quantity int32) (string, error)
	GetOrder(ctx context.Context, id string) (*pb.Order, error)
	UpdateOrder(ctx context.Context, id string, item string, quantity int32) (*pb.Order, error)
	DeleteOrder(ctx context.Context, id string) error
	ListOrders(ctx context.Context) []*pb.Order
}

type OrderService struct {
	pb.UnimplementedOrderServiceServer
	repo    *orderrepo.OrderRepository
	cache   *cache.OrderCache
	log     *slog.Logger
	timeout time.Duration
}

func Register(
	gRPC *grpc.Server,
	repo *orderrepo.OrderRepository,
	cache *cache.OrderCache,
	log *slog.Logger,
	timeout time.Duration,
) {
	pb.RegisterOrderServiceServer(gRPC, New(repo, cache, log, timeout))
}

func New(
	repo *orderrepo.OrderRepository,
	cache *cache.OrderCache,
	log *slog.Logger,
	timeout time.Duration,
) *OrderService {
	return &OrderService{
		repo:    repo,
		cache:   cache,
		log:     log,
		timeout: timeout,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	if req.GetItem() == "" {
		return nil, status.Error(codes.InvalidArgument, "item name cannot be empty")
	}
	if req.GetQuantity() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be positive")
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	id, err := s.repo.CreateOrder(ctx, req.GetItem(), req.GetQuantity())
	if err != nil {
		if errors.Is(err, orderrepo.ErrOrderAlreadyExists) {
			return nil, status.Errorf(codes.AlreadyExists, "order with item '%s' already exists", req.GetItem())
		}
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}

	newOrder := &pb.Order{
		Id:       id,
		Item:     req.GetItem(),
		Quantity: req.GetQuantity(),
	}
	if err = s.cache.Set(ctx, newOrder); err != nil {
		s.log.Error("failed to prime cache after creation", "error", err, "order_id", id)
	}

	return &pb.CreateOrderResponse{Id: id}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	cachedOrder, err := s.cache.Get(ctx, req.GetId())
	if err != nil {
		s.log.Error("failed to get from cache", "error", err)
	}
	if cachedOrder != nil {
		s.log.Debug("got order from cache", "id", req.GetId())
		return &pb.GetOrderResponse{Order: cachedOrder}, nil
	}

	order, err := s.repo.GetOrder(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, orderrepo.ErrOrderNotFound) {
			return nil, status.Errorf(codes.NotFound, "order with id '%s' does not exists", req.GetId())
		}
		return nil, status.Errorf(codes.Internal, "failed to get order: %v", err)
	}

	if err = s.cache.Set(ctx, order); err != nil {
		s.log.Error("failed to set cache", "error", err)
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

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	order, err := s.repo.UpdateOrder(ctx, req.GetId(), req.GetItem(), req.GetQuantity())
	if err != nil {
		if errors.Is(err, orderrepo.ErrOrderNotFound) {
			return nil, status.Errorf(codes.NotFound, "order with id '%s' does not exists", req.GetId())
		}
		return nil, status.Errorf(codes.Internal, "failed to update order: %v", err)
	}

	if err = s.cache.Delete(ctx, order.Id); err != nil {
		s.log.Error("failed to invalidate cache on update", "error", err)
	}

	return &pb.UpdateOrderResponse{Order: order}, nil
}

func (s *OrderService) DeleteOrder(ctx context.Context, req *pb.DeleteOrderRequest) (*pb.DeleteOrderResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	err := s.repo.DeleteOrder(ctx, req.GetId())
	if err != nil {
		if errors.Is(err, orderrepo.ErrOrderNotFound) {
			return nil, status.Errorf(codes.NotFound, "order with id '%s' does not exists", req.GetId())
		}
		return nil, status.Errorf(codes.Internal, "failed to delete order: %v", err)
	}

	if err = s.cache.Delete(ctx, req.GetId()); err != nil {
		s.log.Error("failed to invalidate cache on delete", "error", err)
	}

	return &pb.DeleteOrderResponse{Success: true}, nil
}

func (s *OrderService) ListOrders(ctx context.Context, _ *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	orders := s.repo.ListOrders(ctx)

	return &pb.ListOrdersResponse{Orders: orders}, nil
}
