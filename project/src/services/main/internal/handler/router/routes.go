// Package router ...
package router

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
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
	"github.com/go-redis/cache/v9"
	"github.com/jackc/pgx/v5"
	csrf "github.com/utrack/gin-csrf"
	"go.uber.org/zap"
)

type ChartData struct {
	Labels []string `json:"labels"`
	Values []int    `json:"values"`
}

type DashboardFilialOption struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
}

type DashboardTopItem struct {
	Nome        string  `json:"nome"`
	Faturamento float64 `json:"faturamento"`
	Percentual  float64 `json:"percentual"`
}

type DashboardMensalItem struct {
	Mes         string  `json:"mes"`
	Faturamento float64 `json:"faturamento"`
}

type DashboardVisaoGeralResponse struct {
	Filters struct {
		DataInicio string `json:"dataInicio"`
		DataFim    string `json:"dataFim"`
		Filiais    []int  `json:"filiais"`
	} `json:"filters"`
	Metrics struct {
		FaturamentoTotal float64 `json:"faturamentoTotal"`
		ClientesAtivos   int64   `json:"clientesAtivos"`
		TotalNFs         int64   `json:"totalNFs"`
		TicketMedio      float64 `json:"ticketMedio"`
	} `json:"metrics"`
	Mensal []DashboardMensalItem `json:"mensal"`
	Top    struct {
		Fornecedores []DashboardTopItem `json:"fornecedores"`
		Clientes     []DashboardTopItem `json:"clientes"`
		Produtos     []DashboardTopItem `json:"produtos"`
		Vendedores   []DashboardTopItem `json:"vendedores"`
	} `json:"top"`
	Perf struct {
		Cached  bool  `json:"cached"`
		MsTotal int64 `json:"msTotal"`
	} `json:"perf"`
	Warning string `json:"warning,omitempty"`
}

func parseISODate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

func parseFiliaisCSV(s string) ([]int, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))
	seen := map[int]struct{}{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("filial inválida: %q", p)
		}
		if v <= 0 {
			return nil, fmt.Errorf("filial inválida: %q", p)
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Ints(out)
	return out, nil
}

