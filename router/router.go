package router

import (
	"aliyun-oss-go-web/handler"
	"github.com/gin-gonic/gin"
)

func InitV1(g *gin.RouterGroup) {
	g.GET("/", handler.RootRequest)
}
