#!/bin/bash

# mtool 版本发布脚本 - 自动更新版本号
# 用法: ./release.sh 0.3.0

set -e

VERSION="${1:-}"

if [ -z "$VERSION" ]; then
    echo "❌ 错误：未指定版本号"
    echo "用法: $0 <version>"
    echo "示例: $0 0.3.0"
    exit 1
fi

# 移除 v 前缀（如果有）
VERSION_NUM="${VERSION#v}"

echo "📦 准备发布版本 v$VERSION_NUM"
echo ""

# 1. 更新 Makefile
echo "📝 更新 Makefile..."
sed -i '' "s/^VERSION := .*/VERSION := $VERSION_NUM/" Makefile
if grep -q "VERSION := $VERSION_NUM" Makefile; then
    echo "✓ Makefile 已更新为 v$VERSION_NUM"
else
    echo "❌ 更新 Makefile 失败"
    exit 1
fi

# 2. 更新 main.go
echo "📝 更新 main.go..."
sed -i '' "s/version = \"[^\"]*\"/version = \"$VERSION_NUM\"/" main.go
if grep -q "version = \"$VERSION_NUM\"" main.go; then
    echo "✓ main.go 已更新为 v$VERSION_NUM"
else
    echo "❌ 更新 main.go 失败"
    exit 1
fi

# 3. 在 CHANGELOG.md 添加新版本块
echo "📝 更新 CHANGELOG.md..."
TODAY=$(date +%Y-%m-%d)

# 创建临时文件，在 Unreleased 下方添加新版本
awk -v new_ver="$VERSION_NUM" -v today="$TODAY" '
/^## \[Unreleased\]/ {
    print
    print ""
    print "## [" new_ver "] - " today
    print ""
    print "### Added"
    print "-"
    print ""
    print "### Changed"
    print "-"
    print ""
    print "### Fixed"
    print "-"
    print ""
    next
}
1
' CHANGELOG.md > CHANGELOG.md.tmp

mv CHANGELOG.md.tmp CHANGELOG.md
echo "✓ CHANGELOG.md 已添加新版本块"

echo ""
echo "✅ 文件更新完成！"
echo ""
echo "📋 接下来的步骤："
echo ""
echo "  1️⃣  编辑 CHANGELOG.md 添加具体的改动内容"
echo "     vim CHANGELOG.md"
echo ""
echo "  2️⃣  提交代码"
echo "     git add Makefile main.go CHANGELOG.md"
echo "     git commit -m \"chore: bump version to v$VERSION_NUM\""
echo "     git push origin main"
echo ""
echo "  3️⃣  GitHub Actions 会自动："
echo "     - 编译 ARM64 和 x86_64"
echo "     - 从 CHANGELOG.md 读取版本内容"
echo "     - 创建 Release v$VERSION_NUM"
echo ""
