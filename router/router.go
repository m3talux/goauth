package router

import (
	"net/http"

	"github.com/m3talux/goauth/config"
	"github.com/m3talux/goauth/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Router struct {
	*gin.Engine
	Handlers Handlers
}

type Handlers struct {
	CheckHandler *handler.CheckHandler
}

func NewRouter(handlers Handlers) Router {
	// Use Gin release mode by default
	if config.GinMode() == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := Router{
		Engine:   gin.New(),
		Handlers: handlers,
	}

	// Middlewares
	r.Use(gin.Recovery(), corsMiddleware())

	// Entrypoints
	r.registerMonitoring()
	r.registerAPI()

	r.Static("/openapi", "openapi/")

	r.GET(config.APIPath(), func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/openapi/openapi.html")
	})

	return r
}

func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := config.CorsAllowedOrigins()

	if allowedOrigins == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	corsConfig := cors.Config{
		AllowOrigins:  config.CorsAllowedOrigins(),
		AllowMethods:  []string{"GET"},
		AllowHeaders:  []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders: []string{"Content-Disposition", "Content-Transfer-Encoding", "Content-Description"},
		MaxAge:        config.CorsMaxAge(),
		AllowWildcard: true,
	}

	return cors.New(corsConfig)
}

func (r *Router) registerMonitoring() {
	r.GET("/", r.Handlers.CheckHandler.Alive)
	r.GET("/ready", r.Handlers.CheckHandler.Ready)
}

func (r *Router) registerAPI() {
	_ = r.Group(config.APIPath())
}
