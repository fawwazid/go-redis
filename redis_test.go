package goredis

import (
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestNewAcceptsStandaloneConfig(t *testing.T) {
	client, err := New(Config{Addr: "localhost:6379"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.Close()

	if client.Raw() == nil {
		t.Fatal("Raw() returned a nil client")
	}
}

func TestNewAcceptsSentinelConfig(t *testing.T) {
	client, err := New(Config{
		Mode:          ModeSentinel,
		MasterName:    "mymaster",
		SentinelAddrs: []string{"sentinel-1:26379", "sentinel-2:26379"},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer client.Close()

	if client.Raw() == nil {
		t.Fatal("Raw() returned a nil client")
	}
}

// TestNewPassesRetryAndContextTimeoutSettings checks both client topologies retain safety-critical go-redis options.
func TestNewPassesRetryAndContextTimeoutSettings(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "standalone",
			config: Config{
				Mode: ModeStandalone, Addr: "localhost:6379", MaxRetries: -1, ContextTimeoutEnabled: true,
			},
		},
		{
			name: "sentinel",
			config: Config{
				Mode: ModeSentinel, MasterName: "mymaster", SentinelAddrs: []string{"sentinel-1:26379"},
				MaxRetries: -1, ContextTimeoutEnabled: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.config)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			defer client.Close()
			raw, ok := client.Raw().(*redis.Client)
			if !ok {
				t.Fatalf("Raw() type = %T, want *redis.Client", client.Raw())
			}
			options := raw.Options()
			// go-redis normalizes -1 to zero internal retries; zero means no retry loop.
			if options.MaxRetries != 0 || !options.ContextTimeoutEnabled {
				t.Fatalf("options = {MaxRetries: %d, ContextTimeoutEnabled: %t}, want {0, true}", options.MaxRetries, options.ContextTimeoutEnabled)
			}
		})
	}
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "missing standalone address",
		},
		{
			name: "sentinel settings require explicit mode",
			config: Config{
				MasterName:    "mymaster",
				SentinelAddrs: []string{"sentinel-1:26379"},
			},
		},
		{
			name: "unknown mode",
			config: Config{
				Mode: Mode("cluster"),
				Addr: "localhost:6379",
			},
		},
		{
			name: "missing sentinel master name",
			config: Config{
				Mode:          ModeSentinel,
				SentinelAddrs: []string{"sentinel-1:26379"},
			},
		},
		{
			name: "missing sentinel addresses",
			config: Config{
				Mode:       ModeSentinel,
				MasterName: "mymaster",
			},
		},
		{
			name: "empty sentinel address",
			config: Config{
				Mode:          ModeSentinel,
				MasterName:    "mymaster",
				SentinelAddrs: []string{"sentinel-1:26379", " "},
			},
		},
		{
			name: "nonzero sentinel database",
			config: Config{
				Mode:          ModeSentinel,
				MasterName:    "mymaster",
				SentinelAddrs: []string{"sentinel-1:26379"},
				DB:            1,
			},
		},
		{
			name: "negative database",
			config: Config{
				Mode: ModeStandalone,
				Addr: "localhost:6379",
				DB:   -1,
			},
		},
		{
			name: "negative pool size",
			config: Config{
				Mode:     ModeStandalone,
				Addr:     "localhost:6379",
				PoolSize: -1,
			},
		},
		{
			name: "minimum idle connections exceed pool size",
			config: Config{
				Mode:         ModeStandalone,
				Addr:         "localhost:6379",
				PoolSize:     2,
				MinIdleConns: 3,
			},
		},
		{
			name: "retry count below disabled sentinel",
			config: Config{Mode: ModeStandalone, Addr: "localhost:6379", MaxRetries: -2},
		},
		{
			name: "negative timeout",
			config: Config{
				Mode:        ModeStandalone,
				Addr:        "localhost:6379",
				DialTimeout: -1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.config)
			if client != nil {
				_ = client.Close()
			}
			if !errors.Is(err, ErrInvalidConfig) {
				t.Fatalf("New() error = %v, want %v", err, ErrInvalidConfig)
			}
		})
	}
}

func TestNilClientMethods(t *testing.T) {
	var client *Client

	if err := client.Ping(t.Context()); !errors.Is(err, ErrClientUnavailable) {
		t.Fatalf("Ping() error = %v, want %v", err, ErrClientUnavailable)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}
	if client.Raw() != nil {
		t.Fatal("Raw() returned a client for a nil receiver")
	}
}
