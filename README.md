# go-bdisk

[![Go Version](https://img.shields.io/github/go-mod/go-version/baowuhe/go-bdisk)](go.mod)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

使用 Go 语言开发的百度网盘 CLI 工具和 SDK，支持设备码授权、文件管理、下载上传等功能。

## 项目简述

`go-bdisk` 是一个功能完整的百度网盘客户端工具，包含：

- **CLI 工具** - 命令行界面，支持日常文件管理操作
- **Go SDK** - 百度网盘 API 的 Go 语言封装，可在其他项目中集成使用

本项目采用设备码授权方式，无需在命令行中输入账号密码，通过扫码即可完成授权登录。

## 功能特性

### CLI 工具功能

| 命令 | 功能描述 |
|------|----------|
| `login` | 登录百度网盘（设备码授权） |
| `logout` | 退出登录并清除本地凭证 |
| `ls [path]` | 列出指定路径下的文件和文件夹 |
| `stat <path>` | 查看文件或文件夹的详细信息 |
| `download <remote> [local]` | 下载文件到本地 |
| `upload <local> [remote]` | 上传本地文件到网盘 |
| `mkdir <path>` | 创建文件夹 |
| `rename <path> <new-name>` | 重命名文件/文件夹 |
| `cp <src> <dest>` | 复制文件/文件夹 |
| `mv <src> <dest>` | 移动文件/文件夹 |
| `rm <path> [...]` | 删除文件/文件夹（支持批量） |
| `status` | 查看用户信息 |

### SDK 功能

- ✅ 设备码授权流程（Device Code Flow）
- ✅ Token 自动管理与过期刷新
- ✅ 用户信息查询
- ✅ 网盘配额查询
- ✅ 文件列表与元数据获取
- ✅ 文件复制、移动、删除、重命名
- ✅ 文件夹创建
- ✅ 分片上传（支持进度回调）
- ✅ 文件下载（支持进度回调）

## 安装方法

### 方式一：从源码安装

```bash
# 克隆仓库
git clone https://github.com/baowuhe/go-bdisk.git
cd go-bdisk

# 编译安装
go install .
```

安装完成后，可执行文件位于 `$GOPATH/bin/go-bdisk`

### 方式二：直接使用 Go 命令

```bash
go install github.com/baowuhe/go-bdisk@latest
```

### 方式三：下载预编译二进制

访问 [Releases](https://github.com/baowuhe/go-bdisk/releases) 页面下载对应平台的预编译版本。

## 使用方法

### 快速开始

#### 1. 首次登录

首次使用需要配置百度开放平台的应用凭证：

```bash
# 使用应用 Key 和密钥登录
go-bdisk login --app-key "your_app_key" --secret-key "your_secret_key"
```

**获取应用凭证：**
1. 访问 [百度开放平台](https://pan.baidu.com/union/)
2. 创建应用并获取 `AppKey` 和 `SecretKey`
3. 确保应用权限包含网盘 API

登录流程：
```
================================
请使用浏览器访问以下链接并输入验证码：
验证链接：https://openapi.baidu.com/...
验证码：ABC123
================================
正在等待授权...
登录成功！
```

#### 2. 查看文件列表

```bash
# 列出根目录文件
go-bdisk ls

# 列出指定目录
go-bdisk ls /我的资源

# JSON 格式输出
go-bdisk ls -j
```

#### 3. 上传文件

```bash
# 上传到根目录（使用原文件名）
go-bdisk upload ./本地文件.pdf

# 上传到指定路径
go-bdisk upload ./本地文件.pdf /网盘文件夹/重命名.pdf
```

#### 4. 下载文件

```bash
# 下载到当前目录（使用网盘文件名）
go-bdisk download /网盘/文件.pdf

# 下载到指定路径
go-bdisk download /网盘/文件.pdf ./下载/文件.pdf
```

#### 5. 文件管理

```bash
# 创建文件夹
go-bdisk mkdir /新文件夹

# 重命名文件
go-bdisk rename /旧名称.txt 新名称.txt

# 复制文件
go-bdisk cp /源文件.txt /目标文件夹

# 移动文件
go-bdisk mv /源文件.txt /目标文件夹

# 删除文件
go-bdisk rm /要删除的文件.txt

# 批量删除
go-bdisk rm /文件 1.txt /文件 2.txt /文件夹
```

#### 6. 查看用户信息

```bash
go-bdisk status
```

#### 7. 退出登录

```bash
go-bdisk logout
```

### 全局选项

| 选项 | 简写 | 说明 |
|------|------|------|
| `--json` | `-j` | 使用 JSON 格式输出（便于脚本处理） |

### 命令帮助

```bash
# 查看所有命令
go-bdisk --help

# 查看具体命令帮助
go-bdisk login --help
go-bdisk upload --help
```

### 使用 SDK 开发

在你的 Go 项目中使用百度网盘 SDK：

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

    // 设备码授权
    deviceResp, err := client.Auth.DeviceCodeFlow()
    if err != nil {
        panic(err)
    }

    fmt.Printf("请访问：%s\n", deviceResp.VerificationURL)
    fmt.Printf("验证码：%s\n", deviceResp.UserCode)

    // 轮询获取 token
    token, err := client.Auth.PollToken(deviceResp.DeviceCode, deviceResp.Interval)
    if err != nil {
        panic(err)
    }

    client.SetToken(token)

    // 获取文件列表
    fileList, err := client.File.List("/")
    if err != nil {
        panic(err)
    }

    for _, file := range fileList.Items {
        fmt.Printf("%s (%s)\n", file.Name, file.Type)
    }
}
```

详细 SDK API 文档请参阅 [pkg/API.md](pkg/API.md)

## 配置说明

### 配置文件位置

配置文件存储于用户配置目录：

| 系统 | 路径 |
|------|------|
| Linux | `~/.local/share/go-bdisk/` |
| macOS | `~/.local/share/go-bdisk/` |
| Windows | `%LOCALAPPDATA%\bdisk\` |

### 配置文件

- `bdisk.yml` - 存储应用 Key 和密钥
- `token.json` - 存储访问令牌

## 注意事项

1. **Token 有效期**：Access Token 默认有效期为 30 天，过期后会自动刷新或提示重新登录
2. **API 调用限制**：请遵守百度网盘 API 的调用频率限制
3. **文件大小**：单文件上传最大支持 20GB
4. **路径格式**：所有网盘路径必须使用绝对路径，以 `/` 开头

## 项目结构

```
go-bdisk/
├── cmd/                    # CLI 命令实现
│   ├── root.go            # 根命令
│   ├── login.go           # 登录命令
│   ├── logout.go          # 登出命令
│   ├── ls.go              # 列表命令
│   ├── download.go        # 下载命令
│   ├── upload.go          # 上传命令
│   └── ...
├── pkg/bdisk/             # SDK 核心实现
│   ├── client.go          # 客户端
│   ├── auth.go            # 认证服务
│   ├── file.go            # 文件服务
│   ├── download.go        # 下载服务
│   ├── upload.go          # 上传服务
│   ├── user.go            # 用户服务
│   └── model/             # 数据模型
├── internal/              # 内部包
│   ├── config/            # 配置管理
│   └── cliutil/           # CLI 工具函数
└── main.go                # 程序入口
```

## 开发

```bash
# 运行
go run main.go

# 测试
go test ./...

# 构建
go build -o go-bdisk .
```

## License

MIT License
