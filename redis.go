// Package goredis provides validated Redis client configuration for standalone
// Redis and Redis Sentinel deployments. The underlying go-redis client is
// available for commands beyond the small convenience API in this package.
package goredis

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// Mode selects how the client discovers the Redis server to use.
type Mode string

const (
	// ModeStandalone connects directly to one Redis server.
	ModeStandalone Mode = "standalone"
	// ModeSentinel discovers the current primary through Redis Sentinel and
	// follows primary changes during failover.
	ModeSentinel Mode = "sentinel"
)

var (
	// ErrInvalidConfig indicates that a Config value is incomplete or invalid.
	ErrInvalidConfig = errors.New("redis: invalid configuration")
	// ErrClientUnavailable indicates that a method was called on a nil Client.
	ErrClientUnavailable = errors.New("redis: client is unavailable")
)

// Config contains connection settings for standalone Redis or Redis Sentinel.
// Credentials are passed to go-redis and are never logged by this package.
type Config struct {
	// Mode defaults to ModeStandalone when empty.
	Mode Mode

	// Addr is the Redis server address in standalone mode, such as
	// "localhost:6379".
	Addr string

	// SentinelAddrs and MasterName identify the Sentinel deployment. Sentinel
	// addresses are commonly "localhost:26379". Redis Sentinel deployments
	// support database 0 only.
	SentinelAddrs []string
	MasterName    string

	// Username and Password authenticate to the Redis server. SentinelUsername
	// and SentinelPassword authenticate separately to Sentinel.
	Username         string
	Password         string
	SentinelUsername string
	SentinelPassword string

	// DB selects a Redis logical database. It must be zero in Sentinel mode.
	DB int

	// TLSConfig enables TLS when non-nil. The value is cloned before it is
	// passed to go-redis.
	TLSConfig *tls.Config

	// Optional connection pool and operation settings. Zero leaves the
	// corresponding go-redis default in place. MaxActiveConns limits active
	// connections in the main pool; zero leaves that limit unlimited.
	PoolSize       int
	MinIdleConns   int
	MaxActiveConns int
	DialTimeout    time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration

	// MaxRetries sets the maximum retry attempts. -1 disables retries; zero uses the go-redis default.
	MaxRetries int

	// ContextTimeoutEnabled applies context deadlines to Redis socket read and write timeouts.
	ContextTimeoutEnabled bool
}

// Client owns a go-redis client configured for standalone Redis or Sentinel.
// Use Raw to access the complete go-redis command API.
type Client struct {
	raw redis.UniversalClient
}

// New validates cfg and creates a Redis client. It does not establish a
// connection; call Ping with an application-controlled context to check
// connectivity.
func New(cfg Config) (*Client, error) {
	if cfg.Mode == "" {
		if cfg.MasterName != "" || len(cfg.SentinelAddrs) != 0 {
			return nil, invalidConfig("set ModeSentinel when configuring Sentinel")
		}
		cfg.Mode = ModeStandalone
	}

	if err := validateCommon(cfg); err != nil {
		return nil, err
	}

	var raw redis.UniversalClient
	switch cfg.Mode {
	case ModeStandalone:
		addr := strings.TrimSpace(cfg.Addr)
		if addr == "" {
			return nil, invalidConfig("Addr is required in standalone mode")
		}
		raw = redis.NewClient(&redis.Options{
			Addr:           addr,
			Username:       cfg.Username,
			Password:       cfg.Password,
			DB:             cfg.DB,
			TLSConfig:      cloneTLSConfig(cfg.TLSConfig),
			PoolSize:       cfg.PoolSize,
			MinIdleConns:   cfg.MinIdleConns,
			MaxActiveConns: cfg.MaxActiveConns,
			DialTimeout:    cfg.DialTimeout,
			ReadTimeout:    cfg.ReadTimeout,
			WriteTimeout:   cfg.WriteTimeout,
			MaxRetries:     cfg.MaxRetries,
			ContextTimeoutEnabled: cfg.ContextTimeoutEnabled,
		})
	case ModeSentinel:
		if cfg.DB != 0 {
			return nil, invalidConfig("DB must be 0 in Sentinel mode")
		}
		if strings.TrimSpace(cfg.MasterName) == "" {
			return nil, invalidConfig("MasterName is required in Sentinel mode")
		}
		if len(cfg.SentinelAddrs) == 0 {
			return nil, invalidConfig("SentinelAddrs is required in Sentinel mode")
		}
		addrs := make([]string, len(cfg.SentinelAddrs))
		for i, addr := range cfg.SentinelAddrs {
			addrs[i] = strings.TrimSpace(addr)
			if addrs[i] == "" {
				return nil, invalidConfig("SentinelAddrs cannot contain an empty address")
			}
		}
		raw = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:       strings.TrimSpace(cfg.MasterName),
			SentinelAddrs:    addrs,
			Username:         cfg.Username,
			Password:         cfg.Password,
			SentinelUsername: cfg.SentinelUsername,
			SentinelPassword: cfg.SentinelPassword,
			DB:               cfg.DB,
			TLSConfig:        cloneTLSConfig(cfg.TLSConfig),
			PoolSize:         cfg.PoolSize,
			MinIdleConns:     cfg.MinIdleConns,
			MaxActiveConns:   cfg.MaxActiveConns,
			DialTimeout:      cfg.DialTimeout,
			ReadTimeout:      cfg.ReadTimeout,
			WriteTimeout:     cfg.WriteTimeout,
		})
	default:
		return nil, invalidConfig("Mode must be ModeStandalone or ModeSentinel")
	}

	return &Client{raw: raw}, nil
}

func validateCommon(cfg Config) error {
	if cfg.DB < 0 {
		return invalidConfig("DB cannot be negative")
	}
	if cfg.PoolSize < 0 || cfg.MinIdleConns < 0 || cfg.MaxActiveConns < 0 {
		return invalidConfig("pool sizes and MaxActiveConns cannot be negative")
	}
	if cfg.PoolSize > 0 && cfg.MinIdleConns > cfg.PoolSize {
		return invalidConfig("MinIdleConns cannot exceed PoolSize")
	}
	if cfg.MaxActiveConns > 0 && cfg.MinIdleConns > cfg.MaxActiveConns {
		return invalidConfig("MinIdleConns cannot exceed MaxActiveConns")
	}
	if cfg.DialTimeout < 0 || cfg.ReadTimeout < 0 || cfg.WriteTimeout < 0 {
		return invalidConfig("timeouts cannot be negative")
	}
	if cfg.MaxRetries < -1 {
		return invalidConfig("MaxRetries cannot be less than -1")
	}
	return nil
}

func invalidConfig(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidConfig, message)
}

func cloneTLSConfig(cfg *tls.Config) *tls.Config {
	if cfg == nil {
		return nil
	}
	return cfg.Clone()
}

// Raw returns the underlying go-redis client, which supports the full
// go-redis command API. It returns nil when called on a nil Client.
func (c *Client) Raw() redis.UniversalClient {
	if c == nil {
		return nil
	}
	return c.raw
}

// Ping checks connectivity and returns the Redis server's ping error, if any.
func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.raw == nil {
		return ErrClientUnavailable
	}
	return c.raw.Ping(ctx).Err()
}

// Close closes the underlying client and its connection pool.
func (c *Client) Close() error {
	if c == nil || c.raw == nil {
		return nil
	}
	return c.raw.Close()
}
