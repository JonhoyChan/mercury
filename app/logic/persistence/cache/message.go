package cache

import "outgoing/x"

const userTopicMessageKey = "userTopicMessage:%s"

func (c *Cache) SetTopicMessageUnread(uid, topic string, unread int64) error {
	_, err := c.client.HSet(x.Sprintf(userTopicMessageKey, uid), topic, unread).Result()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) GetTopicMessageUnread(uid, topic string) (int64, error) {
	result, err := c.client.HGet(x.Sprintf(userTopicMessageKey, uid), topic).Int64()
	if err != nil {
		return 0, err
	}

	return result, nil
}

