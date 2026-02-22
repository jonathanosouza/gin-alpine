package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gin-alpine/src/internal/domain/links"
	"gin-alpine/src/internal/domain/users"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type ResetLinkData struct {
	UserID int32  `json:"user_id"`
	Email  string `json:"email"`
}

type ResetPasswordUsecase struct {
	linksRepo links.Repository
	usersRepo users.Repository
}

func NewResetPasswordUsecase(linksRepo links.Repository, usersRepo users.Repository) *ResetPasswordUsecase {
	return &ResetPasswordUsecase{
		linksRepo: linksRepo,
		usersRepo: usersRepo,
	}
}

func (uc *ResetPasswordUsecase) CreateOrReuseResetLink(ctx context.Context, email string) (*uuid.UUID, bool, error) {
	user, err := uc.usersRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, users.ErrUserNotFound
		}
		return nil, false, err
	}
	now := time.Now().UTC()
	existing, err := uc.linksRepo.CheckResetLinkAvailable(ctx, email, now)
	if err == nil && existing != nil {
		return existing, true, nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, err
	}

	linkUUID := uuid.New()
	expiresAt := now
	data, err := json.Marshal(ResetLinkData{
		UserID: user.ID,
		Email:  user.Email,
	})
	if err != nil {
		return nil, false, err
	}
	if _, err := uc.linksRepo.CreateLink(ctx, string(data), expiresAt, linkUUID); err != nil {
		return nil, false, err
	}
	if _, err := uc.linksRepo.CreateUserAvailableLink(ctx, user.ID, linkUUID); err != nil {
		return nil, false, err
	}
	return &linkUUID, false, nil
}

func (uc *ResetPasswordUsecase) GetValidLink(ctx context.Context, linkUUID uuid.UUID) (*ResetLinkData, *int32, error) {
	link, err := uc.linksRepo.GetLink(ctx, linkUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, links.ErrLinkNotFound
		}
		return nil, nil, err
	}
	if link.ExpiresAt.Valid && link.ExpiresAt.Time.Before(time.Now().UTC()) {
		return nil, nil, links.ErrLinkExpired
	}
	var payload ResetLinkData
	if err := json.Unmarshal([]byte(link.Data), &payload); err != nil {
		return nil, nil, err
	}
	return &payload, &link.ID, nil
}

func (uc *ResetPasswordUsecase) ResetPassword(ctx context.Context, linkUUID uuid.UUID, newPassword string) error {
	payload, linkID, err := uc.GetValidLink(ctx, linkUUID)
	if err != nil {
		return err
	}
	if payload == nil || linkID == nil {
		return links.ErrLinkNotFound
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashed := string(hashedPassword)
	update := users.UpdateUserInput{
		ID:       int(payload.UserID),
		Password: &hashed,
	}
	if err := uc.usersRepo.UpdateUser(ctx, update); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return users.ErrUserNotFound
		}
		return err
	}
	return uc.linksRepo.DeleteLink(ctx, *linkID, time.Now())
}
