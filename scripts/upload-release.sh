#!/bin/bash

# Upload compiled binaries to GitHub Release
# 用法: ./scripts/upload-release.sh v0.1.0

set -e

VERSION="${1:-}"

if [ -z "$VERSION" ]; then
    echo "❌ 错误：未指定版本"
    echo "用法: $0 <version>"
    echo "示例: $0 v0.1.0"
    exit 1
fi

# 验证 tag 是否存在
if ! git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo "❌ 错误：tag '$VERSION' 不存在"
    exit 1
fi

# 检查二进制文件是否存在
BINARY_ARM64="dist/darwin-arm64/mtool"
BINARY_AMD64="dist/darwin-amd64/mtool"
PACKAGE_ARM64="dist/mtool-v${VERSION#v}-darwin-arm64.tar.gz"
PACKAGE_AMD64="dist/mtool-v${VERSION#v}-darwin-amd64.tar.gz"

if [ ! -f "$PACKAGE_ARM64" ] && [ ! -f "$PACKAGE_AMD64" ]; then
    echo "❌ 错误：找不到编译产物"
    echo "   预期文件："
    echo "   - $PACKAGE_ARM64"
    echo "   - $PACKAGE_AMD64"
    echo ""
    echo "请先运行: make build-arm64（在 ARM64 Mac 上）"
    echo "      或: make build-amd64（在 Intel Mac 上）"
    echo "      然后: make package"
    exit 1
fi

echo "📤 准备上传 $VERSION 到 GitHub Release..."
echo ""

# 使用 gh CLI 上传文件到 release
FILES=()
if [ -f "$PACKAGE_ARM64" ]; then
    FILES+=("$PACKAGE_ARM64")
    echo "✓ 找到 ARM64 包: $PACKAGE_ARM64"
fi
if [ -f "$PACKAGE_AMD64" ]; then
    FILES+=("$PACKAGE_AMD64")
    echo "✓ 找到 amd64 包: $PACKAGE_AMD64"
fi

echo ""
echo "上传文件到 Release $VERSION..."
gh release upload "$VERSION" "${FILES[@]}" --clobber

echo ""
echo "✅ 上传完成！"
echo "查看 Release: https://github.com/$(gh repo view --json nameWithOwner -q)/releases/tag/$VERSION"
