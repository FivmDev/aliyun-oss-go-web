package router

import (
	"aliyun-oss-go-web/handler"
	"github.com/gin-gonic/gin"
)

func InitV1(g *gin.RouterGroup) {
	g.GET("/", handler.RootRequest)
	g.GET("/get_post_signature_for_oss_upload", handler.GetPostSignature)
}
