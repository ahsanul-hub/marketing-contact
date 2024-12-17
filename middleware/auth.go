package middleware

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected protect routes
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil token dari header Authorization
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Missing or malformed JWT"})
		}

		// Memisahkan "Bearer " dari token
		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		} else {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid token format"})
		}

		// Verifikasi token dan ambil claims
		claims := jwt.MapClaims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			// Pastikan algoritma yang digunakan sesuai
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
			}
			return []byte("dcbsecret"), nil // Ganti dengan secret key Anda
		})

		log.Println("Received token:", token)
		log.Println("Error parsing token:", err)

		if err != nil {
			return jwtError(c, err) // Panggil fungsi error handling
		}

		// Pastikan token valid
		if !parsedToken.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid token"})
		}

		// Periksa klaim role
		role, ok := claims["role"].(string)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Role claim is missing or invalid"})
		}

		c.Locals("role", role)        // Simpan role ke context
		c.Locals("user", parsedToken) // Simpan token ke context
		return c.Next()
	}
}

func jwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"status": "error", "message": "Missing or malformed JWT", "data": nil})
	}
	return c.Status(fiber.StatusUnauthorized).
		JSON(fiber.Map{"status": "error", "message": "Invalid or expired JWT", "data": nil})
}

// Definisikan permission untuk setiap role
var rolesPermissions = map[string][]string{
	"superadmin": {"create", "read", "update", "delete", "manage_users", "manage_merchants"},
	"admin":      {"create", "read", "update", "delete"},
	"merchant":   {"read"},
}

// RBACMiddleware untuk memeriksa permissions berdasarkan role
func RBACMiddleware(neededPermissions []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil role pengguna dari context
		role, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Forbidden"})
		}

		// Ambil permissions yang terkait dengan role
		permissions, exists := rolesPermissions[role]
		if !exists {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Forbidden"})
		}

		// Periksa apakah role mempunyai permission yang diperlukan
		for _, neededPerm := range neededPermissions {
			hasPermission := false
			for _, perm := range permissions {
				if perm == neededPerm {
					hasPermission = true
					break
				}
			}
			if !hasPermission {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Forbidden"})
			}
		}

		return c.Next()
	}
}

func AdminOnly(allowAdmin bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil role pengguna dari context
		role, ok := c.Locals("role").(string)
		log.Println("role", role)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Forbidden: You do not have access to this resource."})
		}

		// Cek role
		if role == "superadmin" {
			return c.Next() // Superadmin selalu diizinkan
		}

		if allowAdmin && role == "admin" {
			return c.Next() // Jika admin diizinkan
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Forbidden: You do not have access to this resource."})
	}
}
