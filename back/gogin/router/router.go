package router

import (
	"gogin/handler"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(AlbumHandler *handler.AlbumHandler) *gin.Engine {
	r := gin.Default()

	r.Use(cors.Default())

	// 핑
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 앨범 라우트
	r.GET("/albums", AlbumHandler.GetAlbums)
	r.GET("/albums/:id", AlbumHandler.GetAlbumByID)
	r.POST("/albums", AlbumHandler.PostAlbum)
	r.PUT("/albums/:id", AlbumHandler.UpdateAlbum)
	r.PATCH("/albums/:id", AlbumHandler.PatchAlbum)
	r.DELETE("/albums/:id", AlbumHandler.DeleteAlbum)

	return r
}
