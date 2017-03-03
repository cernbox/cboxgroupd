package main

import (
	"flag"
	"fmt"
	"github.com/cernbox/cboxgroupd/handlers"
	"github.com/cernbox/cboxgroupd/pkg/ldapgrouplooker"
	"github.com/cernbox/cboxgroupd/pkg/redisgrouplooker"
	gh "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
)

var fPort int
var fLDAPHostname string
var fLDAPPort int
var fRedisHostname string
var fRedisPort int
var fRedisDatabase int
var fRedisTTL int
var fAppLog string
var fHTTPLog string
var fSecret string

func init() {
	flag.IntVar(&fPort, "port", 2002, "Port to listen for connections")
	flag.StringVar(&fLDAPHostname, "ldaphostname", "xldap.cern.ch", "Hostname of the LDAP server")
	flag.IntVar(&fLDAPPort, "ldapport", 389, "Port of LDAP server")
	flag.StringVar(&fRedisHostname, "redishostname", "localhost", "Hostname of the Redis server")
	flag.IntVar(&fRedisPort, "redisport", 6379, "Port of Redis server")
	flag.IntVar(&fRedisDatabase, "redisdb", 0, "Redis number database for keys isolation (0-15)")
	flag.IntVar(&fRedisTTL, "redisttl", 60, "Number of seconds to expire cached entries in Redis")
	flag.StringVar(&fAppLog, "applog", "stderr", "File to log application data")
	flag.StringVar(&fHTTPLog, "httplog", "stderr", "File to log HTTP requests")
	flag.StringVar(&fSecret, "secret", "changeme!!!", "Share secret between services to authenticate requests")
	flag.Parse()
}

func main() {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{fAppLog}
	logger, _ := config.Build()

	lgl := ldapgrouplooker.New(fLDAPHostname, fLDAPPort)
	rgl := redisgrouplooker.New(fRedisHostname, fRedisPort, fRedisDatabase, fRedisTTL, lgl)

	router := mux.NewRouter()

	protectedUsersInGroup := handlers.CheckSharedSecret(logger, fSecret, handlers.UsersInGroup(logger, rgl))
	protectedUsersInGroupTTL := handlers.CheckSharedSecret(logger, fSecret, handlers.UsersInGroupTTL(logger, rgl))
	protectedUserGroups := handlers.CheckSharedSecret(logger, fSecret, handlers.UserGroups(logger, rgl))
	protectedUserGroupsTTL := handlers.CheckSharedSecret(logger, fSecret, handlers.UserGroupsTTL(logger, rgl))

	router.Handle("/api/v1/membership/usersingroup/{gid}", protectedUsersInGroup).Methods("GET")
	router.Handle("/api/v1/membership/usersingroupttl/{gid}", protectedUsersInGroupTTL).Methods("GET")
	router.Handle("/api/v1/membership/usergroups/{uid}", protectedUserGroups).Methods("GET")
	router.Handle("/api/v1/membership/usergroupsttl/{uid}", protectedUserGroupsTTL).Methods("GET")

	out := getHTTPLoggerOut(fHTTPLog)
	loggedRouter := gh.LoggingHandler(out, router)

	logger.Info("server is listening", zap.Int("port", fPort))
	logger.Warn("server stopped", zap.Error(http.ListenAndServe(fmt.Sprintf(":%d", fPort), loggedRouter)))
}

func getHTTPLoggerOut(filename string) *os.File {
	if filename == "stderr" {
		return os.Stderr
	} else if filename == "stdout" {
		return os.Stdout
	} else {
		fd, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return fd
	}
}
