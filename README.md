# aliyun-oss-gin-web
## 项目简介
基于 Go 语言 Gin 框架，封装阿里云 OSS **服务端签名直传** 与 **上传回调** 功能的 API 接口。  
核心优势：通过服务端生成签名（避免前端暴露 AccessKey），前端直接向 OSS 上传文件；文件上传完成后，OSS 自动触发服务端回调接口，实现上传结果校验、业务数据同步（如入库文件信息），完全遵循阿里云 OSS 官方 `PostPolicy` 签名规则。


## 环境要求
1. **基础环境**  
   - Go 1.19+（需支持 Gin 框架、阿里云 OSS Go SDK 及环境变量解析）  
   - Go Modules（依赖管理工具，默认自带）

2. **阿里云 OSS 前置准备**  
   - 已创建 OSS Bucket
   - 已配置 Bucket 跨域规则（前端直传需允许来源域名、请求方法：`GET`/`POST`/`OPTIONS`）  
   - 已获取阿里云账号的 **AccessKey ID/Secret**（需授予 OSS Bucket 的「读写权限」及「回调相关权限」，遵循最小权限原则）  


## 安装与配置
### 1. 克隆项目与初始化依赖
```bash
# 1. 克隆仓库到本地
git clone https://github.com/FivmDev/aliyun-oss-go-web.git
cd aliyun-oss-go-web

# 2. 初始化 Go Modules（首次使用需执行）
go mod init aliyun-oss-go-web  # 模块名可自定义，建议与项目名一致

# 3. 安装依赖（自动拉取 Gin、OSS SDK、dotenv 等）
go mod tidy
```

### 2. 配置文件与全局变量映射
项目核心配置可以直接更改文件内部的变量赋值，也可以通过 **环境变量** 与 `global/global.go` 全局变量关联，完全对应阿里云官方文档的配置逻辑：

#### （1）创建环境变量文件（.env）
在项目根目录新建 `.env` 文件，填写阿里云 OSS 及服务配置（参考 `.env.example`）：
```env
# OSS 基础配置（对应 global/global.go 全局变量）
OSS_REGION=oss-cn-beijing          # 你的 OSS Bucket 地域（如：oss-cn-shanghai）
OSS_BUCKET_NAME=your-bucket-name   # 你的 OSS Bucket 名称
OSS_ACCESS_KEY_ID=your-access-key  # 阿里云账号 AccessKey ID
OSS_ACCESS_KEY_SECRET=your-secret  # 阿里云账号 AccessKey Secret

# 上传回调配置（回调接口为当前服务的接口，需公网可访问）
OSS_CALLBACK_URL=http://你的服务公网地址/api/oss/callback  # OSS 上传完成后触发的回调地址
OSS_CALLBACK_BODY=filename=${object}&size=${size}&mimeType=${mimeType}  # 回调携带的文件信息（可自定义）

# 服务运行配置
PORT=8080  # Gin 服务端口
```

#### （2）全局变量映射说明（global/global.go）
项目启动时，会自动将 `.env` 中的配置读取到 `global/global.go` 的全局变量中，与阿里云官方文档的核心参数一一对应：

| 阿里云官方文档参数 | 项目全局变量（global/global.go） | 配置来源（.env）       | 说明                          |
|--------------------|----------------------------------|------------------------|-------------------------------|
| region             | global.Region                    | OSS_REGION             | OSS 地域（如 oss-cn-beijing） |
| bucketName         | global.BucketName                | OSS_BUCKET_NAME        | OSS Bucket 名称               |
| product            | global.Product                   | 固定为 "oss"           | 阿里云产品名，无需修改        |
| accessKeyId        | global.AccessKeyID               | OSS_ACCESS_KEY_ID      | 阿里云 AccessKey ID           |
| accessKeySecret    | global.AccessKeySecret           | OSS_ACCESS_KEY_SECRET  | 阿里云 AccessKey Secret       |
| callbackUrl        | global.CallbackUrl               | OSS_CALLBACK_URL       | 上传回调接口地址              |
| callbackBody       | global.CallbackBody              | OSS_CALLBACK_BODY      | 回调携带的文件元数据          |

