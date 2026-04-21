package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

const (
	ChanAssetTrades = "asset:%s:trades"
	ChanAssetDepth  = "asset:%s:depth"
	ChanUserOrders  = "user:%s:orders"
)

func AssetTradesChan(symbol string) string { return fmt.Sprintf(ChanAssetTrades, symbol) }
func AssetDepthChan(symbol string) string  { return fmt.Sprintf(ChanAssetDepth, symbol) }
func UserOrdersChan(userID string) string  { return fmt.Sprintf(ChanUserOrders, userID) }

type Bus struct {
	rdb *redis.Client
}

func NewBus(rdb *redis.Client) *Bus {
	return &Bus{rdb: rdb}
}

func (b *Bus) Publish(ctx context.Context, channel string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return b.rdb.Publish(ctx, channel, data).Err()
}

func (b *Bus) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return b.rdb.Subscribe(ctx, channels...)
}

// Cache helpers

func (b *Bus) CacheSet(ctx context.Context, key string, val any) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return b.rdb.Set(ctx, key, data, 0).Err()
}

func (b *Bus) CacheGet(ctx context.Context, key string, dest any) error {
	data, err := b.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (b *Bus) CacheDelete(ctx context.Context, key string) {
	if err := b.rdb.Del(ctx, key).Err(); err != nil {
		log.Printf("cache delete %s: %v", key, err)
	}
}
