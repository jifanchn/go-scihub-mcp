# SciHub-MCP Makefile

# 变量定义
APP_NAME := scihub-mcp
MAIN_FILE := cmd/scihub-mcp/main.go
BUILD_DIR := build
VERSION := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell date +%Y-%m-%d_%H:%M:%S)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Go 相关变量
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# 默认目标
.PHONY: all
all: build

# 构建
.PHONY: build
build:
	@echo "构建 $(APP_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "构建完成: $(BUILD_DIR)/$(APP_NAME)"

# 交叉编译
.PHONY: build-all
build-all: build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	@echo "构建 Linux 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_FILE)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(MAIN_FILE)

.PHONY: build-darwin
build-darwin:
	@echo "构建 macOS 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_FILE)

.PHONY: build-windows
build-windows:
	@echo "构建 Windows 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_FILE)

# 测试
.PHONY: test
test:
	@echo "运行测试..."
	go test -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "运行测试并生成覆盖率报告..."
	go test -v -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 代码检查
.PHONY: lint
lint:
	@echo "运行代码检查..."
	golangci-lint run

.PHONY: fmt
fmt:
	@echo "格式化代码..."
	go fmt ./...

.PHONY: vet
vet:
	@echo "运行 go vet..."
	go vet ./...

# 依赖管理
.PHONY: mod-tidy
mod-tidy:
	@echo "整理模块依赖..."
	go mod tidy

.PHONY: mod-download
mod-download:
	@echo "下载模块依赖..."
	go mod download

# 安装
.PHONY: install
install:
	@echo "安装 $(APP_NAME)..."
	go install $(LDFLAGS) $(MAIN_FILE)

# 清理
.PHONY: clean
clean:
	@echo "清理构建文件..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# 创建发布包
.PHONY: release
release: clean build-all
	@echo "创建发布包..."
	@mkdir -p $(BUILD_DIR)/release
	@for binary in $(shell ls $(BUILD_DIR)/$(APP_NAME)-*); do \
		base=$$(basename $$binary); \
		dir=$(BUILD_DIR)/release/$$base; \
		mkdir -p $$dir; \
		cp $$binary $$dir/$(APP_NAME)$$(echo $$base | grep -o '\.exe$$' || true); \
		cp README.md $$dir/; \
		cp configs/config.yaml $$dir/config.yaml.example; \
		cd $(BUILD_DIR)/release && tar -czf $$base.tar.gz $$base; \
		cd ../..; \
	done
	@echo "发布包已创建在 $(BUILD_DIR)/release/ 目录"

# 运行
.PHONY: run
run: build
	./$(BUILD_DIR)/$(APP_NAME)

.PHONY: run-dev
run-dev:
	@echo "开发模式运行..."
	go run $(MAIN_FILE) --config configs/config.yaml

.PHONY: run-proxy
run-proxy:
	@echo "使用代理运行..."
	go run $(MAIN_FILE) --config configs/config.yaml --proxy-enabled --proxy-host 127.0.0.1 --proxy-port 3080

# 帮助
.PHONY: help
help:
	@echo "可用命令:"
	@echo "  build         构建应用程序"
	@echo "  build-all     交叉编译所有平台"
	@echo "  test          运行测试"
	@echo "  test-coverage 运行测试并生成覆盖率报告"
	@echo "  lint          运行代码检查"
	@echo "  fmt           格式化代码"
	@echo "  vet           运行 go vet"
	@echo "  mod-tidy      整理模块依赖"
	@echo "  install       安装到 GOPATH"
	@echo "  clean         清理构建文件"
	@echo "  release       创建发布包"
	@echo "  run           构建并运行"
	@echo "  run-dev       开发模式运行"
	@echo "  run-proxy     使用代理运行"
	@echo "  help          显示此帮助信息" 