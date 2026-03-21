# mtool 编译指南

## 架构设计

mtool 的编译系统包含两个核心部分：

- **Makefile**：处理所有编译和打包逻辑
- **build.sh**：提供友好的 CLI 界面，调用 Makefile 目标

## 快速开始

### 编译当前架构
```bash
# 使用 Makefile
make build-arm64    # 在 ARM64 Mac 上
make build-amd64    # 在 Intel Mac 上

# 或使用 build.sh
./build.sh current
```

### 打包
```bash
make package
# 或
./build.sh package
```

### 清理
```bash
make clean
# 或
./build.sh clean
```

## Makefile 目标

### `make build-arm64`
在 ARM64 (Apple Silicon M1-M4) Mac 上编译。

```bash
make build-arm64
# 输出：dist/darwin-arm64/mtool (2.2MB)
```

**前置条件**：
- 必须在 ARM64 Mac 上运行
- Go 1.21+
- 启用 CGo（IOKit 需要）

### `make build-amd64`
在 x86_64 (Intel) Mac 上编译。

```bash
make build-amd64
# 输出：dist/darwin-amd64/mtool (2.5MB)
```

**前置条件**：
- 必须在 Intel Mac 上运行
- Go 1.21+
- 启用 CGo（IOKit 需要）

### `make package`
打包所有已编译的二进制文件为 tar.gz。

```bash
make package
# 输出：
# dist/mtool-v0.1.0-darwin-arm64.tar.gz (940KB)
# dist/mtool-v0.1.0-darwin-amd64.tar.gz (1.0MB)
```

**说明**：
- 自动检测 dist/ 目录中存在的架构
- 为每个架构生成压缩包
- 压缩包包含相对路径以便提取

### `make all`
编译所有架构并打包。

```bash
make all
# 依次执行：build-arm64, build-amd64, package
```

**说明**：
- 如果某个架构编译失败，继续进行下一个
- 最后执行 package 命令
- 适合 CI/CD 环境

### `make clean`
删除所有编译产物。

```bash
make clean
# 删除：dist/ 目录及所有内容
```

### `make info`
显示编译配置信息。

```bash
make info
# 显示版本、架构、系统信息等
```

### `make help`
显示帮助信息。

```bash
make help
# 显示所有可用目标和使用说明
```

## build.sh 脚本

build.sh 是 Makefile 的便捷包装，提供更友好的界面和流程引导。

```bash
./build.sh [命令]
```

### 可用命令

| 命令 | 说明 |
|------|------|
| `all` (默认) | 编译所有架构并打包 |
| `current` | 编译当前架构 |
| `arm64` | 编译 ARM64 版本 |
| `amd64` | 编译 x86_64 版本 |
| `package` | 打包所有版本 |
| `clean` | 清理编译产物 |
| `info` | 显示编译信息 |
| `help` | 显示帮助 |

## 编译流程

### 单架构编译（推荐用于开发）

```bash
# 在 Apple Silicon Mac 上：
make build-arm64

# 在 Intel Mac 上：
make build-amd64
```

### 多架构编译发布（推荐用于发布）

**步骤 1：在 ARM64 Mac 上编译**
```bash
git clone <repo>
cd mtool
make build-arm64
# dist/darwin-arm64/mtool 已生成
```

**步骤 2：在 Intel Mac 上编译**
```bash
git clone <repo>
cd mtool
make build-amd64
# dist/darwin-amd64/mtool 已生成
```

**步骤 3：合并并打包**

两个版本都编译完成后，合并 dist 目录：
```bash
# 在任意 Mac 上（确保两个版本都已复制到 dist/）
make package
# 生成：
# dist/mtool-v0.1.0-darwin-arm64.tar.gz
# dist/mtool-v0.1.0-darwin-amd64.tar.gz
```

## 编译输出

