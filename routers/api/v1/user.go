package v1

import (
	"PGCloudDisk/db"
	"PGCloudDisk/errno"
	"PGCloudDisk/utils"
	"PGCloudDisk/utils/lg"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"net/http"
)

type auth struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

type userInfoCanBePublished struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// Auth the user(POST api/v1/auth)
func Auth(c *gin.Context) {
	// 获取username & password
	auth := auth{}
	err := c.ShouldBind(&auth)
	if err != nil || auth.Username == "" || auth.Password == "" {
		utils.Response(c, http.StatusBadRequest, errno.RespCode{Code: errno.RespInvalidParams}, nil)
		return
	}

	username := auth.Username
	pwd := auth.Password

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

// GetUserInfo get user info(GET api/v1/user-infos)
func GetUserInfo(c *gin.Context) {
	uname, ok := c.Get("username")
	if !ok {
		utils.Response(c, http.StatusBadRequest, errno.RespCode{Code: errno.RespInvalidParams}, nil)
		return
	}

	username, ok := uname.(string)
	if !ok {
		utils.Response(c, http.StatusBadRequest, errno.RespCode{Code: errno.RespInvalidParams}, nil)
		return
	}

	user, status := db.GetUserInfo(username)
	if status.Code == errno.UserNotFound {
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespAuthUserNotFound}, nil)
		return
	}

	if status.Code == errno.UserInfoGetFailed {
		utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.RespGetUserInfoFailed}, nil)
		return
	}

	assert.IsEqual(status.Code, errno.Success)
	utils.Response(c, http.StatusOK, errno.RespCode{Code: errno.Success}, gin.H{
		"userinfo": userInfoCanBePublished{
			ID:       user.ID,
			Username: user.Username,
		},
	})
}
