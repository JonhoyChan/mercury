package cache

import (
	"outgoing/app/service/persistence"
	"outgoing/x"
	"outgoing/x/ecode"
	"outgoing/x/rlp"
)

var (
	clientKey = "client:%v"
)

func (c *Cache) GetClient(clientID string) (*persistence.Client, error) {
	key := x.Sprintf(clientKey, clientID)
	clientBytes, err := c.client.Get(key).Bytes()
	if err != nil {
		return nil, err
	}

	var client persistence.Client
	err = rlp.DecodeBytes(clientBytes, &client)
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func (c *Cache) SetClient(clientID string, client *persistence.Client) error {
	if clientID == "" {
		return ecode.NewError("client ID is missing")
	} else if client == nil {
		return ecode.NewError("client can not be nil")
	}

	clientBytes, err := rlp.EncodeToBytes(client)
	if err != nil {
		return err
	}

	key := x.Sprintf(clientKey, clientID)
	if err := c.client.Set(key, clientBytes, 0).Err(); err != nil {
		return err
	}

	return nil
}

func (c *Cache) DeleteClient(clientID string) error {
	return c.client.Del(x.Sprintf(clientKey, clientID)).Err()
}