> 注：`global/global.go` 的变量需与 `.env` 键名对应，确保配置加载逻辑正确（可在 `global/global.go` 中添加 `init()` 函数实现 `.env` 读取，示例：使用 `github.com/joho/godotenv` 包加载）。


## 启动服务
### 1. 启动前检查
- 确认 `.env` 中所有配置已填写正确 或者 在文件中已经给变量正常赋值（尤其是 OSS 地域、Bucket 名称、回调地址）  
- 确认 OSS Bucket 已开启「回调功能」（无需额外开启，只需回调地址可访问）  
- 确认本地网络可访问阿里云 OSS 服务（可通过 `ping oss-cn-beijing.aliyuncs.com` 测试）

### 2. 启动命令
```bash
# 直接启动（适合开发环境）
go run main.go 

# 或编译后启动（适合生产环境）
go build -o aliyun-oss-server
./aliyun-oss-server  # Windows 系统执行：aliyun-oss-server.exe
```

服务启动后，默认监听 `http://localhost:8080`（端口可通过 `.env` 的 `PORT` 配置修改）。


## 核心 API 接口说明
### 1. 服务端签名接口（前端直传前置）
#### 功能
生成 OSS 直传所需的签名（`policy`、`signature` 等参数），前端获取后直接向 OSS 上传文件，无需通过本服务中转。

#### 请求信息
- 方法：`GET`  
- 路径：`/api/oss/get-sign`  
- 请求参数（Query）：  
  | 参数名    | 类型   | 说明                          | 示例                  |
  |-----------|--------|-------------------------------|-----------------------|
  | fileName  | string | 待上传文件名称（含后缀）      | "test.jpg"            |
  | fileType  | string | 文件 MIME 类型                | "image/jpeg"          |
  | fileSize  | int64  | 文件大小（字节）              | 204800（200KB）       |

#### 响应示例（JSON）
```json
{
  "code": 200,
  "msg": "success",
  "data": {
    "accessKeyId": "LTAIxxxxxxxxx",  // 临时 AK（若用 STS 可优化，当前为原 AK）
    "policy": "eyJleHBpcmF0aW9uIjoiMjAyNC0xMC0xMFQxNTowMDowMFoiLCJjb25kaXRpb25zIjpbeyJidWNrZXQiOiJ5b3VyLWJ1Y2tldC1uYW1lIn0sWyJzdGFydHMtd2l0aCIsIiRrZXkiLCJ1c2VyL2ZpbGVzLyJdLFsiY29udGVudC1sZW5ndGgtcmFuZ2UiLDAsMTA0ODU3NjAwMF1dfQ==",
    "signature": "xxxxxxxxx",  // 签名值
    "host": "https://your-bucket-name.oss-cn-beijing.aliyuncs.com",  // OSS 上传地址
    "callback": "eyJjYWxsYmFja1VybCI6Imh0dHA6Ly95b3VyLXNlcnZlci1hZGRyZXNzL2FwaS9vc3MvY2FsbGJhY2siLCJjYWxsYmFja0JvZHkiOiJmaWxlbmFtZT0ke29iamVjdH0mc2l6ZT0ke3NpemV9Jm1pbWVUeXBlPSR7bWltZVR5cGV9In0=",  // 回调参数（Base64 编码）
    "objectKey": "user/files/test.jpg"  // OSS 中存储的文件路径（可自定义）
  }
}
```

> 说明：`policy` 内容需遵循阿里云官方 `PostPolicy` 规则，可在代码中添加文件大小、类型、存储路径的限制（如仅允许图片上传、最大 5MB）。


### 2. 上传回调接口（OSS 触发）
#### 功能
文件上传到 OSS 后，OSS 会自动向本接口发送请求，携带文件元数据，本服务可在此接口中处理业务逻辑（如：将文件信息存入数据库、更新上传状态）。

