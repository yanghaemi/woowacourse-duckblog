package handler

import (
	"encoding/json"
	"gogin/cache"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Album struct {
	ID     uint    `gorm:"primaryKey" json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

type UpdateAlbumRequest struct {
	Title  *string  `json:"title"`
	Artist *string  `json:"artist"`
	Price  *float64 `json:"price"`
}

// 캐시 키/TTL
const (
	AlbumsCacheKey = "albums:all"
	AlbumsCacheTTL = 10 * time.Second
)

type AlbumHandler struct {
	DB  *gorm.DB
	RDB *redis.Client
}

func NewAlbumHandler(db *gorm.DB, rdb *redis.Client) *AlbumHandler {
	return &AlbumHandler{DB: db, RDB: rdb}
}

// 캐시 무효화
func (h *AlbumHandler) invalidateAlbumsCache() {
	if h.RDB != nil {
		if err := h.RDB.Del(cache.Ctx, AlbumsCacheKey).Err(); err != nil {
			log.Printf("Failed to invalidate cache: %v", err)
		}
	}
}

// GET /albums
func (h *AlbumHandler) GetAlbums(c *gin.Context) {
	// 1) Redis 조회
	if h.RDB != nil {
		if data, err := h.RDB.Get(cache.Ctx, AlbumsCacheKey).Bytes(); err == nil && len(data) > 0 {
			var albums []Album
			if err := json.Unmarshal(data, &albums); err == nil {
				c.JSON(http.StatusOK, albums)
				return
			}
		}
	}

	// 2) DB 조회
	var albums []Album
	if err := h.DB.Find(&albums).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch albums"})
		return
	}

	// 3) Redis 캐싱
	if h.RDB != nil {
		if b, err := json.Marshal(albums); err == nil {
			_ = h.RDB.Set(cache.Ctx, AlbumsCacheKey, b, AlbumsCacheTTL).Err()
		}
	}

	c.JSON(http.StatusOK, albums)
}

// GET /albums/:id
func (h *AlbumHandler) GetAlbumByID(c *gin.Context) {
	id := c.Param("id")
	var album Album

	if err := h.DB.First(&album, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": "album not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, album)
}

// POST /albums
func (h *AlbumHandler) PostAlbum(c *gin.Context) {
	var newAlbum Album
	if err := c.BindJSON(&newAlbum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.DB.Create(&newAlbum).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.invalidateAlbumsCache()

	c.JSON(http.StatusCreated, newAlbum)
}

// PUT /albums/:id
// 앨범 전체 수정 (Update)
func (h *AlbumHandler) UpdateAlbum(c *gin.Context) {
	id := c.Param("id")
	var album Album

	// 해당 앨범이 존재하는지 확인
	if err := h.DB.First(&album, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req Album
	// 수정할 데이터 바인딩
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// ID는 유지하고 나머지 필드만 업데이트
	req.ID = album.ID
	result := h.DB.Model(&album).Updates(req)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	h.invalidateAlbumsCache()

	c.IndentedJSON(http.StatusOK, req)
}

// PATCH /albums/:id
// 앨범 일부 수정
func (h *AlbumHandler) PatchAlbum(c *gin.Context) {
	id := c.Param("id")
	var album Album

	// 먼저 해당 앨범이 존재하는지 확인
	if err := h.DB.First(&album, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req UpdateAlbumRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 업데이트할 필드만 포함하는 맵 생성
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Artist != nil {
		updates["artist"] = *req.Artist
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}

	// DB 업데이트
	if len(updates) > 0 {
		result := h.DB.Model(&album).Updates(updates)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
	}

	// 업데이트된 데이터 다시 조회
	h.DB.First(&album, id)

	h.invalidateAlbumsCache() // 캐시 무효화

	c.IndentedJSON(http.StatusOK, album)
}

// DELETE /albums/:id
// 앨범 삭제 (Delete)
func (h *AlbumHandler) DeleteAlbum(c *gin.Context) {
	id := c.Param("id")
	var album Album

	// 먼저 해당 앨범이 존재하는지 확인
	if err := h.DB.First(&album, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// DB에서 삭제
	result := h.DB.Delete(&album)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	h.invalidateAlbumsCache() // 캐시 무효화

	c.Status(http.StatusNoContent) // 204
}
