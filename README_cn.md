# go-scihub-mcp

[中文文档](README_cn.md) | [English](README.md)

一个用 Go 语言编写的 Sci-Hub 镜像管理和文件下载工具，支持 MCP (Model Context Protocol) 兼容的 API 服务。

## 功能特性

- 📋 维护可用的 Sci-Hub 镜像列表
- 🔄 自动检测和更新镜像可用性
- 🌐 支持 SOCKS5 代理配置
- 📁 文件下载和缓存
- 🔗 MCP 兼容的 HTTP API 服务
- 🚀 支持多种传输模式：STDIO 和 SSE (Server-Sent Events)
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
  type: "socks5"      # 支持 socks5, http
  host: "127.0.0.1"
  port: 3080
  username: ""        # 可选：代理用户名
  password: ""        # 可选：代理密码

# 健康检查配置
health_check:
  interval: "30m"     # 检查间隔：30分钟
  timeout: "10s"      # 请求超时：10秒

# MCP 服务配置
mcp:
  port: 8080
  host: "0.0.0.0"     # 监听所有接口
  transport: "stdio"  # 传输模式: stdio (标准输入输出), sse (服务器推送事件)
  sse_path: "/sse"    # SSE端点路径 (仅sse模式)
  
# 下载配置
download:
  cache_dir: "./cache"    # 缓存目录
  max_retries: 3         # 最大重试次数
  timeout: "60s"         # 下载超时时间
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
# 使用默认配置运行（HTTP API 服务）
./scihub-mcp

# 启用全局代理运行（HTTP API 服务）
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

MCP 协议服务器支持两种传输模式：

#### STDIO 传输模式（默认）
```bash
# 启动 MCP 协议服务器（STDIO 通信）
./scihub-mcp mcp

# 启用全局代理启动
./scihub-mcp --proxy-enabled mcp

# 明确指定 stdio 模式
./scihub-mcp mcp --transport stdio

# MCP 协议通过 STDIN/STDOUT 通信
# 工具调用示例：
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | ./scihub-mcp mcp

# 资源访问示例：
echo '{"jsonrpc": "2.0", "id": 2, "method": "resources/read", "params": {"uri": "scihub://cache"}}' | ./scihub-mcp mcp
```

#### SSE 传输模式（服务器推送事件）
```bash
# 启动 MCP 协议服务器（SSE 传输）
./scihub-mcp mcp --transport sse

# 使用自定义端口和 SSE 传输
./scihub-mcp --mcp-port 9090 mcp --transport sse

# 启用代理和 SSE 传输
./scihub-mcp --proxy-enabled mcp --transport sse

# 使用自定义配置文件和 SSE 设置
./scihub-mcp --config configs/config-sse.yaml mcp
```

**SSE 模式端点：**
- SSE 流：`http://localhost:8080/sse`
- 消息端点：`http://localhost:8080/message`
- 健康检查：`http://localhost:8080/health`

**SSE 传输特性：**
- 基于标准 HTTP 的实时双向通信
- 类似 WebSocket 的功能，但使用标准 HTTP
- 更适合 Web 应用和远程客户端
- 支持并发连接
- 内置重连和错误处理
- 兼容 HTTP/2 和现代 Web 基础设施

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

要在 Cursor、Claude Desktop 或其他 MCP 客户端中使用 MCP 协议服务器，需要在客户端设置中配置服务器。服务器支持两种传输模式：

**STDIO 模式**：直接进程通信（推荐本地使用）
**SSE 模式**：基于 HTTP 的通信（推荐远程/Web 使用）

### Cursor AI 配置

1. **打开 Cursor 设置**：
   - 按 `Cmd/Ctrl + ,` 打开设置
   - 转到 "Extensions" -> "MCP Servers" 或搜索 "MCP"

2. **添加 SciHub-MCP 服务器**：

   **STDIO 模式（推荐本地使用）**：
   创建或编辑你的 MCP 配置文件（通常是 `~/.cursor/mcp_servers.json`）：
   ```json
   {
     "mcpServers": {
       "scihub-mcp-stdio": {
         "description": "SciHub MCP 服务器 - STDIO 模式",
         "command": "/path/to/scihub-mcp",
         "args": ["--proxy-enabled", "--proxy-host", "127.0.0.1", "--proxy-port", "3080", "mcp", "--transport", "stdio"],
         "env": {},
         "cwd": "/path/to/working/directory"
       }
     }
   }
   ```

   **SSE 模式（远程/Web 使用）**：
   ```json
   {
     "mcpServers": {
       "scihub-mcp-sse": {
         "description": "SciHub MCP 服务器 - SSE 模式",
         "url": "http://localhost:8080/sse"
       }
     }
   }
   ```

   或者不使用代理：
   ```json
   {
     "mcpServers": {
       "scihub-mcp-stdio": {
         "command": "/path/to/scihub-mcp",
         "args": ["mcp"],
         "env": {},
         "cwd": "/path/to/working/directory"
       },
       "scihub-mcp-sse": {
         "url": "http://localhost:8080/sse"
       }
     }
   }
   ```

