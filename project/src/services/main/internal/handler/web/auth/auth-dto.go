// Package auth
package auth

import "gin-alpine/src/pkg/utils"

type LoginInputDTO struct {
	Email    string `form:"email" binding:"required,email"`
	Password string `form:"password" binding:"required,min=4,max=100"`
}

type ForgotPasswordInputDTO struct {
	Email string `form:"email" json:"email" binding:"required,email"`
}

type ResetPasswordInputDTO struct {
	Password        string `form:"password" json:"password" binding:"required,min=6,max=100"`
	PasswordConfirm string `form:"password_confirm" json:"password_confirm" binding:"required,min=6,max=100"`
}

func GetLoginCustomMessages(t *utils.Translator) *map[string]string {
	return &map[string]string{
		"Email.required":    t.T("errors.user.email_required", nil),
		"Email.email":       t.T("errors.user.valid_email", nil),
		"Password.required": t.T("errors.user.password_required", nil),
		"Password.min":      t.T("errors.user.password_min_length", nil),
	}
}

func GetForgotPasswordCustomMessages(t *utils.Translator) *map[string]string {
	return &map[string]string{
		"Email.required": t.T("errors.user.email_required", nil),
		"Email.email":    t.T("errors.user.valid_email", nil),
	}
}

func GetResetPasswordCustomMessages(t *utils.Translator) *map[string]string {
	return &map[string]string{
		"Password.required":        t.T("errors.user.password_required", nil),
		"Password.min":             t.T("errors.user.password_min_length", nil),
		"PasswordConfirm.required": t.T("errors.user.password_required", nil),
		"PasswordConfirm.min":      t.T("errors.user.password_min_length", nil),
	}
}
