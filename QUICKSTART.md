# mtool 快速开始指南

## 最简单的方式

### 编译当前架构
```bash
make build-arm64    # 在 ARM64 Mac 上
# 或
make build-amd64    # 在 Intel Mac 上
```

### 打包所有架构
```bash
make package
```

### 一键清理
```bash
make clean
```

## 使用 build.sh 脚本

```bash
./build.sh current          # 编译当前架构
./build.sh all              # 编译所有架构并打包
./build.sh arm64            # 编译 ARM64
./build.sh amd64            # 编译 x86_64
./build.sh package          # 打包
./build.sh clean            # 清理
./build.sh help             # 显示帮助
```

## 常用操作

| 任务 | 命令 |
|------|------|
| 编译当前架构 | `make build-arm64` / `make build-amd64` |
| 打包 | `make package` |
| 清理 | `make clean` |
| 查看帮助 | `make help` |

## 编译输出位置

所有产物都在 `dist/` 目录下：

```
dist/
├── darwin-arm64/
│   └── mtool (2.2MB)
├── darwin-amd64/
│   └── mtool (2.5MB)
├── mtool-v0.1.0-darwin-arm64.tar.gz (940KB)
└── mtool-v0.1.0-darwin-amd64.tar.gz (1.0MB)
```

## 验证编译

```bash
# 检查二进制文件
file dist/darwin-arm64/mtool

# 运行版本命令
./dist/darwin-arm64/mtool version

# 快速功能测试
./dist/darwin-arm64/mtool temp -f plain
```

## 系统要求

- Go 1.21+
- macOS 13.0+
- make（通常预装）

## 多架构编译流程

```bash
# 步骤 1：在 Apple Silicon Mac 上编译 ARM64
make build-arm64

# 步骤 2：在 Intel Mac 上编译 x86_64
make build-amd64

# 步骤 3：打包所有架构
make package
```

## 遇到问题？

查看详细文档：
- **编译指南**：见 `BUILD.md`
- **开发指南**：见 `CLAUDE.md`
- **用户指南**：见 `README.md`
