package middleware

import (
	"PGCloudDisk/errno"
	"PGCloudDisk/utils"
	"PGCloudDisk/utils/lg"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Jwt() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := errno.RespCode{Code: errno.RespSuccess}

		// 获取token
		token := c.Request.Header.Get("token")
		if token == "" {
			token = c.Query("t")
			if token == "" {
				code.Code = errno.RespInvalidParams
				utils.Response(c, http.StatusBadRequest, code, nil)
				c.Abort()
				return
			}
		}

		// 解析token
		claims, status := utils.ParseToken(token)
		if status.Success() {
			c.Set("user_id", claims.ID)
			c.Set("username", claims.Username)
		} else {
			lg.Logger.Println(status.Msg())
			code.Code = errno.RespTokenCheckFailed
		}

		if code.Success() {
			c.Next()
		} else {
			utils.Response(c, http.StatusUnauthorized, code, nil)
			c.Abort()
		}
	}
}