func joinIntsCSV(v []int) string {
	if len(v) == 0 {
		return ""
	}
	sb := strings.Builder{}
	for i := range v {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.Itoa(v[i]))
	}
	return sb.String()
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

	if b.Config.Env != gin.ReleaseMode {
		r.Use(func(c *gin.Context) {
			if c.Request.Method == http.MethodGet && strings.HasPrefix(c.Request.URL.Path, "/static/vue/") {
				c.Header("Cache-Control", "no-store")
			}
			c.Next()
		})
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
		if b.Config.Env != gin.ReleaseMode {
			c.Header("Cache-Control", "no-store")
		}
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
			if p == "/" || p == "/login" || p == "/perfil" ||
				p == "/visao-geral" ||
				p == "/visao-geral/calculos" ||
				p == "/comercial" || p == "/comercial/cliente" || p == "/comercial/vendedor" ||
				p == "/logistica" || p == "/financeiro" ||
				p == "/configuracao" || p == "/configuracao/usuarios" {
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
	protected.GET("/visao-geral", serveVueIndex)
	protected.GET("/visao-geral/calculos", serveVueIndex)
	protected.GET("/configuracao", serveVueIndex)
	protected.GET("/configuracao/usuarios", serveVueIndex)
	protected.GET("/configuracao/metas", serveVueIndex)
	protected.GET("/configuracao/automacao", serveVueIndex)
	protected.GET("/perfil", serveVueIndex)
	protected.GET("/comercial", serveVueIndex)
	protected.GET("/comercial/cliente", serveVueIndex)
	protected.GET("/comercial/vendedor", serveVueIndex)
	protected.GET("/comercial/real-meta", serveVueIndex)
	protected.GET("/comercial/fornecedor", serveVueIndex)
	protected.GET("/comercial/vendas-periodo", serveVueIndex)
	protected.GET("/comercial/crescimento-ano", serveVueIndex)
	protected.GET("/comercial/evolucao-vendedor", serveVueIndex)
	protected.GET("/logistica", serveVueIndex)
	protected.GET("/logistica/estoque", serveVueIndex)
	protected.GET("/logistica/sugestao-compras", serveVueIndex)
	protected.GET("/logistica/curva-abc", serveVueIndex)
	protected.GET("/financeiro", serveVueIndex)
	protected.GET("/financeiro/contas-pagar", serveVueIndex)
	protected.GET("/financeiro/contas-receber", serveVueIndex)
	protected.GET("/financeiro/fluxo-caixa", serveVueIndex)
	protected.GET("/financeiro/dre", serveVueIndex)

	protected.GET("/api/dashboard/filiais", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		start := time.Now()
		var exists any
		err := b.DB.DB.QueryRow(ctx, `
-- # Checagem de existência da tabela base
SELECT to_regclass('pcnfsaid')
`).Scan(&exists)
		if err != nil || exists == nil {
			c.JSON(http.StatusOK, gin.H{"filiais": []DashboardFilialOption{}})
			return
		}

		rows, err := b.DB.DB.Query(ctx, `
-- # Filtro de Filial (lista para checkboxes)
SELECT DISTINCT n.codfilial
FROM pcnfsaid n
WHERE n.codfilial IS NOT NULL
ORDER BY n.codfilial
LIMIT 200
`)
		if err != nil {
			b.Logger.Warn("dashboard_filiais_query_error", zap.Error(err))
			c.JSON(http.StatusOK, gin.H{"filiais": []DashboardFilialOption{}})
			return
		}
		defer rows.Close()

		out := make([]DashboardFilialOption, 0, 32)
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				continue
			}
			out = append(out, DashboardFilialOption{ID: id, Label: fmt.Sprintf("Filial %d", id)})
		}
		b.Logger.Info("dashboard_filiais_ok", zap.Int("count", len(out)), zap.Int64("ms", time.Since(start).Milliseconds()))
		c.JSON(http.StatusOK, gin.H{"filiais": out})
	})

	protected.GET("/api/dashboard/visao-geral", func(c *gin.Context) {
		startReq := time.Now()
		dataInicioStr := strings.TrimSpace(c.Query("data_inicio"))
		dataFimStr := strings.TrimSpace(c.Query("data_final"))
		filiaisStr := strings.TrimSpace(c.Query("filiais"))

		if dataInicioStr == "" || dataFimStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Informe data inicial e data final."})
			return
		}

		dataInicio, err := parseISODate(dataInicioStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Data inicial inválida. Use YYYY-MM-DD."})
			return
		}
		dataFim, err := parseISODate(dataFimStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Data final inválida. Use YYYY-MM-DD."})
			return
		}
		if dataInicio.After(dataFim) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Data inicial não pode ser maior que a data final."})
			return
		}
		if dataFim.Sub(dataInicio) > (time.Hour * 24 * 366 * 3) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Intervalo muito grande. Selecione até 3 anos."})
			return
		}

		filiais, err := parseFiliaisCSV(filiaisStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2500*time.Millisecond)
		defer cancel()

		cacheKey := fmt.Sprintf("dash:visao-geral:%s:%s:%s", dataInicioStr, dataFimStr, joinIntsCSV(filiais))
		var cached DashboardVisaoGeralResponse
		if err := b.RedisDB.Cache.Get(ctx, cacheKey, &cached); err == nil {
			cached.Perf.Cached = true
			cached.Perf.MsTotal = time.Since(startReq).Milliseconds()
			c.JSON(http.StatusOK, cached)
			b.Logger.Info("dashboard_visao_geral_cache_hit", zap.Int64("ms", cached.Perf.MsTotal))
			return
		} else if err != nil && err != cache.ErrCacheMiss {
			b.Logger.Warn("dashboard_visao_geral_cache_error", zap.Error(err))
		}

		var exists any
		if err := b.DB.DB.QueryRow(ctx, `
-- # Checagem de existência da tabela base
SELECT to_regclass('pcnfsaid')
`).Scan(&exists); err != nil || exists == nil {
			var res DashboardVisaoGeralResponse
			res.Filters.DataInicio = dataInicioStr
			res.Filters.DataFim = dataFimStr
			res.Filters.Filiais = filiais
			res.Warning = "Sem dados: tabela base não encontrada no banco (pcnfsaid)."
			res.Perf.Cached = false
			res.Perf.MsTotal = time.Since(startReq).Milliseconds()
			c.JSON(http.StatusOK, res)
			return
		}

		if len(filiais) == 0 {
			rows, err := b.DB.DB.Query(ctx, `
-- # Filtro de Filial (fallback: todas as filiais)
SELECT DISTINCT n.codfilial
FROM pcnfsaid n
WHERE n.codfilial IS NOT NULL
ORDER BY n.codfilial
LIMIT 200
`)
			if err == nil {
				for rows.Next() {
					var id int
					if err := rows.Scan(&id); err == nil {
						filiais = append(filiais, id)
					}
				}
				rows.Close()
				sort.Ints(filiais)
			}
		}

		var res DashboardVisaoGeralResponse
		res.Filters.DataInicio = dataInicioStr
		res.Filters.DataFim = dataFimStr
		res.Filters.Filiais = filiais

		// # Cards de Métricas (Faturamento, Clientes, NFs, Ticket Médio)
		// Exemplo de estrutura SQL (NÃO usar concatenação; usar parâmetros):
		// SELECT SUM(vltotal) FROM pcnfsaid WHERE dtsaida BETWEEN '${data_inicial}' AND '${data_final}' AND codfilial IN (${filiais_selecionadas});
		sqlMetrics := `
-- # Cards de Métricas
SELECT
  COALESCE(SUM(n.vltotal), 0)::float8 AS faturamento_total,
  COALESCE(COUNT(DISTINCT n.codcli), 0)::int8 AS clientes_ativos,
  COALESCE(COUNT(*), 0)::int8 AS total_nfs
FROM pcnfsaid n
WHERE n.dtsaida >= $1::date
  AND n.dtsaida <= $2::date
  AND n.codfilial = ANY($3::int[])
`
		t0 := time.Now()
		var faturamento float64
		var clientesAtivos int64
		var totalNFs int64
		err = b.DB.DB.QueryRow(ctx, sqlMetrics, dataInicioStr, dataFimStr, filiais).Scan(&faturamento, &clientesAtivos, &totalNFs)
		b.Logger.Info("dashboard_visao_geral_query", zap.String("name", "metrics"), zap.Int64("ms", time.Since(t0).Milliseconds()))
		if err != nil && err != pgx.ErrNoRows {
			b.Logger.Warn("dashboard_visao_geral_metrics_error", zap.Error(err))
		}
		res.Metrics.FaturamentoTotal = faturamento
		res.Metrics.ClientesAtivos = clientesAtivos
		res.Metrics.TotalNFs = totalNFs
		if totalNFs > 0 {
			res.Metrics.TicketMedio = faturamento / float64(totalNFs)
		}

		// # Gráfico mes a mes
		sqlMensal := `
-- # Gráfico mes a mes
SELECT
  to_char(date_trunc('month', n.dtsaida), 'YYYY-MM-01') AS mes,
  COALESCE(SUM(n.vltotal), 0)::float8 AS faturamento
FROM pcnfsaid n
WHERE n.dtsaida >= $1::date
  AND n.dtsaida <= $2::date
  AND n.codfilial = ANY($3::int[])
GROUP BY 1
ORDER BY 1
`
		t1 := time.Now()
		rows, err := b.DB.DB.Query(ctx, sqlMensal, dataInicioStr, dataFimStr, filiais)
		b.Logger.Info("dashboard_visao_geral_query", zap.String("name", "mensal"), zap.Int64("ms", time.Since(t1).Milliseconds()))
		if err == nil {
			for rows.Next() {
				var mes string
				var v float64
				if err := rows.Scan(&mes, &v); err == nil {
					res.Mensal = append(res.Mensal, DashboardMensalItem{Mes: mes, Faturamento: v})
				}
			}
			rows.Close()
		} else {
			b.Logger.Warn("dashboard_visao_geral_mensal_error", zap.Error(err))
		}

		type topRow struct {
			nome        string
			faturamento float64
		}

		runTop := func(name string, sql string, dest *[]DashboardTopItem) {
			t := time.Now()
			rows, err := b.DB.DB.Query(ctx, sql, dataInicioStr, dataFimStr, filiais)
			b.Logger.Info("dashboard_visao_geral_query", zap.String("name", name), zap.Int64("ms", time.Since(t).Milliseconds()))
			if err != nil {
				b.Logger.Warn("dashboard_visao_geral_top_error", zap.String("name", name), zap.Error(err))
				return
			}
			defer rows.Close()

			tmp := make([]topRow, 0, 10)
			for rows.Next() {
				var nome string
				var v float64
				if err := rows.Scan(&nome, &v); err == nil {
					tmp = append(tmp, topRow{nome: nome, faturamento: v})
				}
			}
			for _, r := range tmp {
				pct := 0.0
				if res.Metrics.FaturamentoTotal > 0 {
					pct = (r.faturamento / res.Metrics.FaturamentoTotal) * 100
				}
				*dest = append(*dest, DashboardTopItem{Nome: r.nome, Faturamento: r.faturamento, Percentual: pct})
			}
		}

		// # Top 10 Clientes
		runTop("top_clientes", `
-- # Top 10 Clientes
SELECT
  COALESCE(c.cliente, CAST(n.codcli AS text)) AS nome,
  COALESCE(SUM(n.vltotal), 0)::float8 AS faturamento
FROM pcnfsaid n
LEFT JOIN pcclient c ON c.codcli = n.codcli
WHERE n.dtsaida >= $1::date
  AND n.dtsaida <= $2::date
  AND n.codfilial = ANY($3::int[])
GROUP BY 1
ORDER BY 2 DESC
LIMIT 10
`, &res.Top.Clientes)

		// # Top 10 Vendedores
		runTop("top_vendedores", `
-- # Top 10 Vendedores
SELECT
  COALESCE(u.nome, CAST(n.codusur AS text)) AS nome,
  COALESCE(SUM(n.vltotal), 0)::float8 AS faturamento
FROM pcnfsaid n
LEFT JOIN pcusuari u ON u.codusur = n.codusur
WHERE n.dtsaida >= $1::date
  AND n.dtsaida <= $2::date
  AND n.codfilial = ANY($3::int[])
GROUP BY 1
ORDER BY 2 DESC
LIMIT 10
`, &res.Top.Vendedores)

		// # Top 10 Produtos
		runTop("top_produtos", `
-- # Top 10 Produtos
SELECT
  COALESCE(p.descricao, CAST(i.codprod AS text)) AS nome,
  COALESCE(SUM(i.vltotal), 0)::float8 AS faturamento
FROM pcnfitem i
JOIN pcnfsaid n ON n.numtransvenda = i.numtransvenda
LEFT JOIN pcprodut p ON p.codprod = i.codprod
WHERE n.dtsaida >= $1::date
  AND n.dtsaida <= $2::date
  AND n.codfilial = ANY($3::int[])
GROUP BY 1
ORDER BY 2 DESC
LIMIT 10
`, &res.Top.Produtos)

		// # Top 10 Fornecedores
		runTop("top_fornecedores", `
-- # Top 10 Fornecedores
SELECT
  COALESCE(f.fornecedor, CAST(p.codfornec AS text)) AS nome,
  COALESCE(SUM(i.vltotal), 0)::float8 AS faturamento
FROM pcnfitem i
JOIN pcnfsaid n ON n.numtransvenda = i.numtransvenda
LEFT JOIN pcprodut p ON p.codprod = i.codprod
LEFT JOIN pcfornec f ON f.codfornec = p.codfornec
WHERE n.dtsaida >= $1::date
  AND n.dtsaida <= $2::date
  AND n.codfilial = ANY($3::int[])
GROUP BY 1
ORDER BY 2 DESC
LIMIT 10
`, &res.Top.Fornecedores)

		res.Perf.Cached = false
		res.Perf.MsTotal = time.Since(startReq).Milliseconds()

		if err := b.RedisDB.Cache.Set(&cache.Item{
			Ctx:   c.Request.Context(),
			Key:   cacheKey,
			Value: &res,
			TTL:   30 * time.Second,
		}); err != nil {
			b.Logger.Warn("dashboard_visao_geral_cache_set_error", zap.Error(err))
		}

		c.JSON(http.StatusOK, res)
		b.Logger.Info("dashboard_visao_geral_ok", zap.Int64("ms", res.Perf.MsTotal))
	})

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
