package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Queue struct {
	client   *redis.Client
	handlers map[string]Handler
	ctx      context.Context
	cancel   context.CancelFunc
}

type Config struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type Handler func(job *Job) error

type Job struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Data       map[string]interface{} `json:"data"`
	CreatedAt  time.Time              `json:"created_at"`
	Attempts   int                    `json:"attempts"`
	MaxRetries int                    `json:"max_retries"`
}

func New(host string, password string, db int) (*Queue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithCancel(context.Background())

	return &Queue{
		client:   client,
		handlers: make(map[string]Handler),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (q *Queue) RegisterHandler(jobType string, handler Handler) {
	q.handlers[jobType] = handler
}

func (q *Queue) Enqueue(jobType string, data map[string]interface{}, maxRetries int) (*Job, error) {
	job := &Job{
		ID:         generateID(),
		Type:       jobType,
		Data:       data,
		CreatedAt:  time.Now(),
		Attempts:   0,
		MaxRetries: maxRetries,
	}

	jobData, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("job:%s", job.ID)
	if err := q.client.Set(q.ctx, key, jobData, 0).Err(); err != nil {
		return nil, err
	}

	if err := q.client.LPush(q.ctx, "queue", job.ID).Err(); err != nil {
		return nil, err
	}

	return job, nil
}

func (q *Queue) Start() {
	go q.processJobs()
}

func (q *Queue) Stop() {
	q.cancel()
}


func (q *Queue) Shutdown() error {
	q.Stop()
	return q.client.Close()
}


func (q *Queue) IsRunning() bool {
	select {
	case <-q.ctx.Done():
		return false
	default:
		return true
	}
}

func (q *Queue) processJobs() {
	for {
		select {
		case <-q.ctx.Done():
			return
		default:
			jobID, err := q.client.RPop(q.ctx, "queue").Result()
			if err != nil {
				if err == redis.Nil {
					time.Sleep(time.Second)
					continue
				}
				continue
			}

			key := fmt.Sprintf("job:%s", jobID)
			jobData, err := q.client.Get(q.ctx, key).Bytes()
			if err != nil {
				continue
			}

			var job Job
			if err := json.Unmarshal(jobData, &job); err != nil {
				continue
			}

			if handler, ok := q.handlers[job.Type]; ok {
				if err := handler(&job); err != nil {
					job.Attempts++
					if job.Attempts < job.MaxRetries {
						q.client.LPush(q.ctx, "queue", job.ID)
					}
				}
			}

			q.client.Del(q.ctx, key)
		}
	}
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
