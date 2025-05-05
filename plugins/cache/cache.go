package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Fluxgo/flux/pkg/flux"
	"github.com/redis/go-redis/v9"
)


type CachePlugin struct {
	app    *flux.Application
	client *redis.Client
	prefix string
}


type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	Prefix   string `yaml:"prefix"`
}


func New(app *flux.Application, config *Config) (*CachePlugin, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &CachePlugin{
		app:    app,
		client: client,
		prefix: config.Prefix,
	}, nil
}


func (p *CachePlugin) Shutdown() error {
	return p.client.Close()
}


func (p *CachePlugin) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return p.client.Set(ctx, p.prefix+key, data, ttl).Err()
}


func (p *CachePlugin) Get(ctx context.Context, key string, value interface{}) error {
	data, err := p.client.Get(ctx, p.prefix+key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return flux.ErrNotFound
		}
		return err
	}

	return json.Unmarshal(data, value)
}


func (p *CachePlugin) Delete(ctx context.Context, key string) error {
	return p.client.Del(ctx, p.prefix+key).Err()
}


func (p *CachePlugin) Clear(ctx context.Context) error {
	pattern := p.prefix + "*"
	iter := p.client.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		if err := p.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	return iter.Err()
}


func (p *CachePlugin) Exists(ctx context.Context, key string) (bool, error) {
	exists, err := p.client.Exists(ctx, p.prefix+key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}


func (p *CachePlugin) Increment(ctx context.Context, key string) (int64, error) {
	return p.client.Incr(ctx, p.prefix+key).Result()
}


func (p *CachePlugin) Decrement(ctx context.Context, key string) (int64, error) {
	return p.client.Decr(ctx, p.prefix+key).Result()
}


func (p *CachePlugin) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return p.client.SetNX(ctx, p.prefix+key, data, ttl).Result()
}


func (p *CachePlugin) GetOrSet(ctx context.Context, key string, value interface{}, ttl time.Duration, fn func() (interface{}, error)) error {
	
	err := p.Get(ctx, key, value)
	if err == nil {
		return nil
	}

	if err != flux.ErrNotFound {
		return err
	}

	
	newValue, err := fn()
	if err != nil {
		return err
	}

	
	if err := p.Set(ctx, key, newValue, ttl); err != nil {
		return err
	}

	
	data, err := json.Marshal(newValue)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, value)
} 
