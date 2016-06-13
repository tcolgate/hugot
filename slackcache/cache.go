package slackcache

import (
	"sync"

	"github.com/nlopes/slack"
)

type Cache struct {
	api *slack.Client

	cacheLock sync.Mutex
	userCache map[string]*slack.User
	chanCache map[string]*slack.Channel
}

func New(api *slack.Client) *Cache {
	return &Cache{
		api: api,
	}
}

func (c *Cache) GetUser(id string) (*slack.User, error) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	if c.userCache == nil {
		c.userCache = make(map[string]*slack.User)
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

func (c *Cache) GetChannel(id string) (*slack.Channel, error) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	if c.chanCache == nil {
		c.chanCache = make(map[string]*slack.Channel)
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
