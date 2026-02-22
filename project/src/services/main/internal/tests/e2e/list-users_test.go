package e2e

import (
	"encoding/json"
	"gin-alpine/src/internal/domain/auth"
	"gin-alpine/src/pkg/utils"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
)

func TestListUsers(t *testing.T) {
	description := "Test if the list users usecases are working correctly"
	defer func() {
		log.Printf("Test: %s\n", description)
		log.Println("Deferred tearing down.")
	}()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("should list users with pagination when authenticated", func(t *testing.T) {
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

		req, _ := sut.NewRequest("GET", "/api/users?page=1&limit=10", nil)
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
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
		}

		var body struct {
			Page         int           `json:"page"`
			ItemsPerPage int           `json:"items_per_page"`
			TotalItems   int           `json:"total_items"`
			TotalPages   int           `json:"total_pages"`
			Items        []interface{} `json:"items"`
		}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}
		if body.Page == 0 || body.ItemsPerPage == 0 {
			t.Errorf("expected pagination fields to be filled")
		}
		if len(body.Items) == 0 {
			t.Errorf("expected at least one user in list")
		}
	})

	t.Run("should return bad request when missing pagination params", func(t *testing.T) {
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

		req, _ := sut.NewRequest("GET", "/api/users?page=1", nil)
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
		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
		}
		var body struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}
		expectedMessage := "Requisição inválida."
		if body.Error != expectedMessage {
			t.Errorf("expected error %q, got %q", expectedMessage, body.Error)
		}
	})

	t.Run("should return empty items when search term is too short", func(t *testing.T) {
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

		req, _ := sut.NewRequest("GET", "/api/users/search?page=1&limit=10&term=ab", nil)
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
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
		}

		var body struct {
			Page         int           `json:"page"`
			ItemsPerPage int           `json:"items_per_page"`
			TotalItems   int           `json:"total_items"`
			TotalPages   int           `json:"total_pages"`
			Items        []interface{} `json:"items"`
		}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}
		if len(body.Items) != 0 {
			t.Errorf("expected empty items for short term, got %d", len(body.Items))
		}
		if body.TotalItems != 0 || body.TotalPages != 0 {
			t.Errorf("expected totals to be zero for short term")
		}
	})

	t.Run("should return users matching the search term", func(t *testing.T) {
		randStr := utils.RandomString(10)
		userName := "search-" + randStr
		userEmail := userName + "@test.com"
		userPassword := defaultPass
		adminLoginResult, err := sut.CreateUserFromAdminLogin(
			t,
			ctx,
			userName,
			userEmail,
			userPassword,
			int32(auth.RoleCustomer))
		if err != nil {
			t.FailNow()
		}
		sut.LogoutCurrentLoggedUser(t, ctx, adminLoginResult.Client)

		client := sut.Login(t, userEmail, userPassword)
		searchTerm := "search"

		req, _ := sut.NewRequest("GET", "/api/users/search?page=1&limit=10&term="+searchTerm, nil)
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
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
		}

		var body struct {
			Items []struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"items"`
		}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}
		found := false
		for _, item := range body.Items {
			if strings.Contains(strings.ToLower(item.Name), strings.ToLower(searchTerm)) ||
				strings.Contains(strings.ToLower(item.Email), strings.ToLower(searchTerm)) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find at least one user matching the search term")
		}
	})
}
