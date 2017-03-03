package main

import (
	"flag"
	"fmt"
	"github.com/cernbox/cboxgroupd/handlers"
	"github.com/cernbox/cboxgroupd/pkg/ldapgrouplooker"
	"github.com/cernbox/cboxgroupd/pkg/redisgrouplooker"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

var fPort int
var fLDAPHostname string
var fLDAPPort int
var fRedisHostname string
var fRedisPort int
var fRedisDatabase int
var fRedisTTL int
var fAppLog string

func init() {
	flag.IntVar(&fPort, "port", 2002, "Port to listen for connections")
	flag.StringVar(&fLDAPHostname, "ldaphostname", "xldap.cern.ch", "Hostname of the LDAP server")
	flag.IntVar(&fLDAPPort, "ldapport", 389, "Port of LDAP server")
	flag.StringVar(&fRedisHostname, "redishostname", "localhost", "Hostname of the Redis server")
	flag.IntVar(&fRedisPort, "redisport", 6379, "Port of Redis server")
	flag.IntVar(&fRedisDatabase, "redisdb", 0, "Redis number database for keys isolation (0-15)")
	flag.IntVar(&fRedisTTL, "redisttl", 60, "Number of seconds to expire cached entries in Redis")
	flag.StringVar(&fAppLog, "applog", "stderr", "File to log application data")
	flag.Parse()
}

func main() {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{fAppLog}
	logger, _ := config.Build()

	lgl := ldapgrouplooker.New(fLDAPHostname, fLDAPPort)
	rgl := redisgrouplooker.New(fRedisHostname, fRedisPort, fRedisDatabase, fRedisTTL, lgl)

	router := mux.NewRouter()

	router.Handle("/api/v1/membership/usersingroup/{gid}", handlers.UsersInGroup(logger, rgl))
	router.Handle("/api/v1/membership/usersingroupttl/{gid}", handlers.UsersInGroupTTL(logger, rgl))
	router.Handle("/api/v1/membership/usergroups/{uid}", handlers.UserGroups(logger, rgl))
	router.Handle("/api/v1/membership/usergroupsttl/{uid}", handlers.UserGroupsTTL(logger, rgl))
	logger.Info("server is listening", zap.Int("port", fPort))
	logger.Warn("server stopped", zap.Error(http.ListenAndServe(fmt.Sprintf(":%d", fPort), router)))
}
