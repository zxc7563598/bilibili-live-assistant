# 项目名称
APP_NAME := GoAdminKit

# Go 参数
GO := go
GO_BUILD := $(GO) build
GO_RUN := $(GO) run

# 目录
CMD_DIR := ./cmd/server
BUILD_DIR := ./bin

# 版本信息
MODULE := $(shell go list -m)

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags "\
-X $(MODULE)/internal/version.Version=$(VERSION) \
-X $(MODULE)/internal/version.Commit=$(COMMIT) \
-X $(MODULE)/internal/version.BuildTime=$(BUILD_TIME) \
-s -w"

.DEFAULT_GOAL := help

## 帮助
help:
	@echo ""
	@echo "GoAdminKit Makefile"
	@echo ""
	@echo "开发命令:"
	@echo "  make dev           启动开发环境 (Go + Web)"
	@echo "  make dev-go        仅启动 Go 服务"
	@echo "  make dev-web       仅启动 Web dev server"
	@echo "  make swagger       生成 Swagger 文档"
	@echo ""
	@echo "构建命令:"
	@echo "  make build         构建当前平台"
	@echo "  make build-web     构建 Web 前端"
	@echo ""
	@echo "打包命令:"
	@echo "  make build-linux       构建 Linux"
	@echo "  make build-macos       构建 macOS Intel"
	@echo "  make build-macos-arm   构建 macOS ARM"
	@echo "  make build-windows     构建 Windows"
	@echo "  make release           构建所有平台"
	@echo ""
	@echo "其他:"
	@echo "  make clean         清理构建文件"
	@echo ""

dev:
	@echo "启动开发环境..."
	@echo "Web: http://localhost:5173"
	@echo "API: http://localhost:9000"
	@make -j2 dev-go dev-web

dev-go:
	@echo "启动 Go 服务..."
	$(GO_RUN) $(CMD_DIR)/main.go

dev-web:
	@echo "启动 Web dev server..."
	@cd ./web && npm install && npm run dev

swagger:
	@command -v swag >/dev/null 2>&1 || { \
		echo "❌ 未安装 swag，请执行: go install github.com/swaggo/swag/cmd/swag@latest"; \
		exit 1; \
	}
	@echo "生成 Swagger 文档..."
	@swag init -g cmd/server/main.go --parseInternal

build-web:
	@echo "构建 Web 站点页面..."
	@command -v npm >/dev/null 2>&1 || { \
	echo "❌ 未检测到 npm，请先安装 Node.js (https://nodejs.org)"; \
	exit 1; \
	}
	@echo "检测到 npm，开始构建 Web 站点..."
	@cd ./web && npm install && npm run build
	@echo "同步 dist 到 internal/webui/dist"
	@rm -rf ./internal/webui/dist
	@mkdir -p ./internal/webui
	@cp -R ./web/dist ./internal/webui/dist

build: build-web
	@echo "构建 $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)

build-linux: clean build-web swagger
	@echo "构建 Linux 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(CMD_DIR)

build-macos: clean build-web swagger
	@echo "构建 macOS Intel 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_DIR)

build-macos-arm: clean build-web swagger
	@echo "构建 macOS ARM 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(CMD_DIR)

build-windows: clean build-web swagger
	@echo "构建 Windows 版本..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GO_BUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_DIR)

release:
	@echo "并行构建所有平台..."
	@mkdir -p $(BUILD_DIR)
	@$(MAKE) build-linux &
	@$(MAKE) build-macos &
	@$(MAKE) build-macos-arm &
	@$(MAKE) build-windows &
	@wait
	@echo ""
	@echo "构建完成，输出目录:"
	@echo "  $(BUILD_DIR)"

test:
	@echo "⚠️  no tests configured"

clean:
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -rf ./internal/webui/dist
	@rm -rf ./docs/docs.go ./docs/swagger.json ./docs/swagger.yaml
	@echo "清理完成"