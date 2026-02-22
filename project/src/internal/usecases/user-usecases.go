package usecases

import (
	"context"
	"errors"
	"strings"

	"gin-alpine/src/internal/domain/auth"
	"gin-alpine/src/internal/domain/users"
	"gin-alpine/src/internal/sqlc/gen"
	"gin-alpine/src/pkg/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	userRepo users.Repository
}

func NewUserUsecase(userRepo users.Repository) *UserUsecase {
	return &UserUsecase{userRepo: userRepo}
}

func (uc *UserUsecase) GetUser(ctx context.Context, id int) (*gen.FindUserByIDRow, error) {
	user, err := uc.userRepo.FindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, users.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (uc *UserUsecase) CreateUser(ctx context.Context, params gen.CreateUserParams) (*gen.User, error) {
	_, err := uc.userRepo.GetUserByEmail(ctx, params.Email)
	if err == nil {
		return nil, users.ErrUserEmailAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	params.Password = string(hashedPassword)
	params.Uuid = uuid.New()

	user, err := uc.userRepo.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *UserUsecase) UpdateUser(ctx context.Context, input users.UpdateUserInput) (*gen.FindUserByIDRow, error) {
	existingUser, err := uc.userRepo.FindUserByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, users.ErrUserNotFound
		}
		return nil, err
	}

	if input.Email != nil {
		found, err := uc.userRepo.GetUserByEmail(ctx, *input.Email)
		if err == nil && found.ID != int32(input.ID) {
			return nil, users.ErrUserEmailAlreadyExists
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}

	if input.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hashed := string(hashedPassword)
		input.Password = &hashed
	}

	if err := uc.userRepo.UpdateUser(ctx, input); err != nil {
		return nil, err
	}

	updated, err := uc.userRepo.FindUserByID(ctx, int(existingUser.ID))
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (uc *UserUsecase) UpdateUserAdmin(ctx context.Context, input users.UpdateUserAdminInput, actorRole auth.Role) (*gen.FindUserByIDRow, error) {
	existingUser, err := uc.userRepo.FindUserByID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, users.ErrUserNotFound
		}
		return nil, err
	}

	if !existingUser.Role.Valid {
		return nil, users.ErrRoleNotAllowed
	}
	if existingUser.Role.PositionType != gen.PositionTypeCUSTOMER && existingUser.Role.PositionType != gen.PositionTypeMANAGER {
		return nil, users.ErrRoleNotAllowed
	}

	if input.RoleID != nil {
		if actorRole == auth.RoleAdmin && *input.RoleID == int(auth.RoleAdmin) {
			return nil, users.ErrRoleNotAllowed
		}
		if actorRole == auth.RoleDev && *input.RoleID == int(auth.RoleDev) {
			return nil, users.ErrRoleNotAllowed
		}
	}

	if input.Email != nil {
		found, err := uc.userRepo.GetUserByEmail(ctx, *input.Email)
		if err == nil && found.ID != int32(input.ID) {
			return nil, users.ErrUserEmailAlreadyExists
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}

	if input.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hashed := string(hashedPassword)
		input.Password = &hashed
	}

	if err := uc.userRepo.UpdateUserAdmin(ctx, input); err != nil {
		return nil, err
	}

	updated, err := uc.userRepo.FindUserByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (uc *UserUsecase) ListUsers(ctx context.Context, page, limit int) (*models.PaginationResult[*gen.GetUsersWithTotalRow], error) {
	pagination := models.NormalizePagination(page, limit)
	items, err := uc.userRepo.GetUsersWithTotal(ctx, int32(pagination.Limit), int32(pagination.Offset))
	if err != nil {
		return nil, err
	}

	totalItems := 0
	if len(items) > 0 {
		totalItems = int(items[0].TotalCount)
	}
	totalPages := 0
	if totalItems > 0 && pagination.Limit > 0 {
		totalPages = (totalItems + pagination.Limit - 1) / pagination.Limit
	}

	result := models.PaginationResult[*gen.GetUsersWithTotalRow]{
		Page:         page,
		ItemsPerPage: pagination.Limit,
		TotalItems:   totalItems,
		TotalPages:   totalPages,
		Items:        items,
	}
	return &result, nil
}

func (uc *UserUsecase) SearchUsers(ctx context.Context, term string, page, limit int) (*models.PaginationResult[*gen.SearchUsersRow], error) {
	pagination := models.NormalizePagination(page, limit)
	normalizedTerm := strings.ToLower(term)
	if len(normalizedTerm) < 3 {
		return &models.PaginationResult[*gen.SearchUsersRow]{
			Page:         page,
			ItemsPerPage: pagination.Limit,
			TotalItems:   0,
			TotalPages:   0,
			Items:        []*gen.SearchUsersRow{},
		}, nil
	}
	items, err := uc.userRepo.SearchUsers(ctx, normalizedTerm, int32(pagination.Limit), int32(pagination.Offset))
	if err != nil {
		return nil, err
	}
	totalItems := 0
	if len(items) > 0 {
		totalItems = int(items[0].TotalCount)
	}
	totalPages := 0
	if totalItems > 0 && pagination.Limit > 0 {
		totalPages = (totalItems + pagination.Limit - 1) / pagination.Limit
	}
	result := models.PaginationResult[*gen.SearchUsersRow]{
		Page:         page,
		ItemsPerPage: pagination.Limit,
		TotalItems:   totalItems,
		TotalPages:   totalPages,
		Items:        items,
	}
	return &result, nil
}
