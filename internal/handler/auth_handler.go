package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lukenguyen/fracture/internal/domain"
	"github.com/lukenguyen/fracture/internal/usecase"
)

type AuthHandler struct {
	authUc *usecase.AuthUseCase
}

// bcrypt only hashes the first 72 bytes, so cap the password there to avoid
// silently ignoring the tail of very long inputs.
type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Name     string `json:"name" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func NewAuthHandler(uc *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUc: uc}
}

// RegisterRoutes gắn các route auth (public) vào group đã cho.
func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/register", h.Register)
	rg.POST("/login", h.Login)
}

// Register godoc
// @Summary Register a new account
// @Description Create a new user account. The password is hashed before storage.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body registerRequest true "Register payload"
// @Success 201 {object} map[string]interface{} "Account created successfully"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 409 {object} map[string]string "Email already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authUc.Register(c.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		switch err {
		case domain.ErrConflict:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case domain.ErrBadRequest:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			respondInternal(c, err)
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

// Login godoc
// @Summary Log in and obtain an access token
// @Description Verify credentials and return a signed JWT access token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login payload"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Invalid email or password"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, accessToken, err := h.authUc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		respondInternal(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         user,
		"access_token": accessToken,
		"token_type":   "Bearer",
	})
}
