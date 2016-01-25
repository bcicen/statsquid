package main

import (
	"gopkg.in/redis.v3"
)

type Transport struct {
	redis   *redis.Client
	channel string
}

func newTransport(redisHost string) (*Transport, error) {
	redisOpts := &redis.Options{Addr: redisHost}

	transport := &Transport{
		redis:   redis.NewClient(redisOpts),
		channel: "statsquid",
	}

	_, err := transport.redis.Ping().Result()
	return transport, err
}

func (t *Transport) Publish(s string) error {
	return t.redis.Publish(t.channel, s).Err()
}

func (t *Transport) Subscribe() (*redis.PubSub, error) {
	return t.redis.Subscribe(t.channel)
}
