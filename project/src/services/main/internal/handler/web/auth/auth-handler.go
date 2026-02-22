package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"gin-alpine/src/pkg/utils"

	authDomain "gin-alpine/src/internal/domain/auth"
	"gin-alpine/src/internal/domain/links"
	"gin-alpine/src/internal/domain/users"
	"gin-alpine/src/internal/infra/redis"
	"gin-alpine/src/internal/tasks"
	"gin-alpine/src/internal/usecases"
	"gin-alpine/src/services/web"

	handler "gin-alpine/src/services/main/internal/handler/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	csrf "github.com/utrack/gin-csrf"
	"go.uber.org/zap"
)

type AuthHandler struct {
	Logger       *zap.Logger
	Renderer     *web.Renderer
	T            *utils.Translator
	RedisDB      *redis.RedisClient
	authUsecases *usecases.AuthUsecases
	resetUsecase *usecases.ResetPasswordUsecase
	asynqClient  *asynq.Client
}

func NewAuthHandler(
	authUseCases *usecases.AuthUsecases,
	resetUsecase *usecases.ResetPasswordUsecase,
	redisDB *redis.RedisClient,
	renderer *web.Renderer,
	logger *zap.Logger,
	t *utils.Translator,
	asynqClient *asynq.Client,
) *AuthHandler {
	return &AuthHandler{
		authUsecases: authUseCases,
		resetUsecase: resetUsecase,
		Renderer:     renderer,
		RedisDB:      redisDB,
		Logger:       logger,
		T:            t,
		asynqClient:  asynqClient,
	}
}

func (h *AuthHandler) LoginPostWeb(c *gin.Context) {
	var loginDTO LoginInputDTO
	if err := c.ShouldBind(&loginDTO); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			messages := utils.CustomErrorTranslator(validationErrors, GetLoginCustomMessages(h.T))
			for field, msg := range messages {
				if errRender := h.Renderer.Render(c.Writer, "auth", "login", gin.H{
					"error": handler.HTTPError{
						Field:   field,
						Message: msg,
					},
					"csrf": csrf.GetToken(c),
				}); errRender != nil {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
				return
			}
		}
		return
	}
	u, err := h.authUsecases.Login(
		c.Request.Context(),
		authDomain.LoginInput{
			Email:    loginDTO.Email,
			Password: loginDTO.Password,
		})
	if err != nil {
		code, httpErr := handler.MapAuthErrorToHTTP(err, h.T)
		h.Logger.Info("login_fail", zap.String("message", httpErr.Message), zap.String("email", loginDTO.Email))
		c.Status(code)
		if h.Renderer.Mode != gin.TestMode {
			if errRender := h.Renderer.Render(c.Writer, "auth", "login", gin.H{
				"error": httpErr,
				"csrf":  csrf.GetToken(c),
			}); errRender != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}
		return
	}
	session := sessions.Default(c)
	session.Clear()
	session.Set("user_id", u.ID)
	if err = session.Save(); err != nil {
		log.Printf("error saving session: %v", err)
		c.Status(http.StatusInternalServerError)
		if errRender := h.Renderer.Render(c.Writer, "auth", "login", gin.H{
			"Error": "We couldn’t start your session. Please try again.",
		}); errRender != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}

	err = authDomain.StoreUserAuth(
		c.Request.Context(),
		h.RedisDB,
		authDomain.UserAuth{
			ID:    u.ID,
			Email: u.Email,
			Name:  u.Name,
			Role:  authDomain.Role(int(u.RoleID)),
		},
		24*time.Hour,
	)
	if err != nil {
		h.Logger.Error("auth_store_fail", zap.String("message", err.Error()))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	h.Logger.Info("login_success:", zap.String("email", loginDTO.Email))
	c.Redirect(http.StatusSeeOther, "/")
}

func (h *AuthHandler) LogoutPostWeb(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	// Remove all session values
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	// clean from redis cache
	if userID != nil {
		key := fmt.Sprintf("auth:user:%v", userID)
		_ = h.RedisDB.Client.Del(c.Request.Context(), key).Err()
	}
	// Persist changes and expire cookie
	err := session.Save()
	if err != nil {
		log.Printf("error saving session %v", err)
	}
	// Redirect user
	h.Logger.Info("logout_success:", zap.Any("user_id", userID))
	c.Redirect(http.StatusFound, "/login")
}

