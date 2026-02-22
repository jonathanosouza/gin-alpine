package e2e

import (
	"gin-alpine/src/internal/sqlc/gen"
	"net/http"
)

type LoginFromAdminResult struct {
	Client *http.Client
	User   *gen.User
}
