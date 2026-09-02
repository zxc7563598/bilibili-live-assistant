# 项目名称
APP_NAME := BiliLiveAssistant

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

# 链接参数：剥离符号表/调试信息(-s -w)、去除 buildid、注入版本信息
LDFLAGS := -s -w -buildid= \
	-X $(MODULE)/internal/version.Version=$(VERSION) \
	-X $(MODULE)/internal/version.Commit=$(COMMIT) \
	-X $(MODULE)/internal/version.BuildTime=$(BUILD_TIME)

# 构建参数：关闭 cgo 生成纯静态二进制 + 去除源码路径(可复现构建)
CGO_ENABLED := 0
GO_BUILD_FLAGS := -trimpath -ldflags "$(LDFLAGS)"

.DEFAULT_GOAL := help

## 帮助
help:
	@echo ""
	@echo "BiliLiveAssistant Makefile"
	@echo ""
	@echo "开发命令:"
	@echo "  make dev           启动开发环境 (Go + Web)"
	@echo "  make dev-go        仅启动 Go 服务"
	@echo "  make dev-web       仅启动 Web dev server"
	@echo "  make dev-shop      仅启动 Shop dev server"
	@echo "  make seed-products 填充商城商品测试数据（手动，不启动服务）"
	@echo "  make swagger       生成 Swagger 文档"
	@echo ""
	@echo "构建命令:"
	@echo "  make build         构建当前平台 (release 优化)"
	@echo "  make build-web     构建 Web 前端"
	@echo "  make build-shop    构建 Shop 前端"
	@echo ""
	@echo "打包命令:"
	@echo "  make build-linux-amd64    构建 linux/amd64"
	@echo "  make build-linux-arm64    构建 linux/arm64"
	@echo "  make build-darwin-amd64   构建 darwin/amd64 (Intel)"
	@echo "  make build-darwin-arm64   构建 darwin/arm64 (Apple Silicon)"
	@echo "  make build-windows-amd64  构建 windows/amd64"
	@echo "  make release              构建全部平台 (并行 + 校验和)"
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
	GIN_MODE=debug $(GO_RUN) $(CMD_DIR)/main.go

dev-web:
	@echo "启动 Web dev server..."
	@cd ./web && npm install && npm run dev

dev-shop:
	@echo "启动 Shop dev server..."
	@cd ./shop && npm install && npm run dev

seed-products:
	@echo "填充商城商品测试数据..."
	$(GO_RUN) $(CMD_DIR)/main.go -seed-products

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

build-shop:
	@echo "构建 Shop 站点页面..."
	@command -v npm >/dev/null 2>&1 || { \
		echo "❌ 未检测到 npm，请先安装 Node.js (https://nodejs.org)"; \
		exit 1; \
	}
	@echo "检测到 npm，开始构建 Shop 站点..."
	@cd ./shop && npm install && npm run build
	@echo "同步 dist 到 internal/webui/shop"
	@rm -rf ./internal/webui/shop
	@mkdir -p ./internal/webui
	@cp -R ./shop/dist ./internal/webui/shop

# 前端资源 + Swagger 文档构建标记：并发 release 时只构建一次
PREPARE_STAMP := $(BUILD_DIR)/.prepared

$(PREPARE_STAMP):
	@$(MAKE) build-web build-shop
	@mkdir -p $(BUILD_DIR)
	@touch $(PREPARE_STAMP)

# 构建当前平台（release 优化，始终刷新前端）
build: build-web build-shop swagger
	@echo "构建 $(APP_NAME)（当前平台）..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO_BUILD) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)

build-linux-amd64: $(PREPARE_STAMP)
	@echo "构建 linux/amd64 ..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GO_BUILD) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(CMD_DIR)

build-linux-arm64: $(PREPARE_STAMP)
	@echo "构建 linux/arm64 ..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) $(GO_BUILD) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(CMD_DIR)

build-darwin-amd64: $(PREPARE_STAMP)
	@echo "构建 darwin/amd64 ..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GO_BUILD) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_DIR)

build-darwin-arm64: $(PREPARE_STAMP)
	@echo "构建 darwin/arm64 ..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) $(GO_BUILD) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(CMD_DIR)

build-windows-amd64: $(PREPARE_STAMP)
	@echo "构建 windows/amd64 ..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GO_BUILD) $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_DIR)

release: clean
	@echo "并行构建全部平台..."
	@$(MAKE) -j5 build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64
	@echo ""
	@echo "生成 SHA256 校验和..."
	@cd $(BUILD_DIR) && shasum -a 256 $(APP_NAME)-* > SHA256SUMS
	@echo ""
	@echo "构建完成，输出目录: $(BUILD_DIR)"
	@ls -lh $(BUILD_DIR)

test:
	@echo "⚠️  no tests configured"

clean:
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -rf ./internal/webui/dist
	@echo "清理完成"