### 目录结构
```
dist/
├── darwin-arm64/
│   └── mtool                                # ARM64 二进制文件 (2.2MB)
├── darwin-amd64/
│   └── mtool                                # x86_64 二进制文件 (2.5MB)
├── mtool-v0.1.0-darwin-arm64.tar.gz        # ARM64 压缩包 (940KB)
└── mtool-v0.1.0-darwin-amd64.tar.gz        # x86_64 压缩包 (1.0MB)
```

### 验证编译结果
```bash
# 查看所有产物
ls -lh dist/

# 验证二进制文件
file dist/darwin-arm64/mtool
file dist/darwin-amd64/mtool

# 验证压缩包
tar -tzf dist/mtool-v0.1.0-darwin-arm64.tar.gz
tar -tzf dist/mtool-v0.1.0-darwin-amd64.tar.gz

# 测试二进制文件
./dist/darwin-arm64/mtool version
```

## 系统要求

### 最小要求
- **Go**：1.21 或更新版本
- **macOS**：13.0 或更新版本
- **make**：标准 Unix/Linux make

### 检查依赖
```bash
go version        # 检查 Go
make --version    # 检查 make
```

## 故障排查

### 错误：无法交叉编译含 CGo 代码

```
❌ 错误：无法交叉编译含 CGo 代码
当前架构：arm64
目标架构：amd64
```

**原因**：mtool 使用 CGo 直接访问 IOKit 硬件接口，无法交叉编译。

**解决方案**：
- 在 Intel Mac 上运行 `make build-amd64`
- 在 ARM64 Mac 上运行 `make build-arm64`

### 错误：Go 命令未找到

```bash
# 安装 Go
brew install go

# 或从官网下载：https://golang.org/dl
```

### 错误：permission denied

```bash
# 确保 build.sh 可执行
chmod +x build.sh
```

## 开发工作流

### 快速迭代

```bash
# 修改代码后重新编译
make clean
make build-arm64    # 或 build-amd64

# 测试二进制文件
./dist/darwin-arm64/mtool temp
```

### 准备发布

```bash
# 在两台 Mac 上分别编译
# ARM64 Mac:
make build-arm64

# Intel Mac:
make build-amd64

# 打包所有版本
make package

# 验证
ls -lh dist/
```

## 版本号管理

编译脚本中的版本号定义在：
- `Makefile`：`VERSION := 0.1.0`
- `main.go`：`const version = "0.1.0"`

更新版本号时，确保两个地方保持一致。

## 性能优化

### 编译大小
使用 `-ldflags="-s -w"` 移除调试符号（默认启用）
- 未优化：~10MB
- 优化后：~2.2MB

### 编译速度
首次编译会较慢（下载依赖），后续编译会使用缓存。

```bash
# 清理缓存后重新编译
make clean
make build-arm64
```

## GitHub Actions 集成

创建 `.github/workflows/build.yml`：

```yaml
name: Build and Package

on:
  push:
    tags:
      - 'v*'

jobs:
  build-arm64:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: make build-arm64
      - uses: actions/upload-artifact@v3
        with:
          name: dist-arm64
          path: dist/

  build-amd64:
    runs-on: macos-13
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: make build-amd64
      - uses: actions/upload-artifact@v3
        with:
          name: dist-amd64
          path: dist/

  package:
    needs: [build-arm64, build-amd64]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/download-artifact@v3
        with:
          name: dist-arm64
          path: dist/
      - uses: actions/download-artifact@v3
        with:
          name: dist-amd64
          path: dist/
      - run: make package
      - uses: softprops/action-gh-release@v1
        with:
          files: dist/mtool-v*.tar.gz
```

## 清单

发布新版本时：
- [ ] 更新 Makefile 中的 VERSION
- [ ] 更新 main.go 中的 version 常量
- [ ] 在两台 Mac 上编译
- [ ] 运行 `make package`
- [ ] 验证压缩包：`tar -tzf dist/mtool-v*.tar.gz`
- [ ] 测试二进制文件：`./dist/darwin-*/mtool temp`
- [ ] 创建 Git 标签：`git tag v0.1.0`
- [ ] 上传压缩包到 GitHub Releases

## 许可证

MIT License
