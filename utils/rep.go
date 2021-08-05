package utils

import (
	"PGCloudDisk/errno"
	"github.com/gin-gonic/gin"
)

type RespMsg struct {
	Code uint32                 `json:"code"`
	Msg  string                 `json:"msg"`
	Data map[string]interface{} `json:"data"`
}

func Response(c *gin.Context, httpCode int, code errno.RespCode, data map[string]interface{}) {
	c.JSON(httpCode, RespMsg{
		code.Code,
		code.Msg(),
		data,
	})
}
