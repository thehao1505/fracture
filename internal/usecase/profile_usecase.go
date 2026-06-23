package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lukenguyen/fracture/internal/domain"
	"github.com/lukenguyen/fracture/internal/repository"
)

// maxBlocksPerProfile chặn lạm dụng — một profile không thể có quá nhiều block.
const maxBlocksPerProfile = 100

type ProfileUseCase struct {
	profileRepo repository.ProfileRepository
	blockRepo   repository.BlockRepository
}

func NewProfileUseCase(p repository.ProfileRepository, b repository.BlockRepository) *ProfileUseCase {
	return &ProfileUseCase{profileRepo: p, blockRepo: b}
}

// ---------------------------------------------------------------------------
// Validation helpers
// ---------------------------------------------------------------------------

var usernameRe = regexp.MustCompile(`^[a-z0-9_.]+$`)

// reservedUsernames là các slug không cho phép vì đụng route/hệ thống.
var reservedUsernames = map[string]struct{}{
	"admin": {}, "api": {}, "app": {}, "login": {}, "register": {},
	"health": {}, "swagger": {}, "me": {}, "p": {}, "www": {},
	"support": {}, "about": {}, "auth": {}, "user": {}, "users": {},
}

// normalizeUsername chuẩn hóa trước khi validate/lưu: luôn lowercase, bỏ khoảng trắng.
func normalizeUsername(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// validateUsername kiểm tra username đã normalize. Trả ErrBadRequest nếu sai.
func validateUsername(u string) error {
	if len(u) < 3 || len(u) > 30 {
		return domain.ErrBadRequest
	}
	if !usernameRe.MatchString(u) {
		return domain.ErrBadRequest
	}
	if strings.HasPrefix(u, ".") || strings.HasSuffix(u, ".") || strings.Contains(u, "..") {
		return domain.ErrBadRequest
	}
	if _, ok := reservedUsernames[u]; ok {
		return domain.ErrBadRequest
	}
	return nil
}

// strictUnmarshal parse JSON và từ chối field thừa (chống lưu dữ liệu tùy tiện).
func strictUnmarshal(raw json.RawMessage, v any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	return d.Decode(v)
}

// validateHTTPURL chỉ cho http/https (chặn javascript:/data: → tránh XSS lúc client render).
func validateHTTPURL(raw string) error {
	if raw == "" {
		return domain.ErrBadRequest
	}
	u, err := url.Parse(raw)
	if err != nil {
		return domain.ErrBadRequest
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return domain.ErrBadRequest
	}
	if u.Host == "" {
		return domain.ErrBadRequest
	}
	return nil
}

// socialPlatforms whitelist nền tảng cho block "socials".
var socialPlatforms = map[string]struct{}{
	"instagram": {}, "github": {}, "x": {}, "twitter": {}, "youtube": {},
	"tiktok": {}, "facebook": {}, "linkedin": {}, "threads": {}, "twitch": {},
	"discord": {}, "telegram": {}, "spotify": {}, "email": {}, "website": {},
}

// validateBlockContent là điểm cốt lõi của mô hình block: validate `content` theo
// `type`. Thêm type mới chỉ cần thêm một nhánh case.
func validateBlockContent(t domain.BlockType, raw json.RawMessage) error {
	switch t {
	case domain.BlockTypeLink:
		var c struct {
			Title     string `json:"title"`
			URL       string `json:"url"`
			Icon      string `json:"icon"`
			Thumbnail string `json:"thumbnail"`
		}
		if err := strictUnmarshal(raw, &c); err != nil {
			return domain.ErrBadRequest
		}
		if c.Title == "" || len(c.Title) > 80 {
			return domain.ErrBadRequest
		}
		return validateHTTPURL(c.URL)

	case domain.BlockTypeSocials:
		var c struct {
			Items []struct {
				Platform string `json:"platform"`
				URL      string `json:"url"`
			} `json:"items"`
		}
		if err := strictUnmarshal(raw, &c); err != nil {
			return domain.ErrBadRequest
		}
		if len(c.Items) < 1 || len(c.Items) > 20 {
			return domain.ErrBadRequest
		}
		for _, it := range c.Items {
			if _, ok := socialPlatforms[it.Platform]; !ok {
				return domain.ErrBadRequest
			}
			if err := validateHTTPURL(it.URL); err != nil {
				return err
			}
		}
		return nil

	case domain.BlockTypeHeader:
		var c struct {
			Text string `json:"text"`
		}
		if err := strictUnmarshal(raw, &c); err != nil {
			return domain.ErrBadRequest
		}
		if c.Text == "" || len(c.Text) > 80 {
			return domain.ErrBadRequest
		}
		return nil

	default:
		return domain.ErrBadRequest // type lạ
	}
}

var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// validateAppearance kiểm tra theme theo schema cố định. Rỗng/`{}` là hợp lệ (default).
func validateAppearance(raw json.RawMessage) error {
	if len(raw) == 0 || string(bytes.TrimSpace(raw)) == "{}" {
		return nil
	}
	var a struct {
		Theme      string `json:"theme"`
		Background *struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"background"`
		Button *struct {
			Style string `json:"style"`
			Color string `json:"color"`
		} `json:"button"`
		Font string `json:"font"`
	}
	if err := strictUnmarshal(raw, &a); err != nil {
		return domain.ErrBadRequest
	}
	if a.Theme != "" && a.Theme != "dark" && a.Theme != "light" {
		return domain.ErrBadRequest
	}
	if a.Background != nil {
		switch a.Background.Type {
		case "color":
			if !hexColorRe.MatchString(a.Background.Value) {
				return domain.ErrBadRequest
			}
		case "", "gradient", "image":
			// chấp nhận, value tự do
		default:
			return domain.ErrBadRequest
		}
	}
	if a.Button != nil {
		if a.Button.Style != "" && a.Button.Style != "rounded" && a.Button.Style != "sharp" && a.Button.Style != "pill" {
			return domain.ErrBadRequest
		}
		if a.Button.Color != "" && !hexColorRe.MatchString(a.Button.Color) {
			return domain.ErrBadRequest
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Profile
// ---------------------------------------------------------------------------

// GetMyProfile trả profile đầy đủ (kể cả nháp) của chủ sở hữu, kèm toàn bộ block.
func (uc *ProfileUseCase) GetMyProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	p, err := uc.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	blocks, err := uc.blockRepo.ListByProfile(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Blocks = blocks
	return p, nil
}

// GetPublicProfile trả profile đã publish theo username, chỉ kèm block đang active.
func (uc *ProfileUseCase) GetPublicProfile(ctx context.Context, username string) (*domain.Profile, error) {
	p, err := uc.profileRepo.GetPublishedByUsername(ctx, normalizeUsername(username))
	if err != nil {
		return nil, err
	}
	blocks, err := uc.blockRepo.ListActiveByProfile(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.Blocks = blocks
	return p, nil
}

// CreateProfile tạo profile lần đầu (1 user chỉ 1 profile).
func (uc *ProfileUseCase) CreateProfile(ctx context.Context, userID uuid.UUID, p *domain.Profile) error {
	username := normalizeUsername(p.Username)
	if err := validateUsername(username); err != nil {
		return err
	}
	if err := validateAppearance(p.Appearance); err != nil {
		return err
	}

	// Chặn tạo profile thứ hai cho cùng user.
	if _, err := uc.profileRepo.GetByUserID(ctx, userID); err == nil {
		return domain.ErrConflict
	} else if err != domain.ErrNotFound {
		return err
	}

	// Chặn trùng username (chưa có id nên loại trừ uuid.Nil).
	exists, err := uc.profileRepo.UsernameExists(ctx, username, uuid.Nil)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrConflict
	}

	now := time.Now().UTC()
	p.ID = uuid.New()
	p.UserID = userID
	p.Username = username
	p.CreatedAt = now
	p.UpdatedAt = now
	return uc.profileRepo.Create(ctx, p)
}

// UpdateProfile cập nhật profile của chủ sở hữu (username/bio/avatar/appearance/publish).
func (uc *ProfileUseCase) UpdateProfile(ctx context.Context, userID uuid.UUID, in *domain.Profile) error {
	existing, err := uc.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if err := validateAppearance(in.Appearance); err != nil {
		return err
	}

	username := normalizeUsername(in.Username)
	if username != existing.Username {
		if err := validateUsername(username); err != nil {
			return err
		}
		exists, err := uc.profileRepo.UsernameExists(ctx, username, existing.ID)
		if err != nil {
			return err
		}
		if exists {
			return domain.ErrConflict
		}
		existing.Username = username
	}

	existing.DisplayName = in.DisplayName
	existing.Bio = in.Bio
	existing.AvatarURL = in.AvatarURL
	if len(in.Appearance) > 0 {
		existing.Appearance = in.Appearance
	}
	existing.IsPublished = in.IsPublished
	existing.UpdatedAt = time.Now().UTC()

	return uc.profileRepo.Update(ctx, existing)
}

// ---------------------------------------------------------------------------
// Block
// ---------------------------------------------------------------------------

// ListMyBlocks liệt kê tất cả block (cả ẩn) của chủ sở hữu.
func (uc *ProfileUseCase) ListMyBlocks(ctx context.Context, userID uuid.UUID) ([]domain.Block, error) {
	p, err := uc.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return uc.blockRepo.ListByProfile(ctx, p.ID)
}

// CreateBlock thêm block mới vào cuối danh sách của chủ sở hữu.
func (uc *ProfileUseCase) CreateBlock(ctx context.Context, userID uuid.UUID, b *domain.Block) error {
	if !b.Type.IsValid() {
		return domain.ErrBadRequest
	}
	if err := validateBlockContent(b.Type, b.Content); err != nil {
		return err
	}

	p, err := uc.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	existing, err := uc.blockRepo.ListByProfile(ctx, p.ID)
	if err != nil {
		return err
	}
	if len(existing) >= maxBlocksPerProfile {
		return domain.ErrBadRequest
	}

	now := time.Now().UTC()
	b.ID = uuid.New()
	b.ProfileID = p.ID
	b.Position = int32(len(existing)) // thêm vào cuối
	b.IsActive = true
	b.ClickCount = 0
	b.CreatedAt = now
	b.UpdatedAt = now
	return uc.blockRepo.Create(ctx, b)
}

// UpdateBlock sửa type/content/is_active của một block thuộc profile chủ sở hữu.
// isActive là con trỏ: nil → giữ nguyên trạng thái hiện tại; ngược lại đặt theo giá trị.
func (uc *ProfileUseCase) UpdateBlock(ctx context.Context, userID uuid.UUID, b *domain.Block, isActive *bool) error {
	if !b.Type.IsValid() {
		return domain.ErrBadRequest
	}
	if err := validateBlockContent(b.Type, b.Content); err != nil {
		return err
	}

	p, err := uc.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Xác minh block thuộc profile của user: chống IDOR và trả 404 nếu không có.
	existing, err := uc.findBlock(ctx, p.ID, b.ID)
	if err != nil {
		return err
	}

	// is_active chỉ đổi khi client gửi rõ; không gửi thì giữ nguyên.
	if isActive != nil {
		b.IsActive = *isActive
	} else {
		b.IsActive = existing.IsActive
	}

	// Không tin profile_id từ client — gắn theo profile của chính user (chống IDOR).
	b.ProfileID = p.ID
	b.UpdatedAt = time.Now().UTC()
	return uc.blockRepo.Update(ctx, b)
}

// findBlock tìm một block theo id trong phạm vi profile. ErrNotFound nếu không thuộc.
func (uc *ProfileUseCase) findBlock(ctx context.Context, profileID, blockID uuid.UUID) (*domain.Block, error) {
	blocks, err := uc.blockRepo.ListByProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}
	for i := range blocks {
		if blocks[i].ID == blockID {
			return &blocks[i], nil
		}
	}
	return nil, domain.ErrNotFound
}

// DeleteBlock xóa block thuộc profile chủ sở hữu.
func (uc *ProfileUseCase) DeleteBlock(ctx context.Context, userID, blockID uuid.UUID) error {
	p, err := uc.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return uc.blockRepo.Delete(ctx, blockID, p.ID)
}

// ReorderBlocks sắp xếp lại block theo thứ tự id truyền vào. Yêu cầu orderedIDs
// đúng bằng tập block hiện có của profile (không thừa/thiếu).
func (uc *ProfileUseCase) ReorderBlocks(ctx context.Context, userID uuid.UUID, orderedIDs []uuid.UUID) error {
	p, err := uc.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	existing, err := uc.blockRepo.ListByProfile(ctx, p.ID)
	if err != nil {
		return err
	}
	if len(orderedIDs) != len(existing) {
		return domain.ErrBadRequest
	}
	current := make(map[uuid.UUID]struct{}, len(existing))
	for _, b := range existing {
		current[b.ID] = struct{}{}
	}
	seen := make(map[uuid.UUID]struct{}, len(orderedIDs))
	for _, id := range orderedIDs {
		if _, ok := current[id]; !ok {
			return domain.ErrBadRequest // id không thuộc profile
		}
		if _, dup := seen[id]; dup {
			return domain.ErrBadRequest // id trùng
		}
		seen[id] = struct{}{}
	}

	return uc.blockRepo.Reorder(ctx, p.ID, orderedIDs)
}

// RecordClick ghi nhận một click trên block public rồi trả URL để client redirect.
// v1 chỉ hỗ trợ block type "link".
func (uc *ProfileUseCase) RecordClick(ctx context.Context, username string, blockID uuid.UUID) (string, error) {
	p, err := uc.profileRepo.GetPublishedByUsername(ctx, normalizeUsername(username))
	if err != nil {
		return "", err
	}

	blocks, err := uc.blockRepo.ListActiveByProfile(ctx, p.ID)
	if err != nil {
		return "", err
	}

	var target *domain.Block
	for i := range blocks {
		if blocks[i].ID == blockID {
			target = &blocks[i]
			break
		}
	}
	if target == nil {
		return "", domain.ErrNotFound
	}
	if target.Type != domain.BlockTypeLink {
		return "", domain.ErrBadRequest
	}

	var c struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(target.Content, &c); err != nil || c.URL == "" {
		return "", domain.ErrBadRequest
	}

	if err := uc.blockRepo.IncrementClick(ctx, blockID); err != nil {
		return "", err
	}
	return c.URL, nil
}
