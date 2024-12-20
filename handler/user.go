package handler

import (
	"app/database"
	"app/dto/model"
	"fmt"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func validToken(t *jwt.Token, id string) bool {
	n, err := strconv.Atoi(id)
	if err != nil {
		return false
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	uid, ok := claims["user_id"].(float64) // Menggunakan float64 karena JWT claims biasanya dalam format float64
	if !ok {
		return false
	}

	return int(uid) == n
}
func validUser(id string, p string) bool {
	db := database.DB
	var user model.User
	db.First(&user, id)
	if user.Username == "" {
		return false
	}
	if !CheckPasswordHash(p, user.Password) {
		return false
	}
	return true
}

func generateJWT(user model.User) (string, error) {

	claims := jwt.MapClaims{}
	claims["user_id"] = user.ID
	claims["username"] = user.Username
	claims["email"] = user.Email
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("dcbsecret"))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// GetUser get a user
func GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	db := database.DB
	var user model.User
	db.Find(&user, id)
	if user.Username == "" {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "No user found with ID", "data": nil})
	}
	return c.JSON(fiber.Map{"status": "success", "message": "User found", "data": user})
}

func CreateUser(c *fiber.Ctx) error {
	type NewUser struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
		// Role     string `json:"role" validate:"oneof=admin superadmin merchant"` // Role harus valid
	}

	user := new(NewUser)
	if err := c.BodyParser(user); err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Review your input", "errors": err.Error()})
	}

	// Validasi request
	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body", "errors": err.Error()})
	}

	// Hash password
	hash, err := hashPassword(user.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Couldn't hash password", "errors": err.Error()})
	}

	// Set hashed password to user
	newUser := model.User{
		Username: user.Username,
		Email:    user.Email,
		Password: hash,
		Role:     "merchant",
	}

	// Save user to database
	if err := database.DB.Create(&newUser).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Couldn't create user", "errors": err.Error()})
	}

	// Generate JWT token untuk pengguna
	token, err := generateJWT(newUser)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Couldn't generate JWT token", "errors": err.Error()})
	}

	type UserResponse struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}

	// Create response without password
	response := UserResponse{
		Username: newUser.Username,
		Email:    newUser.Email,
		Role:     newUser.Role,
	}
	// Return response dengan JWT
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Created user",
		"data":    response,
		"token":   token,
	})
}

func UpdateUser(c *fiber.Ctx) error {
	type UpdateUserInput struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role" validate:"oneof=admin superadmin merchant"`
		Names    string `json:"names"`
	}
	var uui UpdateUserInput
	if err := c.BodyParser(&uui); err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Review your input", "errors": err.Error()})
	}

	id := c.Params("id")
	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userRole := claims["role"].(string)

	if userRole != "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "error", "message": "Only superadmin can update role", "data": nil})
	}

	db := database.DB
	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "User not found", "data": nil})
	}

	if uui.Username != "" {
		user.Username = uui.Username
	}
	if uui.Email != "" {
		user.Email = uui.Email
	}
	if uui.Role != "" {
		user.Role = uui.Role
	}
	if uui.Names != "" {
		user.Names = uui.Names
	}

	if err := db.Save(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Failed to update user", "data": nil})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "User successfully updated", "data": user})
}

// DeleteUser delete user
func DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	token := c.Locals("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	userRole := claims["role"].(string)

	if userRole != "superadmin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "error", "message": "Only superadmin can update role", "data": nil})
	}

	db := database.DB
	var user model.User

	db.First(&user, id)

	if err := db.Delete(&user).Error; err != nil {
		return fmt.Errorf("unable to delete client: %w", err)
	}

	return c.JSON(fiber.Map{"status": "success", "message": "User successfully deleted", "data": nil})
}
