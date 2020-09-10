package cache

import (
	"outgoing/x"
	"outgoing/x/database/redis"
	"outgoing/x/ecode"
	"time"
)

const (
	topicSequenceKey     = "topicSequence:%v"
	defaultTopicLifetime = 60 * time.Minute
)

func (c *Cache) GetTopicSequence(topic string) (int64, error) {
	key := x.Sprintf(topicSequenceKey, topic)
	return c.client.Get(key).Int64()
}

func (c *Cache) SetTopicSequence(topic string, sequence int64, lifetime time.Duration) error {
	if topic == "" {
		return ecode.NewError("topic is missing")
	}
	if lifetime == 0 {
		lifetime = defaultTopicLifetime
	}
	key := x.Sprintf(topicSequenceKey, topic)
	if err := c.client.Set(key, sequence, lifetime).Err(); err != nil {
		return err
	}

	return nil
}

func (c *Cache) IncrTopicSequence(topic string) (int64, error) {
	key := x.Sprintf(topicSequenceKey, topic)
	// Make sure the key already exists
	if c.client.Exists(key).Val() != 1 {
		return 0, redis.RedisNil
	}
	sequence, err := c.client.Incr(key).Result()
	if err != nil {
		return 0, err
	}
	return sequence, nil
}
