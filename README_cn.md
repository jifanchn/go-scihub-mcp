# go-scihub-mcp

[中文文档](README_cn.md) | [English](README.md)

一个用 Go 语言编写的 Sci-Hub 镜像管理和文件下载工具，支持 MCP (Model Context Protocol) 兼容的 API 服务。

## 功能特性

- 📋 维护可用的 Sci-Hub 镜像列表
- 🔄 自动检测和更新镜像可用性
- 🌐 支持 SOCKS5 代理配置
- 📁 文件下载和缓存
- 🔗 MCP 兼容的 HTTP API 服务
- ⚙️ 清晰优先级的灵活命令行配置

## 安装

### 从源码构建

```bash
git clone https://github.com/jifanchn/go-scihub-mcp.git
cd go-scihub-mcp
go mod tidy
go build -o scihub-mcp ./cmd/scihub-mcp
```

### 二进制下载

从 [Releases](https://github.com/jifanchn/go-scihub-mcp/releases) 页面下载对应平台的二进制文件。

## 配置

### 配置优先级

配置遵循清晰的优先级系统：
**命令行参数 > 配置文件 > 默认值**

这意味着全局选项适用于所有命令，并且在适用的情况下可以被命令特定选项覆盖。

### 配置文件

程序支持 YAML 配置文件，默认查找路径：
- `./config.yaml`
- `~/.config/scihub-mcp/config.yaml`
- `/etc/scihub-mcp/config.yaml`

配置文件示例：

```yaml
# Sci-Hub 镜像列表
mirrors:
  - "https://sci-hub.ru"
  - "https://sci-hub.se"
  - "https://sci-hub.st"
  - "https://sci-hub.box"
  - "https://sci-hub.red"
  - "https://sci-hub.al"
  - "https://sci-hub.ee"
  - "https://sci-hub.lu"
  - "https://sci-hub.ren"
  - "https://sci-hub.shop"
  - "https://sci-hub.vg"

# 代理配置
proxy:
  enabled: false
  type: "socks5"
  host: "127.0.0.1"
  port: 3080
  username: ""
  password: ""

# 健康检查配置
health_check:
  interval: "30m"  # 检查间隔
  timeout: "10s"   # 请求超时

# MCP 服务配置
mcp:
  port: 8080
  host: "0.0.0.0"
  
# 下载配置
download:
  cache_dir: "./cache"
  max_retries: 3
  timeout: "60s"
```

### 命令行参数

全局选项适用于所有命令，可以在命令之前指定：

```bash
# 全局代理设置
scihub-mcp --proxy-enabled --proxy-host 127.0.0.1 --proxy-port 3080 [命令]

# 全局配置文件
scihub-mcp --config /path/to/config.yaml [命令]

# 全局 MCP 服务器设置
scihub-mcp --mcp-host 0.0.0.0 --mcp-port 9090 [命令]
```

## 使用方法

### 1. 基本运行

```bash
# 使用默认配置运行
./scihub-mcp

# 启用全局代理运行
./scihub-mcp --proxy-enabled --proxy-host 127.0.0.1 --proxy-port 3080

# 使用自定义配置文件
./scihub-mcp --config /path/to/config.yaml
```

### 2. 文件下载命令

```bash
# 通过 DOI 下载文件
./scihub-mcp fetch --doi "10.1038/nature12373"

# 使用全局代理下载
./scihub-mcp --proxy-enabled fetch --doi "10.1038/nature12373"

# 指定自定义输出路径
./scihub-mcp fetch --doi "10.1038/nature12373" --output "./papers/nature12373.pdf"

# 通过 URL 下载文件
./scihub-mcp fetch --url "https://www.nature.com/articles/nature12373"
```

### 3. HTTP API 服务模式

```bash
# 启动 HTTP API 服务（默认端口 8080）
./scihub-mcp api

# 启用全局代理启动
./scihub-mcp --proxy-enabled api

# 使用全局设置在自定义端口启动
./scihub-mcp --mcp-port 9090 api

# 仅为此实例覆盖端口
./scihub-mcp api --port 9090

# 结合全局代理和自定义端口
./scihub-mcp --proxy-enabled api --port 9090
```

### 4. MCP 协议服务模式

```bash
# 启动 MCP 协议服务器（STDIO 通信）
./scihub-mcp mcp

# 启用全局代理启动
./scihub-mcp --proxy-enabled mcp

# MCP 协议通过 STDIN/STDOUT 通信
# 工具调用示例：
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | ./scihub-mcp mcp

# 资源访问示例：
echo '{"jsonrpc": "2.0", "id": 2, "method": "resources/read", "params": {"uri": "scihub://cache"}}' | ./scihub-mcp mcp
```

### 5. 镜像状态检查

```bash
# 检查所有镜像状态
./scihub-mcp status

# 启用代理检查
./scihub-mcp --proxy-enabled status

# 测试特定镜像
./scihub-mcp test --mirror "https://sci-hub.se"
```

## MCP 服务器配置

要在 Cursor、Claude Desktop 或其他 MCP 客户端中使用 MCP 协议服务器，需要在客户端设置中配置服务器。

### Cursor AI 配置

1. **打开 Cursor 设置**：
   - 按 `Cmd/Ctrl + ,` 打开设置
   - 转到 "Extensions" -> "MCP Servers" 或搜索 "MCP"

2. **添加 SciHub-MCP 服务器**：
   创建或编辑你的 MCP 配置文件（通常是 `~/.cursor/mcp_servers.json`）：
   ```json
   {
     "mcpServers": {
       "scihub-mcp": {
         "command": "/path/to/scihub-mcp",
         "args": ["--proxy-enabled", "--proxy-host", "127.0.0.1", "--proxy-port", "3080", "mcp"],
         "env": {},
         "cwd": "/path/to/working/directory"
       }
     }
   }
   ```

   或者不使用代理：
   ```json
   {
     "mcpServers": {
       "scihub-mcp": {
         "command": "/path/to/scihub-mcp",
         "args": ["mcp"],
         "env": {},
         "cwd": "/path/to/working/directory"
       }
     }
   }
   ```

### Claude Desktop 配置

1. **找到配置文件位置**：
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
   - Linux: `~/.config/Claude/claude_desktop_config.json`

2. **添加服务器配置**：
   ```json
   {
     "mcpServers": {
       "scihub-mcp": {
         "command": "/path/to/scihub-mcp",
         "args": ["mcp"]
       }
     }
   }
   ```

   支持代理的配置：
   ```json
   {
     "mcpServers": {
       "scihub-mcp": {
         "command": "/path/to/scihub-mcp",
         "args": ["--proxy-enabled", "--proxy-host", "127.0.0.1", "--proxy-port", "3080", "mcp"]
       }
     }
   }
   ```

### 通用 MCP 客户端配置

对于其他 MCP 客户端，使用以下设置：

- **服务器命令**: `/path/to/scihub-mcp mcp`
- **通信方式**: STDIO（标准输入/输出）
- **协议**: 带 MCP 扩展的 JSON-RPC 2.0
- **环境变量**: 无需设置
- **工作目录**: 包含可执行文件的目录

### 测试 MCP 连接

你可以手动测试 MCP 服务器：

```bash
# 测试服务器启动和工具列表
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | ./scihub-mcp mcp

# 预期响应应显示可用工具，如：
# {"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"download_paper",...}]}}
```

### 配置示例

**基本配置**（无代理）：
```json
{
  "mcpServers": {
    "scihub-mcp": {
      "command": "/usr/local/bin/scihub-mcp",
      "args": ["mcp"]
    }
  }
}
```

**使用代理**（适用于网络受限用户）：
```json
{
  "mcpServers": {
    "scihub-mcp": {
      "command": "/usr/local/bin/scihub-mcp",
      "args": ["--proxy-enabled", "--proxy-host", "127.0.0.1", "--proxy-port", "3080", "mcp"]
    }
  }
}
```

**自定义配置文件**：
```json
{
  "mcpServers": {
    "scihub-mcp": {
      "command": "/usr/local/bin/scihub-mcp",
      "args": ["--config", "/path/to/custom-config.yaml", "mcp"],
      "cwd": "/path/to/cache/directory"
    }
  }
}
```

### MCP 设置故障排除

1. **服务器无法启动**：
   - 检查二进制文件路径是否正确
   - 确保二进制文件有执行权限
   - 验证工作目录是否存在

2. **代理问题**：
   - 手动测试代理连接
   - 检查代理服务器是否在指定主机/端口运行
   - 先尝试不使用代理以隔离问题

3. **工具/资源未找到**：
   - 配置更改后重启 MCP 客户端
   - 检查服务器日志是否有启动错误
   - 验证配置 JSON 语法是否正确

4. **权限错误**：
   - 确保缓存目录可写
   - 检查二进制文件和配置文件的文件权限

## 服务类型

本工具提供两种不同的服务模式：

### HTTP API 服务 (`api`)
- **类型**: 基于 HTTP 的 REST API
- **通信方式**: 标准 HTTP 请求/响应
- **适用场景**: Web 应用、curl 命令、浏览器访问
- **端点**: `/health`, `/fetch`, `/download/{filename}`, `/mirrors`, `/status`
- **格式**: 带标准 HTTP 状态码的 JSON 响应

**HTTP API 使用示例：**
```bash
# 健康检查
curl http://localhost:8080/health

# 下载论文
curl -X POST http://localhost:8080/fetch \
  -H "Content-Type: application/json" \
  -d '{"doi": "10.1038/nature12373"}'

# 直接下载文件
curl http://localhost:8080/download/paper.pdf --output paper.pdf
```

### MCP 协议服务 (`mcp`)
- **类型**: 基于 STDIO 的模型上下文协议
- **通信方式**: 通过 STDIN/STDOUT 的 JSON-RPC 2.0 消息
- **适用场景**: LLM 应用、MCP 客户端、Cursor AI 集成
- **功能**: 工具、资源和模板
- **格式**: 标准 MCP 协议消息

**可用 MCP 工具：**
- `download_paper` - 下载科学论文
- `check_mirror_status` - 检查镜像可用性
- `test_mirror` - 测试特定镜像
- `list_available_mirrors` - 列出可用镜像

**可用 MCP 资源：**
- `scihub://cache` - 列出缓存的论文（JSON）
- `scihub://mirrors/status` - 镜像状态信息（JSON）
- `scihub://papers/{filename}` - 访问特定论文文件（PDF）

**MCP 使用示例：**
```bash
# 列出可用工具
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | ./scihub-mcp mcp

# 调用下载工具
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "download_paper", "arguments": {"doi": "10.1038/nature12373"}}}' | ./scihub-mcp mcp

# 列出资源
echo '{"jsonrpc": "2.0", "id": 3, "method": "resources/list", "params": {}}' | ./scihub-mcp mcp

# 读取缓存资源
echo '{"jsonrpc": "2.0", "id": 4, "method": "resources/read", "params": {"uri": "scihub://cache"}}' | ./scihub-mcp mcp
```

## API 接口

当运行在 HTTP API 模式（`api`）时，提供以下 HTTP 接口：

### GET /health
检查服务健康状态

```bash
curl http://localhost:8080/health
```

### POST /fetch
下载论文文件并返回 JSON 响应

```bash
curl -X POST http://localhost:8080/fetch \
  -H "Content-Type: application/json" \
  -d '{"doi": "10.1038/nature12373"}'
```

请求格式：
```json
{
  "doi": "10.1038/nature12373",          // DOI (可选)
  "url": "https://example.com/paper",    // 原始URL (可选)
  "title": "Paper Title"                 // 论文标题 (可选)
}
```

响应格式：
```json
{
  "success": true,
  "message": "文件下载成功",
  "data": {
    "filename": "nature12373.pdf",
    "size": 1024000,
    "mirror_used": "https://sci-hub.se",
    "download_url": "https://sci-hub.se/downloads/...",
    "cached": false,
    "download_link": "/download/nature12373.pdf"
  }
}
```

### GET /download/{filename}
通过文件名下载文件

```bash
curl http://localhost:8080/download/nature12373.pdf --output paper.pdf
```

### POST /fetch?return_file=true
下载并直接返回文件内容

```bash
curl -X POST "http://localhost:8080/fetch?return_file=true" \
  -H "Content-Type: application/json" \
  -d '{"doi": "10.1038/nature12373"}' \
  --output paper.pdf
```

### GET /mirrors
获取当前可用镜像状态

```bash
curl http://localhost:8080/mirrors
```

### GET /status
获取系统状态

```bash
curl http://localhost:8080/status
```

## 工作原理

1. **镜像管理**：程序启动时加载配置的镜像列表，后台 goroutine 定期检查每个镜像的可用性
2. **智能选择**：下载时自动选择可用且响应最快的镜像
3. **代理支持**：支持 SOCKS5 代理，在网络受限环境下使用
4. **缓存机制**：已下载的文件会缓存在本地，避免重复下载
5. **错误处理**：具备重试机制和详细的错误日志
6. **灵活配置**：清晰的优先级系统确保可预测的行为

## 开发

### 项目结构

```
go-scihub-mcp/
├── cmd/scihub-mcp/     # 主程序入口
├── internal/           # 内部包
│   ├── config/         # 配置管理
│   ├── mirror/         # 镜像管理
│   ├── downloader/     # 下载器
│   ├── mcp/           # MCP 兼容的 HTTP API 服务
│   └── proxy/         # 代理管理
├── pkg/               # 公共包
├── configs/           # 示例配置文件
├── docs/             # 文档
└── README.md
```

### 运行测试

```bash
go test ./...
```

### 构建发布版本

```bash
# 构建当前平台
make build

# 交叉编译
make build-all

# 创建发布包
make release
```

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 免责声明

本工具仅用于学术研究目的。请确保您的使用符合当地法律法规和相关网站的服务条款。 