// Package user
package user

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	authDomain "gin-alpine/src/internal/domain/auth"
	domain "gin-alpine/src/internal/domain/users"
	"gin-alpine/src/internal/infra/redis"
	"gin-alpine/src/internal/sqlc/gen"
	"gin-alpine/src/internal/usecases"
	"gin-alpine/src/pkg/utils"
	"gin-alpine/src/services/main/internal/handler/web/user/dto"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserHandler struct {
	userUsecase *usecases.UserUsecase
	T           *utils.Translator
	RedisDB     *redis.RedisClient
}

type listUserItem struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	Enabled    bool   `json:"enabled"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	TotalCount int    `json:"total_count"`
}

type listUsersResponse struct {
	Page         int            `json:"page"`
	ItemsPerPage int            `json:"items_per_page"`
	TotalItems   int            `json:"total_items"`
	TotalPages   int            `json:"total_pages"`
	Items        []listUserItem `json:"items"`
}

func formatTimestamp(value pgtype.Timestamp) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02T15:04:05.999999")
}

func NewUserHandler(userUsecase *usecases.UserUsecase, redisDB *redis.RedisClient, t *utils.Translator) *UserHandler {
	return &UserHandler{userUsecase: userUsecase, T: t, RedisDB: redisDB}
}

func (h *UserHandler) handleError(c *gin.Context, status int, key string) {
	c.JSON(status, gin.H{"error": h.T.T(key, nil)})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var userDto dto.CreateUserDTO
	if err := c.ShouldBindJSON(&userDto); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			messages := dto.GetCreateUserCustomMessages(h.T)
			translatedErrors := utils.CustomErrorTranslator(validationErrors, &messages)
			c.JSON(http.StatusBadRequest, gin.H{"errors": translatedErrors})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	params := gen.CreateUserParams{
		Name:     userDto.Name,
		Email:    userDto.Email,
		Password: userDto.Password,
		RoleID:   userDto.RoleID,
	}

	user, err := h.userUsecase.CreateUser(c.Request.Context(), params)
	if err != nil {
		if errors.Is(err, domain.ErrUserEmailAlreadyExists) {
			h.handleError(c, http.StatusBadRequest, "errors.custom.user.email_already_exists")
			return
		}
		h.handleError(c, http.StatusInternalServerError, "create_user_error")
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	pageParam := c.Query("page")
	limitParam := c.Query("limit")
	if pageParam == "" || limitParam == "" {
		h.handleError(c, http.StatusBadRequest, "errors.bad_request")
		return
	}
	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		h.handleError(c, http.StatusBadRequest, "errors.bad_request")
		return
	}
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit < 1 {
		h.handleError(c, http.StatusBadRequest, "errors.bad_request")
		return
	}

	result, err := h.userUsecase.ListUsers(c.Request.Context(), page, limit)
	if err != nil {
		h.handleError(c, http.StatusInternalServerError, "errors.internal")
		return
	}

	items := make([]listUserItem, len(result.Items))
	for i, row := range result.Items {
		role := ""
		if row.Role.Valid {
			role = string(row.Role.PositionType)
		}
		items[i] = listUserItem{
			ID:         int(row.ID),
			Name:       row.Name,
			Email:      row.Email,
			Role:       role,
			Enabled:    row.Enabled,
			CreatedAt:  formatTimestamp(row.CreatedAt),
			UpdatedAt:  formatTimestamp(row.UpdatedAt),
			TotalCount: int(row.TotalCount),
		}
	}

	response := listUsersResponse{
		Page:         result.Page,
		ItemsPerPage: result.ItemsPerPage,
		TotalItems:   result.TotalItems,
		TotalPages:   result.TotalPages,
		Items:        items,
	}
	c.JSON(http.StatusOK, response)
}

type getUserResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Enabled   bool   `json:"enabled"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (h *UserHandler) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil || userID <= 0 {
		h.handleError(c, http.StatusBadRequest, "errors.invalid_id")
		return
	}

	authUserValue, exists := c.Get("auth_user")
	if !exists {
		h.handleError(c, http.StatusUnauthorized, "errors.unauthorized")
		return
	}
	authUser := authUserValue.(authDomain.UserAuth)

	// Only allow user to get their own profile, unless admin?
	// For now strict self-access for this endpoint as per task context "usuário logado poderá atualizar suas informações"
	// But getting info is harmless if authorized properly.
	// Let's stick to self-access for now to be safe.
	if int(authUser.ID) != userID {
		h.handleError(c, http.StatusForbidden, "errors.forbidden")
		return
	}

	user, err := h.userUsecase.GetUser(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			h.handleError(c, http.StatusNotFound, "errors.user.not_found")
			return
		}
		h.handleError(c, http.StatusInternalServerError, "errors.internal")
		return
	}

	role := ""
	if user.Role.Valid {
		role = string(user.Role.PositionType)
	}

	resp := getUserResponse{
		ID:        int(user.ID),
		Name:      user.Name,
		Email:     user.Email,
		Role:      role,
		Enabled:   user.Enabled,
		CreatedAt: formatTimestamp(user.CreatedAt),
		UpdatedAt: formatTimestamp(user.UpdatedAt),
	}

	c.JSON(http.StatusOK, resp)
}

