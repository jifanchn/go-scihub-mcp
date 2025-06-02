# SciHub-MCP Makefile

.PHONY: help build clean install release docker docs lint format check-deps

# 变量定义
BINARY_NAME := scihub-mcp
PACKAGE := github.com/jifanchn/go-scihub-mcp
BUILD_DIR := build
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S%z)

# Go 构建参数
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_TIME)"
GO_BUILD := go build $(LDFLAGS)

# 默认目标
help: ## 显示帮助信息
	@echo "SciHub-MCP 构建工具"
	@echo "==================="
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## 构建二进制文件
	@echo "构建 $(BINARY_NAME)..."
	$(GO_BUILD) -o $(BINARY_NAME) ./cmd/scihub-mcp
	@echo "构建完成: $(BINARY_NAME)"

build-all: ## 交叉编译所有平台
	@echo "交叉编译所有平台..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux amd64
	GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/scihub-mcp
	
	# Linux arm64
	GOOS=linux GOARCH=arm64 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/scihub-mcp
	
	# macOS amd64
	GOOS=darwin GOARCH=amd64 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/scihub-mcp
	
	# macOS arm64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/scihub-mcp
	
	# Windows amd64
	GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/scihub-mcp
	
	# Windows arm64
	GOOS=windows GOARCH=arm64 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-arm64.exe ./cmd/scihub-mcp
	
	@echo "交叉编译完成，文件位于 $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

test: ## 运行 Go 单元测试
	@echo "运行单元测试..."
	go test -v ./...

clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@rm -rf cache/
	@echo "清理完成"

install: build ## 安装到系统路径
	@echo "安装 $(BINARY_NAME) 到 /usr/local/bin/..."
	@sudo cp $(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "安装完成"

uninstall: ## 从系统路径卸载
	@echo "卸载 $(BINARY_NAME)..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "卸载完成"

release: clean build-all ## 创建发布包
	@echo "创建发布包..."
	@mkdir -p $(BUILD_DIR)/release
	
	# 为每个平台创建压缩包
	@cd $(BUILD_DIR) && \
	for binary in $(BINARY_NAME)-*; do \
		if [[ "$$binary" == *".exe" ]]; then \
			platform=$${binary%%.exe}; \
			zip -q release/$$platform.zip $$binary; \
		else \
			platform=$$binary; \
			tar -czf release/$$platform.tar.gz $$binary; \
		fi; \
	done
	
	# 复制配置文件到发布目录
	@cp -r configs $(BUILD_DIR)/release/
	@cp README.md README_cn.md LICENSE $(BUILD_DIR)/release/ 2>/dev/null || true
	
	@echo "发布包创建完成，位于 $(BUILD_DIR)/release/"
	@ls -la $(BUILD_DIR)/release/

docker: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	@docker build -t scihub-mcp:$(VERSION) .
	@docker tag scihub-mcp:$(VERSION) scihub-mcp:latest
	@echo "Docker 镜像构建完成: scihub-mcp:$(VERSION)"

docker-run: docker ## 运行 Docker 容器
	@echo "运行 Docker 容器..."
	@docker run -it --rm -p 8080:8080 scihub-mcp:latest

docs: ## 生成文档
	@echo "生成文档..."
	@go doc -all ./... > docs/api.md 2>/dev/null || echo "Go doc 生成失败，请检查代码"
	@echo "文档生成完成"

lint: ## 运行代码检查
	@echo "运行代码检查..."
	@which golangci-lint >/dev/null 2>&1 || { echo "请先安装 golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; exit 1; }
	@golangci-lint run

format: ## 格式化代码
	@echo "格式化代码..."
	@go fmt ./...
	@goimports -w . 2>/dev/null || echo "goimports 未安装，跳过导入排序"

check-deps: ## 检查和更新依赖
	@echo "检查依赖..."
	@go mod tidy
	@go mod verify
	@echo "依赖检查完成"

vet: ## 运行 go vet
	@echo "运行 go vet..."
	@go vet ./...

security: ## 运行安全检查
	@echo "运行安全检查..."
	@which gosec >/dev/null 2>&1 || { echo "请先安装 gosec: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; exit 1; }
	@gosec ./...

coverage: ## 生成测试覆盖率报告
	@echo "生成测试覆盖率报告..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告生成完成: coverage.html"

dev: build ## 开发模式：构建并启动开发服务器
	@echo "启动开发服务器..."
	@./$(BINARY_NAME) --config configs/config.yaml api

dev-mcp: build ## 开发模式：启动 MCP 服务器
	@echo "启动 MCP 开发服务器..."
	@./$(BINARY_NAME) --config configs/config.yaml mcp

# 性能测试
benchmark: build ## 运行性能测试
	@echo "运行性能测试..."
	@go test -bench=. -benchmem ./... 2>/dev/null || echo "没有找到性能测试"

# 快速开始
quick-start: build ## 快速开始：构建并显示帮助
	@echo "快速开始指南:"
	@echo "=============="
	@echo "1. 配置文件已准备好："
	@echo "   - configs/config.yaml (默认配置)"
	@echo ""
	@echo "2. 运行测试："
	@echo "   make test"
	@echo ""
	@echo "3. 启动服务："
	@echo "   - HTTP API: make dev"
	@echo "   - MCP SSE: make dev-mcp"
	@echo ""
	@echo "4. 下载论文："
	@echo "   ./$(BINARY_NAME) fetch --doi \"10.1038/nature12373\""
	@echo ""
	@./$(BINARY_NAME) --help