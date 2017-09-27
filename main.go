package main

import (
	"flag"
	"fmt"
	"github.com/cernbox/cboxgroupd/handlers"
	"github.com/cernbox/cboxgroupd/pkg/ldapgrouplooker"
	"github.com/cernbox/cboxgroupd/pkg/redisgrouplooker"
	gh "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
)

// Build information obtained with the help of -ldflags
var (
	appName       string
	buildDate     string // date -u
	gitTag        string // git describe --exact-match HEAD
	gitNearestTag string // git describe --abbrev=0 --tags HEAD
	gitCommit     string // git rev-parse HEAD
)

var fVersion bool

func init() {
	viper.SetDefault("port", 2002)
	viper.SetDefault("ldaphostname", "xldap.cern.ch")
	viper.SetDefault("ldapport", 389)
	viper.SetDefault("ldappagelimit", 1000)
	viper.SetDefault("redishostname", "localhost")
	viper.SetDefault("redisport", 6379)
	viper.SetDefault("redisdb", 0)
	viper.SetDefault("redisttl", 60)
	viper.SetDefault("applog", "stderr")
	viper.SetDefault("httplog", "stderr")
	viper.SetDefault("secret", "change_me!!!")
	viper.SetDefault("ldapmaxconcurrency", 10)

	viper.SetConfigName("cboxgroupd")
	viper.AddConfigPath("/etc/cboxgroupd/")

	flag.BoolVar(&fVersion, "version", false, "Show version")
	flag.Int("port", 2002, "Port to listen for connections")
	flag.String("ldaphostname", "xldap.cern.ch", "Hostname of the LDAP server")
	flag.Int("ldapport", 389, "Port of LDAP server")
	flag.Uint("ldappagelimit", 1000, "Page limit for paged searchs")
	flag.String("redishostname", "localhost", "Hostname of the Redis server")
	flag.Int("redisport", 6379, "Port of Redis server")
	flag.Int("redisdb", 0, "Redis number database for keys isolation (0-15)")
	flag.Int("redisttl", 60, "Number of seconds to expire cached entries in Redis")
	flag.String("applog", "stderr", "File to log application data")
	flag.String("httplog", "stderr", "File to log HTTP requests")
	flag.String("secret", "changeme!!!", "Share secret between services to authenticate requests")
	flag.Int("ldapmaxconcurrency", 100, "Number of concurrent connections to LDAP for update operations")
	flag.String("config", "", "Configuration file to use")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}

func main() {

	if fVersion {
		showVersion()
	}

	if viper.GetString("config") != "" {
		viper.SetConfigFile(viper.GetString("config"))
	}

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	config := zap.NewProductionConfig()
	config.OutputPaths = []string{viper.GetString("applog")}
	logger, _ := config.Build()

	lgl := ldapgrouplooker.New(viper.GetString("ldaphostname"), viper.GetInt("ldapport"), uint32(viper.GetInt("ldappagelimit")))
	rgl := redisgrouplooker.New(viper.GetString("redishostname"), viper.GetInt("redisport"), viper.GetInt("redisdb"), viper.GetInt("redisttl"), lgl)

	router := mux.NewRouter()

	protectedUsersInGroup := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UsersInGroup(logger, rgl))
	protectedUsersInComputingGroup := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UsersInComputingGroup(logger, rgl))
	protectedUserGroups := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UserGroups(logger, rgl))
	protectedUserComputingGroups := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UserComputingGroups(logger, rgl))
	protectedUsersInGroupTTL := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UsersInGroupTTL(logger, rgl))
	protectedUsersInComputingGroupTTL := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UsersInComputingGroupTTL(logger, rgl))
	protectedUserGroupsTTL := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UserGroupsTTL(logger, rgl))
	protectedUserComputingGroupsTTL := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UserComputingGroupsTTL(logger, rgl))

	protectedUpdateUsersInGroup := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UpdateUsersInGroup(logger, rgl, viper.GetInt("ldapmaxconcurrency")))
	protectedUpdateUserGroups := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UpdateUserGroups(logger, rgl, viper.GetInt("ldapmaxconcurrency")))

	router.Handle("/api/v1/membership/usersingroup/{gid}", protectedUsersInGroup).Methods("GET")
	router.Handle("/api/v1/membership/usersincomputinggroup/{gid}", protectedUsersInComputingGroup).Methods("GET")
	router.Handle("/api/v1/membership/usergroups/{uid}", protectedUserGroups).Methods("GET")
	router.Handle("/api/v1/membership/usercomputinggroups/{uid}", protectedUserComputingGroups).Methods("GET")

	router.Handle("/api/v1/membership/usersingroupttl/{gid}", protectedUsersInGroupTTL).Methods("GET")
	router.Handle("/api/v1/membership/usersincomputinggroupttl/{gid}", protectedUsersInComputingGroupTTL).Methods("GET")
	router.Handle("/api/v1/membership/usergroupsttl/{uid}", protectedUserGroupsTTL).Methods("GET")
	router.Handle("/api/v1/membership/usercomputinggroupsttl/{uid}", protectedUserComputingGroupsTTL).Methods("GET")

	router.Handle("/api/v1/update/usersingroup", protectedUpdateUsersInGroup).Methods("POST")
	router.Handle("/api/v1/update/usergroups", protectedUpdateUserGroups).Methods("POST")

	out := getHTTPLoggerOut(viper.GetString("httplog"))
	loggedRouter := gh.LoggingHandler(out, router)

	logger.Info("server is listening", zap.Int("port", viper.GetInt("port")))
	logger.Warn("server stopped", zap.Error(http.ListenAndServe(fmt.Sprintf(":%d", viper.GetInt("port")), loggedRouter)))
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

func showVersion() {
	// if gitTag is not empty we are on release build
	if gitTag != "" {
		fmt.Printf("%s %s commit:%s release-build\n", appName, gitNearestTag, gitCommit)
		os.Exit(0)
	}
	fmt.Printf("%s %s commit:%s dev-build\n", appName, gitNearestTag, gitCommit)
	os.Exit(0)
}
