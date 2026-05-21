# 百度网盘 Go SDK API 文档

百度网盘 Go 语言 SDK，提供设备码授权、文件管理、下载上传等功能的完整实现。

## 目录

- [快速开始](#快速开始)
- [配置](#配置)
- [客户端](#客户端)
- [认证服务](#认证服务)
- [用户服务](#用户服务)
- [文件服务](#文件服务)
- [下载服务](#下载服务)
- [上传服务](#上传服务)
- [错误处理](#错误处理)
- [数据模型](#数据模型)

---

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/baowuhe/go-bdisk/pkg/bdisk"
)

func main() {
    // 创建配置
    config := bdisk.NewConfig("your_app_key", "your_secret_key")

    // 创建客户端
    client, err := bdisk.NewClient(config)
    if err != nil {
        panic(err)
    }

    // 设备码授权流程
    deviceResp, err := client.Auth.DeviceCodeFlow()
    if err != nil {
        panic(err)
    }

    fmt.Printf("请访问: %s\n", deviceResp.VerificationURL)
    fmt.Printf("验证码：%s\n", deviceResp.UserCode)

    // 轮询获取 token
    token, err := client.Auth.PollToken(deviceResp.DeviceCode, deviceResp.Interval)
    if err != nil {
        panic(err)
    }

    client.SetToken(token)
    fmt.Println("登录成功！")

    // 获取用户信息
    userInfo, err := client.User.GetInfo()
    if err != nil {
        panic(err)
    }
    fmt.Printf("欢迎，%s\n", userInfo.NetdiskName)
}
```

---

## 配置

### Config

SDK 配置结构体，包含应用凭证信息。

```go
type Config struct {
    AppKey    string  // 应用 Key
    SecretKey string  // 应用密钥
}
```

### NewConfig

创建新的 SDK 配置。

**函数签名：**
```go
func NewConfig(appKey, secretKey string) *Config
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| appKey | string | 百度开放平台应用 Key |
| secretKey | string | 百度开放平台应用密钥 |

**返回值：**
- `*Config` - 配置对象

**示例：**
```go
config := bdisk.NewConfig("your_app_key", "your_secret_key")
```

### Validate

验证配置是否有效。

**方法签名：**
```go
func (c *Config) Validate() error
```

**返回值：**
- `error` - 配置无效时返回错误

---

## 客户端

### Client

百度网盘 SDK 客户端，提供所有 API 服务的访问入口。

```go
type Client struct {
    Auth     *AuthService     // 认证服务
    User     *UserService     // 用户服务
    File     *FileService     // 文件服务
    Download *DownloadService // 下载服务
    Upload   *UploadService   // 上传服务
}
```

### NewClient

创建新的 SDK 客户端。

**函数签名：**
```go
func NewClient(config *Config) (*Client, error)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| config | *Config | SDK 配置对象 |

**返回值：**
- `*Client` - 客户端对象
- `error` - 创建失败时返回错误

**示例：**
```go
config := bdisk.NewConfig("app_key", "secret_key")
client, err := bdisk.NewClient(config)
if err != nil {
    panic(err)
}
```

### SetToken

设置访问令牌。

**方法签名：**
```go
func (c *Client) SetToken(token *Token)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| token | *Token | 访问令牌对象 |

### GetToken

获取当前访问令牌。

**方法签名：**
```go
func (c *Client) GetToken() *Token
```

**返回值：**
- `*Token` - 当前访问令牌

### ClearToken

清除访问令牌。

**方法签名：**
```go
func (c *Client) ClearToken()
```

---

## 认证服务

### AuthService

认证服务，处理设备码授权流程和 Token 管理。

### DeviceCodeFlow

设备码授权流程，获取设备验证码。

**方法签名：**
```go
func (a *AuthService) DeviceCodeFlow() (*DeviceCodeResponse, error)
```

**返回值：**
- `*DeviceCodeResponse` - 设备码响应
- `error` - 请求失败时返回错误

**示例：**
```go
deviceResp, err := client.Auth.DeviceCodeFlow()
if err != nil {
    panic(err)
}

fmt.Printf("请访问：%s\n", deviceResp.VerificationURL)
fmt.Printf("验证码：%s\n", deviceResp.UserCode)
```

### PollToken

轮询获取访问令牌。

**方法签名：**
```go
func (a *AuthService) PollToken(deviceCode string, interval int) (*Token, error)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| deviceCode | string | 设备码 |
| interval | int | 轮询间隔（秒） |

**返回值：**
- `*Token` - 访问令牌
- `error` - 获取失败时返回错误

**示例：**
```go
token, err := client.Auth.PollToken(deviceResp.DeviceCode, deviceResp.Interval)
if err != nil {
    panic(err)
}
client.SetToken(token)
```

### IsTokenValid

检查当前 token 是否有效。

**方法签名：**
```go
func (a *AuthService) IsTokenValid() bool
```

**返回值：**
- `bool` - token 有效返回 true

### ClearToken

清除 token。

**方法签名：**
```go
func (a *AuthService) ClearToken()
```

### RefreshToken

使用 refresh_token 刷新 access_token。

**方法签名：**
```go
func (a *AuthService) RefreshToken(refreshToken string) (*Token, error)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| refreshToken | string | 刷新令牌 |

**返回值：**
- `*Token` - 新的访问令牌
- `error` - 刷新失败时返回错误

### Token

访问令牌结构体。

```go
type Token struct {
    AccessToken  string    // 访问令牌
    RefreshToken string    // 刷新令牌
    ExpiresIn    int64     // 有效期（秒）
    CreatedAt    int64     // 创建时间戳
    ExpiresAt    time.Time // 过期时间（运行时计算）
}
```

**方法：**
- `IsExpired() bool` - 检查 token 是否已过期
- `IsValid() bool` - 检查 token 是否有效

### DeviceCodeResponse

设备码响应结构体。

```go
type DeviceCodeResponse struct {
    DeviceCode      string  // 设备码
    UserCode        string  // 用户验证码
    VerificationURL string  // 验证 URL
    QrcodeURL       string  // 二维码 URL
    ExpiresIn       int     // 有效期（秒）
    Interval        int     // 轮询间隔（秒）
}
```

---

## 用户服务

### UserService

用户服务，提供用户信息和配额查询功能。

### GetInfo

获取用户信息。

**方法签名：**
```go
func (u *UserService) GetInfo() (*model.UserInfo, error)
```

**返回值：**
- `*model.UserInfo` - 用户信息
- `error` - 请求失败时返回错误

**示例：**
```go
userInfo, err := client.User.GetInfo()
if err != nil {
    panic(err)
}
fmt.Printf("用户名：%s\n", userInfo.NetdiskName)
fmt.Printf("VIP 类型：%d\n", userInfo.Vip)
```

### GetQuota

获取网盘配额信息。

**方法签名：**
```go
func (u *UserService) GetQuota() (*model.QuotaInfo, error)
```

**返回值：**
- `*model.QuotaInfo` - 配额信息
- `error` - 请求失败时返回错误

**示例：**
```go
quota, err := client.User.GetQuota()
if err != nil {
    panic(err)
}
fmt.Printf("总空间：%d GB\n", quota.Total/1024/1024/1024)
fmt.Printf("已使用：%d GB\n", quota.Used/1024/1024/1024)
fmt.Printf("剩余：%d GB\n", quota.Free/1024/1024/1024)
```

---

## 文件服务

### FileService

文件服务，提供文件列表、信息获取、复制、移动、删除、重命名、创建文件夹等功能。

### List

列出指定路径下的文件和文件夹。

**方法签名：**
```go
func (f *FileService) List(path string) (*model.FileList, error)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| path | string | 要列出的目录路径 |

**返回值：**
- `*model.FileList` - 文件列表
- `error` - 请求失败时返回错误

**示例：**
```go
fileList, err := client.File.List("/我的资源")
if err != nil {
    panic(err)
}
for _, file := range fileList.Items {
    fmt.Printf("%s (%s)\n", file.Name, file.Type)
}
```

### GetInfo

获取文件或文件夹的详细信息。

**方法签名：**
```go
func (f *FileService) GetInfo(path string) (*model.FileInfo, error)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| path | string | 文件或文件夹路径 |

**返回值：**
- `*model.FileInfo` - 文件信息
- `error` - 请求失败时返回错误

**示例：**
```go
info, err := client.File.GetInfo("/我的资源/文档.pdf")
if err != nil {
    panic(err)
}
fmt.Printf("文件大小：%d 字节\n", info.Size)
fmt.Printf("MD5: %s\n", info.MD5)
```

### Copy

复制文件或文件夹。

**方法签名：**
```go
func (f *FileService) Copy(srcPath, destPath string, ondup ...string) error
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| srcPath | string | 源文件/文件夹路径 |
| destPath | string | 目标文件夹路径 |
| ondup | string | 重名处理策略（可选）：`fail`(默认)、`newcopy`(覆盖)、`skip`(跳过) |

**示例：**
```go
// 默认重名失败
err := client.File.Copy("/源/文件.txt", "/目标文件夹")

// 覆盖已存在的文件
err := client.File.Copy("/源/文件.txt", "/目标文件夹", "newcopy")
```

### Move

移动文件或文件夹。

**方法签名：**
```go
func (f *FileService) Move(srcPath, destPath string, ondup ...string) (string, error)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| srcPath | string | 源文件/文件夹路径 |
| destPath | string | 目标文件夹路径，若不存在则直接移动到该路径 |
| ondup | string | 重名处理策略（可选）：`fail`(默认)、`newcopy`(覆盖)、`skip`(跳过) |

**返回值：**
- `string` - 实际移动到的目标路径
- `error` - 移动失败时返回错误

**示例：**
```go
actualPath, err := client.File.Move("/旧位置/文件.txt", "/新位置")
```

### Delete

删除文件或文件夹。

**方法签名：**
```go
func (f *FileService) Delete(paths ...string) error
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| paths | ...string | 要删除的文件或文件夹路径列表 |

**示例：**
```go
// 删除单个文件
err := client.File.Delete("/要删除的文件.txt")

// 删除多个文件
err := client.File.Delete("/文件 1.txt", "/文件 2.txt", "/文件夹")
```

### Rename

重命名文件或文件夹。

**方法签名：**
```go
func (f *FileService) Rename(path, newName string) error
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| path | string | 当前文件/文件夹路径 |
| newName | string | 新名称（不包含路径） |

**示例：**
```go
err := client.File.Rename("/旧名称.txt", "新名称.txt")
```

### CreateDir

创建文件夹。

**方法签名：**
```go
func (f *FileService) CreateDir(path string, rtype ...int) (*CreateDirReturn, error)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| path | string | 要创建的文件夹绝对路径 |
| rtype | int | 文件命名策略（可选）：`0`(不重命名返回冲突)、`1`(冲突即重命名，默认) |

**返回值：**
- `*CreateDirReturn` - 创建结果
- `error` - 创建失败时返回错误

**示例：**
```go
result, err := client.File.CreateDir("/我的资源/新文件夹")
if err != nil {
    panic(err)
}
fmt.Printf("文件夹 FSID: %d\n", result.FSID)
```

---

## 下载服务

### DownloadService

下载服务，提供文件下载功能，支持进度回调。

### Start

下载文件到本地。

**方法签名：**
```go
func (d *DownloadService) Start(remotePath, localPath string) error
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| remotePath | string | 网盘文件路径 |
| localPath | string | 本地保存路径（可以是文件或文件夹） |

**返回值：**
- `error` - 下载失败时返回错误

**示例：**
```go
err := client.Download.Start("/网盘/文件.pdf", "./本地/文件.pdf")
if err != nil {
    panic(err)
}
```

### StartWithProgress

下载文件到本地，带进度回调。

**方法签名：**
```go
func (d *DownloadService) StartWithProgress(remotePath, localPath string, callback ProgressCallback) error
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| remotePath | string | 网盘文件路径 |
| localPath | string | 本地保存路径 |
| callback | ProgressCallback | 进度回调函数 |

**示例：**
```go
callback := func(progress bdisk.DownloadProgress) {
    fmt.Printf("下载进度：%.2f%% (%d/%d)\n", 
        progress.Percent, progress.Downloaded, progress.Total)
}

err := client.Download.StartWithProgress("/网盘/大文件.zip", "./大文件.zip", callback)
```

### GetFileFSID

获取文件的 fs_id、大小和名称。

**方法签名：**
```go
func (d *DownloadService) GetFileFSID(path string) (int64, int64, string, error)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| path | string | 网盘文件路径 |

**返回值：**
- `int64` - 文件 fs_id
- `int64` - 文件大小（字节）
- `string` - 文件名
- `error` - 获取失败时返回错误

### DownloadProgress

下载进度结构体。

```go
type DownloadProgress struct {
    Downloaded int64   // 已下载字节数
    Total      int64   // 总字节数
    Percent    float64 // 下载百分比 (0-100)
}
```

### ProgressCallback

进度回调函数类型。

```go
type ProgressCallback func(progress DownloadProgress)
```

---

## 上传服务

### UploadService

上传服务，提供文件分片上传功能，支持进度回调。

### Start

上传本地文件到网盘。

**方法签名：**
```go
func (u *UploadService) Start(localPath, remotePath string) (string, error)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| localPath | string | 本地文件路径 |
| remotePath | string | 网盘保存路径（包含文件名），若为目标文件夹则自动拼接文件名 |

**返回值：**
- `string` - 实际保存的网盘路径
- `error` - 上传失败时返回错误

**示例：**
```go
err := client.Upload.Start("./本地/文件.pdf", "/网盘/文件.pdf")
if err != nil {
    panic(err)
}
```

### StartWithProgress

上传本地文件到网盘，带进度回调。

**方法签名：**
```go
func (u *UploadService) StartWithProgress(localPath, remotePath string, callback UploadProgressCallback) (string, error)
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| localPath | string | 本地文件路径 |
| remotePath | string | 网盘保存路径，若为目标文件夹则自动拼接文件名 |
| callback | UploadProgressCallback | 进度回调函数 |

**返回值：**
- `string` - 实际保存的网盘路径
- `error` - 上传失败时返回错误

**示例：**
```go
callback := func(progress bdisk.UploadProgress) {
    fmt.Printf("上传进度：%.2f%% (分片 %d/%d)\n", 
        progress.Percent, progress.CurrentPart, progress.TotalParts)
}

err := client.Upload.StartWithProgress("./大文件.zip", "/网盘/大文件.zip", callback)
```

### UploadProgress

上传进度结构体。

```go
type UploadProgress struct {
    Uploaded    int64   // 已上传字节数
    Total       int64   // 总字节数
    Percent     float64 // 上传百分比 (0-100)
    CurrentPart int     // 当前分片序号
    TotalParts  int     // 总分片数
}
```

### UploadProgressCallback

上传进度回调函数类型。

```go
type UploadProgressCallback func(progress UploadProgress)
```

---

## 错误处理

### 预定义错误

```go
var (
    ErrTokenExpired  = errors.New("access token has expired")  // Token 已过期
    ErrNoToken       = errors.New("no access token found")     // 未找到 Token
    ErrInvalidConfig = errors.New("invalid configuration")     // 配置无效
)
```

### APIError

API 错误结构体。

```go
type APIError struct {
    ErrCode int    `json:"errno"`   // 错误码
    ErrMsg  string `json:"errmsg"`  // 错误信息
}
```

### IsTokenExpiredError

检查错误是否为 token 过期错误。

**函数签名：**
```go
func IsTokenExpiredError(err error) bool
```

**参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| err | error | 要检查的错误 |

**返回值：**
- `bool` - 是 token 过期错误返回 true

**示例：**
```go
userInfo, err := client.User.GetInfo()
if err != nil {
    if bdisk.IsTokenExpiredError(err) {
        // token 已过期，清除并重新登录
        client.Auth.ClearToken()
        // 重新执行授权流程...
    }
}
```

### Token 过期处理

```go
// 检查 token 是否有效
if !client.Auth.IsTokenValid() {
    // token 已过期，需要重新登录
    client.Auth.ClearToken()
    // 引导用户重新授权...
}
```

---

## 数据模型

### model.FileInfo

文件信息结构体。

```go
type FileInfo struct {
    FSID       int64           // 文件唯一标识
    Path       string          // 文件路径
    Name       string          // 文件名
    Type       FileType        // 文件类型 (file/dir)
    Size       int64           // 文件大小（字节）
    ModifiedAt time.Time       // 修改时间
    CreatedAt  time.Time       // 创建时间
    MD5        string          // 文件 MD5
}
```

### model.FileList

文件列表响应结构体。

```go
type FileList struct {
    Path  string      // 当前路径
    Items []*FileInfo // 文件列表
    Total int         // 总数
}
```

### model.FileType

文件类型枚举。

```go
type FileType string

const (
    FileTypeFile FileType = "file"  // 文件
    FileTypeDir  FileType = "dir"   // 文件夹
)
```

### model.UserInfo

用户信息结构体。

```go
type UserInfo struct {
    BaiduName   string  // 百度账号
    NetdiskName string  // 网盘账号
    AvatarURL   string  // 头像地址
    Vip         int     // 会员类型
}
```

### model.QuotaInfo

网盘配额信息结构体。

```go
type QuotaInfo struct {
    Total int64  // 总空间大小（字节）
    Used  int64  // 已用空间大小（字节）
    Free  int64  // 剩余空间大小（字节）
}
```

---

## 常见错误码

| 错误码 | 说明 |
|--------|------|
| 110 | Token 已过期 |
| 40023 | 设备授权码已过期 |
| 40024 | 用户尚未授权（轮询中） |
| 40025 | 设备授权码已使用 |

---

## 注意事项

1. **Token 有效期**：Access Token 默认有效期为 30 天，过期后需要使用 RefreshToken 刷新或重新授权
2. **API 调用频率**：请注意百度网盘 API 的调用频率限制
3. **文件大小限制**：单文件上传最大支持 20GB
4. **分片上传**：SDK 默认使用 4MB 分片大小进行上传
5. **路径格式**：所有路径必须使用绝对路径，以 `/` 开头