func (h *UserHandler) SearchUsers(c *gin.Context) {
	pageParam := c.Query("page")
	limitParam := c.Query("limit")
	if pageParam == "" || limitParam == "" {
		h.handleError(c, http.StatusBadRequest, "errors.bad_request")
		return
	}
	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		h.handleError(c, http.StatusBadRequest, "errors.bad_request")
		return
	}
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit < 1 {
		h.handleError(c, http.StatusBadRequest, "errors.bad_request")
		return
	}

	term := strings.TrimSpace(c.Query("term"))
	if len(term) < 3 {
		c.JSON(http.StatusOK, listUsersResponse{
			Page:         page,
			ItemsPerPage: limit,
			TotalItems:   0,
			TotalPages:   0,
			Items:        []listUserItem{},
		})
		return
	}

	result, err := h.userUsecase.SearchUsers(c.Request.Context(), term, page, limit)
	if err != nil {
		h.handleError(c, http.StatusInternalServerError, "errors.internal")
		return
	}

	items := make([]listUserItem, len(result.Items))
	for i, row := range result.Items {
		role := ""
		if row.Role.Valid {
			role = string(row.Role.PositionType)
		}
		items[i] = listUserItem{
			ID:         int(row.ID),
			Name:       row.Name,
			Email:      row.Email,
			Role:       role,
			Enabled:    row.Enabled,
			CreatedAt:  formatTimestamp(row.CreatedAt),
			UpdatedAt:  formatTimestamp(row.UpdatedAt),
			TotalCount: int(row.TotalCount),
		}
	}

	response := listUsersResponse{
		Page:         result.Page,
		ItemsPerPage: result.ItemsPerPage,
		TotalItems:   result.TotalItems,
		TotalPages:   result.TotalPages,
		Items:        items,
	}
	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil || userID <= 0 {
		h.handleError(c, http.StatusBadRequest, "errors.invalid_id")
		return
	}

	authUserValue, exists := c.Get("auth_user")
	if !exists {
		h.handleError(c, http.StatusUnauthorized, "errors.unauthorized")
		return
	}
	authUser := authUserValue.(authDomain.UserAuth)
	if int(authUser.ID) != userID {
		h.handleError(c, http.StatusForbidden, "errors.forbidden")
		return
	}

	var updateDto dto.UpdateUserDTO
	if err := c.ShouldBindJSON(&updateDto); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			messages := dto.GetUpdateUserCustomMessages(h.T)
			translatedErrors := utils.CustomErrorTranslator(validationErrors, &messages)
			c.JSON(http.StatusBadRequest, gin.H{"errors": translatedErrors})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updateDto.Enable != nil {
		h.handleError(c, http.StatusForbidden, "errors.forbidden")
		return
	}

	if updateDto.Name == nil && updateDto.Email == nil && updateDto.Password == nil {
		h.handleError(c, http.StatusBadRequest, "errors.update_no_fields")
		return
	}

	input := domain.UpdateUserInput{
		ID:       userID,
		Name:     updateDto.Name,
		Email:    updateDto.Email,
		Password: updateDto.Password,
	}
	updatedUser, err := h.userUsecase.UpdateUser(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrUserEmailAlreadyExists) {
			h.handleError(c, http.StatusBadRequest, "errors.custom.user.email_already_exists")
			return
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			h.handleError(c, http.StatusNotFound, "errors.user.not_found")
			return
		}
		h.handleError(c, http.StatusInternalServerError, "errors.internal")
		return
	}

	role := authDomain.RoleCustomer
	if updatedUser.Role.Valid {
		switch updatedUser.Role.PositionType {
		case gen.PositionTypeMANAGER:
			role = authDomain.RoleManager
		case gen.PositionTypeADMIN:
			role = authDomain.RoleAdmin
		case gen.PositionTypeDEV:
			role = authDomain.RoleDev
		default:
			role = authDomain.RoleCustomer
		}
	}
	_ = authDomain.StoreUserAuth(
		c.Request.Context(),
		h.RedisDB,
		authDomain.UserAuth{
			ID:    updatedUser.ID,
			Email: updatedUser.Email,
			Name:  updatedUser.Name,
			Role:  role,
		},
		24*time.Hour,
	)

	c.JSON(http.StatusOK, updatedUser)
}

