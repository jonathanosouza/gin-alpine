// Package e2e
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"gin-alpine/src/internal/sqlc/gen"
	"gin-alpine/src/pkg/utils"
	"gin-alpine/src/services/main/internal/bootstrap"
	"io"
	"strings"
	"testing"

	"gin-alpine/src/services/main/internal/handler/router"
	"gin-alpine/src/services/main/internal/handler/web/user/dto"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

type SUT struct {
	Bootstrap *bootstrap.Bootstrap
	Server    *httptest.Server
	Router    http.Handler
	Addr      string
}

func NewSUT() *SUT {
	b := bootstrap.MustGetBootstrapInstance()
	b.Config.Env = gin.TestMode
	b.SetInitialData()
	router := router.NewRouter(b)
	server := httptest.NewServer(router)
	return &SUT{
		Bootstrap: b,
		Router:    router,
		Server:    server,
		Addr:      server.URL,
	}
}

func (s *SUT) NewRequest(method, path string, body io.Reader) (*http.Request, error) {
	// ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	url := s.Addr + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (s *SUT) LogoutCurrentLoggedUser(t *testing.T, ctx context.Context, client *http.Client) {
	req, err := s.NewRequest("POST", "/logout", nil)
	if err != nil {
		utils.FatalResult("Failed to create logout request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("error executing logout request %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Logout failed with status")
	}
}

func (s *SUT) CreateUserFromAdminLogin(t *testing.T, ctx context.Context, name, email, password string, roleID int32) (*LoginFromAdminResult, error) {
	newUserPayload := dto.CreateUserDTO{
		Name:     name,
		Email:    name + "@test.com",
		Password: "password",
		RoleID:   roleID,
	}
	payloadBytes, err := json.Marshal(newUserPayload)
	if err != nil {
		return nil, err
	}
	client := s.LoginAdmin(t)
	req, err := s.NewRequest("POST", "/api/users", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			utils.FatalResult("Failed to close response body", err)
		}
	}(res.Body)
	var user gen.User
	err = json.NewDecoder(res.Body).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &LoginFromAdminResult{
		Client: client,
		User:   &user,
	}, nil
}

func (s *SUT) LoginAdmin(t *testing.T) *http.Client {
	return s.Login(t, s.Bootstrap.Config.AdminEmail, s.Bootstrap.Config.AdminPass)
}

func (s *SUT) Login(t *testing.T, email, password string) *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("Failed to create cookie jar %v", err)
	}
	client := &http.Client{Jar: jar}
	loginPayload := map[string]string{
		"email":    email,
		"password": password,
	}
	payloadBytes, _ := json.Marshal(loginPayload)

	req, err := s.NewRequest("POST", "/login", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("Failed to create login request")
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute login request %v", err)
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			t.Fatalf("Failed to close response body: %v", err)
		}
	}(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Login failed with status")
	}
	// The jar is passed by reference, so it's automatically updated.
	return client
}
