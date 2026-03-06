package handlers

import (
	"net/http"
	"strings"
	"war-room-backend/internal/db"
	"war-room-backend/internal/models"
	"war-room-backend/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type AuthHandler struct {
	AuthService *services.AuthService
}

func NewAuthHandler(as *services.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: as}
}

type RegisterRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	BatchCode string `json:"batchCode"`
}

type LoginRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	BatchCode string `json:"batchCode"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	req := new(RegisterRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Validate batch code
	batchCode := strings.ToUpper(strings.TrimSpace(req.BatchCode))
	if batchCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Batch code is required"})
	}
	var batch models.Batch
	if err := db.DB.Where("code = ? AND active = ?", batchCode, true).First(&batch).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid or inactive batch code"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}

	// Check if user exists
	var existingUser models.User
	if err := db.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Email already registered"})
	}

	hashedPassword, err := h.AuthService.HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not hash password"})
	}

	user := models.User{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword,
		BatchCode: batchCode,
		Role:      "participant",
	}

	if err := db.DB.Create(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create user"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

func (h *AuthHandler) Login(c echo.Context) error {
	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Look up user by email first
	var user models.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}

	if !h.AuthService.CheckPasswordHash(req.Password, user.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	// Admin login: batch code is optional
	if user.Role == "admin" {
		token, err := h.AuthService.GenerateToken(&user)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not generate token"})
		}
		return c.JSON(http.StatusOK, map[string]any{
			"token": token,
			"user": map[string]any{
				"id":    user.ID,
				"email": user.Email,
				"name":  user.Name,
				"role":  user.Role,
			},
		})
	}

	// Participant login: batch code is required
	batchCode := strings.ToUpper(strings.TrimSpace(req.BatchCode))
	if batchCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Batch code is required"})
	}
	var batch models.Batch
	if err := db.DB.Where("code = ? AND active = ?", batchCode, true).First(&batch).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or inactive batch code"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}

	if user.BatchCode != batchCode {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Batch code does not match your account"})
	}

	token, err := h.AuthService.GenerateToken(&user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not generate token"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"token": token,
		"user": map[string]any{
			"id":        user.ID,
			"email":     user.Email,
			"name":      user.Name,
			"batchCode": user.BatchCode,
			"role":      user.Role,
		},
		"batch": map[string]any{
			"code":  batch.Code,
			"name":  batch.Name,
			"level": batch.Level,
		},
	})
}

// ============================================
// Admin Middleware
// ============================================

// AdminOnly is an Echo middleware that restricts access to admin users.
// Must be applied AFTER the JWT auth middleware.
func AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userToken, ok := c.Get("user").(*jwt.Token)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		claims := userToken.Claims.(jwt.MapClaims)
		role, _ := claims["role"].(string)
		if role != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Admin access required"})
		}
		return next(c)
	}
}

func (h *AuthHandler) Me(c echo.Context) error {
	userToken, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	claims := userToken.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	var user models.User
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"id":        user.ID,
		"email":     user.Email,
		"name":      user.Name,
		"batchCode": user.BatchCode,
		"role":      user.Role,
	})
}
