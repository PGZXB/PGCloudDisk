package routers

import (
	"PGCloudDisk/middleware"
	"PGCloudDisk/routers/api/v1"
	"github.com/gin-gonic/gin"
)

func Init() *gin.Engine {
	g := gin.New()
	g.Use(gin.Logger())
	g.Use(gin.Recovery())

	v1Group := g.Group("api/v1")
	{
		// 无需验证
		v1Group.POST("/auth", v1.Auth)

		// 需要验证
		v1Group.Use(middleware.Jwt())

		v1Group.GET("/user-infos", v1.GetUserInfo)

		// files
		v1Group.POST("/files", v1.UploadFile)       // 上传文件
		v1Group.GET("/files/:id", v1.DownloadFile)  // 下载文件
		v1Group.DELETE("/files/:id", v1.DeleteFile) // 删除文件

		// file-infos
		v1Group.GET("file-infos/:id", v1.GetFileInfo)        // 查看文件信息
		v1Group.GET("file-infos", v1.GetFileInfosWithFilter) // 获取文件信息
	}
	return g
}
