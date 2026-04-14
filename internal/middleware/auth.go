package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ysrn87/warung-ku/internal/models"
)

type Claims struct {
	UserID string      `json:"userId"`
	Name   string      `json:"name"`
	Role   models.Role `json:"role"`
	jwt.RegisteredClaims
}

// validateAuth reads & validates the JWT cookie.
// It sets claims in locals and returns nil on success.
// It does NOT call c.Next() — that is the caller's responsibility.
func validateAuth(c *fiber.Ctx) error {
	tokenStr := c.Cookies("auth_token")

	if tokenStr == "" {
		auth := c.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr = strings.TrimPrefix(auth, "Bearer ")
		}
	}

	if tokenStr == "" {
		return c.Redirect("/login")
	}

	claims, err := parseToken(tokenStr)
	if err != nil {
		c.ClearCookie("auth_token")
		return c.Redirect("/login")
	}

	c.Locals("claims", claims)
	return nil // success — caller decides what to do next
}

// RequireAuth can be used as standalone middleware on a single route.
// (When used inside RequireRole, use validateAuth instead.)
func RequireAuth(c *fiber.Ctx) error {
	if err := validateAuth(c); err != nil {
		return err
	}
	// Check that validateAuth actually set claims (redirect returns nil too)
	if GetClaims(c) == nil {
		return nil // redirect was already queued, stop chain
	}
	return c.Next()
}

// RequireRole checks that the caller has one of the given roles.
// It calls validateAuth (not RequireAuth) to avoid a double c.Next().
func RequireRole(roles ...models.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := validateAuth(c); err != nil {
			return err
		}

		claims := GetClaims(c)
		if claims == nil {
			// validateAuth sent a redirect and returned nil — stop here
			return nil
		}

		for _, role := range roles {
			if claims.Role == role {
				return c.Next() // ← only one c.Next(), here, by the middleware
			}
		}

		return c.Status(fiber.StatusForbidden).SendString("Access denied")
	}
}

// GetClaims extracts JWT claims from the Fiber context.
func GetClaims(c *fiber.Ctx) *Claims {
	claims, _ := c.Locals("claims").(*Claims)
	return claims
}

func parseToken(tokenStr string) (*Claims, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return token.Claims.(*Claims), nil
}
