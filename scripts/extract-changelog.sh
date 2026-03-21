#!/bin/bash

# 从 CHANGELOG.md 中提取指定版本的变更内容
# 用法: ./scripts/extract-changelog.sh v0.1.0

VERSION="${1:-}"
VERSION_NUM="${VERSION#v}"

# 如果没有指定版本，使用 Makefile 中的版本
if [ -z "$VERSION" ]; then
    if [ -f "Makefile" ]; then
        VERSION_NUM=$(grep "VERSION := " Makefile | awk '{print $3}')
    fi
fi

if [ -z "$VERSION_NUM" ]; then
    echo "❌ 错误：无法确定版本号" >&2
    exit 1
fi

# 从 CHANGELOG.md 提取内容
if [ ! -f "CHANGELOG.md" ]; then
    echo "❌ 错误：找不到 CHANGELOG.md" >&2
    exit 1
fi

# 提取版本内容
in_section=0
output=""

while IFS= read -r line; do
    # 检查是否是目标版本的开始
    if echo "$line" | grep -q "^## \[$VERSION_NUM\]"; then
        in_section=1
        continue
    fi

    # 如果已经在目标版本中，检查是否到达下一个版本
    if [ $in_section -eq 1 ]; then
        if echo "$line" | grep -q "^## "; then
            # 到达下一个版本，停止提取
            break
        fi
        # 跳过空行但保留内容行
        if [ -n "$line" ]; then
            output="$output$line"$'\n'
        fi
    fi
done < "CHANGELOG.md"

# 输出内容
if [ -z "$output" ]; then
    echo "❌ 错误：在 CHANGELOG.md 中找不到版本 $VERSION_NUM" >&2
    exit 1
fi

echo "$output"
