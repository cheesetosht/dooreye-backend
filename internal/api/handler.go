package api

import (
	"context"
	"dooreye-backend/internal/store"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	db     *store.DB
	log    *slog.Logger
	router *gin.Engine
	srv    *http.Server
}

func NewHandler(db *store.DB, log *slog.Logger) *Handler {
	h := &Handler{
		db:  db,
		log: log,
	}

	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(h.LoggerMiddleware())

	router.GET("/health", h.handleHealth())

	api := router.Group("/api")
	api.Use(h.AuthMiddleware())
	{
		api.GET("/visitors", h.getVisitorByPhone)
		api.GET("/visits", h.getVisits)

		api.POST("/visits/security", h.createVisitAsSecurity)
		api.POST(
			"/visitors/pre-approved",
			h.createPreApprovedVisitor,
		)
		api.POST("/users/activate",
			h.createUser)
	}
	h.router = router
	return h
}

func (h *Handler) Run(addr string) error {
	h.srv = &http.Server{
		Addr:         addr,
		Handler:      h.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return h.srv.ListenAndServe()
}

func (h *Handler) Shutdown(ctx context.Context) error {
	return h.srv.Shutdown(ctx)
}

func (h *Handler) Close() error {
	return h.srv.Close()
}

func (h *Handler) handleHealth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now(),
		})
	}
}

// Helper methods
func (h *Handler) respondError(c *gin.Context, status int, err error) {
	h.log.Error("handler error",
		"error", err,
		"path", c.Request.URL.Path,
		"client_ip", c.ClientIP(),
	)
	c.JSON(status, gin.H{"error": err.Error()})
}