#### 请求信息
- 方法：`POST`  
- 路径：`/api/oss/callback`（需与 `.env` 中 `OSS_CALLBACK_URL` 一致）  
- 请求体（由 OSS 自动携带，格式为 `multipart/form-data`）：  
  | 参数名    | 类型   | 说明                          | 示例                  |
  |-----------|--------|-------------------------------|-----------------------|
  | filename  | string | 上传后的文件名称              | "user/files/test.jpg" |
  | size      | string | 文件大小（字符串格式）        | "204800"              |
  | mimeType  | string | 文件 MIME 类型                | "image/jpeg"          |
  | signature | string | OSS 回调签名（用于验证合法性）| "xxxxxxxxx"           |

#### 响应要求（需返回 JSON，OSS 会校验）
```json
{
  "code": 200,
  "msg": "callback success",
  "data": {
    "fileId": "123456"  // 业务自定义字段（如数据库中的文件ID）
  }
}
```

> 关键：需在接口中验证 OSS 回调的签名（参考阿里云官方文档的「回调签名验证」逻辑），防止伪造回调请求。


## 关键配置参考（对应官方文档）
1. **PostPolicy 规则配置**  
   代码中生成 `policy` 时，需按官方文档定义生效时间、存储路径、文件限制等，示例逻辑：
   ```go
   // 参考官方 PostPolicy 构造
   policyStr := fmt.Sprintf(`{
     "expiration": "%s",  // 签名有效期（如 1小时后）
     "conditions": [
       {"bucket": "%s"},  // 限制上传到指定 Bucket
       ["content-length-range", 0, 5242880],  // 限制文件大小（0~5MB）
       ["starts-with", "$key", "user/files/"]  // 限制存储路径前缀
     ]
   }`, expirationTime, global.BucketName)
   ```

2. **回调参数编码**  
   回调的 `callbackUrl` 和 `callbackBody` 需按官方要求进行 Base64 编码，示例：
   ```go
   // 回调参数结构体
   callbackParam := struct {
     CallbackUrl string `json:"callbackUrl"`
     CallbackBody string `json:"callbackBody"`
   }{
     CallbackUrl: global.CallbackUrl,
     CallbackBody: global.CallbackBody,
   }
   // Base64 编码
   callbackBytes, _ := json.Marshal(callbackParam)
   callbackBase64 := base64.StdEncoding.EncodeToString(callbackBytes)
   ```


## 注意事项
1. **AccessKey 安全**  
   - 生产环境建议使用 **STS 临时 AK**（而非永久 AK），通过阿里云 STS 服务生成短期有效、权限受限的 AK，降低泄露风险（参考官方「STS 授权」文档）。  
   - 禁止将 `.env` 文件提交到 Git，需在 `.gitignore` 中添加 `.env`。

2. **跨域配置**  
   在阿里云 OSS 控制台「Bucket 权限管理 → 跨域设置」中添加规则，示例：
   | 允许来源       | 允许 Methods       | 允许 Headers       | 暴露 Headers                |
   |----------------|--------------------|--------------------|-----------------------------|
   | http://localhost:3000（前端地址） | GET,POST,OPTIONS | *                  | x-oss-request-id,ETag       |

3. **回调接口可访问性**  
   OSS 触发回调时，要求 `callbackUrl` 为公网可访问（本地开发可使用内网穿透工具，如 ngrok，生成临时公网地址）。

4. **错误排查**  
   - 若前端上传报错「403 Forbidden」，优先检查：AK 权限、policy 规则、跨域配置。  
   - 若回调无响应，检查：回调地址是否可达、回调签名验证是否通过、响应格式是否符合 JSON 要求。


## 参考文档
- [阿里云 OSS 服务端签名直传并设置回调](https://help.aliyun.com/zh/oss/python-1?spm=a2c4g.11186623.help-menu-search-31815.d_1#3673e04109v8d)（核心规则参考，语言适配为 Go）  
- [阿里云 OSS Go SDK 官方文档](https://github.com/aliyun/aliyun-oss-go-sdk)（SDK 调用细节）  
- [Gin 框架官方文档](https://gin-gonic.com/zh-cn/docs/)（API 接口开发参考）
