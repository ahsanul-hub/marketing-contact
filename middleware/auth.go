package middleware

import (
	"app/repository"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Missing or malformed JWT"})
		}
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

		// log.Println("Error parsing token:", err)

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

func RBACMiddleware(neededPermissions []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Forbidden"})
		}

		permissions, exists := rolesPermissions[role]
		if !exists {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Forbidden"})
		}

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

func AdminOnly(superAdminOnly bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Forbidden: You do not have access to this resource."})
		}

		// Superadmin always has access
		if role == "superadmin" {
			return c.Next()
		}

		// If superAdminOnly is true, only superadmin can access
		if superAdminOnly {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Forbidden: Superadmin access required."})
		}

		// If superAdminOnly is false, both superadmin and admin can access
		if role == "admin" {
			return c.Next()
		}

		// Merchant cannot access admin endpoints at all
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Forbidden: Admin access required."})
	}
}

// ClientAuth middleware untuk validasi client berdasarkan token dan header appkey/appid
func ClientAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil token dari context yang sudah divalidasi oleh middleware Protected
		token := c.Locals("user")
		if token == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Unauthorized access",
			})
		}

		// Ambil header appkey dan appid
		appKey := c.Get("appkey")
		appID := c.Get("appid")

		if appKey == "" || appID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Missing required headers: appkey and appid",
			})
		}

		// Validasi bahwa client dengan appkey dan appid ini ada
		client, err := repository.FindClient(c.Context(), appKey, appID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Client not found",
			})
		}

		// Simpan client data ke context untuk digunakan di handler
		c.Locals("client", client)
		return c.Next()
	}
}
