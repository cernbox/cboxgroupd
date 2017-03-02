package redisgrouplooker

import (
	"context"
	"fmt"
	"github.com/cernbox/cboxgroupd/pkg"
	"gopkg.in/redis.v5"
)

// redisgrouplooker is a wrapper around any GroupLooker that will cache
// resglts for a given TTL.
// If the query cannot be found in the cache, it will call the wrapped GroupLooker
// for getting the resglts and it will cache the resglts for the configured TTL
func New(hostname string, port, db int, wrapped pkg.GroupLooker) pkg.GroupLooker {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", hostname, port),
		DB:   db,
	})
	return &groupLooker{
		client:  client,
		wrapped: wrapped,
	}
}

type groupLooker struct {
	client  *redis.Client
	wrapped pkg.GroupLooker
}

// GetUsersInGroups returns the uids (users) members of the given gid
// In redis, the keys follow the pattern <uid>:<gid>, like hugo:cernbox-admins
// To query for all groups of a given user we query redis for the prefix hugo:*
func (gl *groupLooker) GetUsersInGroup(ctx context.Context, gid string) ([]string, error) {
	uids, err := gl.wrapped.GetUsersInGroup(ctx, gid)
	if err != nil {
		return nil, err
	}

	for _, uid := range uids {
		if err := gl.client.SAdd("egroup:"+gid, uid).Err(); err != nil {
			return nil, err
		}
	}
	return uids, nil
}

func (gl *groupLooker) GetUserGroups(ctx context.Context, uid string) ([]string, error) {
	gids, err := gl.wrapped.GetUserGroups(ctx, uid)
	if err != nil {
		return nil, err
	}

	for _, gid := range gids {
		if err := gl.client.SAdd("u:"+uid, gid).Err(); err != nil {
			return nil, err
		}
	}
	return gids, nil
}
