package ldapgrouplooker

import (
	"context"
	"fmt"
	"github.com/cernbox/cboxgroupd/pkg"
	"gopkg.in/ldap.v2"
	"strings"
)

func New(hostname string, port int) pkg.GroupLooker {
	return &groupLooker{
		hostname: hostname,
		port:     port,
	}
}

type groupLooker struct {
	hostname string
	port     int
}

func (ul *groupLooker) GetUsersInGroup(ctx context.Context, gid string) ([]string, error) {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", ul.hostname, ul.port))
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

	sr, err := l.Search(searchRequest)
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
	// if there are not any uids we assume the group does not exist
	if len(uids) == 0 {
		return nil, pkg.NewGroupLookerError(pkg.GroupLookerErrorNotFound).WithMessage(gid)
	}

	return uids, nil
}

func (ul *groupLooker) GetUserGroups(ctx context.Context, uid string) ([]string, error) {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", ul.hostname, ul.port))
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

	sr, err := l.Search(searchRequest)
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
						if tokens[1] == "OU=e-groups" {
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

	// if there are not any uids we assume the group does not exist
	if len(gids) == 0 {
		return nil, pkg.NewGroupLookerError(pkg.GroupLookerErrorNotFound).WithMessage(uid)
	}
	return gids, nil
}
