package goredis

import (
	"errors"
	"testing"
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
