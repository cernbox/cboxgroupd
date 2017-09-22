package ldapgrouplooker

import (
	"context"
	"fmt"
	"github.com/cernbox/cboxgroupd/pkg"
	"gopkg.in/ldap.v2"
	"strings"
	"time"
)

func New(hostname string, port int, pageLimit uint32) pkg.GroupLooker {
	return &groupLooker{
		hostname:  hostname,
		port:      port,
		pageLimit: pageLimit,
	}
}

type groupLooker struct {
	hostname  string
	port      int
	pageLimit uint32
}

func (gl *groupLooker) GetUsersInGroup(ctx context.Context, gid string) ([]string, error) {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", gl.hostname, gl.port))
	if err != nil {
		return nil, err
	}
	defer l.Close()

	searchRequest := ldap.NewSearchRequest(
		"OU=Users,OU=Organic Units,DC=cern,DC=ch",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(memberOf:1.2.840.113556.1.4.1941:=CN=%s,OU=e-groups,OU=Workgroups,DC=cern,DC=ch)", gid),
		[]string{"dn", "sAMAccountName", "memberOf"},
		nil,
	)

	sr, err := l.SearchWithPaging(searchRequest, gl.pageLimit)
	if err != nil {
		return nil, err
	}

	var uids []string
	for _, entry := range sr.Entries {
		for _, attr := range entry.Attributes {
			if attr.Name == "sAMAccountName" {
				if len(attr.Values) > 0 {
					if attr.Values[0] != "" {
						uids = append(uids, attr.Values[0])
					}
				}
			}
		}
	}

	return uids, nil
}

func (gl *groupLooker) GetUserGroups(ctx context.Context, uid string) ([]string, error) {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", gl.hostname, gl.port))
	if err != nil {
		return nil, err
	}
	defer l.Close()

	searchRequest := ldap.NewSearchRequest(
		"OU=e-groups,OU=Workgroups,DC=cern,DC=ch",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(member:1.2.840.113556.1.4.1941:=CN=%s,OU=Users,OU=Organic Units,DC=cern,DC=ch)", uid),
		[]string{"dn", "sAMAccountName"},
		nil,
	)

	sr, err := l.SearchWithPaging(searchRequest, gl.pageLimit)
	if err != nil {
		return nil, err
	}

	var gids []string
	for _, entry := range sr.Entries {
		for _, attr := range entry.Attributes {
			if attr.Name == "sAMAccountName" {
				if len(attr.Values) > 0 {
					if attr.Values[0] != "" {
						gids = append(gids, attr.Values[0])
					}
				}
			}
		}
	}

	return gids, nil
}

func (gl *groupLooker) GetUsersInComputingGroup(ctx context.Context, gid string) ([]string, error) {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", gl.hostname, gl.port))
	if err != nil {
		return nil, err
	}
	defer l.Close()

	searchRequest := ldap.NewSearchRequest(
		"OU=Users,OU=Organic Units,DC=cern,DC=ch",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(memberOf:1.2.840.113556.1.4.1941:=CN=%s,OU=unix,OU=Workgroups,DC=cern,DC=ch)", gid),
		[]string{"dn", "sAMAccountName", "memberOf"},
		nil,
	)

	sr, err := l.SearchWithPaging(searchRequest, gl.pageLimit)
	if err != nil {
		return nil, err
	}

	var uids []string
	for _, entry := range sr.Entries {
		for _, attr := range entry.Attributes {
			if attr.Name == "sAMAccountName" {
				if len(attr.Values) > 0 {
					if attr.Values[0] != "" {
						uids = append(uids, attr.Values[0])
					}
				}
			}
		}
	}

	return uids, nil
}

func (gl *groupLooker) GetUserComputingGroups(ctx context.Context, uid string) ([]string, error) {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", gl.hostname, gl.port))
	if err != nil {
		return nil, err
	}
	defer l.Close()

	searchRequest := ldap.NewSearchRequest(
		"OU=Users,OU=Organic Units,DC=cern,DC=ch",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(cn=%s)", uid),
		[]string{"dn", "memberOf"},
		nil,
	)

	sr, err := l.SearchWithPaging(searchRequest, gl.pageLimit)
	if err != nil {
		return nil, err
	}

	var gids []string
	for _, entry := range sr.Entries {
		for _, attr := range entry.Attributes {
			if attr.Name == "memberOf" {
				for _, v := range attr.Values {
					// v is in form CN=cern-fellows,OU=e-groups,OU=Workgroups,DC=cern,DC=ch
					// check that we only include e-groups in the response
					tokens := strings.Split(v, ",")
					if len(tokens) == 5 {
						if tokens[1] == "OU=unix" {
							cnTokens := strings.Split(tokens[0], "=")
							if len(cnTokens) == 2 && cnTokens[0] == "CN" && cnTokens[1] != "" {
								gids = append(gids, cnTokens[1])
							}
						}
					}
				}
			}
		}
	}

	return gids, nil
}

func (gl *groupLooker) GetTTLForUser(ctx context.Context, uid string) (time.Duration, error) {
	return time.Duration(-1), nil
}

func (gl *groupLooker) GetTTLForGroup(ctx context.Context, gid string) (time.Duration, error) {
	return time.Duration(-1), nil
}

func (gl *groupLooker) GetTTLForComputingUser(ctx context.Context, uid string) (time.Duration, error) {
	return time.Duration(-1), nil
}

func (gl *groupLooker) GetTTLForComputingGroup(ctx context.Context, gid string) (time.Duration, error) {
	return time.Duration(-1), nil
}