### Claude Desktop 配置

1. **找到配置文件位置**：
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\\Claude\\claude_desktop_config.json`
   - Linux: `~/.config/Claude/claude_desktop_config.json`

2. **添加服务器配置**：

   **STDIO 模式（默认）**：
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

   **SSE 模式**：
   ```json
   {
     "mcpServers": {
       "scihub-mcp-sse": {
         "url": "http://localhost:8080/sse"
       }
     }
   }
   ```

   **包含代理支持的两种模式**：
   ```json
   {
     "mcpServers": {
       "scihub-mcp-stdio": {
         "command": "/path/to/scihub-mcp",
         "args": ["--proxy-enabled", "--proxy-host", "127.0.0.1", "--proxy-port", "3080", "mcp", "--transport", "stdio"]
       },
       "scihub-mcp-sse": {
         "url": "http://localhost:8080/sse"
       }
     }
   }
   ```

   注意：对于 SSE 模式，需要单独启动服务器：
   ```bash
   # 后台启动 SSE 服务器
   ./scihub-mcp --proxy-enabled mcp --transport sse &
   ```

### 通用 MCP 客户端配置

对于其他 MCP 客户端，使用这些设置：

- **服务器命令**: `/path/to/scihub-mcp mcp`
- **通信方式**: STDIO（标准输入/输出）
- **协议**: JSON-RPC 2.0 with MCP 扩展
- **环境变量**: 无需设置
- **工作目录**: 包含可执行文件的目录

### 测试 MCP 连接

可以手动测试 MCP 服务器：

```bash
# 测试服务器启动和工具列表
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | ./scihub-mcp mcp

# 期望响应应显示可用工具，如：
# {"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"download_paper",...}]}}
```

### 配置示例

为了方便使用，我们在 `configs/` 目录中提供了示例配置文件：

- [`configs/cursor_mcp_config.json`](configs/cursor_mcp_config.json) - 基本 Cursor 配置
- [`configs/config-sse.yaml`](configs/config-sse.yaml) - SSE 模式配置

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

**包含代理**（适用于网络受限的用户）：
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

**使用自定义配置文件**：
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
- **类型**: 基于多种传输的模型上下文协议
- **传输模式**: 
  - **STDIO**: 通过 STDIN/STDOUT 的 JSON-RPC 2.0 消息（默认）
  - **SSE**: 通过 HTTP Server-Sent Events 的 JSON-RPC 2.0 消息
- **适用场景**: LLM 应用、MCP 客户端、Cursor AI 集成
- **功能**: 工具、资源和模板
- **格式**: 标准 MCP 协议消息

**传输模式对比：**

| 功能 | STDIO 模式 | SSE 模式 |
|------|------------|----------|
| 通信方式 | 进程管道 | HTTP/SSE |
| 适用场景 | 本地 CLI 工具 | Web/远程应用 |
| 设置方式 | 直接执行 | 服务器 + 客户端 |
| 并发性 | 单会话 | 多会话 |
| 网络需求 | 不需要 | HTTP 网络 |
| 防火墙 | 不受影响 | 可能需要端口访问 |

**可用 MCP 工具：**
- `download_paper` - 下载科学论文
- `check_mirror_status` - 检查镜像可用性
- `test_mirror` - 测试特定镜像
- `list_available_mirrors` - 列出可用镜像

**可用 MCP 资源：**
- `scihub://cache` - 列出缓存的论文（JSON）
- `scihub://mirrors/status` - 镜像状态信息（JSON）
- `scihub://papers/{filename}` - 访问特定论文文件（PDF）

**MCP 使用示例（STDIO 模式）：**
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

**MCP 使用示例（SSE 模式）：**
```bash
# 启动 SSE 服务器
./scihub-mcp mcp --transport sse &

# 连接到 SSE 端点获取事件流
curl -N http://localhost:8080/sse

# 向消息端点发送 JSON-RPC 消息
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}'
```