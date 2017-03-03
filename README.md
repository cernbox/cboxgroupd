# CERNBox Group Daemon

The CERNBox group daemon is a dameon that caches LDAP information from central
CERN LDAP for finding:

- all the egroups a user belongs to 
- all the users inside an egroup

Redis is used as shared cache allowing multiple instances of cboxgroupd to be
run simultaneously
