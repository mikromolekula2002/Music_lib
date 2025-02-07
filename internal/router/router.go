package router

import (
	"mikromolekula2002/music_library_ver1.0/internal/controller"
	"mikromolekula2002/music_library_ver1.0/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Router struct {
	Gin            *gin.Engine
	MusicCotroller *controller.MusicLibController
}

func NewRouter(service *service.MusicLibService) *Router {
	r := gin.Default()

	return &Router{
		Gin:            r,
		MusicCotroller: controller.NewMusicLibController(service),
	}
}

func (r *Router) SetRoutes(envType string) {
	r.Gin.POST("/create-song", r.MusicCotroller.SaveSong)
	r.Gin.GET("/song", r.MusicCotroller.GetSongTextByGroup)
	r.Gin.PUT("/song", r.MusicCotroller.UpdateSong)
	r.Gin.DELETE("/song", r.MusicCotroller.DeleteSong)
	r.Gin.GET("/songs", r.MusicCotroller.GetAllSongs)

	if envType == "debug" {
		r.Gin.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		gin.SetMode(gin.DebugMode)
	}
}
