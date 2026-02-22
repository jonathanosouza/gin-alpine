// Package links contains the domain logic for the links module.
package links

import (
	"context"
	"errors"
	"time"

	"gin-alpine/src/internal/sqlc/gen"

	"github.com/google/uuid"
)

var (
	ErrLinkNotFound = errors.New("link not found")
	ErrLinkExpired  = errors.New("link expired")
)

type Repository interface {
	CheckResetLinkAvailable(ctx context.Context, email string, now time.Time) (*uuid.UUID, error)
	CreateLink(ctx context.Context, data string, expiresAt time.Time, linkUUID uuid.UUID) (int32, error)
	CreateUserAvailableLink(ctx context.Context, userID int32, linkUUID uuid.UUID) (int32, error)
	GetLink(ctx context.Context, linkUUID uuid.UUID) (*gen.GetLinkRow, error)
	DeleteLink(ctx context.Context, id int32, deletedAt time.Time) error
}
