package initialize

import (
	"aliyun-oss-go-web/router"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	g := gin.Default()
	// 确保 templates 目录存在，且 index.html 在该目录下
	g.LoadHTMLGlob("templates/*")
	g.Static("/static", "../static")
	v1 := g.Group("v1")
	router.InitV1(v1)
	return g
}
