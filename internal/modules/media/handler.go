package media

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"gopress/internal/http/response"
	"gopress/internal/model"
	"gopress/internal/modules/auth"
)

var extensionPattern = regexp.MustCompile(`^\.[a-z0-9]{1,10}$`)

type Handler struct {
	db       *gorm.DB
	dir      string
	maxBytes int64
}

func NewHandler(db *gorm.DB, dir string, maxBytes int64) *Handler {
	return &Handler{db: db, dir: dir, maxBytes: maxBytes}
}

type listQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1,max=200"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Keyword  string `form:"keyword" binding:"omitempty,max=200"`
	Kind     string `form:"kind" binding:"omitempty,oneof=image video audio other"`
}

func (h *Handler) Upload(c *gin.Context) {
	header, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "file is required")
		return
	}
	if header.Size <= 0 {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "file is empty")
		return
	}
	if header.Size > h.maxBytes {
		response.Error(c, http.StatusRequestEntityTooLarge, response.CodeBadRequest,
			fmt.Sprintf("file exceeds the %d MB upload limit", h.maxBytes/(1024*1024)))
		return
	}
	source, err := header.Open()
	if err != nil {
		h.handleError(c, err)
		return
	}
	defer source.Close()

	sniff := make([]byte, 512)
	n, readErr := io.ReadFull(source, sniff)
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		h.handleError(c, readErr)
		return
	}
	mimeType := http.DetectContentType(sniff[:n])
	if mimeType == "application/octet-stream" {
		if guessed := mime.TypeByExtension(strings.ToLower(filepath.Ext(header.Filename))); guessed != "" {
			mimeType = guessed
		}
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		h.handleError(c, err)
		return
	}

	if err := os.MkdirAll(h.dir, 0o755); err != nil {
		h.handleError(c, err)
		return
	}
	storedName := randomName() + safeExtension(header.Filename)
	path := filepath.Join(h.dir, storedName)
	target, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		h.handleError(c, err)
		return
	}
	written, copyErr := io.Copy(target, io.LimitReader(source, h.maxBytes+1))
	closeErr := target.Close()
	if copyErr != nil || closeErr != nil || written > h.maxBytes {
		_ = os.Remove(path)
		if copyErr != nil {
			h.handleError(c, copyErr)
		} else if closeErr != nil {
			h.handleError(c, closeErr)
		} else {
			response.Error(c, http.StatusRequestEntityTooLarge, response.CodeBadRequest,
				fmt.Sprintf("file exceeds the %d MB upload limit", h.maxBytes/(1024*1024)))
		}
		return
	}

	media := model.Media{
		OriginalName: truncateName(header.Filename),
		StoredName:   storedName,
		MimeType:     mimeType,
		Size:         written,
		UploaderID:   auth.CurrentUserID(c),
	}
	if err := h.db.WithContext(c).Create(&media).Error; err != nil {
		_ = os.Remove(path)
		h.handleError(c, err)
		return
	}
	response.Created(c, model.ToMediaDTO(media))
}

func (h *Handler) List(c *gin.Context) {
	var query listQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid media query")
		return
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 24
	}
	db := h.db.WithContext(c).Model(&model.Media{})
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("original_name LIKE ? OR mime_type LIKE ?", like, like)
	}
	switch query.Kind {
	case "image":
		db = db.Where("mime_type LIKE ?", "image/%")
	case "video":
		db = db.Where("mime_type LIKE ?", "video/%")
	case "audio":
		db = db.Where("mime_type LIKE ?", "audio/%")
	case "other":
		db = db.Where("mime_type NOT LIKE ? AND mime_type NOT LIKE ? AND mime_type NOT LIKE ?",
			"image/%", "video/%", "audio/%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		h.handleError(c, err)
		return
	}
	items := make([]model.Media, 0, query.PageSize)
	if err := db.Order("created_at DESC").Order("id DESC").
		Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&items).Error; err != nil {
		h.handleError(c, err)
		return
	}
	result := make([]model.MediaDTO, 0, len(items))
	for _, item := range items {
		result = append(result, model.ToMediaDTO(item))
	}
	response.OK(c, response.NewPage(result, total, query.Page, query.PageSize))
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var item model.Media
	if err := h.db.WithContext(c).First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, response.CodeNotFound, "media not found")
			return
		}
		h.handleError(c, err)
		return
	}
	if err := h.db.WithContext(c).Delete(&model.Media{}, id).Error; err != nil {
		h.handleError(c, err)
		return
	}
	if err := os.Remove(filepath.Join(h.dir, item.StoredName)); err != nil && !os.IsNotExist(err) {
		log.Printf("remove media file %s: %v", item.StoredName, err)
	}
	response.OK(c, gin.H{"deleted": true})
}

// Serve streams an uploaded file. Stored names are flat and server-generated,
// and the response headers make active formats such as HTML or SVG inert.
func (h *Handler) Serve(c *gin.Context) {
	name := c.Param("name")
	if name == "" || filepath.Base(name) != name || strings.HasPrefix(name, ".") {
		c.Status(http.StatusNotFound)
		return
	}
	path := filepath.Join(h.dir, name)
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		c.Status(http.StatusNotFound)
		return
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; img-src data:")
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.File(path)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	log.Printf("media operation: %v", err)
	response.Error(c, http.StatusInternalServerError, response.CodeInternal, "media operation failed")
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, response.CodeBadRequest, "invalid media id")
		return 0, false
	}
	return uint(id), true
}

func safeExtension(name string) string {
	ext := strings.ToLower(filepath.Ext(filepath.Base(name)))
	if !extensionPattern.MatchString(ext) {
		return ""
	}
	return ext
}

func randomName() string {
	buffer := make([]byte, 12)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
	}
	return hex.EncodeToString(buffer)
}

func truncateName(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "file"
	}
	if utf8.RuneCountInString(name) <= 200 {
		return name
	}
	runes := []rune(name)
	return string(runes[:200])
}
