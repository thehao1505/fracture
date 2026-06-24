package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lukenguyen/fracture/internal/domain"
	"github.com/lukenguyen/fracture/internal/handler/middleware"
	"github.com/lukenguyen/fracture/internal/usecase"
)

type ProfileHandler struct {
	profileUc *usecase.ProfileUseCase
}

func NewProfileHandler(uc *usecase.ProfileUseCase) *ProfileHandler {
	return &ProfileHandler{profileUc: uc}
}

// RegisterPublic gắn các route public (không cần token) vào group "/p".
func (h *ProfileHandler) RegisterPublic(rg *gin.RouterGroup) {
	rg.GET("/:username", h.GetPublic)
	rg.POST("/:username/blocks/:id/click", h.RecordClick)
}

// RegisterMe gắn các route profile của chủ sở hữu vào group "/me" (cần AuthRequired).
func (h *ProfileHandler) RegisterMe(rg *gin.RouterGroup) {
	rg.GET("/profile", h.GetMine)
	rg.POST("/profile", h.Create)
	rg.PUT("/profile", h.Update)
}

type createProfileRequest struct {
	Username    string          `json:"username" binding:"required"`
	DisplayName string          `json:"display_name"`
	Bio         string          `json:"bio"`
	AvatarURL   string          `json:"avatar_url"`
	Appearance  json.RawMessage `json:"appearance"`
}

type updateProfileRequest struct {
	Username    string          `json:"username" binding:"required"`
	DisplayName string          `json:"display_name"`
	Bio         string          `json:"bio"`
	AvatarURL   string          `json:"avatar_url"`
	Appearance  json.RawMessage `json:"appearance"`
	IsPublished bool            `json:"is_published"`
}

// currentUserID đọc id user mà AuthRequired đã set vào context.
func currentUserID(c *gin.Context) uuid.UUID {
	return c.MustGet(middleware.ContextUserID).(uuid.UUID)
}

// respondDomainError map lỗi domain sang HTTP status. Lỗi lạ → 500 (qua respondInternal).
func respondDomainError(c *gin.Context, err error) {
	switch err {
	case domain.ErrBadRequest, domain.ErrInvalidID:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case domain.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case domain.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		respondInternal(c, err)
	}
}

// toPublicProfile chuyển domain.Profile sang DTO public (ẩn user_id/is_published/click_count).
func toPublicProfile(p *domain.Profile) domain.PublicProfile {
	blocks := make([]domain.PublicBlock, 0, len(p.Blocks))
	for _, b := range p.Blocks {
		blocks = append(blocks, domain.PublicBlock{
			ID:      b.ID,
			Type:    b.Type.String(),
			Content: b.Content,
		})
	}
	return domain.PublicProfile{
		Username:    p.Username,
		DisplayName: p.DisplayName,
		Bio:         p.Bio,
		AvatarURL:   p.AvatarURL,
		Appearance:  p.Appearance,
		Blocks:      blocks,
	}
}

// GetPublic godoc
// @Summary Get a public profile by username
// @Description Public link-in-bio page. Returns the profile and its active blocks only when published.
// @Tags profiles
// @Produce json
// @Param username path string true "Profile username"
// @Success 200 {object} map[string]interface{} "Public profile"
// @Failure 404 {object} map[string]string "Profile not found or not published"
// @Router /p/{username} [get]
func (h *ProfileHandler) GetPublic(c *gin.Context) {
	username := c.Param("username")

	p, err := h.profileUc.GetPublicProfile(c.Request.Context(), username)
	if err != nil {
		respondDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toPublicProfile(p)})
}

// RecordClick godoc
// @Summary Record a click on a public block
// @Description Increments the block click count and returns the destination URL for the client to redirect to.
// @Tags profiles
// @Produce json
// @Param username path string true "Profile username"
// @Param id path string true "Block ID (UUID)"
// @Success 200 {object} map[string]string "Destination URL"
// @Failure 400 {object} map[string]string "Invalid block ID or non-clickable block"
// @Failure 404 {object} map[string]string "Profile or block not found"
// @Router /p/{username}/blocks/{id}/click [post]
func (h *ProfileHandler) RecordClick(c *gin.Context) {
	username := c.Param("username")

	blockID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrInvalidID.Error()})
		return
	}

	url, err := h.profileUc.RecordClick(c.Request.Context(), username, blockID)
	if err != nil {
		respondDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

// GetMine godoc
// @Summary Get my profile
// @Description Returns the authenticated user's full profile (including drafts and all blocks).
// @Tags profiles
// @Produce json
// @Success 200 {object} map[string]interface{} "Profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Profile not found"
// @Security BearerAuth
// @Router /me/profile [get]
func (h *ProfileHandler) GetMine(c *gin.Context) {
	p, err := h.profileUc.GetMyProfile(c.Request.Context(), currentUserID(c))
	if err != nil {
		respondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": p})
}

// Create godoc
// @Summary Create my profile
// @Description Creates the profile for the authenticated user (one profile per user).
// @Tags profiles
// @Accept json
// @Produce json
// @Param request body createProfileRequest true "Create profile payload"
// @Success 201 {object} map[string]interface{} "Profile created"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 409 {object} map[string]string "Profile or username already exists"
// @Security BearerAuth
// @Router /me/profile [post]
func (h *ProfileHandler) Create(c *gin.Context) {
	var req createProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p := domain.Profile{
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
		AvatarURL:   req.AvatarURL,
		Appearance:  req.Appearance,
	}

	if err := h.profileUc.CreateProfile(c.Request.Context(), currentUserID(c), &p); err != nil {
		respondDomainError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": p})
}

// Update godoc
// @Summary Update my profile
// @Description Updates username/bio/avatar/appearance/publish state for the authenticated user.
// @Tags profiles
// @Accept json
// @Produce json
// @Param request body updateProfileRequest true "Update profile payload"
// @Success 200 {object} map[string]string "Profile updated"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 404 {object} map[string]string "Profile not found"
// @Failure 409 {object} map[string]string "Username already exists"
// @Security BearerAuth
// @Router /me/profile [put]
func (h *ProfileHandler) Update(c *gin.Context) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p := domain.Profile{
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Bio:         req.Bio,
		AvatarURL:   req.AvatarURL,
		Appearance:  req.Appearance,
		IsPublished: req.IsPublished,
	}

	if err := h.profileUc.UpdateProfile(c.Request.Context(), currentUserID(c), &p); err != nil {
		respondDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully"})
}