func (h *AuthHandler) ForgotPasswordGet(c *gin.Context) {
	data := h.buildAuthPageData(c, "Recuperar senha", "")
	session := sessions.Default(c)
	if session.Get("forgot_password_success") != nil {
		data["success"] = "Se o e-mail existir, você receberá as instruções em instantes."
		session.Delete("forgot_password_success")
		_ = session.Save()
	}
	if err := h.Renderer.Render(c.Writer, "auth", "forgot-password", data); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

func (h *AuthHandler) ForgotPasswordPost(c *gin.Context) {
	var dto ForgotPasswordInputDTO
	if err := c.ShouldBind(&dto); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			messages := utils.CustomErrorTranslator(validationErrors, GetForgotPasswordCustomMessages(h.T))
			for _, msg := range messages {
				data := h.buildAuthPageData(c, "Recuperar senha", msg)
				if errRender := h.Renderer.Render(c.Writer, "auth", "forgot-password", data); errRender != nil {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
				return
			}
		}
		return
	}

	linkUUID, _, err := h.resetUsecase.CreateOrReuseResetLink(c.Request.Context(), dto.Email)
	if err != nil && !errors.Is(err, users.ErrUserNotFound) {
		data := h.buildAuthPageData(c, "Recuperar senha", "Não foi possível processar sua solicitação.")
		if errRender := h.Renderer.Render(c.Writer, "auth", "forgot-password", data); errRender != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}
	if linkUUID != nil {
		resetURL := fmt.Sprintf("%s/reset-senha/%s", h.getBaseURL(c), linkUUID.String())
		tmpl, err := tasks.GetEmailTemplate("reset-password.html")
		if err != nil {
			h.Logger.Error("reset_password_template_error", zap.String("error", err.Error()))
		} else {
			rendered, err := tasks.RenderResetPasswordTemplate(tmpl, &utils.ResetPasswordVars{Link: resetURL})
			if err != nil {
				h.Logger.Error("reset_password_render_error", zap.String("error", err.Error()))
			} else {
				payload := utils.EmailPayload{
					Addresses: &[]string{dto.Email},
					Subject:   tasks.ResetPasswordSubject,
					Body:      rendered.String(),
				}
				task, err := tasks.NewEmailTask(&payload)
				if err == nil {
					_, err = h.asynqClient.Enqueue(task)
				}
				if err != nil {
					h.Logger.Error("reset_password_enqueue_error", zap.String("error", err.Error()))
				}
			}
		}
	}

	session := sessions.Default(c)
	session.Set("forgot_password_success", true)
	_ = session.Save()
	c.Redirect(http.StatusSeeOther, "/recuperar-senha")
}

func (h *AuthHandler) ResetPasswordGet(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("reset_password_success") != nil {
		session.Delete("reset_password_success")
		_ = session.Save()
		data := h.buildResetPageData(c, "", "Senha atualizada com sucesso. Você já pode entrar.", "")
		if err := h.Renderer.Render(c.Writer, "auth", "reset-password", data); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}
	linkUUID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		data := h.buildResetPageData(c, "Link inválido ou expirado.", "", "")
		if errRender := h.Renderer.Render(c.Writer, "auth", "reset-password", data); errRender != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}
	_, _, err = h.resetUsecase.GetValidLink(c.Request.Context(), linkUUID)
	if err != nil {
		data := h.buildResetPageData(c, "Link inválido ou expirado.", "", "")
		if errRender := h.Renderer.Render(c.Writer, "auth", "reset-password", data); errRender != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}
	data := h.buildResetPageData(c, "", "", linkUUID.String())
	if err := h.Renderer.Render(c.Writer, "auth", "reset-password", data); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

func (h *AuthHandler) ResetPasswordPost(c *gin.Context) {
	linkUUID, err := uuid.Parse(c.Param("uuid"))
	if err != nil {
		data := h.buildResetPageData(c, "Link inválido ou expirado.", "", "")
		if errRender := h.Renderer.Render(c.Writer, "auth", "reset-password", data); errRender != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}
	var dto ResetPasswordInputDTO
	if err := c.ShouldBind(&dto); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			messages := utils.CustomErrorTranslator(validationErrors, GetResetPasswordCustomMessages(h.T))
			for _, msg := range messages {
				data := h.buildResetPageData(c, msg, "", linkUUID.String())
				if errRender := h.Renderer.Render(c.Writer, "auth", "reset-password", data); errRender != nil {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
				return
			}
		}
		return
	}
	if dto.Password != dto.PasswordConfirm {
		data := h.buildResetPageData(c, "As senhas não conferem.", "", linkUUID.String())
		if errRender := h.Renderer.Render(c.Writer, "auth", "reset-password", data); errRender != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}
	err = h.resetUsecase.ResetPassword(c.Request.Context(), linkUUID, dto.Password)
	if err != nil {
		if errors.Is(err, links.ErrLinkExpired) || errors.Is(err, links.ErrLinkNotFound) {
			data := h.buildResetPageData(c, "Link inválido ou expirado.", "", "")
			if errRender := h.Renderer.Render(c.Writer, "auth", "reset-password", data); errRender != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}
		data := h.buildResetPageData(c, "Não foi possível atualizar a senha.", "", linkUUID.String())
		if errRender := h.Renderer.Render(c.Writer, "auth", "reset-password", data); errRender != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		return
	}
	session := sessions.Default(c)
	session.Set("reset_password_success", true)
	_ = session.Save()
	c.Redirect(http.StatusSeeOther, "/reset-senha/"+linkUUID.String())
}

func (h *AuthHandler) buildAuthPageData(c *gin.Context, title, errMsg string) gin.H {
	data := gin.H{
		"Title": title,
	}
	if h.Renderer.Mode != gin.TestMode {
		data["csrf"] = csrf.GetToken(c)
	}
	if errMsg != "" {
		data["error"] = errMsg
	}
	return data
}

func (h *AuthHandler) buildResetPageData(c *gin.Context, errMsg, successMsg, linkUUID string) gin.H {
	data := h.buildAuthPageData(c, "Redefinir senha", errMsg)
	if successMsg != "" {
		data["success"] = successMsg
	}
	if linkUUID != "" {
		data["link_uuid"] = linkUUID
	}
	return data
}

func (h *AuthHandler) getBaseURL(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if forwarded := c.GetHeader("X-Forwarded-Proto"); forwarded != "" {
		scheme = strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	return fmt.Sprintf("%s://%s", scheme, host)
}
