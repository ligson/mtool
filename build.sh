#!/bin/bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 脚本信息
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_NAME="mtool"
VERSION="0.1.0"

# 检测系统架构
detect_arch() {
  local arch=$(uname -m)
  case $arch in
    arm64)
      echo "arm64"
      ;;
    x86_64)
      echo "amd64"
      ;;
    *)
      echo "unknown"
      ;;
  esac
}

# 打印横幅
print_banner() {
  echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
  echo -e "${BLUE}  ${PROJECT_NAME} 编译脚本 v${VERSION}${NC}"
  echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# 打印使用说明
print_usage() {
  cat << EOF
使用方法：
  ./build.sh [选项]

选项：
  all              编译所有架构并打包（默认）
  current          仅编译当前架构
  arm64            编译 ARM64（Apple Silicon M1-M4）
  amd64            编译 x86_64（Intel Mac）
  package          打包所有已编译的二进制文件
  clean            清理编译产物
  info             显示系统和编译信息
  help             显示此帮助信息

示例：
  ./build.sh                    # 编译所有架构并打包
  ./build.sh current            # 编译当前架构
  ./build.sh arm64              # 在 ARM64 Mac 上编译
  ./build.sh amd64              # 在 Intel Mac 上编译
  ./build.sh package            # 打包所有已编译版本
  ./build.sh clean              # 清理编译产物

多架构编译流程：
  # 在 Apple Silicon Mac 上运行
  ./build.sh arm64

  # 在 Intel Mac 上运行
  ./build.sh amd64

  # 两个架构都编译完成后，打包所有版本
  ./build.sh package

EOF
}

# 显示系统信息
print_system_info() {
  local arch=$(detect_arch)
  local os=$(uname -s)
  local arch_name="未知"

  case $arch in
    arm64)
      arch_name="ARM64 (Apple Silicon M1-M4)"
      ;;
    amd64)
      arch_name="x86_64 (Intel Mac)"
      ;;
  esac

  echo -e "${GREEN}系统信息：${NC}"
  echo "  操作系统：$os"
  echo "  架构：$arch_name"
  echo "  Go 版本：$(go version | awk '{print $3}')"
  echo ""
}

# 验证依赖
check_dependencies() {
  echo -e "${BLUE}→ 检查依赖...${NC}"

  if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go 未安装，请访问 https://golang.org/dl${NC}"
    exit 1
  fi

  if ! command -v make &> /dev/null; then
    echo -e "${RED}✗ make 未安装${NC}"
    exit 1
  fi

  echo -e "${GREEN}✓ 依赖检查完成${NC}"
}

# 主函数
main() {
  print_banner
  echo ""

  # 检查依赖
  check_dependencies
  echo ""

  # 显示系统信息
  print_system_info

  # 处理命令行参数
  local cmd="${1:-all}"

  case "$cmd" in
    all)
      echo -e "${GREEN}→ 编译所有架构并打包${NC}"
      cd "$SCRIPT_DIR"
      make build-arm64 || true
      make build-amd64 || true
      make package
      ;;
    current)
      local arch=$(detect_arch)
      if [ "$arch" = "arm64" ]; then
        echo -e "${GREEN}→ 编译当前架构（ARM64）${NC}"
        cd "$SCRIPT_DIR"
        make build-arm64
      elif [ "$arch" = "amd64" ]; then
        echo -e "${GREEN}→ 编译当前架构（x86_64）${NC}"
        cd "$SCRIPT_DIR"
        make build-amd64
      else
        echo -e "${RED}✗ 不支持的架构：$arch${NC}"
        exit 1
      fi
      ;;
    arm64)
      echo -e "${GREEN}→ 编译 ARM64 版本${NC}"
      cd "$SCRIPT_DIR"
      make build-arm64
      ;;
    amd64)
      echo -e "${GREEN}→ 编译 x86_64 版本${NC}"
      cd "$SCRIPT_DIR"
      make build-amd64
      ;;
    package)
      echo -e "${GREEN}→ 打包所有已编译的版本${NC}"
      cd "$SCRIPT_DIR"
      make package
      ;;
    clean)
      echo -e "${YELLOW}→ 清理编译产物${NC}"
      cd "$SCRIPT_DIR"
      make clean
      echo -e "${GREEN}✓ 清理完成${NC}"
      ;;
    info)
      cd "$SCRIPT_DIR"
      make info
      ;;
    help)
      print_usage
      ;;
    *)
      echo -e "${RED}✗ 未知命令：$cmd${NC}"
      echo ""
      print_usage
      exit 1
      ;;
  esac

  echo ""
  echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# 如果直接调用此脚本
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
  main "$@"
fi
