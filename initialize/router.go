package initialize

import (
	"aliyun-oss-go-web/router"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	g := gin.Default()
	v1 := g.Group("/v1")
	router.InitV1Router(v1)
	return g
}
