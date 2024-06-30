package middleware

import (
	"hime-backend/repository"
	"hime-backend/utility"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func AuthByRoleLevel(roleLevel int) fiber.Handler {
	// refer DB for role level mapping
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := utility.ParseJWTToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		isValid, err := repository.CheckAuthByID(int(claims.UserID), roleLevel)
		if isValid {
			c.Locals("user_id", claims.UserID)
			c.Locals("role_level", claims.RoleLevel)
			return c.Next()
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "access denied"})
	}
}
