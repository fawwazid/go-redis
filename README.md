# go-redis

Install with `go get github.com/fawwazid/go-redis`.

A small configuration layer for Go Redis clients. It supports direct
standalone connections and primary discovery with Redis Sentinel failover,
using [`go-redis`](https://github.com/redis/go-redis) for the Redis protocol
and complete command API. Requires Go 1.27.3.

## Standalone

```go
import (
    "context"
    "time"

    goredis "github.com/fawwazid/go-redis"
)

func storeGreeting(ctx context.Context) error {
    client, err := goredis.New(goredis.Config{
        Mode: goredis.ModeStandalone,
        Addr: "localhost:6379",
    })
    if err != nil {
        return err
    }
    defer client.Close()

    if err := client.Ping(ctx); err != nil {
        return err
    }
    return client.Raw().Set(ctx, "greeting", "hello", time.Minute).Err()
}
```

## High availability with Sentinel

```go
import goredis "github.com/fawwazid/go-redis"

func newSentinelClient(redisUsername, redisPassword string) (*goredis.Client, error) {
    return goredis.New(goredis.Config{
        Mode:          goredis.ModeSentinel,
        MasterName:    "mymaster",
        SentinelAddrs: []string{"sentinel-1:26379", "sentinel-2:26379", "sentinel-3:26379"},
        Username:      redisUsername,
        Password:      redisPassword,
        // Set SentinelUsername and SentinelPassword when Sentinel has separate auth.
    })
}
```

Sentinel mode discovers and follows the current primary during failover. It
uses logical database 0, as Redis Sentinel replication requires. This is
Sentinel support, not Redis Cluster support. Set `TLSConfig` to enable TLS;
certificate validation remains controlled by Go's TLS defaults unless the
application explicitly changes them. Load credentials from a secret store or
environment configuration instead of hard-coding them.

`Config` also accepts pool sizes, `MaxActiveConns`, and dial/read/write
timeouts. `MaxActiveConns` caps connections in the main pool; zero means no
cap. The dedicated pipeline pool is separate and can add up to its configured
`PipelinePoolSize` connections. `PoolSize` is a base size, not a hard limit.
`Client.Raw()` exposes the complete `redis.UniversalClient` API; `Ping` and
`Close` are convenience methods.

License: MIT.
