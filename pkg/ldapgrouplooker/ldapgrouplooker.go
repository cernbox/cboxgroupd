package ldapgrouplooker

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/cernbox/cboxgroupd/pkg"
	"gopkg.in/ldap.v2"
	"strconv"
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

// GetUsersInGroup is an expensive query that can put the cluster down if there are a lot of concurrent connections.
// Try to minimize its usage.
func (gl *groupLooker) GetUsersInGroup(ctx context.Context, gid string, cached bool) ([]string, error) {
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

// GetUserGroups returns the list of e-groups a user belongs to.
// The implementation relies on the unpacking of security identifiers, what is used by kerberos also to perform the authentication.
// The black magic can be checked in this blog: http://blogs.perl.org/users/initself/2013/09/netldap-active-directory-sid-unpack.html
// Jose from OpenStack provided me with a python recipe to perform the decoding of the binary SID to objectSID.
// This implementation asks for the user tokenGroups attribute, decodes the SIDs and then perform a big query to resolve the SID to human name (cn)
// To obtain the tokenGroups one can issue this:
//  #ldapsearch -x -LLL -H ldap://xldap.cern.ch -b 'CN=gonzalhu,OU=Users,OU=Organic Units,DC=cern,DC=ch' -s base '(&(objectClass=user))' tokenGroups
// The filter can be huge and hit the maximum allowed size imposed by AD, so in case we see failures we need to run it in chunks.
// The decoding is based on little endian, in something does not seem to work, probably is because of the architecture. Be aware.
func (gl *groupLooker) GetUserGroups(ctx context.Context, uid string, cached bool) ([]string, error) {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", gl.hostname, gl.port))
	if err != nil {
		return nil, err
	}
	defer l.Close()

	searchRequest := ldap.NewSearchRequest(
		fmt.Sprintf("CN=%s,OU=Users,OU=Organic Units,DC=cern,DC=ch", uid),
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=User)",
		[]string{"tokenGroups"},
		nil,
	)

	sr, err := l.SearchWithPaging(searchRequest, gl.pageLimit)
	if err != nil {
		return nil, err
	}

	var sids []string
	for _, entry := range sr.Entries {
		for _, attr := range entry.Attributes {
			if attr.Name == "tokenGroups" {
				for _, binarySID := range attr.ByteValues {
					numSubIDs, _ := strconv.ParseUint(fmt.Sprintf("%d", binarySID[1]), 16, 64)
					auth, _ := strconv.ParseUint(fmt.Sprintf("%x", binarySID[2:8]), 16, 64)

					sidObject := fmt.Sprintf("S-%x-%d", binarySID[0], auth)
					//authorities := []uint32{}
					for i := uint64(0); i < numSubIDs; i++ {
						part := binarySID[8+4*i : 12+4*i]
						a := binary.LittleEndian.Uint32(part)
						sidObject += fmt.Sprintf("-%d", a)
						//authorities = append(authorities, a)
					}

					//fmt.Printf("SID_RAW=%X SRL=%x NUM_SUB_ID=%d AUTH=%d AUTHORITIES=%+v SID=%s\n", binarySID, binarySID[0], numSubIDs, auth, authorities, sidObject)
					sids = append(sids, sidObject)
				}
			}
		}
	}

	groupsFilter := "(&(objectClass=Group)(|%s))"
	var query string
	for _, sid := range sids {
		query += fmt.Sprintf("(objectSID=%s)", sid)
	}
	groupsFilter = fmt.Sprintf(groupsFilter, query)

	searchRequest = ldap.NewSearchRequest(
		"OU=e-groups,OU=Workgroups,DC=cern,DC=ch",
		ldap.ScopeSingleLevel, ldap.NeverDerefAliases, 0, 0, false,
		groupsFilter,
		[]string{"cn"},
		nil,
	)

	sr, err = l.SearchWithPaging(searchRequest, gl.pageLimit)
	if err != nil {
		return nil, err
	}

	var gids []string
	for _, entry := range sr.Entries {
		for _, attr := range entry.Attributes {
			if attr.Name == "cn" {
				for _, cn := range attr.Values {
					gids = append(gids, cn)
				}
			}
		}
	}
	return gids, nil
}

func (gl *groupLooker) GetUsersInComputingGroup(ctx context.Context, gid string, cached bool) ([]string, error) {
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

func (gl *groupLooker) GetUserComputingGroups(ctx context.Context, uid string, cached bool) ([]string, error) {
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
