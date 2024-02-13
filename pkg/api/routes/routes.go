package routes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/metal-toolbox/fleet-rest-skeleton/internal/app"
	"github.com/metal-toolbox/fleet-rest-skeleton/internal/metrics"
	"github.com/metal-toolbox/fleet-rest-skeleton/internal/version"
	"go.hollow.sh/toolbox/ginauth"
	"go.hollow.sh/toolbox/ginjwt"
	"go.uber.org/zap"
)

var (
	readTimeout  = 10 * time.Second
	writeTimeout = 20 * time.Second

	authMiddleWare *ginauth.MultiTokenMiddleware
	ginNoOp        = func(_ *gin.Context) {}
)

// apiHandler is a function that performs real work for this API.
type apiHandler func(map[string]any) (map[string]any, error)

func composeAppLogging(l *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next() // call the next function in the chain
		code := c.Writer.Status()
		metrics.APICallEpilog(start, path, code)

		fields := []zap.Field{
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status-code", code),
			zap.Time("start", start),
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.Strings("errors", c.Errors.Errors()))
			l.Error("errors on API request",
				fields...,
			)
			return
		}

		l.Info("api call complete", fields...)
	}
}

// ComposeHTTPServer returns an http.Server that handles our API
func ComposeHTTPServer(app *app.App) *http.Server {
	if len(app.Cfg.JWTAuth) != 0 {
		var err error
		authMiddleWare, err = ginjwt.NewMultiTokenMiddlewareFromConfigs(app.Cfg.JWTAuth...)
		if err != nil {
			app.Log.Fatal(
				"failed to initialize auth middleware",
				zap.Error(err),
			)
		}
	}

	g := gin.New()

	if !app.Cfg.DeveloperMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// set up common middleware for logging and metrics
	g.Use(composeAppLogging(app.Log), gin.Recovery())

	// some boilerplate setup
	g.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound,
			gin.H{
				"message": "invalid request - route not found",
			},
		)
	})

	// a liveness endpoint
	g.GET("/_health/liveness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"time": time.Now()})
	})

	g.GET("/api/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, version.Current())
	})

	g.POST("/api/echo",
		composeAuthHandler(createScopes("response")), // auth handler
		wrapAPICall(apiEcho))                         // api function, wrapped into middleware

	g.POST("/api/error",
		composeAuthHandler(createScopes("response")),
		wrapAPICall(apiError))

	// add other API endpoints to the gin Engine as required

	return &http.Server{
		Addr:         app.Cfg.ListenAddress,
		Handler:      g,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}
}

// wrapAPICall is an adapter for any arbitrary code so that you can isolate your
// logic from having to take gin-specific data structures and whatnot. It assumes
// your API function takes a map[string]any and returns a JSON-serializable result
// and an error. This function could be altered to pull any kind of parameter out
// of the raw JSON input.
func wrapAPICall(fn apiHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var responseCode int

		m := make(map[string]any)
		if err := ctx.BindJSON(&m); err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]any{
				"error": err.Error(),
			})
		}

		obj, err := fn(m)
		if err == nil {
			responseCode = http.StatusOK
		} else {
			responseCode = http.StatusInternalServerError
			obj = map[string]any{
				"error": err.Error(),
			}
		}
		ctx.JSON(responseCode, obj)
	}
}

func composeAuthHandler(scopes []string) gin.HandlerFunc {
	if authMiddleWare == nil {
		return ginNoOp
	}
	return authMiddleWare.AuthRequired(scopes)
}

func createScopes(items ...string) []string {
	s := []string{"write", "create"}
	for _, i := range items {
		s = append(s, fmt.Sprintf("create:%s", i))
	}

	return s
}

func readScopes(items ...string) []string {
	s := []string{"read"}
	for _, i := range items {
		s = append(s, fmt.Sprintf("read:%s", i))
	}

	return s
}

func updateScopes(items ...string) []string {
	s := []string{"write", "update"}
	for _, i := range items {
		s = append(s, fmt.Sprintf("update:%s", i))
	}

	return s
}

func deleteScopes(items ...string) []string {
	s := []string{"write", "delete"}
	for _, i := range items {
		s = append(s, fmt.Sprintf("delete:%s", i))
	}

	return s
}
