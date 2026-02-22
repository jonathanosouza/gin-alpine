package postgres

import (
	"context"
	"errors"
	"time"

	"gin-alpine/src/internal/domain/links"
	"gin-alpine/src/internal/sqlc/gen"
	"gin-alpine/src/pkg/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type LinksRepository struct {
	R *PgRepository
}

func NewLinksRepository(p *PgRepository) *LinksRepository {
	return &LinksRepository{R: p}
}

func (r *LinksRepository) CheckResetLinkAvailable(ctx context.Context, email string, now time.Time) (*uuid.UUID, error) {
	q := gen.New(r.R.DB)
	result, err := q.CheckIfUserHasResetPasswordLinksAvailable(ctx, gen.CheckIfUserHasResetPasswordLinksAvailableParams{
		Email:     email,
		ExpiresAt: pgtype.Timestamp{Time: now, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	if !result.Valid {
		return nil, pgx.ErrNoRows
	}
	linkUUID := uuid.UUID(result.Bytes)
	return &linkUUID, nil
}

func (r *LinksRepository) CreateLink(ctx context.Context, data string, expiresAt time.Time, linkUUID uuid.UUID) (int32, error) {
	expiresAt = expiresAt.UTC().Add(utils.LinkResetPasswordExpiration)
	q := gen.New(r.R.DB)
	return q.CreateLink(ctx, gen.CreateLinkParams{
		Data:      data,
		ExpiresAt: pgtype.Timestamp{Time: expiresAt, Valid: true},
		Uuid:      linkUUID,
	})
}

func (r *LinksRepository) CreateUserAvailableLink(ctx context.Context, userID int32, linkUUID uuid.UUID) (int32, error) {
	q := gen.New(r.R.DB)
	return q.CreateUserAvailableLinks(ctx, gen.CreateUserAvailableLinksParams{
		UserID:   userID,
		LinkUuid: linkUUID,
		Type:     gen.LinkTypeRESETPASS,
	})
}

func (r *LinksRepository) GetLink(ctx context.Context, linkUUID uuid.UUID) (*gen.GetLinkRow, error) {
	q := gen.New(r.R.DB)
	link, err := q.GetLink(ctx, linkUUID)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *LinksRepository) DeleteLink(ctx context.Context, id int32, deletedAt time.Time) error {
	q := gen.New(r.R.DB)
	return q.DeleteLink(ctx, gen.DeleteLinkParams{
		ID:        id,
		DeletedAt: pgtype.Timestamp{Time: deletedAt, Valid: true},
	})
}

var _ links.Repository = (*LinksRepository)(nil)
