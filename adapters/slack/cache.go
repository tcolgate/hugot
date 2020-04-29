package slack

import (
	"sync"

	client "github.com/slack-go/slack"
)

type cache struct {
	api *client.Client

	cacheLock sync.Mutex
	userCache map[string]*client.User
	chanCache map[string]*client.Channel
}

func newCache(api *client.Client) *cache {
	return &cache{
		api: api,
	}
}

func (c *cache) GetUser(id string) (*client.User, error) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	if c.userCache == nil {
		c.userCache = make(map[string]*client.User)
	}

	if u, ok := c.userCache[id]; ok {
		return u, nil
	}

	u, err := c.api.GetUserInfo(id)
	if err != nil {
		return nil, err
	}

	c.userCache[id] = u
	return u, nil
}

func (c *cache) GetChannel(id string) (*client.Channel, error) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	if c.chanCache == nil {
		c.chanCache = make(map[string]*client.Channel)
	}

	if sc, ok := c.chanCache[id]; ok {
		return sc, nil
	}

	sc, err := c.api.GetChannelInfo(id)
	if err != nil {
		return nil, err
	}

	c.chanCache[id] = sc
	return sc, nil
}
