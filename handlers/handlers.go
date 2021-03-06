package handlers

import (
	"encoding/json"
	"github.com/cernbox/cboxgroupd/pkg"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

var searchTermRegexp = regexp.MustCompile(`^[a-zA-Z0-9_.\-:\s]*$`)

func CheckSharedSecret(logger *zap.Logger, secret string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The secret is passed in the header: Authorization: Bearer mysecret
		h := r.Header.Get("Authorization")
		secret := "bearer " + secret
		if secret != strings.ToLower(h) {
			logger.Warn("wrong secret")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func isValidFilter(s string) bool {
	if s == "" {
		return false
	}

	return searchTermRegexp.MatchString(s)
}

func Search(logger *zap.Logger, groupLooker pkg.GroupLooker) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filter := mux.Vars(r)["filter"]
		if !isValidFilter(filter) {
			logger.Error("filter is invalid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		entries, err := groupLooker.Search(r.Context(), filter, true)
		if err != nil {
			logger.Info("error getting entries", zap.Error(err), zap.String("filter", filter))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Info("entries found", zap.Int("numentries", len(entries)), zap.String("filter", filter))
		json.NewEncoder(w).Encode(entries)
	})
}

func UsersInGroup(logger *zap.Logger, groupLooker pkg.GroupLooker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gid := mux.Vars(r)["gid"]
		if !isValidFilter(gid) {
			logger.Error("gid is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		uids, err := groupLooker.GetUsersInGroup(r.Context(), gid, true)
		if err != nil {
			if gle, ok := err.(pkg.GroupLookerError); ok {
				if gle.Code == pkg.GroupLookerErrorNotFound {
					logger.Warn("group not found", zap.String("gid", gid))
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}
			logger.Info("error getting users", zap.Error(err), zap.String("gid", gid))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Info("users found", zap.Int("numusers", len(uids)), zap.String("gid", gid))
		json.NewEncoder(w).Encode(uids)
	})
}

func UsersInComputingGroup(logger *zap.Logger, groupLooker pkg.GroupLooker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gid := mux.Vars(r)["gid"]
		if !isValidFilter(gid) {
			logger.Error("gid is invalid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		uids, err := groupLooker.GetUsersInComputingGroup(r.Context(), gid, true)
		if err != nil {
			if gle, ok := err.(pkg.GroupLookerError); ok {
				if gle.Code == pkg.GroupLookerErrorNotFound {
					logger.Warn("computing group not found", zap.String("gid", gid))
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}
			logger.Info("error getting users", zap.Error(err), zap.String("gid", gid))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Info("users found", zap.Int("numusers", len(uids)), zap.String("gid", gid))
		json.NewEncoder(w).Encode(uids)
	})
}

func UserGroups(logger *zap.Logger, groupLooker pkg.GroupLooker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := mux.Vars(r)["uid"]
		if !isValidFilter(uid) {
			logger.Error("uid is invalid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		gids, err := groupLooker.GetUserGroups(r.Context(), uid, true)
		if err != nil {
			if gle, ok := err.(pkg.GroupLookerError); ok {
				if gle.Code == pkg.GroupLookerErrorNotFound {
					logger.Warn("user not found", zap.String("uid", uid))
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}
			logger.Info("error getting users", zap.Error(err), zap.String("uid", uid))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Info("groups found", zap.Int("numgroups", len(gids)), zap.String("uid", uid))
		json.NewEncoder(w).Encode(gids)
	})
}

func UserComputingGroups(logger *zap.Logger, groupLooker pkg.GroupLooker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := mux.Vars(r)["uid"]
		if !isValidFilter(uid) {
			logger.Error("uid is invalid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		gids, err := groupLooker.GetUserComputingGroups(r.Context(), uid, true)
		if err != nil {
			if gle, ok := err.(pkg.GroupLookerError); ok {
				if gle.Code == pkg.GroupLookerErrorNotFound {
					logger.Warn("user not found", zap.String("uid", uid))
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}
			logger.Info("error getting users", zap.Error(err), zap.String("uid", uid))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logger.Info("unix groups found", zap.Int("numgroups", len(gids)), zap.String("uid", uid))
		json.NewEncoder(w).Encode(gids)
	})
}

func UserGroupsTTL(logger *zap.Logger, groupLooker pkg.GroupLooker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := mux.Vars(r)["uid"]
		if !isValidFilter(uid) {
			logger.Error("uid is invalid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ttl, err := groupLooker.GetTTLForUser(r.Context(), uid)
		if err != nil {
			logger.Info("error getting ttl for user", zap.Error(err), zap.String("uid", uid))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		res := struct {
			UID string  `json:"uid"`
			TTL float64 `json:"ttl"`
		}{uid, ttl.Seconds()}
		logger.Info("ttl retrieved", zap.String("uid", uid), zap.Float64("ttl", res.TTL))
		json.NewEncoder(w).Encode(res)
	})
}

func UserComputingGroupsTTL(logger *zap.Logger, groupLooker pkg.GroupLooker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := mux.Vars(r)["uid"]
		if !isValidFilter(uid) {
			logger.Error("uid is invalid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ttl, err := groupLooker.GetTTLForComputingUser(r.Context(), uid)
		if err != nil {
			logger.Info("error getting ttl for computing user", zap.Error(err), zap.String("uid", uid))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		res := struct {
			UID string  `json:"uid"`
			TTL float64 `json:"ttl"`
		}{uid, ttl.Seconds()}
		logger.Info("ttl retrieved", zap.String("uid", uid), zap.Float64("ttl", res.TTL))
		json.NewEncoder(w).Encode(res)
	})
}

func UsersInGroupTTL(logger *zap.Logger, groupLooker pkg.GroupLooker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gid := mux.Vars(r)["gid"]
		if !isValidFilter(gid) {
			logger.Error("gid is invalid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ttl, err := groupLooker.GetTTLForGroup(r.Context(), gid)
		if err != nil {
			logger.Info("error getting ttl for group", zap.Error(err), zap.String("gid", gid))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		res := struct {
			GID string  `json:"gid"`
			TTL float64 `json:"ttl"`
		}{gid, ttl.Seconds()}
		logger.Info("ttl retrieved", zap.String("gid", gid), zap.Float64("ttl", res.TTL))
		json.NewEncoder(w).Encode(res)
	})
}

func UsersInComputingGroupTTL(logger *zap.Logger, groupLooker pkg.GroupLooker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gid := mux.Vars(r)["gid"]
		if !isValidFilter(gid) {
			logger.Error("gid is invalid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ttl, err := groupLooker.GetTTLForComputingGroup(r.Context(), gid)
		if err != nil {
			logger.Info("error getting ttl for computing group", zap.Error(err), zap.String("gid", gid))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		res := struct {
			GID string  `json:"gid"`
			TTL float64 `json:"ttl"`
		}{gid, ttl.Seconds()}
		logger.Info("ttl retrieved", zap.String("gid", gid), zap.Float64("ttl", res.TTL))
		json.NewEncoder(w).Encode(res)
	})
}

// UpdateUsersInGroups allows to trigger a refresh of users belonfing to a group
func UpdateUsersInGroup(logger *zap.Logger, groupLooker pkg.GroupLooker, maxConcurrency int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			Groups []string `json:"groups"`
		}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		req := &request{}
		err = json.Unmarshal(data, req)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		go func() {
			var throttle = make(chan int, maxConcurrency)
			var wg sync.WaitGroup
			for _, gid := range req.Groups {
				if gid != "" {
					throttle <- 1
					wg.Add(1)
					go func(wg *sync.WaitGroup, throttle chan int, gid string) {
						defer wg.Done()
						defer func() { <-throttle }()
						uids, err := groupLooker.GetUsersInGroup(r.Context(), gid, false)
						if err != nil {
							if gle, ok := err.(pkg.GroupLookerError); ok {
								if gle.Code == pkg.GroupLookerErrorNotFound {
									logger.Warn("async: group not found", zap.String("gid", gid))
									return
								}
							}
							logger.Info("async: error getting users", zap.Error(err), zap.String("gid", gid))
						}
						logger.Info("async: users found", zap.Int("numusers", len(uids)), zap.String("gid", gid))
					}(&wg, throttle, gid)
				}
			}
			wg.Wait()
		}()
		w.WriteHeader(http.StatusAccepted)
	})
}

// UpdateUserGroups allows to trigger a refresh of groups belonfing to an user
func UpdateUserGroups(logger *zap.Logger, groupLooker pkg.GroupLooker, maxConcurrency int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			Users []string `json:"users"`
		}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		req := &request{}
		err = json.Unmarshal(data, req)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		go func() {
			var throttle = make(chan int, maxConcurrency)
			var wg sync.WaitGroup
			for _, uid := range req.Users {
				if uid != "" {
					throttle <- 1
					wg.Add(1)
					go func(wg *sync.WaitGroup, throttle chan int, uid string) {
						defer wg.Done()
						defer func() { <-throttle }()
						gids, err := groupLooker.GetUserGroups(r.Context(), uid, false)
						if err != nil {
							if gle, ok := err.(pkg.GroupLookerError); ok {
								if gle.Code == pkg.GroupLookerErrorNotFound {
									logger.Warn("async: user not found", zap.String("uid", uid))
								}
							}
							logger.Info("async: error getting groups", zap.Error(err), zap.String("uid", uid))
							return
						}
						logger.Info("async: groups found", zap.Int("numgroups", len(gids)), zap.String("uid", uid))
					}(&wg, throttle, uid)
				}
			}
			wg.Wait()
		}()
		w.WriteHeader(http.StatusAccepted)
	})
}
