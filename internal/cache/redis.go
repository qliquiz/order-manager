package cache

import (
	pb "349877-artemkagor05-course-1478/gen/api"
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

type OrderCache struct {
	client *redis.Client
}

func New(client *redis.Client) *OrderCache {
	return &OrderCache{client: client}
}

func (c *OrderCache) Get(ctx context.Context, id string) (*pb.Order, error) {
	result, err := c.client.Get(ctx, id).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var order pb.Order
	if err = json.Unmarshal([]byte(result), &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (c *OrderCache) Set(ctx context.Context, order *pb.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, order.Id, data, 5*time.Minute).Err()
}

func (c *OrderCache) Delete(ctx context.Context, id string) error {
	return c.client.Del(ctx, id).Err()
}
