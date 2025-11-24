package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"

	"gorm.io/driver/mysql"

	"github.com/gin-contrib/cors"

	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"

	"github.com/joho/godotenv"
)

type Album struct {
	ID     uint    `gorm:"primaryKey" json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// 부분 수정용 dto
type UpdateAlbumRequest struct {
	Title  *string  `json:"title"`
	Artist *string  `json:"artist"`
	Price  *float64 `json:"price"`
}

var (
	rdb  *redis.Client
	rctx = context.Background()
)

func initRedis() {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := getEnv("REDIS_PASSWORD", "")

	rdb = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		PoolSize:     200, // 동시 연결 수
		MinIdleConns: 20,
		MaxRetries:   5,
	})

	if err := pingRedis(); err != nil {
		log.Printf("Warning: Redis connection failed: %v. Running without cache.", err)
		if rdb != nil {
			_ = rdb.Close()
		}
		rdb = nil // Redis 사용 불가 표시
		return    // 실패 시 함수 끝
	}
	log.Println("Redis connected successfully")
}

func pingRedis() error {
	ctx, cancel := context.WithTimeout(rctx, 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}
	return nil
}

// 캐시 무효화 함수
func invalidateAlbumsCache() {
	if rdb != nil {
		if err := rdb.Del(rctx, albumsCacheKey).Err(); err != nil {
			log.Printf("Failed to invalidate cache: %v", err)
		}
	}
}

// 캐시 키 상수
const albumsCacheKey = "albums:all"
const albumsCacheTTL = 10 * time.Second

var db *gorm.DB

func initDB() {
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3310")
	dbName := getEnv("DB_NAME", "duckblog")
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASSWORD", "")

	// DSN 구성
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	// DB 연결 재시도 (최대 10초)
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database, retrying... (%d/10)", i+1)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Fatalf("failed to connect database after retries: %v", err)
	}
	log.Println("Database connected successfully")

	// 커넥션 풀 설정
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get database instance: %v", err)
	}

	// 최대 연결 수
	sqlDB.SetMaxOpenConns(300) // 동시에 열 수 있는 최대 연결 = 100개

	// 유휴 연결 수
	sqlDB.SetMaxIdleConns(10) // 대기 상태 연결 = 10개 (재사용)

	// 연결 재사용 시간
	sqlDB.SetConnMaxLifetime(time.Hour) // 1시간 후 연결 재생성

	// 테이블 자동 생성 (마이그레이션)
	err = db.AutoMigrate(&Album{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
}

// GET /albums
// 모든 앨범 조회 (Read - List)
func getAlbums(c *gin.Context) {

	// 1) Redis 캐시 먼저 조회
	// 처음엔 DB로 접근하지만 이후 동일한 /albums 요청은 Redis에서 바로 꺼내서 리턴하여 훨씬 가벼워진다.
	if rdb != nil {
		if data, err := rdb.Get(rctx, albumsCacheKey).Bytes(); err == nil && len(data) > 0 {
			var albums []Album
			if err := json.Unmarshal(data, &albums); err == nil {
				c.JSON(http.StatusOK, albums)
				return
			}
		}
	}

	// 2) 캐시 미스 -> DB 조회
	var albums []Album
	if err := db.Find(&albums).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch albums"})
		return
	}

	// 3) DB 결과물 Redis에 캐싱
	if rdb != nil {
		if b, err := json.Marshal(albums); err == nil {
			_ = rdb.Set(rctx, albumsCacheKey, b, albumsCacheTTL).Err()
		}
	}

	c.JSON(http.StatusOK, albums)
}

// GET /albums/:id
// 특정 앨범 조회 (Read - Detail)
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")
	var album Album

	result := db.First(&album, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// 404
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, album)
}

// POST /albums
// 새 앨범 생성 (Create) + Redis 캐시 무효화
func postAlbums(c *gin.Context) {
	var newAlbum Album

	// 요청 바디(JSON)를 newAlbum에 바인딩
	if err := c.BindJSON(&newAlbum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// DB에 저장
	result := db.Create(&newAlbum)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	invalidateAlbumsCache() // 캐시 무효화

	c.IndentedJSON(http.StatusCreated, newAlbum)
}

// PUT /albums/:id
// 앨범 전체 수정 (Update)
func updateAlbum(c *gin.Context) {
	id := c.Param("id")
	var album Album

	// 해당 앨범이 존재하는지 확인
	if err := db.First(&album, id).Error; err != nil {
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
	result := db.Model(&album).Updates(req)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	invalidateAlbumsCache() // 캐시 무효화

	c.IndentedJSON(http.StatusOK, req)
}

// PATCH /albums/:id
// 앨범 일부 수정
func patchAlbum(c *gin.Context) {
	id := c.Param("id")
	var album Album

	// 먼저 해당 앨범이 존재하는지 확인
	if err := db.First(&album, id).Error; err != nil {
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
		result := db.Model(&album).Updates(updates)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
	}

	// 업데이트된 데이터 다시 조회
	db.First(&album, id)

	invalidateAlbumsCache() // 캐시 무효화

	c.IndentedJSON(http.StatusOK, album)
}

// DELETE /albums/:id
// 앨범 삭제 (Delete)
func deleteAlbum(c *gin.Context) {
	id := c.Param("id")
	var album Album

	// 먼저 해당 앨범이 존재하는지 확인
	if err := db.First(&album, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// DB에서 삭제
	result := db.Delete(&album)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	invalidateAlbumsCache() // 캐시 무효화

	c.Status(http.StatusNoContent) // 204
}

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	initDB()
	initRedis() // Redis 초기화

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// CORS 설정 추가
	r.Use(cors.Default())

	// -----라우터들-----
	r.GET("/albums", getAlbums)
	r.GET("/albums/:id", getAlbumByID)
	r.POST("/albums", postAlbums)
	r.PUT("/albums/:id", updateAlbum)
	r.PATCH("/albums/:id", patchAlbum)
	r.DELETE("/albums/:id", deleteAlbum)
	// -----------------

	r.Run()
}

// 환경변수 읽기 헬퍼 함수
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
