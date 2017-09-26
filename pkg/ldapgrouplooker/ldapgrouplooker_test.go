package ldapgrouplooker

import (
	"context"
	"fmt"
	"testing"
)

func TestUserGroups(t *testing.T) {
	ctx := context.Background()
	gl := New("xldap.cern.ch", 389, 1000)
	gids, err := gl.GetUserGroups(ctx, "gonzalhu", false)
	if err != nil {
		t.Error(err)
	}
	for _, g := range gids {
		fmt.Println(g)
	}
}

func TestComputingGroups(t *testing.T) {
	ctx := context.Background()
	gl := New("xldap.cern.ch", 389, 1000)
	gids, err := gl.GetUserComputingGroups(ctx, "gonzalhu", false)
	if err != nil {
		t.Error(err)
	}
	for _, g := range gids {
		fmt.Println(g)
	}
}
