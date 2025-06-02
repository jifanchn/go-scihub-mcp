# go-scihub-mcp

[中文文档](README_cn.md) | [English](README.md)

一个用 Go 语言编写的 Sci-Hub 镜像管理和文件下载工具，支持 MCP (Model Context Protocol) 兼容的 API 服务。

## 功能特性

- 📋 维护可用的 Sci-Hub 镜像列表
- 🔄 自动检测和更新镜像可用性
- 🌐 支持 SOCKS5 代理配置
- 📁 文件下载和缓存
- 🔗 MCP 兼容的 SSE API 服务
- ⚙️ 清晰优先级的灵活命令行配置
- 🐳 Docker 支持和 docker-compose

## 安装

### 从源码构建

```bash
git clone https://github.com/jifanchn/go-scihub-mcp.git
cd go-scihub-mcp
go mod tidy
go build -o scihub-mcp ./cmd/scihub-mcp
```

### Docker

```bash
# 使用 docker-compose（推荐）
docker-compose up -d

# 或手动构建运行
docker build -t scihub-mcp .
docker run -d -p 8088:8088 --name scihub-mcp scihub-mcp
```

### 二进制下载

从 [Releases](https://github.com/jifanchn/go-scihub-mcp/releases) 页面下载对应平台的二进制文件。

## 配置

### 配置优先级

配置遵循清晰的优先级系统：
**命令行参数 > 配置文件 > 默认值**

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
  interval: "30m"
  timeout: "10s"

# MCP 服务配置
mcp:
  port: 8080
  host: "0.0.0.0"
  transport: "sse"
  sse_path: "/sse"
  
# 下载配置
download:
  cache_dir: "./cache"
  max_retries: 3
  timeout: "60s"
```

## 使用方法

### 1. 基本运行

```bash
# 使用默认配置运行（HTTP API 服务）
./scihub-mcp

# 启用代理运行
./scihub-mcp --proxy-enabled --proxy-host 127.0.0.1 --proxy-port 3080

# 使用自定义配置文件
./scihub-mcp --config /path/to/config.yaml
```

### 2. 文件下载命令

```bash
# 通过 DOI 下载文件
./scihub-mcp fetch --doi "10.1038/nature12373"

# 使用代理下载
./scihub-mcp --proxy-enabled fetch --doi "10.1038/nature12373"

# 指定自定义输出路径
./scihub-mcp fetch --doi "10.1038/nature12373" --output "./papers/nature12373.pdf"

# 通过 URL 下载文件
./scihub-mcp fetch --url "https://www.nature.com/articles/nature12373"
```

### 3. MCP 协议服务模式 (SSE)

```bash
# 启动 MCP 协议服务器
./scihub-mcp mcp

# 使用自定义端口
./scihub-mcp --mcp-port 9090 mcp

# 启用代理
./scihub-mcp --proxy-enabled --proxy-host 127.0.0.1 --proxy-port 3080 mcp
```

**SSE 模式端点：**
- SSE 流：`http://localhost:8080/sse`
- 消息端点：`http://localhost:8080/message`  
- 健康检查：`http://localhost:8080/health`

### 4. 镜像状态检查

```bash
# 检查所有镜像状态
./scihub-mcp status

# 启用代理检查
./scihub-mcp --proxy-enabled status
```

### 5. Docker 部署

```bash
# 使用 docker-compose（推荐）
docker-compose up -d

# 检查状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

Docker 服务将：
- 在端口 8088 上运行（映射到主机）
- 使用主机的 `127.0.0.1:3080` SOCKS5 代理
- 在 `./cache` 目录中持久化缓存数据
- 除非停止否则自动重启

## MCP 服务器配置

### Cursor AI 配置

创建或编辑您的 MCP 配置文件（通常为 `~/.cursor/mcp_servers.json`）：

```json
{
  "mcpServers": {
    "scihub-mcp": {
      "description": "SciHub MCP Server - SSE Mode",
      "url": "http://localhost:8080/sse"
    }
  }
}
```

对于 Docker 部署：
```json
{
  "mcpServers": {
    "scihub-mcp": {
      "description": "SciHub MCP Server - Docker",  
      "url": "http://localhost:8088/sse"
    }
  }
}
```

### Claude Desktop 配置

配置文件位置：
- macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: `%APPDATA%\Claude\claude_desktop_config.json`
- Linux: `~/.config/Claude/claude_desktop_config.json`

配置内容：

```json
{
  "mcpServers": {
    "scihub-mcp": {
      "url": "http://localhost:8080/sse"
    }
  }
}
```

## MCP 工具和资源

### 可用工具

1. **download_paper**: 下载科学论文 PDF 文件
2. **check_mirror_status**: 检查 Sci-Hub 镜像的可用性状态  
3. **test_mirror**: 测试特定 Sci-Hub 镜像的可用性
4. **list_available_mirrors**: 获取当前可用的 Sci-Hub 镜像列表

### 可用资源

1. **scihub://cache**: 缓存的论文文件列表
2. **scihub://mirrors/status**: 所有 Sci-Hub 镜像的实时状态
3. **scihub://papers/{filename}**: 访问缓存的论文 PDF 文件

## 开发

### 构建

```bash
# 构建
make build

# 交叉编译
make build-all

# Docker 构建
make docker
```

### 测试

```bash
# 运行测试
make test

# 覆盖率测试
make coverage
```

## 许可证

此项目根据 MIT 许可证进行许可。

## 免责声明

此工具仅用于教育和研究目的。用户有责任确保遵守其管辖区域内的适用法律法规。 