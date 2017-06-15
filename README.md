# CERNBox Group Daemon

The CERNBox group daemon is a dameon that caches LDAP information from central
CERN LDAP for finding:

- all the egroups a user belongs to 
- all the users inside an egroup

Redis is used as shared cache allowing multiple instances of cboxgroupd to be
run simultaneously

## Options

```
Usage of cboxgroupd:
  -applog string
        File to log application data (default "stderr")
  -httplog string
        File to log HTTP requests (default "stderr")
  -ldaphostname string
        Hostname of the LDAP server (default "xldap.cern.ch")
  -ldappagelimit uint
        Page limit for paged searchs (default 1000)
  -ldapport int
        Port of LDAP server (default 389)
  -port int
        Port to listen for connections (default 2002)
  -redisdb int
        Redis number database for keys isolation (0-15)
  -redishostname string
        Hostname of the Redis server (default "localhost")
  -redisport int
        Port of Redis server (default 6379)
  -redisttl int
        Number of seconds to expire cached entries in Redis (default 60)
  -secret string
        Share secret between services to authenticate requests (default "changeme!!!")
  -version
        Show version

```

## How to run it

```
$ cboxgroupd --secret abc --port 2002

{"level":"info","ts":1497533157.7921486,"caller":"/home/labkode/go/src/github.com/cernbox/cboxgroupd/main.go:81","msg":"server is listening","port":3000}
....
```

## Some example requests

```
curl -i localhost:2002/api/v1/membership/usersingroup/cernbox-admins -H "Authorization: Bearer abc"

curl -i localhost:2002/api/v1/membership/usersincomputinggroup/zp -H "Authorization: Bearer abc"

curl -i localhost:2002/api/v1/membership/usergroups/gonzalhu -H "Authorization: Bearer abc"

```

