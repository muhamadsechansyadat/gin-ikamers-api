package health

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type checkerFunc struct {
	name string
	fn   func(ctx context.Context) error
}

func (c *checkerFunc) Name() string                    { return c.name }
func (c *checkerFunc) Check(ctx context.Context) error { return c.fn(ctx) }

func NewChecker(name string, fn func(ctx context.Context) error) Checker {
	return &checkerFunc{name: name, fn: fn}
}

func Postgres(name string, db *sql.DB) Checker {
	return NewChecker(name, func(ctx context.Context) error {
		return db.PingContext(ctx)
	})
}

func Redis(name string, rdb *redis.Client) Checker {
	return NewChecker(name, func(ctx context.Context) error {
		return rdb.Ping(ctx).Err()
	})
}

func SMTP(name, host string, port int) Checker {
	addr := fmt.Sprintf("%s:%d", host, port)
	return NewChecker(name, func(ctx context.Context) error {
		var d net.Dialer
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			return err
		}
		return conn.Close()
	})
}

func HTTP(name, url string) Checker {
	client := &http.Client{Timeout: 5 * time.Second}
	return NewChecker(name, func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 500 {
			return fmt.Errorf("upstream returned %d", resp.StatusCode)
		}
		return nil
	})
}
