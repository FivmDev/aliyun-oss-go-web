package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	// 定义默认的IP和端口字符串
	strIPPort := ":8080"
	if len(os.Args) == 3 {
		strIPPort = fmt.Sprintf("%s:%s", os.Args[1], os.Args[2])
	} else if len(os.Args) != 1 {
		fmt.Println("Usage   : go run test1.go                ")
		fmt.Println("Usage   : go run test1.go ip port        ")
		fmt.Println("")
		os.Exit(0)
	}
	// 打印服务器运行的地址和端口
	fmt.Printf("server is running on %s \n", strIPPort)
	// 注册处理根路径请求的函数
	http.HandleFunc("/", handlerRequest)
	// 注册处理获取签名请求的函数
	http.HandleFunc("/get_post_signature_for_oss_upload", handleGetPostSignature)
	// 启动HTTP服务器
	err := http.ListenAndServe(strIPPort, nil)
	if err != nil {
		strError := fmt.Sprintf("http.ListenAndServe failed : %s \n", err.Error())
		panic(strError)
	}
}
