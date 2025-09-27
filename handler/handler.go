package handler

import (
	"aliyun-oss-go-web/forms"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aliyun/credentials-go/credentials"

	"aliyun-oss-go-web/global"
	"github.com/gin-gonic/gin"
)

// RootRequest 函数处理根路径请求
func RootRequest(c *gin.Context) {
	data := map[string]interface{}{
		"msg": "请求成功",
	}
	// 渲染模板（模板名需与加载时一致，如 "index.html"）
	c.HTML(http.StatusOK, "index.html", data)
	// 检查渲染过程中是否有错误（如模板被意外删除）
	if len(c.Errors) > 0 {
		// 获取错误信息
		err := c.Errors.Last()
		// 返回友好的错误提示
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "模板文件不存在或渲染失败",
			"detail": err.Error(),
		})
		// 终止后续处理
		c.Abort()
	}
}

// GetPostSignature 函数处理获取签名请求
func GetPostSignature(c *gin.Context) {
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("Access-Control-Allow-Origin", "*") // 允许跨域
	c.JSON(http.StatusOK, getPolicyToken())
}

// getPolicyToken 函数生成 OSS 上传所需的签名和凭证
func getPolicyToken() string {
	// 设置bucket所处地域
	global.Region = "cn-hangzhou"
	// 设置bucket名称
	global.BucketName = "examplebucket"
	// 设置 OSS 上传地址
	host := fmt.Sprintf("https://%s.oss-%s.aliyuncs.com", global.BucketName, global.Region)
	// 设置上传目录
	dir := "user-dir"
	// callbackUrl为 上传回调服务器的URL，请将下面的IP和Port配置为您自己的真实信息。
	callbackUrl := "http://oss-demo.aliyuncs.com:23450/callback"

	config := new(credentials.Config).
		SetType("ram_role_arn").
		SetAccessKeyId(os.Getenv("OSS_ACCESS_KEY_ID")).
		SetAccessKeySecret(os.Getenv("OSS_ACCESS_KEY_SECRET")).
		SetRoleArn(os.Getenv("OSS_STS_ROLE_ARN")).
		SetRoleSessionName("Role_Session_Name").
		SetPolicy("").
		SetRoleSessionExpiration(3600)

	// 根据配置创建凭证提供器
	provider, err := credentials.NewCredential(config)
	if err != nil {
		log.Fatalf("NewCredential fail, err:%v", err)
	}

	// 从凭证提供器获取凭证
	cred, err := provider.GetCredential()
	if err != nil {
		log.Fatalf("GetCredential fail, err:%v", err)
	}

	// 构建policy
	utcTime := time.Now().UTC()
	date := utcTime.Format("20060102")
	expiration := utcTime.Add(1 * time.Hour)
	policyMap := map[string]any{
		"expiration": expiration.Format("2006-01-02T15:04:05.000Z"),
		"conditions": []any{
			map[string]string{"bucket": global.BucketName},
			map[string]string{"x-oss-signature-version": "OSS4-HMAC-SHA256"},
			map[string]string{"x-oss-credential": fmt.Sprintf("%v/%v/%v/%v/aliyun_v4_request", *cred.AccessKeyId, date, global.Region, global.Product)},
			map[string]string{"x-oss-date": utcTime.Format("20060102T150405Z")},
			map[string]string{"x-oss-security-token": *cred.SecurityToken},
		},
	}

	// 将policy转换为 JSON 格式
	policy, err := json.Marshal(policyMap)
	if err != nil {
		log.Fatalf("json.Marshal fail, err:%v", err)
	}

	// 构造待签名字符串（StringToSign）
	stringToSign := base64.StdEncoding.EncodeToString([]byte(policy))

	hmacHash := func() hash.Hash { return sha256.New() }
	// 构建signing key
	signingKey := "aliyun_v4" + *cred.AccessKeySecret
	h1 := hmac.New(hmacHash, []byte(signingKey))
	io.WriteString(h1, date)
	h1Key := h1.Sum(nil)

	h2 := hmac.New(hmacHash, h1Key)
	io.WriteString(h2, global.Region)
	h2Key := h2.Sum(nil)

	h3 := hmac.New(hmacHash, h2Key)
	io.WriteString(h3, global.Product)
	h3Key := h3.Sum(nil)

	h4 := hmac.New(hmacHash, h3Key)
	io.WriteString(h4, "aliyun_v4_request")
	h4Key := h4.Sum(nil)

	// 生成签名
	h := hmac.New(hmacHash, h4Key)
	io.WriteString(h, stringToSign)
	signature := hex.EncodeToString(h.Sum(nil))

	var callbackParam forms.CallbackParam
	callbackParam.CallbackUrl = callbackUrl
	callbackParam.CallbackBody = "filename=${object}&size=${size}&mimeType=${mimeType}&height=${imageInfo.height}&width=${imageInfo.width}"
	callbackParam.CallbackBodyType = "application/x-www-form-urlencoded"
	callback_str, err := json.Marshal(callbackParam)
	if err != nil {
		fmt.Println("callback json err:", err)
	}
	callbackBase64 := base64.StdEncoding.EncodeToString(callback_str)
	// 构建返回给前端的表单
	policyToken := forms.PolicyToken{
		Policy:           stringToSign,
		SecurityToken:    *cred.SecurityToken,
		SignatureVersion: "OSS4-HMAC-SHA256",
		Credential:       fmt.Sprintf("%v/%v/%v/%v/aliyun_v4_request", *cred.AccessKeyId, date, global.Region, global.Product),
		Date:             utcTime.UTC().Format("20060102T150405Z"),
		Signature:        signature,
		Host:             host,           // 返回 OSS 上传地址
		Dir:              dir,            // 返回上传目录
		Callback:         callbackBase64, // 返回上传回调参数
	}

	response, err := json.Marshal(policyToken)
	if err != nil {
		fmt.Println("json err:", err)
	}
	return string(response)
}
