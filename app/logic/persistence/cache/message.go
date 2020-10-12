package cache

import (
	"mercury/x"
	"strconv"
)

const (
	// keys
	userTopicSequenceKey = "userTopicLastSequence:%s"
	userTopicsKey        = "userTopics:%s"
)

func (c *Cache) SetUserTopicLastSequence(uid, topic string, sequence int64) error {
	return c.client.HSet(x.Sprintf(userTopicSequenceKey, uid), topic, sequence).Err()
}

func (c *Cache) GetUserTopicLastSequence(uid, topic string) (int64, error) {
	result, err := c.client.HGet(x.Sprintf(userTopicSequenceKey, uid), topic).Int64()
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (c *Cache) GetUserTopicsLastSequence(uid string) (map[string]int64, error) {
	topics := make(map[string]int64)
	result, err := c.client.HGetAll(x.Sprintf(userTopicSequenceKey, uid)).Result()
	if err != nil {
		return nil, err
	}

	for k, v := range result {
		sequence, _ := strconv.ParseInt(v, 10, 64)
		topics[k] = sequence
	}

	return topics, nil
}

func (c *Cache) SetUsersTopic(uids []string, topic string) error {
	for _, uid := range uids {
		key := x.Sprintf(userTopicsKey, uid)
		if !c.client.HExists(key, topic).Val() {
			if err := c.client.SAdd(key, topic).Err(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Cache) GetUserTopics(uid string) ([]string, error) {
	return c.client.SMembers(x.Sprintf(userTopicsKey, uid)).Result()
}
