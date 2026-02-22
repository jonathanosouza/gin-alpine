package e2e

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"gin-alpine/src/internal/domain/auth"
	"gin-alpine/src/internal/sqlc/gen"
	"gin-alpine/src/pkg/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestResetPasswordFlow(t *testing.T) {
	description := "Test if reset password flow works correctly"
	defer func() {
		t.Logf("Test: %s", description)
	}()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("should reuse the same link while it is active", func(t *testing.T) {
		randStr := utils.RandomString(10)
		email := randStr + "@test.com"
		result, err := sut.CreateUserFromAdminLogin(
			t,
			ctx,
			randStr,
			email,
			defaultPass,
			int32(auth.RoleCustomer),
		)
		if err != nil {
			t.FailNow()
		}
		sut.LogoutCurrentLoggedUser(t, ctx, result.Client)

		if err := requestResetPassword(t, email); err != nil {
			t.Fatalf("request reset password failed: %v", err)
		}
		firstLink := fetchResetLinkUUID(t, email)

		if err := requestResetPassword(t, email); err != nil {
			t.Fatalf("request reset password failed: %v", err)
		}
		secondLink := fetchResetLinkUUID(t, email)

		if firstLink.String() != secondLink.String() {
			t.Fatalf("expected same link uuid, got %s and %s", firstLink, secondLink)
		}
	})

	t.Run("should reset password and invalidate the link", func(t *testing.T) {
		randStr := utils.RandomString(10)
		email := randStr + "@test.com"
		result, err := sut.CreateUserFromAdminLogin(
			t,
			ctx,
			randStr,
			email,
			defaultPass,
			int32(auth.RoleCustomer),
		)
		if err != nil {
			t.FailNow()
		}
		sut.LogoutCurrentLoggedUser(t, ctx, result.Client)

		if err := requestResetPassword(t, email); err != nil {
			t.Fatalf("request reset password failed: %v", err)
		}
		linkUUID := fetchResetLinkUUID(t, email)

		newPassword := "novaSenha123"
		if err := submitResetPassword(t, linkUUID, newPassword); err != nil {
			t.Fatalf("reset password failed: %v", err)
		}

		_ = sut.Login(t, email, newPassword)

		q := gen.New(sut.Bootstrap.DB.DB)
		_, err = q.CheckIfUserHasResetPasswordLinksAvailable(ctx, gen.CheckIfUserHasResetPasswordLinksAvailableParams{
			Email:     email,
			ExpiresAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		})
		if err == nil || !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("expected no active links after reset, got %v", err)
		}
	})
}

func requestResetPassword(t *testing.T, email string) error {
	payload := map[string]string{
		"email": email,
	}
	payloadBytes, _ := json.Marshal(payload)
	req, err := sut.NewRequest("POST", "/recuperar-senha", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Logf("failed to close response body: %v", err)
		}
	}()
	if res.StatusCode != http.StatusOK {
		return errors.New("unexpected status code")
	}
	return nil
}

func submitResetPassword(t *testing.T, linkUUID uuid.UUID, password string) error {
	payload := map[string]string{
		"password":         password,
		"password_confirm": password,
	}
	payloadBytes, _ := json.Marshal(payload)
	req, err := sut.NewRequest("POST", "/reset-senha/"+linkUUID.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Logf("failed to close response body: %v", err)
		}
	}()
	if res.StatusCode != http.StatusOK {
		return errors.New("unexpected status code")
	}
	return nil
}

func fetchResetLinkUUID(t *testing.T, email string) uuid.UUID {
	q := gen.New(sut.Bootstrap.DB.DB)
	linkUUID, err := q.CheckIfUserHasResetPasswordLinksAvailable(ctx, gen.CheckIfUserHasResetPasswordLinksAvailableParams{
		Email:     email,
		ExpiresAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to get reset link uuid: %v", err)
	}
	if !linkUUID.Valid {
		t.Fatalf("expected valid reset link uuid")
	}
	return uuid.UUID(linkUUID.Bytes)
}
