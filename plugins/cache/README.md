# flux Cache Plugin

A Redis-based caching plugin for the flux framework.

## Features

- Simple key-value caching with Redis
- Support for TTL (Time To Live)
- Atomic operations (Increment, Decrement)
- Conditional operations (SetNX)
- Cache prefixing for isolation
- Automatic connection management

## Installation

1. Add the Redis dependency to your project:
```bash
go get github.com/redis/go-redis/v9
```

2. Configure the plugin in your `flux.yaml`:
```yaml
plugins:
  cache:
    host: localhost
    port: 6379
    password: ""  
    db: 0
    prefix: "flux:cache:"
```

## Usage

```go
import "github.com/Fluxgo/flux/plugins/cache"

// In your controller or service
func (c *YourController) SomeAction(ctx *flux.Context) error {
    // Get the cache plugin
    cachePlugin := c.App.GetPlugin("cache").(*cache.CachePlugin)
    
    // Set a value
    err := cachePlugin.Set(ctx.Context, "key", "value", time.Hour)
    if err != nil {
        return err
    }
    
    // Get a value
    var value string
    err = cachePlugin.Get(ctx.Context, "key", &value)
    if err != nil {
        return err
    }
    
    // Check if key exists
    exists, err := cachePlugin.Exists(ctx.Context, "key")
    if err != nil {
        return err
    }
    
    // Increment a counter
    count, err := cachePlugin.Increment(ctx.Context, "counter")
    if err != nil {
        return err
    }
    
    // Delete a key
    err = cachePlugin.Delete(ctx.Context, "key")
    if err != nil {
        return err
    }
    
    // Clear all cache
    err = cachePlugin.Clear(ctx.Context)
    if err != nil {
        return err
    }
    
    return nil
}
```

## Available Methods

- `Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error`
- `Get(ctx context.Context, key string, value interface{}) error`
- `Delete(ctx context.Context, key string) error`
- `Clear(ctx context.Context) error`
- `Exists(ctx context.Context, key string) (bool, error)`
- `Increment(ctx context.Context, key string) (int64, error)`
- `Decrement(ctx context.Context, key string) (int64, error)`
- `SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error)`
- `GetOrSet(ctx context.Context, key string, value interface{}, ttl time.Duration, fn func() (interface{}, error)) error`

## Configuration Options

| Option   | Type    | Default     | Description                    |
|----------|---------|-------------|--------------------------------|
| host     | string  | localhost   | Redis server host              |
| port     | int     | 6379        | Redis server port              |
| password | string  | ""          | Redis server password          |
| db       | int     | 0           | Redis database number          |
| prefix   | string  | flux:cache:| Prefix for all cache keys      |

## License

MIT License 
