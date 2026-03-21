# 发布指南

## 自动发布流程（ARM64）

当你创建 version tag 时，GitHub Actions 会自动：
1. 编译 ARM64 版本
2. 打包为 `.tar.gz`
3. 创建 GitHub Release 并上传

```bash
# 创建版本标签
git tag v0.2.0
git push origin v0.2.0
```

GitHub Actions 会自动编译并发布 ARM64 版本。

## 完整发布流程（ARM64 + amd64）

要同时发布两个架构的包，需要在两台不同的机器上编译：

### 步骤 1：在 ARM64 Mac 上编译

```bash
# 检出代码
git clone <repo>
cd mtool

# 编译 ARM64
make build-arm64

# 打包
make package
```

### 步骤 2：在 Intel Mac 上编译

```bash
# 检出相同的代码版本
git clone <repo>
cd mtool
git checkout v0.2.0  # 切换到相同的 tag

# 编译 amd64
make build-amd64

# 打包
make package
```

### 步骤 3：合并包文件

在任意一台机器上，将另一台机器的包文件合并：

```bash
# 例如在 ARM64 Mac 上，从 Intel Mac 复制 amd64 包
scp user@intel-mac:~/mtool/dist/mtool-v0.2.0-darwin-amd64.tar.gz dist/
```

或者手动下载 GitHub Actions artifacts：
1. 访问 https://github.com/YOUR_REPO/actions
2. 找到最新的 workflow 运行
3. 下载 `mtool-darwin-arm64` 和 `mtool-darwin-amd64` artifacts

### 步骤 4：上传到 Release

使用 `upload-release.sh` 脚本将两个包上传到 GitHub Release：

```bash
# 确保有两个包文件：
# dist/mtool-v0.2.0-darwin-arm64.tar.gz
# dist/mtool-v0.2.0-darwin-amd64.tar.gz

# 运行上传脚本
./scripts/upload-release.sh v0.2.0
```

脚本会自动：
- 验证 tag 存在
- 检查包文件
- 上传到对应的 GitHub Release

## 单架构发布

如果只需要发布 ARM64 版本：

```bash
git tag v0.2.0
git push origin v0.2.0
# GitHub Actions 自动编译并发布
```

## 环境要求

- **ARM64 Mac**：运行 `make build-arm64`（Apple Silicon M1-M4）
- **Intel Mac**：运行 `make build-amd64`（Intel 处理器）
- **GitHub CLI**：安装 `gh` 用于上传脚本（可选，可使用网页界面上传）

```bash
# 安装 gh CLI
brew install gh

# 认证
gh auth login
```

## 故障排查

### `gh: command not found`

安装 GitHub CLI：
```bash
brew install gh
gh auth login
```

### "tag not found"

确保 tag 已创建和推送：
```bash
git tag v0.2.0
git push origin v0.2.0
```

### "Package not found"

确保已编译和打包：
```bash
make build-arm64    # 或 build-amd64
make package
```

## GitHub Actions 状态

- **Continuous Build**：每次 push 到 main 自动编译 ARM64
  - Release 标签：`continuous`
  - 预发布版本

- **Version Release**：创建 version tag 时自动发布
  - Release 标签：`v0.2.0` 等
  - 稳定版本
