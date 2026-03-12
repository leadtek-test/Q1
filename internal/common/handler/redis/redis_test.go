package redis

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestRedisClientHelpers(t *testing.T) {
	err := SetNX(context.Background(), nil, "k", "v", time.Second)
	if err == nil {
		t.Fatalf("expected nil client error")
	}

	err = Del(context.Background(), nil, "k")
	if err == nil {
		t.Fatalf("expected nil client error")
	}
}

func TestSupplierAndClient(t *testing.T) {
	viper.Set("redis.test.ip", "127.0.0.1")
	viper.Set("redis.test.port", "6379")
	viper.Set("redis.test.pool_size", 1)
	viper.Set("redis.test.max_conn", 1)
	viper.Set("redis.test.conn_timeout", 100)
	viper.Set("redis.test.read_timeout", 100)
	viper.Set("redis.test.write_timeout", 100)

	c := supplier("test")
	if c == nil {
		t.Fatalf("supplier should return redis client")
	}
}
