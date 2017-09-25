package redisgrouplooker

import (
	"context"
	"fmt"
	"github.com/cernbox/cboxgroupd/pkg"
	"gopkg.in/redis.v5"
	"time"
)

// redisgrouplooker is a wrapper around any GroupLooker that will cache
// resglts for a given TTL.
// If the query cannot be found in the cache, it will call the wrapped GroupLooker
// for getting the resglts and it will cache the resglts for the configured TTL
func New(hostname string, port, db, ttl int, wrapped pkg.GroupLooker) pkg.GroupLooker {
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", hostname, port),
		DB:   db,
	})
	return &groupLooker{
		ttl:     ttl,
		client:  client,
		wrapped: wrapped,
	}
}

type groupLooker struct {
	ttl     int
	client  *redis.Client
	wrapped pkg.GroupLooker
}

// GetUsersInGroups returns the uids (users) members of the given gid
// In redis, the keys follow the pattern <uid>:<gid>, like hugo:cernbox-admins
// To query for all groups of a given user we query redis for the prefix hugo:*
func (gl *groupLooker) GetUsersInGroup(ctx context.Context, gid string, cached bool) ([]string, error) {
	key := fmt.Sprintf("egroup:%s", gid)

	// check if it is cached
	if cached {
		if gl.client.Exists(key).Val() == true {
			cmd := gl.client.SMembers(key)
			if cmd.Err() == nil {
				uids, err := cmd.Result()
				if err == nil {
					return uids, nil
				}
			}
		}
	}

	uids, err := gl.wrapped.GetUsersInGroup(ctx, gid, false)
	if err != nil {
		return nil, err
	}

	pipeline := gl.client.TxPipeline()
	defer pipeline.Close()
	for _, uid := range uids {
		pipeline.SAdd(key, uid)
	}
	pipeline.Expire(key, time.Second*time.Duration(gl.ttl))
	_, err = pipeline.Exec()
	if err != nil {
		return nil, err
	}
	return uids, nil
}

func (gl *groupLooker) GetUsersInComputingGroup(ctx context.Context, gid string, cached bool) ([]string, error) {
	key := fmt.Sprintf("unixgroup:%s", gid)

	// check if it is cached
	if cached {
		if gl.client.Exists(key).Val() == true {
			cmd := gl.client.SMembers(key)
			if cmd.Err() == nil {
				uids, err := cmd.Result()
				if err == nil {
					return uids, nil
				}
			}
		}
	}

	uids, err := gl.wrapped.GetUsersInComputingGroup(ctx, gid, false)
	if err != nil {
		return nil, err
	}

	pipeline := gl.client.TxPipeline()
	defer pipeline.Close()
	for _, uid := range uids {
		pipeline.SAdd(key, uid)
	}
	pipeline.Expire(key, time.Second*time.Duration(gl.ttl))
	_, err = pipeline.Exec()
	if err != nil {
		return nil, err
	}
	return uids, nil
}

func (gl *groupLooker) GetUserGroups(ctx context.Context, uid string, cached bool) ([]string, error) {
	key := fmt.Sprintf("u:%s", uid)

	// check if it is cached
	if cached {
		if gl.client.Exists(key).Val() == true {
			cmd := gl.client.SMembers(key)
			if cmd.Err() == nil {
				gids, err := cmd.Result()
				if err == nil {
					return gids, nil
				}
			}
		}
	}

	gids, err := gl.wrapped.GetUserGroups(ctx, uid, false)
	if err != nil {
		return nil, err
	}

	pipeline := gl.client.TxPipeline()
	defer pipeline.Close()
	for _, gid := range gids {
		pipeline.SAdd(key, gid)
	}
	pipeline.Expire(key, time.Second*time.Duration(gl.ttl))
	_, err = pipeline.Exec()
	if err != nil {
		return nil, err
	}
	return gids, nil
}

func (gl *groupLooker) GetUserComputingGroups(ctx context.Context, uid string, cached bool) ([]string, error) {
	key := fmt.Sprintf("unixuser:%s", uid)

	// check if it is cached
	if cached {
		if gl.client.Exists(key).Val() == true {
			cmd := gl.client.SMembers(key)
			if cmd.Err() == nil {
				gids, err := cmd.Result()
				if err == nil {
					return gids, nil
				}
			}
		}
	}

	gids, err := gl.wrapped.GetUserComputingGroups(ctx, uid, false)
	if err != nil {
		return nil, err
	}

	pipeline := gl.client.TxPipeline()
	defer pipeline.Close()
	for _, gid := range gids {
		pipeline.SAdd(key, gid)
	}
	pipeline.Expire(key, time.Second*time.Duration(gl.ttl))
	_, err = pipeline.Exec()
	if err != nil {
		return nil, err
	}
	return gids, nil
}

func (gl *groupLooker) GetTTLForUser(ctx context.Context, uid string) (time.Duration, error) {
	key := fmt.Sprintf("u:%s", uid)
	return gl.client.TTL(key).Result()
}

func (gl *groupLooker) GetTTLForGroup(ctx context.Context, gid string) (time.Duration, error) {
	key := fmt.Sprintf("egroup:%s", gid)
	return gl.client.TTL(key).Result()
}

func (gl *groupLooker) GetTTLForComputingGroup(ctx context.Context, gid string) (time.Duration, error) {
	key := fmt.Sprintf("unixgroup:%s", gid)
	return gl.client.TTL(key).Result()
}

func (gl *groupLooker) GetTTLForComputingUser(ctx context.Context, gid string) (time.Duration, error) {
	key := fmt.Sprintf("unixuser:%s", gid)
	return gl.client.TTL(key).Result()
}
