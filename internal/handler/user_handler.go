package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lukenguyen/fracture/internal/domain"
	"github.com/lukenguyen/fracture/internal/usecase"
)

type UserHandler struct {
	userUc *usecase.UserUseCase
}

type createUserRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required"`
}

type updateUserRequest struct {
	Email string `json:"email" binding:"omitempty,email"`
	Name  string `json:"name"`
}

func NewUserHandler(uc *usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUc: uc}
}

// RegisterRoutes gắn các route user vào group đã cho (group này cần AuthRequired).
func (h *UserHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/:id", h.GetUser)
	rg.POST("", h.CreateUser)
	rg.PUT("/:id", h.UpdateUser)
	rg.DELETE("/:id", h.DeleteUser)
	rg.GET("", h.ListUsers)
}

// GetUser godoc
// @Summary Get a user by ID
// @Description Retrieve a user from the database by providing their UUID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]interface{} "User retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 404 {object} map[string]string "User not found"
// @Security BearerAuth
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := h.userUc.GetUser(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrInvalidID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with email and name
// @Tags users
// @Accept json
// @Produce json
// @Param request body createUserRequest true "Create user payload"
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 409 {object} map[string]string "Email already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := domain.User{
		Email: req.Email,
		Name:  req.Name,
	}

	if err := h.userUc.CreateUser(c.Request.Context(), &user); err != nil {
		if err == domain.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrBadRequest {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		respondInternal(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update a user by ID with email and/or name
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param request body updateUserRequest true "Update user payload"
// @Success 200 {object} map[string]string "User updated successfully"
// @Failure 400 {object} map[string]string "Invalid user ID or request body"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 409 {object} map[string]string "Email already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := domain.User{
		Email: req.Email,
		Name:  req.Name,
	}

	if err := h.userUc.UpdateUser(c.Request.Context(), id, &user); err != nil {
		if err == domain.ErrInvalidID || err == domain.ErrBadRequest {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrConflict {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		respondInternal(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]string "User deleted successfully"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := h.userUc.DeleteUser(c.Request.Context(), id); err != nil {
		if err == domain.ErrInvalidID {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		respondInternal(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

// ListUsers godoc
// @Summary List users
// @Description List users with pagination and optional keyword search on name/email
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 20, max 100)"
// @Param keyword query string false "Search keyword (matches name or email)"
// @Success 200 {object} map[string]interface{} "Users retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security BearerAuth
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page")) // "" -> 0, usecase tự set default
	limit, _ := strconv.Atoi(c.Query("limit"))
	keyword := c.Query("keyword")

	users, total, err := h.userUc.ListUsers(c.Request.Context(), usecase.ListUsersParams{
		Page:    page,
		Limit:   limit,
		Keyword: keyword,
	})
	if err != nil {
		respondInternal(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
