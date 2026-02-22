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

func TestCreateUser(t *testing.T) {
	description := "Test if the create user usecases are working correctely"
	defer func() {
		log.Printf("Test: %s\n", description)
		log.Println("Deferred tearing down.")
	}()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// -- success manager create user --
	t.Run("should create user with success when authenticated as manager", func(t *testing.T) {
		// 1. Create Manager User
		randStr := utils.RandomString(10)
		managerEmail := randStr + "@test.com"
		managerPassword := defaultPass
		// will create a new user but using the
		// default admin login credentials, then logs outs from
		// this main admin account
		adminLoginResult, err := sut.CreateUserFromAdminLogin(
			t,
			ctx,
			randStr,
			managerEmail,
			managerPassword,
			int32(auth.RoleManager))
		if err != nil {
			t.FailNow()
		}
		// logout current admin initial logged user
		sut.LogoutCurrentLoggedUser(t, ctx, adminLoginResult.Client)

		// 2. Login as Manager
		client := sut.Login(t, managerEmail, managerPassword)

		// 3. Create New User
		randStr = utils.RandomString(10)
		newUserPayload := dto.CreateUserDTO{
			Name:     randStr,
			Email:    randStr + "@test.com",
			Password: defaultPass,
			RoleID:   int32(auth.RoleCustomer),
		}
		payloadBytes, err := json.Marshal(newUserPayload)
		if err != nil {
			t.Fatalf("Failed to marshal payload: %v", err)
		}
		req, err := sut.NewRequest("POST", "/api/users/", bytes.NewBuffer(payloadBytes))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer func(body io.ReadCloser) {
			if err := body.Close(); err != nil {
				t.Fatalf("Failed to close response body: %v", err)
			}
		}(res.Body)

		// 4. Assert
		if res.StatusCode != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, res.StatusCode)
		}
	})

	// -- fail customer user attempt to create another user
	t.Run("should return forbidden when authenticated as user", func(t *testing.T) {
		randStr := utils.RandomString(10)
		// 1. Create User
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
		// logout current admin initial logged user
		sut.LogoutCurrentLoggedUser(t, ctx, adminLoginResult.Client)

		// 2. Login as User
		client := sut.Login(t, userEmail, userPassword)

		// 3. Attempt to Create New User
		randStr = utils.RandomString(10)
		anotherUserPayload := dto.CreateUserDTO{
			Name:     randStr,
			Email:    randStr + "@test.com",
			Password: defaultPass,
			RoleID:   int32(auth.RoleCustomer),
		}
		payloadBytes, _ := json.Marshal(anotherUserPayload)

		req, _ := sut.NewRequest("POST", "/api/users/", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer func(body io.ReadCloser) {
			if err := body.Close(); err != nil {
				t.Fatalf("Failed to close response body: %v", err)
			}
		}(res.Body)

		// 4. Assert
		if res.StatusCode != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, res.StatusCode)
		}
	})

	t.Run("should return bad request when email already exists", func(t *testing.T) {
		randStr := utils.RandomString(10)
		// 1. Create Manager User
		managerEmail := randStr + "@test.com"
		managerPassword := defaultPass
		adminLoginResult, err := sut.CreateUserFromAdminLogin(
			t,
			ctx,
			randStr,
			managerEmail,
			managerPassword,
			int32(auth.RoleManager))
		if err != nil {
			t.FailNow()
		}
		sut.LogoutCurrentLoggedUser(t, ctx, adminLoginResult.Client)

		// 2. Login as Manager
		client := sut.Login(t, managerEmail, managerPassword)

		// 3. Create a user
		newUserEmail := randStr + "-new@test.com"
		newUserPayload := dto.CreateUserDTO{
			Name:     "new-user",
			Email:    newUserEmail,
			Password: defaultPass,
			RoleID:   int32(auth.RoleCustomer),
		}
		payloadBytes, _ := json.Marshal(newUserPayload)
		req, _ := sut.NewRequest("POST", "/api/users/", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer func(body io.ReadCloser) {
			if err := body.Close(); err != nil {
				t.Fatalf("Failed to close response body: %v", err)
			}
		}(res.Body)
		if res.StatusCode != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, res.StatusCode)
		}

		// 4. Try to create the same user again
		payloadBytes, _ = json.Marshal(newUserPayload)
		req, _ = sut.NewRequest("POST", "/api/users/", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		res, err = client.Do(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}
		defer func(body io.ReadCloser) {
			if err := body.Close(); err != nil {
				t.Fatalf("Failed to close response body: %v", err)
			}
		}(res.Body)

		// 5. Assert
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
