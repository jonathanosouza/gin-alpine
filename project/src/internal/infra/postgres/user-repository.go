package postgres

import (
	"context"
	"time"

	"gin-alpine/src/internal/domain/users"
	"gin-alpine/src/internal/sqlc/gen"

	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	R *PgRepository
}

func NewUserRepository(p *PgRepository) *UserRepository {
	return &UserRepository{R: p}
}

func (r *UserRepository) FindUserByID(ctx context.Context, id int) (*gen.FindUserByIDRow, error) {
	q := gen.New(r.R.DB)
	user, err := q.FindUserByID(ctx, int32(id))
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *UserRepository) UpdateUser(ctx context.Context, input users.UpdateUserInput) error {
	q := gen.New(r.R.DB)
	params := gen.UpdateUserAdminParams{
		Name:      pgtype.Text{},
		Email:     pgtype.Text{},
		Password:  pgtype.Text{},
		UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		RoleID:    pgtype.Int4{},
		DeletedAt: pgtype.Timestamp{},
		Enabled:   pgtype.Bool{},
		ID:        int32(input.ID),
	}
	if input.Name != nil {
		params.Name = pgtype.Text{String: *input.Name, Valid: true}
	}
	if input.Email != nil {
		params.Email = pgtype.Text{String: *input.Email, Valid: true}
	}
	if input.Password != nil {
		params.Password = pgtype.Text{String: *input.Password, Valid: true}
	}
	if input.Enabled != nil {
		params.Enabled = pgtype.Bool{Bool: *input.Enabled, Valid: true}
	}
	return q.UpdateUserAdmin(ctx, params)
}
func (r *UserRepository) UpdateUserAdmin(ctx context.Context, input users.UpdateUserAdminInput) error {
	q := gen.New(r.R.DB)
	params := gen.UpdateUserAdminParams{
		Name:      pgtype.Text{},
		Email:     pgtype.Text{},
		Password:  pgtype.Text{},
		UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		RoleID:    pgtype.Int4{},
		DeletedAt: pgtype.Timestamp{},
		Enabled:   pgtype.Bool{},
		ID:        int32(input.ID),
	}
	if input.Name != nil {
		params.Name = pgtype.Text{String: *input.Name, Valid: true}
	}
	if input.Email != nil {
		params.Email = pgtype.Text{String: *input.Email, Valid: true}
	}
	if input.Password != nil {
		params.Password = pgtype.Text{String: *input.Password, Valid: true}
	}
	if input.RoleID != nil {
		params.RoleID = pgtype.Int4{Int32: int32(*input.RoleID), Valid: true}
	}
	if input.Enabled != nil {
		params.Enabled = pgtype.Bool{Bool: *input.Enabled, Valid: true}
	}
	return q.UpdateUserAdmin(ctx, params)
}
func (r *UserRepository) GetUsersWithTotal(ctx context.Context, limit, offset int32) ([]*gen.GetUsersWithTotalRow, error) {
	q := gen.New(r.R.DB)
	rows, err := q.GetUsersWithTotal(ctx, gen.GetUsersWithTotalParams{
		Limit:  int(limit),
		Offset: int(offset),
	})
	if err != nil {
		return nil, err
	}
	items := make([]*gen.GetUsersWithTotalRow, len(rows))
	for i := range rows {
		items[i] = &rows[i]
	}
	return items, nil
}

func (r *UserRepository) SearchUsers(ctx context.Context, term string, limit, offset int32) ([]*gen.SearchUsersRow, error) {
	q := gen.New(r.R.DB)
	searchText := pgtype.Text{String: term, Valid: true}
	rows, err := q.SearchUsers(ctx, gen.SearchUsersParams{
		Column1: searchText,
		Column2: searchText,
		Limit:   int(limit),
		Offset:  int(offset),
		Column5: searchText,
		Column6: searchText,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*gen.SearchUsersRow, len(rows))
	for i := range rows {
		items[i] = &rows[i]
	}
	return items, nil
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (*gen.FindUserByEmailRow, error) {
	q := gen.New(r.R.DB)
	u, err := q.FindUserByEmail(ctx, email)
	return &u, err
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*gen.User, error) {
	q := gen.New(r.R.DB)
	user, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, arg gen.CreateUserParams) (*gen.User, error) {
	q := gen.New(r.R.DB)
	user, err := q.CreateUser(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
