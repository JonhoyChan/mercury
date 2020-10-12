package cache

import (
	"mercury/x"
	"time"

	"github.com/go-redis/redis/v7"
)

const (
	// keys
	userSessionServerKey = "userSessionServer:%s"
	sessionServerKey     = "sessionServer:%s"

	// scripts
	addMappingLUA = `
		if redis.call("HEXISTS", KEYS[1], ARGV[1]) == 1 then
            return 0
        end
		redis.call("HSET", KEYS[1], ARGV[1], ARGV[2])
		redis.call("EXPIRE", KEYS[1], ARGV[3])
		redis.call("SET", KEYS[2], ARGV[2]);
        redis.call("EXPIRE", KEYS[2], ARGV[3])
        return 1
    `
)

// Mapping expiration time
var mappingExpire = 1800 * time.Second

// key: uid; field: sid; value: serverID
func (c *Cache) AddMapping(uid, sid, serverID string) error {
	keys := []string{
		x.Sprintf(userSessionServerKey, uid),
		x.Sprintf(sessionServerKey, sid),
	}
	args := []interface{}{sid, serverID, mappingExpire.Seconds()}
	err := redis.NewScript(addMappingLUA).Run(c.client, keys, args...).Err()
	return err
}

func (c *Cache) ExpireMapping(uid, sid string) (bool, error) {
	var (
		expired bool
		err     error
	)
	expired, err = c.client.Expire(x.Sprintf(userSessionServerKey, uid), mappingExpire).Result()
	if err != nil {
		return false, err
	}
	if !expired {
		return expired, err
	}

	expired, err = c.client.Expire(x.Sprintf(sessionServerKey, sid), mappingExpire).Result()
	if err != nil {
		return false, err
	}

	return expired, nil
}

// Delete the mapping
func (c *Cache) DeleteMapping(uid, sid string) error {
	var err error
	if err = c.client.HDel(x.Sprintf(userSessionServerKey, uid), sid).Err(); err != nil {
		return err
	}

	if err = c.client.Del(x.Sprintf(sessionServerKey, sid)).Err(); err != nil {
		return err
	}

	return nil
}

func (c *Cache) GetSessions(uids ...string) (map[string]string, []string, error) {
	sessions := make(map[string]string)
	var onlineUIDs []string
	for _, uid := range uids {
		result, err := c.client.HGetAll(x.Sprintf(userSessionServerKey, uid)).Result()
		if err != nil {
			return nil, nil, err
		}

		if len(result) > 0 {
			onlineUIDs = append(onlineUIDs, uid)
		}

		for k, v := range result {
			sessions[k] = v
		}
	}

	return sessions, onlineUIDs, nil
}

func (c *Cache) GetServerIDs(sids ...string) ([]string, error) {
	var servers []string
	if len(sids) > 0 {
		var keys []string
		for _, sid := range sids {
			keys = append(keys, x.Sprintf(sessionServerKey, sid))
		}
		result, err := c.client.MGet(keys...).Result()
		if err != nil {
			return nil, err
		}

		for _, v := range result {
			var serverID string
			if v != nil {
				serverID = v.(string)
			}
			servers = append(servers, serverID)
		}
	}

	return servers, nil
}
