package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

var sut *SUT
var ctx context.Context
var defaultPass = "password"

func TestMain(m *testing.M) {
	testCtx := context.Background()
	pgContainer, pgConn, pgErr := startPostgres(testCtx)
	if pgErr != nil {
		fmt.Printf("failed to start postgres: %v\n", pgErr)
		os.Exit(1)
	}
	redisContainer, redisAddr, redisErr := startRedis(testCtx)
	if redisErr != nil {
		fmt.Printf("failed to start redis: %v\n", redisErr)
		_ = pgContainer.Terminate(testCtx)
		os.Exit(1)
	}
	if err := os.Setenv("GOOSE_DBSTRING", pgConn); err != nil {
		fmt.Printf("failed to set GOOSE_DBSTRING: %v\n", err)
		_ = redisContainer.Terminate(testCtx)
		_ = pgContainer.Terminate(testCtx)
		os.Exit(1)
	}
	if err := os.Setenv("REDIS_URL", redisAddr); err != nil {
		fmt.Printf("failed to set REDIS_URL: %v\n", err)
		_ = redisContainer.Terminate(testCtx)
		_ = pgContainer.Terminate(testCtx)
		os.Exit(1)
	}

	sut = NewSUT()
	ctx = context.Background()
	code := m.Run()
	sut.Server.Close()
	_ = redisContainer.Terminate(testCtx)
	_ = pgContainer.Terminate(testCtx)
	os.Exit(code)
}

func startPostgres(ctx context.Context) (*postgres.PostgresContainer, string, error) {
	container, err := postgres.Run(
		ctx,
		"postgres:15.3-alpine",
		postgres.WithDatabase("gin_alpine_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("ps_secret"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, "", err
	}
	conn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, "", err
	}
	return container, conn, nil
}

func startRedis(ctx context.Context) (*redis.RedisContainer, string, error) {
	container, err := redis.Run(
		ctx,
		"redis:alpine",
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp").WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, "", err
	}
	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, "", err
	}
	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, "", err
	}
	return container, fmt.Sprintf("%s:%s", host, port.Port()), nil
}
