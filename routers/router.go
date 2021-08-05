package routers

import (
	"PGCloudDisk/errno"
	"PGCloudDisk/middleware"
	"PGCloudDisk/routers/api/v1"
	"PGCloudDisk/utils"
	"github.com/gin-gonic/gin"
	"net/http"
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

		f := func(c *gin.Context) {
			name, _ := c.Get("username")

			utils.Response(c, http.StatusOK, errno.RespCode{}, gin.H{
				"username": name,
				"Test2":    100,
				"Test3":    []int{1, 2, 3},
			})
		}

		v1Group.GET("/test", f)
		v1Group.POST("/test", f)
		v1Group.PUT("/test", f)
		v1Group.DELETE("/test", f)
	}
	return g
}
