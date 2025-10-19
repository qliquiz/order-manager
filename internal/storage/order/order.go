package orderstore

import (
	pb "349877-artemkagor05-course-1478/api/gen"
	"context"
	"errors"
	"github.com/google/uuid"
	"sync"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderAlreadyExists = errors.New("order already exists")
)

type OrderStorage struct {
	mu     sync.RWMutex
	orders map[string]*pb.Order
}

func New() *OrderStorage {
	return &OrderStorage{
		orders: make(map[string]*pb.Order),
	}
}

func (s *OrderStorage) CreateOrder(ctx context.Context, item string, quantity int32) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, order := range s.orders {
		if order.Item == item {
			return "", ErrOrderAlreadyExists
		}
	}

	order := &pb.Order{
		Id:       uuid.NewString(),
		Item:     item,
		Quantity: quantity,
	}
	s.orders[order.Id] = order

	return order.Id, nil
}

func (s *OrderStorage) GetOrder(ctx context.Context, id string) (*pb.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, order := range s.orders {
		if order.Id == id {
			return order, nil
		}
	}

	return nil, ErrOrderNotFound
}

func (s *OrderStorage) UpdateOrder(ctx context.Context, id string, item string, quantity int32) (*pb.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[id]
	if !ok {
		return nil, ErrOrderNotFound
	}

	for _, existingOrder := range s.orders {
		if existingOrder.Id != id && existingOrder.Item == item {
			return nil, ErrOrderAlreadyExists
		}
	}

	order.Item = item
	order.Quantity = quantity

	return order, nil
}

func (s *OrderStorage) DeleteOrder(ctx context.Context, id string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.orders[id]; !ok {
		return ErrOrderNotFound
	}
	delete(s.orders, id)

	return nil
}

func (s *OrderStorage) ListOrders(ctx context.Context) []*pb.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]*pb.Order, 0, len(s.orders))
	for _, order := range s.orders {
		orders = append(orders, order)
	}

	return orders
}
