# GitHub Workflows

mtool 项目包含两个自动化工作流程。

## 1. build.yml - 编译和发布工作流

**触发条件：**
- Push 到 main/master 分支
- 创建 v* 标签（版本发布）
- Pull Request 到 main/master 分支

**流程：**

```
build-arm64              build-amd64
   ↓                         ↓
   └─────────────┬───────────┘
                 ↓
             package
                 ↓
             release (仅在创建 tag 时)
```

**详细步骤：**

1. **build-arm64** (macOS-latest)
   - 检出代码
   - 设置 Go 环境
   - 执行 `make build-arm64`
   - 上传 ARM64 二进制文件到 artifacts

2. **build-amd64** (macOS-13)
   - 检出代码
   - 设置 Go 环境
   - 执行 `make build-amd64`
   - 上传 x86_64 二进制文件到 artifacts

3. **package** (需要前两个完成)
   - 下载两个二进制文件
   - 执行 `make package` 打包
   - 上传压缩包到 artifacts

4. **release** (仅当推送 tag 时)
   - 下载所有 artifacts
   - 创建 GitHub Release
   - 上传 tar.gz 到 Release assets

**发布流程：**

```bash
# 更新版本号
# 编辑 Makefile, main.go 中的 VERSION

# 提交更改
git add .
git commit -m "chore: bump version to v0.2.0"
git push

# 创建标签（触发 release 工作流）
git tag v0.2.0
git push origin v0.2.0
```

## 2. continuous-release.yml - 持续发布工作流

**触发条件：**
- 每次 Push 到 main/master 分支

**流程：**

```
build-arm64              build-amd64
   ↓                         ↓
   └─────────────┬───────────┘
                 ↓
        package-and-release
                 ↓
         更新 "continuous" Release
```

**特点：**

- 自动在每次 main 分支提交时编译和打包
- 创建一个名为 "continuous" 的特殊 Release
- 每次 push 时自动更新该 Release 的 assets
- 标记为 **prerelease**（预发布）
- 包含最新的构建信息和提交哈希

**assets 包含：**

- mtool-v{VERSION}-darwin-arm64.tar.gz
- mtool-v{VERSION}-darwin-amd64.tar.gz

## 工作流使用场景

### 场景 1：日常开发

```
git push origin feature-branch  →  build.yml (编译但不发布)
                                  ↓
                                编译成功
                                artifacts 中有可用的二进制文件
```

### 场景 2：发布新版本

```
git tag v0.2.0                  →  build.yml (完整流程)
git push origin v0.2.0             ↓
                                build-arm64, build-amd64
                                   ↓
                                  package
                                   ↓
                                 release
                                   ↓
                        GitHub Release with assets
```

### 场景 3：持续构建

```
git push origin main            →  continuous-release.yml
                                   ↓
                                编译所有架构
                                   ↓
                                打包并更新 continuous Release
                                   ↓
                    最新的二进制文件可在 Release 下载
```

## 下载构建产物

### 从 GitHub Actions Artifacts

1. 打开 Actions 标签页
2. 选择最新的工作流运行
3. 下载所需的 artifact（mtool-packages）

### 从 GitHub Releases

**从 Continuous Release（最新）：**
```bash
https://github.com/{owner}/{repo}/releases/tag/continuous
```

**从版本 Release：**
```bash
https://github.com/{owner}/{repo}/releases/tag/v0.2.0
```

## 环境要求

### macOS-latest (ARM64)
- 用于编译 Apple Silicon 版本
- 自动使用最新 macOS（通常是最新的 macOS 版本）

### macOS-13 (x86_64)
- 用于编译 Intel 版本
- 使用 macOS 13 确保与较旧的 Intel Mac 兼容

### Ubuntu-latest (打包)
- 用于打包和发布
- 需要 Makefile 中的 make 和相关工具

## 故障排查

### 编译失败

检查：
1. Go 版本是否正确（1.21+）
2. Makefile 中的编译命令是否正确
3. CGo 依赖是否正确

### 发布失败

检查：
1. `GITHUB_TOKEN` 权限是否正确
2. Release 名称是否正确
3. 文件路径是否存在

### 权限问题

确保仓库设置中：
- Settings → Actions → General → Workflow permissions 设置为 "Read and write"

## 自定义工作流

### 修改触发条件

编辑 workflow 文件中的 `on:` 部分：

```yaml
on:
  push:
    branches: [main, develop]
    paths: ['src/**', 'Makefile']
  pull_request:
    branches: [main]
```

### 修改 Go 版本

```yaml
- uses: actions/setup-go@v4
  with:
    go-version: '1.22'  # 修改此处
```

### 添加自定义步骤

在 `package-and-release` job 中添加步骤：

```yaml
- name: Custom step
  run: |
    echo "做一些其他的事情"
```

## 最佳实践

1. **使用版本标签发布稳定版本**
   ```bash
   git tag v0.2.0 -m "Release version 0.2.0"
   git push origin v0.2.0
   ```

2. **定期检查 continuous Release**
   - 用于获取最新的开发版本
   - 不适合生产环境

3. **清理旧的 artifacts**
   - 设置 `retention-days` 自动删除旧 artifacts
   - 防止占用存储空间

4. **监控工作流状态**
   - 在 README 中添加工作流徽章
   ```markdown
   ![Build](https://github.com/{owner}/{repo}/actions/workflows/build.yml/badge.svg)
   ```

## 相关文件

- `.github/workflows/build.yml` - 标签发布工作流
- `.github/workflows/continuous-release.yml` - 持续发布工作流
- `Makefile` - 编译命令定义
- `build.sh` - 本地编译脚本

## 许可证

MIT License
