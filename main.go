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

func init() {
	flag.IntVar(&fPort, "port", 2002, "Port to listen for connections")
	flag.StringVar(&fLDAPHostname, "ldaphostname", "xldap.cern.ch", "Hostname of the LDAP server")
	flag.IntVar(&fLDAPPort, "ldapport", 389, "Port of LDAP server")
	flag.StringVar(&fRedisHostname, "redishostname", "localhost", "Hostname of the Redis server")
	flag.IntVar(&fRedisPort, "redisport", 6379, "Port of Redis server")
	flag.IntVar(&fRedisDatabase, "redisdb", 0, "Redis number database for keys isolation (0-15)")
	flag.Parse()
}

func main() {
	logger, _ := zap.NewDevelopment()

	lgl := ldapgrouplooker.New(fLDAPHostname, fLDAPPort)
	rgl := redisgrouplooker.New(fRedisHostname, fRedisPort, fRedisDatabase, lgl)

	router := mux.NewRouter()

	router.Handle("/api/v1/membership/usersingroup/{gid}", handlers.UsersInGroup(logger, rgl))
	router.Handle("/api/v1/membership/usergroups/{uid}", handlers.UserGroups(logger, rgl))
	logger.Info("Server is listening", zap.Int("port", fPort))
	logger.Fatal("Server stopped", zap.Error(http.ListenAndServe(fmt.Sprintf(":%d", fPort), router)))
}
