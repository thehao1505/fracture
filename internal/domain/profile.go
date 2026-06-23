package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type BlockType string

func (b BlockType) String() string { return string(b) }

// IsValid báo BlockType có nằm trong tập type được hỗ trợ không.
func (b BlockType) IsValid() bool {
	switch b {
	case BlockTypeLink, BlockTypeSocials, BlockTypeHeader:
		return true
	default:
		return false
	}
}

const (
	BlockTypeLink    BlockType = "link"
	BlockTypeSocials BlockType = "socials"
	BlockTypeHeader  BlockType = "header"
)

type Block struct {
	ID         uuid.UUID       `json:"id"`
	ProfileID  uuid.UUID       `json:"profile_id"`
	Type       BlockType       `json:"type"`
	Content    json.RawMessage `json:"content"`
	Position   int32           `json:"position"`
	IsActive   bool            `json:"is_active"`
	ClickCount int64           `json:"click_count"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type Profile struct {
	ID          uuid.UUID       `json:"id"`
	UserID      uuid.UUID       `json:"user_id"`
	Username    string          `json:"username"`
	DisplayName string          `json:"display_name"`
	Bio         string          `json:"bio"`
	AvatarURL   string          `json:"avatar_url"`
	Appearance  json.RawMessage `json:"appearance"`
	IsPublished bool            `json:"is_published"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Blocks      []Block         `json:"blocks,omitempty"`
}

type PublicProfile struct {
	Username    string          `json:"username"`
	DisplayName string          `json:"display_name"`
	Bio         string          `json:"bio"`
	AvatarURL   string          `json:"avatar_url"`
	Appearance  json.RawMessage `json:"appearance"`
	Blocks      []PublicBlock   `json:"blocks"`
}

type PublicBlock struct {
	ID      uuid.UUID       `json:"id"`
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}
