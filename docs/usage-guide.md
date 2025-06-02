# SciHub-MCP 使用指南

本指南详细介绍了如何使用 SciHub-MCP 工具的各种功能和模式。

## 目录

1. [快速开始](#快速开始)
2. [传输模式详解](#传输模式详解)
3. [代理配置](#代理配置)
4. [MCP 客户端集成](#mcp-客户端集成)
5. [API 接口使用](#api-接口使用)
6. [故障排除](#故障排除)
7. [高级配置](#高级配置)

## 快速开始

### 1. 安装和设置

```bash
# 下载并安装
git clone https://github.com/jifanchn/go-scihub-mcp.git
cd go-scihub-mcp
go build -o scihub-mcp ./cmd/scihub-mcp

# 创建配置文件
cp configs/config.yaml ./config.yaml

# 编辑配置文件（可选）
nano config.yaml
```

### 2. 基本下载

```bash
# 通过 DOI 下载论文
./scihub-mcp fetch --doi "10.1038/nature12373"

# 通过 URL 下载论文
./scihub-mcp fetch --url "https://www.nature.com/articles/nature12373"

# 指定输出路径
./scihub-mcp fetch --doi "10.1038/nature12373" --output "./papers/nature_paper.pdf"
```

### 3. 启动服务

```bash
# 启动 HTTP API 服务
./scihub-mcp api

# 启动 MCP 协议服务器（STDIO 模式）
./scihub-mcp mcp

# 启动 MCP 协议服务器（SSE 模式）
./scihub-mcp mcp --transport sse
```

## 传输模式详解

### STDIO 模式

STDIO 模式是默认的 MCP 传输模式，适用于：
- 本地命令行工具
- 直接进程调用
- 简单的客户端集成

**优点：**
- 简单直接，无需网络配置
- 低延迟通信
- 不受防火墙影响

**使用方式：**
```bash
# 启动 STDIO 服务器
./scihub-mcp mcp --transport stdio

# 测试工具列表
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | ./scihub-mcp mcp
```

### SSE 模式

SSE（Server-Sent Events）模式适用于：
- Web 应用集成
- 远程客户端连接
- 需要并发连接的场景

**优点：**
- 支持多个并发连接
- 基于标准 HTTP 协议
- 内置重连机制
- 兼容现代 Web 基础设施

**使用方式：**
```bash
# 启动 SSE 服务器
./scihub-mcp mcp --transport sse

# 在另一个终端测试连接
curl -N http://localhost:8080/sse

# 发送 JSON-RPC 消息
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}'
```

## 代理配置

### 配置文件方式

编辑 `config.yaml`：
```yaml
proxy:
  enabled: true
  type: "socks5"      # 或 "http"
  host: "127.0.0.1"
  port: 3080
  username: ""        # 可选
  password: ""        # 可选
```

### 命令行方式

```bash
# 全局代理设置
./scihub-mcp --proxy-enabled --proxy-host 127.0.0.1 --proxy-port 3080 [命令]

# 示例：使用代理下载文件
./scihub-mcp --proxy-enabled fetch --doi "10.1038/nature12373"

# 示例：使用代理启动 MCP 服务器
./scihub-mcp --proxy-enabled mcp --transport sse
```

### 代理类型对比

| 代理类型 | 协议 | 安全性 | 速度 | 适用场景 |
|----------|------|--------|------|----------|
| SOCKS5   | TCP  | 高     | 快   | 通用代理 |
| HTTP     | HTTP | 中     | 中   | Web 代理 |

## MCP 客户端集成

### Cursor AI 集成

1. **创建配置文件**：
   ```bash
   mkdir -p ~/.cursor
   cat > ~/.cursor/mcp_servers.json << 'EOF'
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
   EOF
   ```

2. **重启 Cursor**：重启应用以加载新配置

3. **测试连接**：在 Cursor 中询问有关论文下载的问题

### Claude Desktop 集成

1. **找到配置文件**：
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
   - Linux: `~/.config/Claude/claude_desktop_config.json`

2. **添加配置**：
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

3. **重启 Claude Desktop**

### 自定义客户端

对于自定义 MCP 客户端，需要实现：

1. **进程启动**：启动 `scihub-mcp mcp` 进程
2. **JSON-RPC 通信**：通过 STDIN/STDOUT 发送 JSON-RPC 2.0 消息
3. **工具调用**：实现 MCP 工具调用协议
4. **资源访问**：实现 MCP 资源读取协议

示例代码（Python）：
```python
import subprocess
import json

# 启动 MCP 服务器
process = subprocess.Popen(
    ["/path/to/scihub-mcp", "mcp"],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
    text=True
)

# 发送工具列表请求
request = {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list",
    "params": {}
}

process.stdin.write(json.dumps(request) + "\n")
process.stdin.flush()

# 读取响应
response = process.stdout.readline()
print(json.loads(response))
```

## API 接口使用

### HTTP API 模式

启动 HTTP API 服务：
```bash
./scihub-mcp api --port 8080
```

#### 健康检查

```bash
curl http://localhost:8080/health
```

响应：
```json
{
  "status": "ok",
  "mirrors_available": 5,
  "cache_size": "128MB"
}
```

#### 下载论文

```bash
curl -X POST http://localhost:8080/fetch \
  -H "Content-Type: application/json" \
  -d '{
    "doi": "10.1038/nature12373",
    "title": "Optional paper title"
  }'
```

响应：
```json
{
  "success": true,
  "message": "文件下载成功",
  "data": {
    "filename": "nature12373.pdf",
    "size": 1024000,
    "mirror_used": "https://sci-hub.se",
    "cached": false,
    "download_link": "/download/nature12373.pdf"
  }
}
```

#### 下载文件

```bash
curl http://localhost:8080/download/nature12373.pdf --output paper.pdf
```

#### 镜像状态

```bash
curl http://localhost:8080/mirrors
```

响应：
```json
{
  "mirrors": {
    "https://sci-hub.ru": {
      "status": "online",
      "response_time": "1.2s"
    },
    "https://sci-hub.se": {
      "status": "online", 
      "response_time": "0.8s"
    }
  },
  "summary": {
    "total": 11,
    "online": 8,
    "offline": 2,
    "slow": 1
  }
}
```

## 故障排除

### 常见问题

1. **下载失败**
   ```bash
   # 检查镜像状态
   ./scihub-mcp status
   
   # 测试特定镜像
   ./scihub-mcp test --mirror "https://sci-hub.se"
   
   # 启用详细日志
   ./scihub-mcp --verbose fetch --doi "10.1038/nature12373"
   ```

2. **代理问题**
   ```bash
   # 测试代理连接
   curl --socks5 127.0.0.1:3080 http://www.google.com
   
   # 不使用代理测试
   ./scihub-mcp fetch --doi "10.1038/nature12373"
   ```

3. **MCP 连接问题**
   ```bash
   # 测试 MCP 服务器启动
   echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | ./scihub-mcp mcp
   
   # 检查权限
   ls -la ./scihub-mcp
   chmod +x ./scihub-mcp
   ```

4. **SSE 模式问题**
   ```bash
   # 检查端口是否被占用
   lsof -i :8080
   
   # 测试 SSE 连接
   curl -v http://localhost:8080/sse
   
   # 检查防火墙设置
   sudo ufw status
   ```

### 调试模式

启用详细日志：
```bash
# 设置日志级别
export LOG_LEVEL=debug
./scihub-mcp fetch --doi "10.1038/nature12373"
```

### 配置验证

验证配置文件：
```bash
# 验证 YAML 语法
./scihub-mcp --config config.yaml --dry-run

# 显示当前配置
./scihub-mcp config show
```

## 高级配置

### 自定义缓存目录

```yaml
download:
  cache_dir: "/custom/cache/path"
  max_retries: 5
  timeout: "120s"
```

### 自定义健康检查

```yaml
health_check:
  interval: "15m"     # 每15分钟检查一次
  timeout: "5s"       # 5秒超时
```

### 多环境配置

创建不同的配置文件：

```bash
# 开发环境
cp config.yaml config-dev.yaml

# 生产环境
cp config.yaml config-prod.yaml

# 使用特定配置
./scihub-mcp --config config-prod.yaml api
```

### 性能优化

1. **并发下载**：
   ```yaml
   download:
     max_concurrent: 3
     timeout: "30s"
   ```

2. **缓存优化**：
   ```yaml
   download:
     cache_dir: "/fast/ssd/cache"
     cache_max_size: "10GB"
   ```

3. **镜像优化**：
   ```yaml
   mirrors:
     - "https://fast-mirror.com"
     - "https://backup-mirror.com"
   
   health_check:
     interval: "5m"    # 更频繁的检查
   ```

## 最佳实践

1. **生产环境部署**：
   - 使用配置文件而非命令行参数
   - 设置适当的缓存目录和大小限制
   - 配置日志轮转
   - 使用进程管理器（如 systemd）

2. **开发环境**：
   - 使用较短的健康检查间隔
   - 启用详细日志
   - 使用本地缓存目录

3. **网络受限环境**：
   - 配置可靠的代理服务器
   - 增加重试次数和超时时间
   - 使用较少但稳定的镜像

4. **高并发场景**：
   - 使用 SSE 传输模式
   - 配置负载均衡
   - 监控系统资源使用 