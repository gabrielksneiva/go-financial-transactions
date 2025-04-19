package repositories

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gabrielksneiva/go-financial-transactions/domain"
	"github.com/redis/go-redis/v9"
)

var (
	RedisCli *redis.Client
	RedisCtx = context.Background()
)

const (
	transactionLimit  = 10
	transactionWindow = time.Minute
)

type RedisClient struct{}

type RedisRateLimiter struct {
	Client domain.RedisClientInterface // Interface do cliente Redis
}

// InitRedis conecta ao Redis e testa a conexão
func InitRedis(Host string, DB int) domain.RedisClientInterface {
	RedisCli = redis.NewClient(&redis.Options{
		Addr:     Host,
		Password: "",
		DB:       DB,
	})

	_, err := RedisCli.Ping(RedisCtx).Result()
	if err != nil {
		log.Fatalf("❌ Erro ao conectar no Redis: %v", err)
	}
	fmt.Println("✅ Redis conectado com sucesso.")

	return &RedisClient{}
}

func NewRedisRateLimiter(client domain.RedisClientInterface) *RedisRateLimiter {
	return &RedisRateLimiter{Client: client}
}

func (r *RedisClient) Get(ctx context.Context, key string) (int, error) {
	val, err := RedisCli.Get(RedisCtx, key).Int()
	if err != nil && err != redis.Nil {
		return 0, fmt.Errorf("erro ao obter valor do Redis: %w", err)
	}
	return val, nil
}

func (r *RedisClient) Set(ctx context.Context, key string, value int) error {
	_, err := RedisCli.Set(RedisCtx, key, value, 0).Result()
	if err != nil {
		return fmt.Errorf("erro ao definir valor no Redis: %w", err)
	}
	return nil
}

func (r *RedisClient) Incr(ctx context.Context, key string) (int, error) {
	val, err := RedisCli.Incr(RedisCtx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("erro ao incrementar valor no Redis: %w", err)
	}
	return int(val), nil
}

func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	_, err := RedisCli.Expire(RedisCtx, key, expiration).Result()
	return err
}

func (r *RedisRateLimiter) CheckTransactionRateLimit(userID uint) error {
	key := fmt.Sprintf("rate_limit:user:%d", userID)

	val, err := r.Client.Get(RedisCtx, key)
	if err != nil {
		return fmt.Errorf("erro no Redis: %w", err)
	}

	if val >= transactionLimit {
		return fmt.Errorf("limite de transações atingido")
	}

	_, err = r.Client.Incr(RedisCtx, key)
	if err != nil {
		return err
	}

	return r.Client.Expire(RedisCtx, key, transactionWindow)
}
