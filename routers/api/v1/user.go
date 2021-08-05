package v1

import (
	"PGCloudDisk/db"
	"PGCloudDisk/errno"
	"PGCloudDisk/utils"
	"PGCloudDisk/utils/lg"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Auth(c *gin.Context) {
	// 获取username & password
	username, ok1 := c.GetPostForm("username")
	pwd, ok2 := c.GetPostForm("password")

	if !ok1 || !ok2 {
		utils.Response(c, http.StatusBadRequest, errno.RespCode{Code: errno.RespFailed}, nil)
		return
	}

	// 验证, 如果成功返回Token
	status := db.UserCheck(username, pwd)

	if status.Success() {
		token, s := utils.GetToken(username)
		if s.Success() {
			utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespSuccess}, gin.H{
				"token": token,
			})
			lg.Logger.Printf("user %s auth successfully\n", username)
			return
		}
	}

	// 失败返回错误
	lg.Logger.Printf("user %s auth failed\n", username)
	switch status.Code {
	case errno.UserNotFound:
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespAuthUserNotFound}, nil)
	case errno.UserNamePwdNotMatched:
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespAuthUserNamePwdNotMatched}, nil)
	default:
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespFailed}, nil)
	}
}
