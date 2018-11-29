package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cernbox/cboxgroupd/handlers"
	"github.com/cernbox/cboxgroupd/pkg/ldapgrouplooker"
	"github.com/cernbox/cboxgroupd/pkg/redisgrouplooker"
	gh "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
	viper.SetDefault("ldapbindusername", "")
	viper.SetDefault("ldapbindpassword", "")
	viper.SetDefault("ldappagelimit", 1000)
	viper.SetDefault("ldapbasesearch", "OU=Users,OU=Organic Units,DC=cern,DC=ch")
	viper.SetDefault("redishostname", "localhost")
	viper.SetDefault("groupsearch", true)
	viper.SetDefault("redisenable", true)
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
	flag.String("ldapbindusername", "", "Username to access LDAP")
	flag.String("ldapbindpassword", "", "password to access LDAP")
	flag.Uint("ldappagelimit", 1000, "Page limit for paged searchs")
	flag.String("ldapbasesearch", "OU=Users,OU=Organic Units,DC=cern,DC=ch", "Base LDAP search path")
	flag.String("redishostname", "localhost", "Hostname of the Redis server")
	flag.String("redispassword", "foo", "Password for the Redis server")
	flag.Bool("groupsearch", false, "Allow user groups searches")
	flag.Bool("redisenable", false, "Use or not redis as a cache between LDAP and cboxgroup")
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

	lgl := ldapgrouplooker.New(viper.GetString("ldaphostname"), viper.GetInt("ldapport"), viper.GetString("ldapbindusername"), viper.GetString("ldapbindpassword"), uint32(viper.GetInt("ldappagelimit")), viper.GetString("ldapbasesearch"))
	gl := lgl

	if viper.GetBool("redisenable") {
		logger.Info("Using REDIS")
		gl = redisgrouplooker.New(viper.GetString("redishostname"), viper.GetInt("redisport"), viper.GetInt("redisdb"), viper.GetInt("redisttl"), viper.GetString("redispassword"), lgl)
	}

	fGroupSearch := viper.GetBool("groupsearch")
	if fGroupSearch {
		logger.Info("Using Group search")
	}

	logger.Info("Configs", zap.String("ldaphostname", viper.GetString("ldaphostname")), zap.String("ldapbasesearch", viper.GetString("ldapbasesearch")), zap.Int("ldapport", viper.GetInt("ldapport")))

	router := mux.NewRouter()

	protectedUsersInGroup := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UsersInGroup(logger, gl, fGroupSearch))
	protectedUsersInComputingGroup := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UsersInComputingGroup(logger, gl, fGroupSearch))
	protectedUserGroups := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UserGroups(logger, gl, fGroupSearch))
	protectedUserComputingGroups := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UserComputingGroups(logger, gl, fGroupSearch))
	protectedUsersInGroupTTL := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UsersInGroupTTL(logger, gl, fGroupSearch))
	protectedUsersInComputingGroupTTL := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UsersInComputingGroupTTL(logger, gl, fGroupSearch))
	protectedUserGroupsTTL := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UserGroupsTTL(logger, gl, fGroupSearch))
	protectedUserComputingGroupsTTL := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UserComputingGroupsTTL(logger, gl, fGroupSearch))

	protectedUpdateUsersInGroup := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UpdateUsersInGroup(logger, gl, viper.GetInt("ldapmaxconcurrency"), fGroupSearch))
	protectedUpdateUserGroups := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.UpdateUserGroups(logger, gl, viper.GetInt("ldapmaxconcurrency"), fGroupSearch))

	protectedSearch := handlers.CheckSharedSecret(logger, viper.GetString("secret"), handlers.Search(logger, gl))

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

	router.Handle("/api/v1/search/{filter}", protectedSearch).Methods("GET")

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
