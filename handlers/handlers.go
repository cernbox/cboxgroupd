package handlers

import (
	"encoding/json"
	"github.com/cernbox/cboxgroupd/pkg"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

func UsersInGroup(logger *zap.Logger, groupLooker pkg.GroupLooker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gid := mux.Vars(r)["gid"]
		if gid == "" {
			logger.Error("gid is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		uids, err := groupLooker.GetUsersInGroup(r.Context(), gid)
		if err != nil {
			if gle, ok := err.(pkg.GroupLookerError); ok {
				if gle.Code == pkg.GroupLookerErrorNotFound {
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}
			logger.Info("error getting users", zap.Error(err), zap.String("gid", gid))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(uids)
	})
}

func UsersInGroupTTL(logger *zap.Logger, groupLooker pkg.GroupLooker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gid := mux.Vars(r)["gid"]
		if gid == "" {
			logger.Error("gid is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ttl, err := groupLooker.GetTTLForGroup(r.Context(), gid)
		if err != nil {
			logger.Info("error getting ttl for group", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		res := struct {
			GID string  `json:"gid"`
			TTL float64 `json:"ttl"`
		}{gid, ttl.Seconds()}

		json.NewEncoder(w).Encode(res)
	})
}

func UserGroups(logger *zap.Logger, groupLooker pkg.GroupLooker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := mux.Vars(r)["uid"]
		if uid == "" {
			logger.Error("uid is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		gids, err := groupLooker.GetUserGroups(r.Context(), uid)
		if err != nil {
			if gle, ok := err.(pkg.GroupLookerError); ok {
				if gle.Code == pkg.GroupLookerErrorNotFound {
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}
			logger.Info("error getting users", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(gids)
	})
}

func UserGroupsTTL(logger *zap.Logger, groupLooker pkg.GroupLooker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := mux.Vars(r)["uid"]
		if uid == "" {
			logger.Error("uid is empty")
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

		json.NewEncoder(w).Encode(res)
	})
}
