package mattermost

import (
	"sync"

	mm "github.com/mattermost/mattermost-server/model"
)

type cache struct {
	api  *mm.Client4
	team *mm.Team

	cacheLock     sync.Mutex
	userCache     map[string]*mm.User
	userNameCache map[string]*mm.User
	chanCache     map[string]*mm.Channel
	chanNameCache map[string]*mm.Channel
}

func newCache(api *mm.Client4, team *mm.Team) *cache {
	return &cache{
		api:  api,
		team: team,
	}
}

func (c *cache) GetUser(id string) (*mm.User, error) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	if c.userCache == nil {
		c.userCache = make(map[string]*mm.User)
	}
	if c.userNameCache == nil {
		c.userNameCache = make(map[string]*mm.User)
	}

	if u, ok := c.userCache[id]; ok {
		return u, nil
	}

	u, resp := c.api.GetUser(id, "")
	if resp.Error != nil {
		return nil, resp.Error
	}

	c.userCache[id] = u
	c.userNameCache[u.Username] = u
	return u, nil
}

func (c *cache) GetUserByName(name string) (*mm.User, error) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	if c.userCache == nil {
		c.userCache = make(map[string]*mm.User)
	}
	if c.userNameCache == nil {
		c.userNameCache = make(map[string]*mm.User)
	}

	if u, ok := c.userNameCache[name]; ok {
		return u, nil
	}

	u, resp := c.api.GetUserByUsername(name, "")
	if resp.Error != nil {
		return nil, resp.Error
	}

	c.userNameCache[name] = u
	c.userCache[u.Id] = u
	return u, nil
}

func (c *cache) GetChannel(id string) (*mm.Channel, error) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	if c.chanCache == nil {
		c.chanCache = make(map[string]*mm.Channel)
	}
	if c.chanNameCache == nil {
		c.chanNameCache = make(map[string]*mm.Channel)
	}

	if sc, ok := c.chanCache[id]; ok {
		return sc, nil
	}

	sc, resp := c.api.GetChannel(id, "")
	if resp.Error != nil {
		return nil, resp.Error
	}

	c.chanCache[id] = sc
	return sc, nil
}

func (c *cache) GetChannelByName(name string) (*mm.Channel, error) {
	c.cacheLock.Lock()
	defer c.cacheLock.Unlock()

	if c.chanCache == nil {
		c.chanCache = make(map[string]*mm.Channel)
	}
	if c.chanNameCache == nil {
		c.chanNameCache = make(map[string]*mm.Channel)
	}

	if sc, ok := c.chanNameCache[name]; ok {
		return sc, nil
	}

	sc, resp := c.api.GetChannelByName(name, c.team.Id, "")
	if resp.Error != nil {
		return nil, resp.Error
	}

	c.chanNameCache[name] = sc
	c.chanCache[sc.Id] = sc
	return sc, nil
}
