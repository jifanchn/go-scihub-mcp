# 更新日志

## [2.0.0] - 2024-12-28

### 重大变更 🚨
- **移除 stdio 模式支持**：项目现在只支持 SSE (Server-Sent Events) 模式的 MCP 通信
- **简化架构**：移除冗余的 HTTP API 服务器实现，统一使用 MCP 协议服务器

### 新增功能 ✨
- **Docker 支持**：添加完整的 Docker 和 docker-compose 支持
- **健康检查**：Docker 容器包含内置健康检查
- **代理配置优化**：改进代理设置，支持容器内外网络环境

### 架构清理 🧹
- 删除 `internal/mcp` 包（HTTP REST API 服务器）
- 保留 `internal/mcpserver` 包（MCP 协议 SSE 服务器）
- 移除所有测试脚本和 stdio 模式相关代码
- 清理配置文件，只保留 `configs/config.yaml`

### 配置变更 ⚙️
- MCP 传输模式：默认且仅支持 `sse`
- 配置验证：只接受 `sse` 传输模式
- Docker 部署：默认端口从 8080 改为 8088

### 删除的文件 📁
- `scripts/` 目录及所有测试脚本
- `configs/cursor_mcp_*.json` 配置示例
- `configs/claude_desktop_config.json`
- `configs/config-sse.yaml`
- `internal/mcp/server.go`
- `test_stdio.sh`

### 更新的文档 📚
- 完全重写 README.md 和 README_cn.md
- 添加 Docker 部署说明
- 更新 MCP 配置示例
- 清理过时的 stdio 模式文档

### 命令行变更 💻
- 移除 `test` 命令
- `api` 命令现在启动 SSE 模式的 MCP 服务器
- `mcp` 命令专门用于 MCP 协议服务器
- 保留 `fetch` 和 `status` 命令

### Docker 支持 🐳
- 新增 `Dockerfile` 多阶段构建
- 新增 `docker-compose.yml` 编排文件
- 新增 `.dockerignore` 文件
- 支持代理配置和健康检查

### 破坏性变更 ⚠️
- **移除 stdio 模式**：如果您在使用 stdio 模式的 MCP 连接，需要迁移到 SSE 模式
- **端口变更**：Docker 部署默认使用端口 8088 而非 8080
- **配置格式**：移除对 `transport: "stdio"` 的支持

### 迁移指南 🔄
1. 更新 MCP 客户端配置使用 SSE 端点：`http://localhost:8080/sse`
2. 如果使用 Docker，注意端口变更为 8088
3. 移除任何 stdio 模式的配置引用
4. 更新配置文件中的 `transport` 字段为 `"sse"`

### 技术改进 🔧
- 简化代码结构，提高可维护性
- 统一服务器实现，减少代码重复
- 改进错误处理和日志记录
- 优化构建流程和依赖管理 