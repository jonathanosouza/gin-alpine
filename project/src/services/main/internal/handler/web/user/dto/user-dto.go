// Package dto
package dto

import "gin-alpine/src/pkg/utils"

type CreateUserDTO struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	RoleID   int32  `json:"role_id" binding:"required"`
}

type UpdateUserDTO struct {
	Name     *string `json:"name" binding:"omitempty"`
	Email    *string `json:"email" binding:"omitempty,email"`
	Password *string `json:"password" binding:"omitempty,min=4"`
	Enable   *bool   `json:"enable"`
}

type UpdateUserAdminDTO struct {
	Name     *string `json:"name" binding:"omitempty"`
	Email    *string `json:"email" binding:"omitempty,email"`
	Password *string `json:"password" binding:"omitempty,min=4"`
	Enable   *bool   `json:"enable"`
	RoleID   *int32  `json:"role_id" binding:"omitempty"`
}

func GetCreateUserCustomMessages(t *utils.Translator) map[string]string {
	return map[string]string{
		"Name.required":     t.T("errors.validation.user.name_required", nil),
		"Email.required":    t.T("errors.validation.user.email_required", nil),
		"Email.email":       t.T("errors.validation.user.email_invalid", nil),
		"Password.required": t.T("errors.validation.user.password_required", nil),
		"Password.min":      t.T("errors.validation.user.password_min", nil),
		"RoleID.required":   t.T("errors.validation.user.role_id_required", nil),
	}
}

func GetUpdateUserCustomMessages(t *utils.Translator) map[string]string {
	return map[string]string{
		"Email.email":  t.T("errors.validation.user.email_invalid", nil),
		"Password.min": t.T("errors.validation.user.password_min", nil),
	}
}
