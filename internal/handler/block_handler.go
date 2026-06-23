package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lukenguyen/fracture/internal/domain"
	"github.com/lukenguyen/fracture/internal/usecase"
)

type BlockHandler struct {
	profileUc *usecase.ProfileUseCase
}

func NewBlockHandler(uc *usecase.ProfileUseCase) *BlockHandler {
	return &BlockHandler{profileUc: uc}
}

type createBlockRequest struct {
	Type    string          `json:"type" binding:"required"`
	Content json.RawMessage `json:"content" binding:"required"`
}

type updateBlockRequest struct {
	Type    string          `json:"type" binding:"required"`
	Content json.RawMessage `json:"content" binding:"required"`
	// IsActive là con trỏ để phân biệt "không gửi" (giữ nguyên) với false (ẩn).
	IsActive *bool `json:"is_active"`
}

type reorderRequest struct {
	Order []string `json:"order" binding:"required"`
}

// List godoc
// @Summary List my blocks
// @Description Returns all blocks (including hidden ones) for the authenticated user's profile.
// @Tags blocks
// @Produce json
// @Success 200 {object} map[string]interface{} "Blocks"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Profile not found"
// @Security BearerAuth
// @Router /me/blocks [get]
func (h *BlockHandler) List(c *gin.Context) {
	blocks, err := h.profileUc.ListMyBlocks(c.Request.Context(), currentUserID(c))
	if err != nil {
		respondDomainError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": blocks})
}

// Create godoc
// @Summary Create a block
// @Description Adds a block to the authenticated user's profile.
// @Tags blocks
// @Accept json
// @Produce json
// @Param request body createBlockRequest true "Create block payload"
// @Success 201 {object} map[string]interface{} "Block created"
// @Failure 400 {object} map[string]string "Invalid request body or content"
// @Failure 404 {object} map[string]string "Profile not found"
// @Security BearerAuth
// @Router /me/blocks [post]
func (h *BlockHandler) Create(c *gin.Context) {
	var req createBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b := domain.Block{
		Type:    domain.BlockType(req.Type),
		Content: req.Content,
	}

	if err := h.profileUc.CreateBlock(c.Request.Context(), currentUserID(c), &b); err != nil {
		respondDomainError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": b})
}

// Update godoc
// @Summary Update a block
// @Description Updates type/content/visibility of a block owned by the authenticated user.
// @Tags blocks
// @Accept json
// @Produce json
// @Param id path string true "Block ID (UUID)"
// @Param request body updateBlockRequest true "Update block payload"
// @Success 200 {object} map[string]string "Block updated"
// @Failure 400 {object} map[string]string "Invalid block ID, request body or content"
// @Failure 404 {object} map[string]string "Profile not found"
// @Security BearerAuth
// @Router /me/blocks/{id} [put]
func (h *BlockHandler) Update(c *gin.Context) {
	blockID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrInvalidID.Error()})
		return
	}

	var req updateBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b := domain.Block{
		ID:      blockID,
		Type:    domain.BlockType(req.Type),
		Content: req.Content,
	}

	if err := h.profileUc.UpdateBlock(c.Request.Context(), currentUserID(c), &b, req.IsActive); err != nil {
		respondDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "block updated successfully"})
}

// Delete godoc
// @Summary Delete a block
// @Description Deletes a block owned by the authenticated user.
// @Tags blocks
// @Produce json
// @Param id path string true "Block ID (UUID)"
// @Success 200 {object} map[string]string "Block deleted"
// @Failure 400 {object} map[string]string "Invalid block ID"
// @Failure 404 {object} map[string]string "Profile not found"
// @Security BearerAuth
// @Router /me/blocks/{id} [delete]
func (h *BlockHandler) Delete(c *gin.Context) {
	blockID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrInvalidID.Error()})
		return
	}

	if err := h.profileUc.DeleteBlock(c.Request.Context(), currentUserID(c), blockID); err != nil {
		respondDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "block deleted successfully"})
}

// Reorder godoc
// @Summary Reorder blocks
// @Description Sets the position of every block from the given ordered list of block IDs.
// @Tags blocks
// @Accept json
// @Produce json
// @Param request body reorderRequest true "Ordered block IDs"
// @Success 200 {object} map[string]string "Blocks reordered"
// @Failure 400 {object} map[string]string "Invalid block ID or order does not match existing blocks"
// @Failure 404 {object} map[string]string "Profile not found"
// @Security BearerAuth
// @Router /me/blocks/reorder [patch]
func (h *BlockHandler) Reorder(c *gin.Context) {
	var req reorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ids := make([]uuid.UUID, 0, len(req.Order))
	for _, raw := range req.Order {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": domain.ErrInvalidID.Error()})
			return
		}
		ids = append(ids, id)
	}

	if err := h.profileUc.ReorderBlocks(c.Request.Context(), currentUserID(c), ids); err != nil {
		respondDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "blocks reordered successfully"})
}
