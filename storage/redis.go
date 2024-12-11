package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(host string, port int, password string, db int) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisClient{client: client}, nil
}

func (rc *RedisClient) StoreKline(ctx context.Context, symbol, interval string, kline interface{}) error {
	key := fmt.Sprintf("kline:%s:%s", symbol, interval)
	return rc.storeWithLimit(ctx, key, kline, "kline_max_items")
}

func (rc *RedisClient) StoreTrade(ctx context.Context, symbol string, trade interface{}) error {
	key := fmt.Sprintf("trades:%s", symbol)
	return rc.storeWithLimit(ctx, key, trade, "trades_max_items")
}

func (rc *RedisClient) StoreOrderBook(ctx context.Context, symbol string, orderbook interface{}) error {
	key := fmt.Sprintf("orderbook:%s", symbol)
	data, err := json.Marshal(orderbook)
	if err != nil {
		return err
	}
	return rc.client.Set(ctx, key, data, 5*time.Second).Err()
}

func (rc *RedisClient) StoreBookTicker(ctx context.Context, symbol string, bookTicker interface{}) error {
	key := fmt.Sprintf("bookticker:%s", symbol)
	data, err := json.Marshal(bookTicker)
	if err != nil {
		return err
	}
	return rc.client.Set(ctx, key, data, 1*time.Second).Err()
}

func (rc *RedisClient) storeWithLimit(ctx context.Context, key string, value interface{}, limitKey string) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Get the limit from config or use default
	limit := int64(999)
	if limitKey == "kline_max_items" {
		limit = 999  // Use config value here if needed
	} else if limitKey == "trades_max_items" {
		limit = 4999 // Use config value here if needed
	}

	pipe := rc.client.Pipeline()
	pipe.LPush(ctx, key, data)
	pipe.LTrim(ctx, key, 0, limit)
	_, err = pipe.Exec(ctx)
	return err
}

func (rc *RedisClient) Close() error {
	return rc.client.Close()
} 