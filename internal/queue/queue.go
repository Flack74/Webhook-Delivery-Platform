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
