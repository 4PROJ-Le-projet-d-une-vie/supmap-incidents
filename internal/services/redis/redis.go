package redis

import (
	"context"
	json2 "encoding/json"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type Redis struct {
	log    *slog.Logger
	client *redis.Client
	send   chan redis.Message
	read   chan redis.Message
}

func NewRedis(client *redis.Client, log *slog.Logger) *Redis {
	return &Redis{
		log:    log,
		client: client,
		send:   make(chan redis.Message, 1),
		read:   make(chan redis.Message, 1),
	}
}

func (r *Redis) Run(ctx context.Context) {
	go r.publisher(ctx)
}

func (r *Redis) publisher(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			r.log.Info("stopping redis publisher")
			return
		case msg := <-r.send:
			r.log.Info("message send to redis", "channel", msg.Channel, "message", msg.Payload)
			err := r.client.Publish(ctx, msg.Channel, msg.Payload).Err()
			if err != nil {
				r.log.Error("redis publish message error", "error", err)
			}
		}
	}
}

func (r *Redis) PublishMessage(channel string, payload any) error {
	json, err := json2.Marshal(payload)
	if err != nil {
		return err
	}

	select {
	case r.send <- redis.Message{
		Channel: channel,
		Payload: string(json),
	}:
	default:
		r.log.Warn("channel is full, message dropped")
	}

	return nil
}
