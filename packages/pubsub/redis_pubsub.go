// Package pubsub provides Redis-based pub/sub for job updates across services.
package pubsub

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// JobUpdatesChannel is the Redis channel for job updates.
	JobUpdatesChannel = "job:updates"
)

// JobUpdate represents a job status change event.
type JobUpdate struct {
	JobID      string     `json:"job_id"`
	TenantID   string     `json:"tenant_id"`
	Type       string     `json:"type"`
	Input      string     `json:"input"`
	Status     string     `json:"status"`
	Result     *string    `json:"result,omitempty"`
	Error      *string    `json:"error,omitempty"`
	Provider   *string    `json:"provider,omitempty"`
	TokensIn   int        `json:"tokens_in"`
	TokensOut  int        `json:"tokens_out"`
	Cost       float64    `json:"cost"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// Publisher publishes job updates to Redis.
type Publisher struct {
	client *redis.Client
}

// NewPublisher creates a new Redis publisher.
func NewPublisher(redisURL string) (*Publisher, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Publisher{client: client}, nil
}

// Publish sends a job update event.
func (p *Publisher) Publish(ctx context.Context, update *JobUpdate) error {
	data, err := json.Marshal(update)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, JobUpdatesChannel, data).Err()
}

// Close closes the Redis connection.
func (p *Publisher) Close() error {
	return p.client.Close()
}

// Subscriber subscribes to job updates from Redis.
type Subscriber struct {
	client *redis.Client
	pubsub *redis.PubSub
}

// NewSubscriber creates a new Redis subscriber.
func NewSubscriber(redisURL string) (*Subscriber, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Subscriber{client: client}, nil
}

// Subscribe starts listening for job updates.
// Returns a channel that receives JobUpdate events.
func (s *Subscriber) Subscribe(ctx context.Context) (<-chan *JobUpdate, error) {
	s.pubsub = s.client.Subscribe(ctx, JobUpdatesChannel)

	// Wait for subscription confirmation
	_, err := s.pubsub.Receive(ctx)
	if err != nil {
		return nil, err
	}

	updates := make(chan *JobUpdate, 100)

	go func() {
		defer close(updates)
		ch := s.pubsub.Channel()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}

				var update JobUpdate
				if err := json.Unmarshal([]byte(msg.Payload), &update); err != nil {
					continue // Skip malformed messages
				}

				select {
				case updates <- &update:
				default:
					// Channel full, skip
				}
			}
		}
	}()

	return updates, nil
}

// Close closes the Redis connection.
func (s *Subscriber) Close() error {
	if s.pubsub != nil {
		s.pubsub.Close()
	}
	return s.client.Close()
}
