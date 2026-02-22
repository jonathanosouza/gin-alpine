package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gin-alpine/src/internal/domain/auth"
	"gin-alpine/src/pkg/utils"
	"gin-alpine/src/services/main/internal/handler/web/user/dto"
	"log"
	"net/http"
	"testing"
)

func TestUpdateUser(t *testing.T) {
	description := "Test if the update user usecases are working correctly"
	defer func() {
		log.Printf("Test: %s\n", description)
		log.Println("Deferred tearing down.")
	}()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("should update own user data with success", func(t *testing.T) {
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

		client := sut.Login(t, userEmail, userPassword)

		newName := "updated-" + randStr
		newEmail := randStr + "-updated@test.com"
		payload := dto.UpdateUserDTO{
			Name:   &newName,
			Email:  &newEmail,
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("Failed to marshal payload: %v", err)
		}
		req, err := sut.NewRequest("PUT", "/api/users/"+int32ToString(adminLoginResult.User.ID), bytes.NewBuffer(payloadBytes))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer func() {
			if err := res.Body.Close(); err != nil {
				t.Logf("failed to close response body: %v", err)
			}
		}()
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
		}
	})

	t.Run("should return forbidden when trying to update another user", func(t *testing.T) {
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
		payload := dto.UpdateUserDTO{
			Name: &newName,
		}
		payloadBytes, _ := json.Marshal(payload)
		req, _ := sut.NewRequest("PUT", "/api/users/"+int32ToString(adminLoginResult.User.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		res, err := adminClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer func() {
			if err := res.Body.Close(); err != nil {
				t.Logf("failed to close response body: %v", err)
			}
		}()
		if res.StatusCode != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, res.StatusCode)
		}
		var body struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}
		expectedMessage := "Ação não permitida para seu perfil."
		if body.Error != expectedMessage {
			t.Errorf("expected error %q, got %q", expectedMessage, body.Error)
		}
	})

	t.Run("should return bad request when email already exists", func(t *testing.T) {
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

		client := sut.Login(t, firstEmail, firstPassword)

		payload := dto.UpdateUserDTO{
			Email: &secondEmail,
		}
		payloadBytes, _ := json.Marshal(payload)
		req, _ := sut.NewRequest("PUT", "/api/users/"+int32ToString(firstResult.User.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer func() {
			if err := res.Body.Close(); err != nil {
				t.Logf("failed to close response body: %v", err)
			}
		}()
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
}

func int32ToString(v int32) string {
	return fmt.Sprintf("%d", v)
}
