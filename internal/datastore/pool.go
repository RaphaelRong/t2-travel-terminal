// Package datastore 负责数据库连接池管理。
// 支持两种连接方式：
//   1. 通过 DATABASE_URL 直接连接（本地开发或 Cloud SQL Proxy）
//   2. 通过 GCP Cloud SQL Connector 连接（生产环境）
package datastore

import (
	"context"
	"fmt"
	"net"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/config"
)

// Pool 包装 pgxpool.Pool，并在使用 Cloud SQL Connector 时负责关闭 dialer。
type Pool struct {
	pool   *pgxpool.Pool
	dialer *cloudsqlconn.Dialer
}

// NewPool 根据配置创建数据库连接池。
// 默认使用 GCP Cloud SQL Connector；如需本地开发可设置 DATABASE_URL。
func NewPool(ctx context.Context, cfg *config.Config) (*Pool, error) {
	if cfg.DatabaseInstanceConnectionName != "" {
		return newCloudSQLPool(ctx, cfg)
	}

	if cfg.DatabaseURL != "" {
		return newURLPool(ctx, cfg.DatabaseURL)
	}

	return nil, fmt.Errorf(
		"database not configured: set DATABASE_INSTANCE_CONNECTION_NAME for GCP Cloud SQL " +
			"or DATABASE_URL for direct connection",
	)
}

func newURLPool(ctx context.Context, databaseURL string) (*Pool, error) {
	pgCfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, pgCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	return &Pool{pool: pool}, nil
}

func newCloudSQLPool(ctx context.Context, cfg *config.Config) (*Pool, error) {
	if cfg.DatabaseUser == "" || cfg.DatabasePassword == "" || cfg.DatabaseName == "" {
		return nil, fmt.Errorf(
			"cloud sql requires DATABASE_USER, DATABASE_PASSWORD and DATABASE_NAME",
		)
	}

	dsn := fmt.Sprintf(
		"user=%s password=%s database=%s sslmode=disable",
		cfg.DatabaseUser, cfg.DatabasePassword, cfg.DatabaseName,
	)
	pgCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse cloud sql config: %w", err)
	}

	dialer, err := cloudsqlconn.NewDialer(ctx)
	if err != nil {
		return nil, fmt.Errorf("create cloud sql dialer: %w", err)
	}

	pgCfg.ConnConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.Dial(ctx, cfg.DatabaseInstanceConnectionName)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgCfg)
	if err != nil {
		_ = dialer.Close()
		return nil, fmt.Errorf("create cloud sql pool: %w", err)
	}

	return &Pool{pool: pool, dialer: dialer}, nil
}

// Acquire 从连接池获取一个连接。
func (p *Pool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	return p.pool.Acquire(ctx)
}

// Begin 开启一个事务。
func (p *Pool) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.pool.Begin(ctx)
}

// Exec 执行一条不返回行的 SQL。
func (p *Pool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return p.pool.Exec(ctx, sql, args...)
}

// Query 执行返回多行的查询。
func (p *Pool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return p.pool.Query(ctx, sql, args...)
}

// QueryRow 执行只返回一行的查询。
func (p *Pool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

// Ping 检查数据库连通性。
func (p *Pool) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

// Close 关闭连接池（以及 Cloud SQL Connector dialer）。
func (p *Pool) Close() {
	p.pool.Close()
	if p.dialer != nil {
		_ = p.dialer.Close()
	}
}
