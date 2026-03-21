package queue

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client              *redis.Client
	mainQueueName       string
	processingQueueName string
	delayedQueueName    string
}

func NewRedisQueue(
	client *redis.Client,
	mainQueueName string,
	processingQueueName string,
	delayedQueueName string,
) *RedisQueue {
	return &RedisQueue{
		client:              client,
		mainQueueName:       mainQueueName,
		processingQueueName: processingQueueName,
		delayedQueueName:    delayedQueueName,
	}
}

func (q *RedisQueue) Enqueue(ctx context.Context, deliveryID string) error {
	return q.client.LPush(ctx, q.mainQueueName, deliveryID).Err()
}

func (q *RedisQueue) EnqueueDelayed(ctx context.Context, deliveryID string, deliverAt time.Time) error {
	return q.client.ZAdd(ctx, q.delayedQueueName, redis.Z{
		Score:  float64(deliverAt.UnixMilli()),
		Member: deliveryID,
	}).Err()
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

func (q *RedisQueue) Nack(ctx context.Context, deliveryID string) error {
	pipe := q.client.Pipeline()
	pipe.LRem(ctx, q.processingQueueName, 0, deliveryID)
	pipe.RPush(ctx, q.mainQueueName, deliveryID)
	_, err := pipe.Exec(ctx)
	return err
}

func (q *RedisQueue) NackDelayed(ctx context.Context, deliveryID string, deliverAt time.Time) error {
	pipe := q.client.Pipeline()
	pipe.LRem(ctx, q.processingQueueName, 0, deliveryID)
	pipe.ZAdd(ctx, q.delayedQueueName, redis.Z{
		Score:  float64(deliverAt.UnixMilli()),
		Member: deliveryID,
	})
	_, err := pipe.Exec(ctx)
	return err
}

func (q *RedisQueue) PromoteDue(ctx context.Context, now time.Time, batchSize int64) error {
	if batchSize <= 0 {
		batchSize = 100
	}

	dueIDs, err := q.client.ZRangeByScore(ctx, q.delayedQueueName, &redis.ZRangeBy{
		Min:   "-inf",
		Max:   strconv.FormatInt(now.UnixMilli(), 10),
		Count: batchSize,
	}).Result()
	if err != nil || len(dueIDs) == 0 {
		return err
	}

	pipe := q.client.TxPipeline()
	for _, deliveryID := range dueIDs {
		pipe.ZRem(ctx, q.delayedQueueName, deliveryID)
		pipe.RPush(ctx, q.mainQueueName, deliveryID)
	}
	_, err = pipe.Exec(ctx)
	return err
}

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

	if err := q.PromoteDue(ctx, time.Now().UTC(), 1000); err != nil {
		return err
	}

	return nil
}
