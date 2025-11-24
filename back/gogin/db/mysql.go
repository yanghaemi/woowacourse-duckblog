package db

import (
	"fmt"
	"gogin/config"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB() *gorm.DB {
	dbHost := config.GetEnv("DB_HOST", "127.0.0.1")
	dbPort := config.GetEnv("DB_PORT", "3310")
	dbName := config.GetEnv("DB_NAME", "duckblog")
	dbUser := config.GetEnv("DB_USER", "root")
	dbPass := config.GetEnv("DB_PASSWORD", "")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	var (
		db  *gorm.DB
		err error
	)

	// 재시도 로직
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database, retrying... (%d/10): %v", i+1, err)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Fatalf("failed to connect database after retries: %v", err)
	}

	log.Println("Database connected successfully")

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

	return db
}
