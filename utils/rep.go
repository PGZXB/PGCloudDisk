package utils

import (
	"PGCloudDisk/errno"
	"github.com/gin-gonic/gin"
)

type RespMsg struct {
	Code uint32      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Response(c *gin.Context, httpCode int, code errno.RespCode, data interface{}) {
	c.JSON(httpCode, RespMsg{
		code.Code,
		code.Msg(),
		data,
	})
}
