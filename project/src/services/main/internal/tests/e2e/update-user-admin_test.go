package e2e

import (
	"bytes"
	"encoding/json"
	"gin-alpine/src/internal/domain/auth"
	"gin-alpine/src/pkg/utils"
	"gin-alpine/src/services/main/internal/handler/web/user/dto"
	"io"
	"log"
	"net/http"
	"testing"
)

func TestUpdateUserAdmin(t *testing.T) {
	description := "Test if the update user admin usecases are working correctly"
	defer func() {
		log.Printf("Test: %s\n", description)
		log.Println("Deferred tearing down.")
	}()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("should update customer when authenticated as admin", func(t *testing.T) {
		randStr := utils.RandomString(10)
		userEmail := randStr + "@test.com"
		userPassword := defaultPass
		adminLoginResult, err := sut.CreateUserFromAdminLogin(
			t,
			ctx,
			randStr,
			userEmail,
			userPassword,
			int32(auth.RoleCustomer))
		if err != nil {
			t.FailNow()
		}
		sut.LogoutCurrentLoggedUser(t, ctx, adminLoginResult.Client)

		adminClient := sut.LoginAdmin(t)

		newName := "updated-" + randStr
		newRole := int32(auth.RoleManager)
		payload := dto.UpdateUserAdminDTO{
			Name:   &newName,
			RoleID: &newRole,
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("Failed to marshal payload: %v", err)
		}
		req, err := sut.NewRequest("PUT", "/api/users/admin/"+int32ToString(adminLoginResult.User.ID), bytes.NewBuffer(payloadBytes))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		res, err := adminClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer func(body io.ReadCloser) {
			if err := body.Close(); err != nil {
				t.Fatalf("Failed to close response body: %v", err)
			}
		}(res.Body)
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
		}
	})

	t.Run("should return forbidden when admin tries to set role to admin", func(t *testing.T) {
		randStr := utils.RandomString(10)
		userEmail := randStr + "@test.com"
		userPassword := defaultPass
		adminLoginResult, err := sut.CreateUserFromAdminLogin(
			t,
			ctx,
			randStr,
			userEmail,
			userPassword,
			int32(auth.RoleCustomer))
		if err != nil {
			t.FailNow()
		}
		sut.LogoutCurrentLoggedUser(t, ctx, adminLoginResult.Client)

		adminClient := sut.LoginAdmin(t)

		roleID := int32(auth.RoleAdmin)
		payload := dto.UpdateUserAdminDTO{
			RoleID: &roleID,
		}
		payloadBytes, _ := json.Marshal(payload)
		req, _ := sut.NewRequest("PUT", "/api/users/admin/"+int32ToString(adminLoginResult.User.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		res, err := adminClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer func(body io.ReadCloser) {
			if err := body.Close(); err != nil {
				t.Fatalf("Failed to close response body: %v", err)
			}
		}(res.Body)
		if res.StatusCode != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, res.StatusCode)
		}
		var body struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}
		expectedMessage := "O seu cargo não permite a execução desta ação."
		if body.Error != expectedMessage {
			t.Errorf("expected error %q, got %q", expectedMessage, body.Error)
		}
	})

	t.Run("should return bad request when admin sets existing email", func(t *testing.T) {
		firstRand := utils.RandomString(10)
		firstEmail := firstRand + "@test.com"
		firstPassword := defaultPass
		firstResult, err := sut.CreateUserFromAdminLogin(
			t,
			ctx,
			firstRand,
			firstEmail,
			firstPassword,
			int32(auth.RoleCustomer))
		if err != nil {
			t.FailNow()
		}
		sut.LogoutCurrentLoggedUser(t, ctx, firstResult.Client)

		secondRand := utils.RandomString(10)
		secondEmail := secondRand + "@test.com"
		secondPassword := defaultPass
		secondResult, err := sut.CreateUserFromAdminLogin(
			t,
			ctx,
			secondRand,
			secondEmail,
			secondPassword,
			int32(auth.RoleCustomer))
		if err != nil {
			t.FailNow()
		}
		sut.LogoutCurrentLoggedUser(t, ctx, secondResult.Client)

		adminClient := sut.LoginAdmin(t)

		payload := dto.UpdateUserAdminDTO{
			Email: &secondEmail,
		}
		payloadBytes, _ := json.Marshal(payload)
		req, _ := sut.NewRequest("PUT", "/api/users/admin/"+int32ToString(firstResult.User.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		res, err := adminClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer func(body io.ReadCloser) {
			if err := body.Close(); err != nil {
				t.Fatalf("Failed to close response body: %v", err)
			}
		}(res.Body)
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
		}
		var body struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}
		expectedMessage := "Este e-mail já está em uso."
		if body.Error != expectedMessage {
			t.Errorf("expected error %q, got %q", expectedMessage, body.Error)
		}
	})

	t.Run("should return forbidden when authenticated as manager", func(t *testing.T) {
		randStr := utils.RandomString(10)
		managerEmail := randStr + "@test.com"
		managerPassword := defaultPass
		managerResult, err := sut.CreateUserFromAdminLogin(
			t,
			ctx,
			randStr,
			managerEmail,
			managerPassword,
			int32(auth.RoleManager))
		if err != nil {
			t.FailNow()
		}
		sut.LogoutCurrentLoggedUser(t, ctx, managerResult.Client)

		managerClient := sut.Login(t, managerEmail, managerPassword)

		newName := "updated-" + randStr
		payload := dto.UpdateUserAdminDTO{
			Name: &newName,
		}
		payloadBytes, _ := json.Marshal(payload)
		req, _ := sut.NewRequest("PUT", "/api/users/admin/"+int32ToString(managerResult.User.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		res, err := managerClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer func(body io.ReadCloser) {
			if err := body.Close(); err != nil {
				t.Fatalf("Failed to close response body: %v", err)
			}
		}(res.Body)
		if res.StatusCode != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, res.StatusCode)
		}
	})
}
