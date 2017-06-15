package pkg

import (
	"context"
	"fmt"
	"time"
)

type GroupLookerErrorCode string

const (
	GroupLookerErrorNotFound GroupLookerErrorCode = "GROUPLOOKER_ERROR_NOT_FOUND"
)

func NewGroupLookerError(code GroupLookerErrorCode) GroupLookerError {
	return GroupLookerError{Code: code}
}

type GroupLookerError struct {
	Code    GroupLookerErrorCode
	Message string
}

func (sr GroupLookerError) WithMessage(msg string) GroupLookerError {
	sr.Message = msg
	return sr
}

func (sr GroupLookerError) Error() string {
	return fmt.Sprintf("%s: %s", sr.Code, sr.Message)
}

type GroupLooker interface {
	GetUsersInGroup(ctx context.Context, gid string) ([]string, error)
	GetUserGroups(ctx context.Context, uid string) ([]string, error)
	GetUsersInComputingGroup(ctx context.Context, gid string) ([]string, error)
	GetUserComputingGroups(ctx context.Context, gid string) ([]string, error)
	GetTTLForUser(ctx context.Context, uid string) (time.Duration, error)
	GetTTLForGroup(ctx context.Context, gid string) (time.Duration, error)
	GetTTLForComputingUser(ctx context.Context, uid string) (time.Duration, error)
	GetTTLForComputingGroup(ctx context.Context, gid string) (time.Duration, error)
}
