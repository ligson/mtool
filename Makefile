.PHONY: all deps build-arm64 build-amd64 package clean help info

# 版本号
VERSION := 0.7.0
BINARY_NAME := mtool
DIST_DIR := dist

# 当前系统信息
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

# 检测当前架构
ifeq ($(UNAME_M), arm64)
  CURRENT_ARCH := arm64
else ifeq ($(UNAME_M), x86_64)
  CURRENT_ARCH := amd64
else
  CURRENT_ARCH := unknown
endif

# 设置 Go 构建标志
GO := go
GOFLAGS := -ldflags="-s -w"

help:
	@echo "mtool Makefile - 多架构编译和打包"
	@echo ""
	@echo "使用方法："
	@echo "  make build-arm64    - 构建 ARM64（Apple Silicon M1-M4）"
	@echo "  make build-amd64    - 构建 x86_64（Intel Mac）"
	@echo "  make package        - 打包所有已编译的架构"
	@echo "  make all            - 编译所有架构并打包"
	@echo "  make clean          - 清理编译产物"
	@echo "  make info           - 显示编译配置信息"
	@echo ""
	@echo "当前系统信息："
	@echo "  OS: $(UNAME_S)"
	@echo "  架构: $(UNAME_M) ($(CURRENT_ARCH))"
	@echo ""
	@echo "示例流程："
	@echo "  # 在 Apple Silicon Mac 上："
	@echo "  make build-arm64"
	@echo ""
	@echo "  # 在 Intel Mac 上："
	@echo "  make build-amd64"
	@echo ""
	@echo "  # 打包所有架构："
	@echo "  make package"

# 全部编译并打包
all: build-arm64 build-amd64 package
	@echo ""
	@echo "✓ 完整编译和打包完成"

# 下载 Go 模块依赖（保证 go.sum 与构建缓存完整）
deps:
	@echo "📥 正在下载 Go 模块依赖..."
	$(GO) mod download
	@echo "✓ 依赖准备完成"

# 创建 dist 目录结构
$(DIST_DIR):
	@mkdir -p $(DIST_DIR)/darwin-arm64
	@mkdir -p $(DIST_DIR)/darwin-amd64

# 构建 ARM64（Apple Silicon）
build-arm64: deps $(DIST_DIR)
	@echo "📦 正在编译 ARM64（Apple Silicon M1-M4）..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 $(GO) build $(GOFLAGS) -o $(DIST_DIR)/darwin-arm64/$(BINARY_NAME) .
	@chmod +x $(DIST_DIR)/darwin-arm64/$(BINARY_NAME)
	@echo "✓ ARM64 编译完成: $(DIST_DIR)/darwin-arm64/$(BINARY_NAME)"
	@ls -lh $(DIST_DIR)/darwin-arm64/$(BINARY_NAME)

# 构建 x86_64（Intel）
build-amd64: deps $(DIST_DIR)
	@echo "📦 正在编译 x86_64（Intel Mac）..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 $(GO) build $(GOFLAGS) -o $(DIST_DIR)/darwin-amd64/$(BINARY_NAME) .
	@chmod +x $(DIST_DIR)/darwin-amd64/$(BINARY_NAME)
	@echo "✓ x86_64 编译完成: $(DIST_DIR)/darwin-amd64/$(BINARY_NAME)"
	@ls -lh $(DIST_DIR)/darwin-amd64/$(BINARY_NAME)

# 打包所有��编译的架构
package: $(DIST_DIR)
	@echo "📦 正在打包所有架构..."
	@if [ -f "$(DIST_DIR)/darwin-arm64/$(BINARY_NAME)" ]; then \
		cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-v$(VERSION)-darwin-arm64.tar.gz darwin-arm64/$(BINARY_NAME) && cd .. ; \
		echo "✓ ARM64 压缩包: $(DIST_DIR)/$(BINARY_NAME)-v$(VERSION)-darwin-arm64.tar.gz" ; \
		ls -lh $(DIST_DIR)/$(BINARY_NAME)-v$(VERSION)-darwin-arm64.tar.gz ; \
	fi
	@if [ -f "$(DIST_DIR)/darwin-amd64/$(BINARY_NAME)" ]; then \
		cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-v$(VERSION)-darwin-amd64.tar.gz darwin-amd64/$(BINARY_NAME) && cd .. ; \
		echo "✓ x86_64 压缩包: $(DIST_DIR)/$(BINARY_NAME)-v$(VERSION)-darwin-amd64.tar.gz" ; \
		ls -lh $(DIST_DIR)/$(BINARY_NAME)-v$(VERSION)-darwin-amd64.tar.gz ; \
	fi
	@echo ""
	@echo "✓ 打包完成"
	@echo ""
	@echo "编译产物:"
	@ls -lh $(DIST_DIR)/ | grep -E "(mtool|\.tar\.gz)"

# 清理
clean:
	@echo "🧹 正在清理编译产物..."
	@rm -rf $(DIST_DIR)
	@echo "✓ 清理完成"

# 显示配置信息
info:
	@echo "mtool 编译配置"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "版本号：        $(VERSION)"
	@echo "二进制名：      $(BINARY_NAME)"
	@echo "输出目录：      $(DIST_DIR)/"
	@echo "Go 编译器：     $(GO)"
	@echo "系统：          $(UNAME_S)"
	@echo "当前架构：      $(UNAME_M) ($(CURRENT_ARCH))"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "可用目标："
	@echo "  make build-arm64     在 ARM64 Mac 上编译 ARM64 版本"
	@echo "  make build-amd64     在 Intel Mac 上编译 x86_64 版本"
	@echo "  make package         打包所有已编译的架构"
	@echo "  make all             编译所有架构并打包"
	@echo "  make clean           清理所有编译产物"

.DEFAULT_GOAL := help
