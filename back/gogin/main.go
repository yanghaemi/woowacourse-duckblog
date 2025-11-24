package main

import (
	"gogin/cache"
	"gogin/config"
	"gogin/db"
	"gogin/handler"
	"gogin/router"
	"log"
)

func main() {
	config.LoadEnv()

	database := db.NewDB()
	redisClient := cache.NewRedis() // Redis 초기화

	// 핸들러 생성
	albumHandler := handler.NewAlbumHandler(database, redisClient)

	r := router.NewRouter(albumHandler)

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