func (h *UserHandler) UpdateUserAdmin(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil || userID <= 0 {
		h.handleError(c, http.StatusBadRequest, "errors.invalid_id")
		return
	}

	authUserValue, exists := c.Get("auth_user")
	if !exists {
		h.handleError(c, http.StatusUnauthorized, "errors.unauthorized")
		return
	}
	authUser := authUserValue.(authDomain.UserAuth)

	var updateDto dto.UpdateUserAdminDTO
	if err := c.ShouldBindJSON(&updateDto); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			messages := dto.GetUpdateUserCustomMessages(h.T)
			translatedErrors := utils.CustomErrorTranslator(validationErrors, &messages)
			c.JSON(http.StatusBadRequest, gin.H{"errors": translatedErrors})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updateDto.Name == nil && updateDto.Email == nil && updateDto.Password == nil && updateDto.Enable == nil && updateDto.RoleID == nil {
		h.handleError(c, http.StatusBadRequest, "errors.update_no_fields")
		return
	}

	var roleID *int
	if updateDto.RoleID != nil {
		value := int(*updateDto.RoleID)
		roleID = &value
	}

	input := domain.UpdateUserAdminInput{
		ID:       userID,
		Name:     updateDto.Name,
		Email:    updateDto.Email,
		Password: updateDto.Password,
		Enabled:  updateDto.Enable,
		RoleID:   roleID,
	}
	updatedUser, err := h.userUsecase.UpdateUserAdmin(c.Request.Context(), input, authUser.Role)
	if err != nil {
		if errors.Is(err, domain.ErrRoleNotAllowed) {
			h.handleError(c, http.StatusForbidden, "errors.user.role_forbidden")
			return
		}
		if errors.Is(err, domain.ErrUserEmailAlreadyExists) {
			h.handleError(c, http.StatusBadRequest, "errors.custom.user.email_already_exists")
			return
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			h.handleError(c, http.StatusNotFound, "errors.user.not_found")
			return
		}
		h.handleError(c, http.StatusInternalServerError, "errors.internal")
		return
	}

	if updateDto.Enable != nil && !*updateDto.Enable {
		key := fmt.Sprintf("auth:user:%v", userID)
		_ = h.RedisDB.Client.Del(c.Request.Context(), key).Err()
	}

	c.JSON(http.StatusOK, updatedUser)
}
