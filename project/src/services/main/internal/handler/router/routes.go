// Package router ...
package router

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"gin-alpine/src/internal/domain/auth"
	"gin-alpine/src/pkg/utils"
	"gin-alpine/src/services/main/internal/bootstrap"
	"gin-alpine/src/services/main/internal/handler/middleware"

	"io/fs"
	"log"

	"gin-alpine/src/services/web"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
)

type ChartData struct {
	Labels []string `json:"labels"`
	Values []int    `json:"values"`
}

func NewRouter(b *bootstrap.Bootstrap) *gin.Engine {
	gin.SetMode(b.Config.Env)
	r := gin.New()
	r.RedirectTrailingSlash = true
	r.RedirectFixedPath = true

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5174"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-API-Key", "X-CSRF-Token", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	store := cookie.NewStore([]byte("very-secret-key"))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   b.Config.Env == gin.ReleaseMode, // enable in production (HTTPS)
	})

	r.Use(sessions.Sessions("session", store))
	r.Use(middleware.RateLimitMiddleware(b))
	// Then csrf
	if b.Config.Env != gin.TestMode {
		csrfMiddleware := csrf.Middleware(csrf.Options{
			Secret: b.Config.CSRFSecret,
			ErrorFunc: func(c *gin.Context) {
				c.String(http.StatusForbidden, "CSRF token mismatch")
				c.Abort()
			},
		})
		r.Use(csrfMiddleware)
		r.Use(middleware.CSRFTpl())
	}

	var staticFS fs.FS
	if b.Config.Env == gin.ReleaseMode {
		embedSub, err := fs.Sub(web.StaticFilesAll, "static")
		if err != nil {
			log.Fatalf("error creating static sub filesystem: %v", err)
		}
		staticFS = embedSub
	} else {
		staticPath, err := utils.GetFilePath([]string{"src", "services", "web", "static"})
		if err != nil {
			log.Fatalf("error getting static path: %v", err)
		}
		staticFS = os.DirFS(staticPath)
	}

	r.StaticFS("/static", http.FS(staticFS))

	r.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("favicon.ico", http.FS(staticFS))
	})
	r.GET("/logo4.png", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.File("logo4.png")
	})

	serveVueIndex := func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		data, err := fs.ReadFile(staticFS, "vue/index.html")
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if b.Config.Env != gin.TestMode {
			token := csrf.GetToken(c)
			if token != "" {
				html := string(data)
				if !strings.Contains(html, `meta name="csrf-token"`) {
					meta := fmt.Sprintf(`<meta name="csrf-token" content="%s" />`, token)
					if strings.Contains(html, "</head>") {
						html = strings.Replace(html, "</head>", meta+"\n  </head>", 1)
					} else {
						html = meta + "\n" + html
					}
					data = []byte(html)
				}
			}
		}
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(data)
	}

	// force-serve SPA index for known shell routes to avoid any implicit redirects
	r.Use(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			p := c.Request.URL.Path
			if p == "/" || p == "/login" || p == "/configuracoes" || p == "/perfil" {
				serveVueIndex(c)
				c.Abort()
				return
			}
		}
		c.Next()
	})

	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			path := c.Request.URL.Path
			if !strings.HasPrefix(path, "/api") && !strings.HasPrefix(path, "/static") && path != "/favicon.ico" {
				serveVueIndex(c)
				return
			}
		}
		if err := b.Renderer.Render(c.Writer, "base", "404", gin.H{
			"Title": "Page not found",
		}); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	})
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Println("panic:", recovered)
		if err := b.Renderer.Render(c.Writer, "base", "500", gin.H{
			"Title": "Internal Server Error",
		}); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}))

	// PUBLIC ROUTES
	public := r.Group("")
	public.GET("/api/context", func(c *gin.Context) {
		csrfToken := csrf.GetToken(c)
		res := gin.H{
			"csrf":    csrfToken,
			"env":     b.Config.Env,
			"AppData": nil,
			"User":    nil,
			"Can": gin.H{
				"Customer": false,
				"Manager":  false,
				"Admin":    false,
				"Dev":      false,
			},
			"IsAuth": false,
		}

		session := sessions.Default(c)
		userID := session.Get("user_id")
		if userID == nil {
			c.JSON(http.StatusOK, res)
			return
		}

		key := fmt.Sprintf("auth:user:%v", userID)
		var user auth.UserAuth
		var appData utils.AppData
		errUser := b.RedisDB.Cache.Get(c.Request.Context(), key, &user)
		errAppData := b.RedisDB.Cache.Get(c.Request.Context(), b.RedisDB.GetAppDataKey(), &appData)
		if errUser != nil || errAppData != nil {
			c.JSON(http.StatusOK, res)
			return
		}

		res["AppData"] = appData
		res["User"] = gin.H{
			"ID":    user.ID,
			"Name":  user.Name,
			"Email": user.Email,
			"Role":  user.Role,
		}
		res["IsAuth"] = true
		res["Can"] = gin.H{
			"Customer": user.Role >= auth.RoleCustomer,
			"Manager":  user.Role >= auth.RoleManager,
			"Admin":    user.Role >= auth.RoleAdmin,
			"Dev":      user.Role >= auth.RoleDev,
		}

		c.JSON(http.StatusOK, res)
	})

	public.GET("/login", serveVueIndex)
	public.POST("/login", b.AuthWebHandler.LoginPostWeb)
	public.GET("/recuperar-senha", serveVueIndex)
	public.POST("/recuperar-senha", b.AuthWebHandler.ForgotPasswordPost)
	public.GET("/reset-senha/:uuid", serveVueIndex)
	public.POST("/reset-senha/:uuid", b.AuthWebHandler.ResetPasswordPost)

	// PROTECTED ROUTES
	protected := r.Group("")
	protected.Use(middleware.AuthWeb(b.RedisDB))
	protected.POST("/logout", b.AuthWebHandler.LogoutPostWeb)
	// handle root explicitly
	protected.GET("/", serveVueIndex)
	protected.GET("/configuracoes", serveVueIndex)
	protected.GET("/perfil", serveVueIndex)
	protected.PUT("/api/users/:id", b.UserWebHandler.UpdateUser)
	protected.GET("/api/users/:id", b.UserWebHandler.GetUser)
	protected.GET("/api/users", b.UserWebHandler.ListUsers)
	protected.GET("/api/users/search", b.UserWebHandler.SearchUsers)

	users := protected.Group("/api/users")
	users.Use(middleware.RequireRoleAtLeast(auth.RoleManager))
	{
		users.POST("/", b.UserWebHandler.CreateUser)
	}

	adminUsers := protected.Group("/api/users/admin")
	adminUsers.Use(middleware.RequireAnyRole(auth.RoleAdmin, auth.RoleDev))
	{
		adminUsers.PUT("/:id", b.UserWebHandler.UpdateUserAdmin)
	}
	// API ROUTES
	// protected.GET("/api/patterns/:id/draws", b.PatternsHTTPHandler.ListPatternAndDrawsHTTP)

	// ONLY FOR DEV USERS
	// dev := r.Group("/api/system-admin/")
	// dev.Use(middleware.AuthWeb(b.RedisDB))
	// dev.Use(middleware.RequireRoleAtLeast(auth.RoleDev))
	// dev.POST("/draws", b.DrawsHTTPHandler.AddMostRecentDrawAndSyncHTTP)

	return r
}
