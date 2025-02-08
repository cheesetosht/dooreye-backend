package api

import (
	"dooreye-backend/internal/model"
	"dooreye-backend/internal/store"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrUserInactive      = errors.New("user is inactive")
	ErrInvalidAuthHeader = errors.New("invalid authorization header")
	ErrUnauthorizedRole  = errors.New("unauthorized role")
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserRoleKey  contextKey = "user_role"
	SocietyIDKey contextKey = "society_id"
)

type AuthUser struct {
	ID        string
	Role      model.UserRole
	SocietyID *int64
	IsActive  bool
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(token, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			h.respondError(c, http.StatusUnauthorized, ErrInvalidAuthHeader)
			c.Abort()
			return
		}

		deviceID := parts[1]

		user, err := h.db.GetUserByDeviceID(c.Request.Context(), deviceID)
		if err != nil {
			var status int
			switch {
			case errors.Is(err, store.ErrNotFound):
				status = http.StatusNotFound
				err = ErrInvalidToken
			default:
				status = http.StatusUnauthorized
			}
			c.JSON(status, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set(string(UserIDKey), user.ID)
		c.Set(string(UserRoleKey), user.Role)
		if user.SocietyID != nil {
			c.Set(string(SocietyIDKey), *user.SocietyID)
		}

		c.Next()
	}
}

func GetAuthUser(c *gin.Context) (*AuthUser, error) {
	userID, exists := c.Get(string(UserIDKey))
	if !exists {
		return nil, ErrInvalidToken
	}

	userRole, exists := c.Get(string(UserRoleKey))
	if !exists {
		return nil, ErrInvalidToken
	}

	user := &AuthUser{
		ID:   userID.(string),
		Role: userRole.(model.UserRole),
	}

	if societyID, exists := c.Get(string(SocietyIDKey)); exists {
		sid := societyID.(int64)
		user.SocietyID = &sid
	}

	return user, nil
}

func (s *Handler) LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		s.log.Info("request completed",
			"path", path,
			"method", c.Request.Method,
			"status", c.Writer.Status(),
			"latency", time.Since(start),
			"client_ip", c.ClientIP(),
		)
	}
}
