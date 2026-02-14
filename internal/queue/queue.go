package queue

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client              *redis.Client
	mainQueueName       string
	processingQueueName string
}

func NewRedisQueue(client *redis.Client, mainQueueName string, processingQueueName string) *RedisQueue {
	return &RedisQueue{
		client:              client,
		mainQueueName:       mainQueueName,
		processingQueueName: processingQueueName,
	}
}

func (q *RedisQueue) Enqueue(ctx context.Context, deliveryID string) error {
	return q.client.LPush(ctx, q.mainQueueName, deliveryID).Err()
}

func (q *RedisQueue) Reserve(ctx context.Context) (string, error) {
	deliveryID, err := q.client.BRPopLPush(ctx, q.mainQueueName, q.processingQueueName, 0).Result()
	if err != nil {
		return "", err
	}
	return deliveryID, nil
}

func (q *RedisQueue) Ack(ctx context.Context, deliveryID string) error {
	return q.client.LRem(ctx, q.processingQueueName, 0, deliveryID).Err()
}

// Nack moves a job back to the main queue for retry
func (q *RedisQueue) Nack(ctx context.Context, deliveryID string) error {
	pipe := q.client.Pipeline()
	pipe.LRem(ctx, q.processingQueueName, 0, deliveryID)
	pipe.RPush(ctx, q.mainQueueName, deliveryID)
	_, err := pipe.Exec(ctx)
	return err
}

// RecoverStalled moves jobs from processing back to main queue (run on startup)
func (q *RedisQueue) RecoverStalled(ctx context.Context) error {
	for {
		deliveryID, err := q.client.RPopLPush(ctx, q.processingQueueName, q.mainQueueName).Result()
		if err != nil {
			break
		}
		if deliveryID == "" {
			break
		}
	}
	return nil
}
