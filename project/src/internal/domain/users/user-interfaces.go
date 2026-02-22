package users

import (
	"context"

	"gin-alpine/src/internal/sqlc/gen"
)

type Repository interface {
	CreateUser(ctx context.Context, arg gen.CreateUserParams) (*gen.User, error)
	GetUserByEmail(ctx context.Context, email string) (*gen.User, error)
	FindUserByEmail(ctx context.Context, email string) (*gen.FindUserByEmailRow, error)
	FindUserByID(ctx context.Context, id int) (*gen.FindUserByIDRow, error)
	UpdateUser(ctx context.Context, input UpdateUserInput) error
	UpdateUserAdmin(ctx context.Context, input UpdateUserAdminInput) error
	GetUsersWithTotal(ctx context.Context, limit, offset int32) ([]*gen.GetUsersWithTotalRow, error)
	SearchUsers(ctx context.Context, term string, limit, offset int32) ([]*gen.SearchUsersRow, error)
}
