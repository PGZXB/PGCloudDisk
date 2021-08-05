package middleware

import (
	"PGCloudDisk/errno"
	"PGCloudDisk/utils"
	"PGCloudDisk/utils/lg"
	"github.com/gin-gonic/gin"
	"net/http"
)

type tok struct {
	Token string `json:"token"`
}

func Jwt() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := errno.RespCode{Code: errno.RespSuccess}

		// 获取token
		tk := tok{}
		err := c.ShouldBindJSON(&tk)
		if err != nil {
			code.Code = errno.RespInvalidParams
			utils.Response(c, http.StatusBadRequest, code, nil)
			c.Abort()
			return
		}

		// 解析token
		claims, status := utils.ParseToken(tk.Token)
		if status.Success() {
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
