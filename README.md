# mtool - Mac 传感器监控工具

[![Build Status](https://github.com/ligson/mtool/actions/workflows/build.yml/badge.svg)](https://github.com/ligson/mtool/actions/workflows/build.yml)
[![Latest Release](https://img.shields.io/github/v/release/ligson/mtool?color=blue)](https://github.com/ligson/mtool/releases)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

一个轻量级 CLI 工具，用于监控 Apple Silicon（M1-M4）的传感器数据：CPU/GPU 温度、风扇转速和热学指标。

## 快速下载

从 [GitHub Releases](https://github.com/ligson/mtool/releases) 下载最新的编译好的二进制文件。

**支持架构：**
- ARM64 (Apple Silicon M1-M4) - ✅ 支持
- x86_64 (Intel Mac) - ✅ 支持

```bash
# 解压并安装
tar -xzf mtool-v*.tar.gz
cp mtool-v*/mtool /usr/local/bin/

# 验证安装
mtool temp
```

## 功能特性

✨ **无需 sudo** - 直接通过 IOKit 读取 SMC
🌡️ **CPU 温度** - 按核心和平均值读取
💨 **风扇转速** - 实际 RPM、最小/最大/目标值
📊 **多种输出格式** - 表格、JSON、CSV、纯文本
🎯 **灵活过滤** - 按类型（CPU/GPU）、key 或分组
⚡ **速度快** - 每次查询约 50ms，开销最小

## 从源代码编译

### 前置要求
- Go 1.21+
- macOS 13.0+（M1-M4 或 Intel Mac）
- macOS 编译工具（Xcode Command Line Tools）

### 编译步骤

**1. 克隆仓库**
```bash
git clone https://github.com/ligson/mtool.git
cd mtool
```

**2. 选择编译方式**

**仅编译当前架构（推荐）：**
```bash
# 在 ARM64 Mac 上
make build-arm64

# 在 Intel Mac 上
make build-amd64
```

**打包成压缩包：**
```bash
make package
# 输出: dist/mtool-v*.tar.gz
```

**编译两个架构（需要两台不同架构的 Mac）：**
```bash
# 在 ARM64 Mac 上
make build-arm64
make package

# 在 Intel Mac 上
make build-amd64
make package
```

**3. 安装**
```bash
# 安装到系统路径
sudo cp dist/darwin-arm64/mtool /usr/local/bin/  # ARM64
# 或
sudo cp dist/darwin-amd64/mtool /usr/local/bin/  # x86_64

# 或直接运行编译后的二进制
./dist/darwin-arm64/mtool temp
```

**4. 验证**
```bash
mtool temp
```

### Makefile 命令

```bash
make help              # 显示所有命令
make build-arm64       # 编译 ARM64（需要在 ARM64 Mac 上）
make build-amd64       # 编译 x86_64（需要在 Intel Mac 上）
make package           # 打包已编译的二进制
make all               # 编译所有架构并打包
make clean             # 清理编译产物
make info              # 显示编译配置
```

### 快速测试（不安装）
```bash
make build-arm64
./dist/darwin-arm64/mtool temp
```

## 快速开始

### 显示所有温度传感器
```bash
mtool temp
```
输出：
```
SENSOR             KEY   TEMPERATURE
------             ---   -----------
CPU Core 1         Tp01  40.0 °C  [████░░░░░░]
CPU Core 2         Tp05  40.0 °C  [████░░░░░░]
...
GPU Core 2         Tg0j  45.9 °C  [████░░░░░░]
```

### 仅显示 CPU 平均温度
```bash
mtool temp -t cpu -f plain
```
输出：
```
40.0
```

### 显示分组的传感器平均值
```bash
mtool temp -g
```
输出：
```
GROUP    AVG (°C)  COUNT  DETAILS
-----    --------  -----  -------
CPU      40.0      10     Tp01 40.0°C, Tp05 40.0°C, ...
GPU      45.9      2      Tg0j 45.9°C, Tg0d 45.9°C
SoC      53.2      2      Te05 53.5°C, Te0S 52.9°C
```

### 显示风扇转速
```bash
mtool fan
```
输出：
```
FAN    ACTUAL (RPM)  MIN (RPM)  MAX (RPM)  TARGET (RPM)
---    ------------  ---------  ---------  -----------
Fan 0  0             2317       7826       0
Fan 1  0             2317       7826       0
```

## 命令详解

### `mtool temp [选项]`
显示温度传感器。

**选项：**
- `-f, --format=<fmt>` - 输出格式：`table`（默认）、`json`、`plain`、`csv`
- `-g, --group` - 按传感器类型分组并显示平均值
- `-k, --key=<key>` - 仅显示特定传感器 key（例如 `Tp01`）
- `-t, --type=<type>` - 仅显示特定传感器类型：`cpu`、`gpu`、`soc`、`battery`、`ambient`

**示例：**
```bash
mtool temp                          # 所有传感器，表格格式
mtool temp -f json                  # JSON 格式
mtool temp -f plain                 # 一行一个数值
mtool temp -t cpu                   # 仅 CPU 传感器
mtool temp -t cpu -f plain          # CPU 平均温度（用于脚本）
mtool temp -t gpu -f json           # GPU 数据（JSON 格式）
mtool temp -g -f csv                # 分组平均值（CSV 格式）
mtool temp -k Tp01                  # 单个传感器（CPU 核心 1）
```

### `mtool fan [选项]`
显示风扇转速。

**选项：**
- `-f, --format=<fmt>` - 输出格式：`table`、`json`、`plain`、`csv`

**示例：**
```bash
mtool fan                           # 风扇转速，表格格式
mtool fan -f json                   # JSON 格式
mtool fan -f plain                  # 仅 RPM 值
```

### `mtool all [选项]`
显示所有传感器数据（温度、风扇、powermetrics）。

**选项：** 与 `temp` 和 `fan` 相同

**示例：**
```bash
mtool all                           # 完整报告
mtool all -f json                   # 所有数据（JSON 格式）
mtool all -t cpu -f plain           # 仅 CPU 平均值
```

### `mtool power`
显示来自 `powermetrics` 的功耗/热学数据（CPU 瓦数、集群频率）。

**注意：** 需要 `powermetrics` ���进制文件（通常在 M1+ 上可用）

**示例：**
```bash
mtool power                         # 功耗和热学数据
```

### `mtool diag`
调试：显示原始 SMC key 类型和值。

**示例：**
```bash
mtool diag                          # 所有 SMC key 及其类型
```

## 用例

### 1. 监控 CPU 温度
```bash
mtool temp -t cpu -f plain
# 输出：40.0
```

### 2. 检查系统是否发生热节流
```bash
mtool temp -t gpu -f plain
# 如果 > 90°C，很可能发生热节流
```

### 3. 每 10 秒记录一次温度数据
```bash
while true; do
  echo "$(date '+%Y-%m-%d %H:%M:%S') CPU: $(mtool temp -t cpu -f plain)°C GPU: $(mtool temp -t gpu -f plain)°C"
  sleep 10
done
```

### 4. 导出为 JSON 供监控工具使用
```bash
mtool temp -f json > /tmp/sensor_data.json
# 用 jq 解析、发送到监控系统等
```

### 5. 创建简单仪表板（需要 watch）
```bash
watch -n 1 'mtool all -f plain -g'
```

### 6. CPU 超过 85°C 时告警
```bash
cpu_temp=$(mtool temp -t cpu -f plain)
if (( $(echo "$cpu_temp > 85" | bc -l) )); then
  echo "⚠️ CPU 温度过高：${cpu_temp}°C"
fi
```

## 输出格式

### 表格（默认）
人类可读，带有彩色编码的温度条。
```
SENSOR      KEY   TEMPERATURE
------      ---   -----------
CPU Core 1  Tp01  40.0 °C  [████░░░░░░]
```

### 纯文本
一行一个数值。适合 shell 脚本和管道。
```
40.0
40.0
40.0
```

### JSON
结构化数据，包含完整详细信息。
```json
[
  {
    "name": "CPU Core 1",
    "key": "Tp01",
    "celsius": 40.0
  }
]
```

### CSV
电子表格兼容格式。
```
name,key,celsius
CPU Core 1,Tp01,40.0
CPU Core 2,Tp05,40.0
```

## 传感器类型

| 类型 | Key 前缀 | 数量 | 示例 |
|------|---------|------|------|
| CPU | `Tp` | 10 | CPU 核心 + 芯片总体 |
| GPU | `Tg` | 2-4 | GPU 核心 |
| SoC | `Te` | 2-4 | 芯片能耗区域 |
| 电池 | `TB` | 3-4 | 电池单元 |
| 环境 | `Ta` | 2-3 | 环境温度 |
| 其他 | 各种 | - | WiFi、NAND 等 |

## 技术说明

### 精度
- 温度读数：±0.5°C（SMC 报告精度）
- 风扇转速：±1 RPM
- SMC 每 ~100ms 更新一次

### 兼容性
- **M4**：✅ 已全面测试
- **M1-M3**：⚠️ 可能兼容（未测试）
- **Intel Mac**：❌ 不支持（不同的 SMC 格式）

### 数据格式
- 温度编码为 IEEE 754 float32（小端序）
- 风扇转速用 RPM 表示（浮点数）
- 所有值已解码为摄氏度/RPM

### 为什么不需要 sudo
mtool 直接通过 IOKit 从系统管理控制器读取，用户进程可用。只有内核级性能监控（PMU）需要 root。

## 故障排查

### "未找到温度传感器"
- 检查 SMC 可用性：`./mtool diag`
- 验证你在 Apple Silicon 上（M1+）：`uname -m` 应输出 `arm64`

### 所有核心温度都显示 40°C
- CPU 空闲/冷却时这是正常的
- 运行重计算来查看变化的读数：`yes > /dev/null &`

### 风扇转速显示 0 RPM
- 被动冷却模式（空闲/轻度负载时正常）
- 除非伴有节流，否则��需担心

### `powermetrics` 显示不可用
- 在 PATH 中找不到 `powermetrics` 二进制文件
- 这是可选的；所有核心功能无需它也能工作

## 开发

### 修改代码后运行
```bash
make build-arm64  # 或 make build-amd64（取决于你的 Mac 架构）
./dist/darwin-arm64/mtool temp
```

### 运行测试
```bash
go test ./...
```

## 贡献

欢迎提交 Issue 和 PR！

### 开发流程

1. Fork 仓库
2. 创建开发分支：`git checkout -b feature/xxx`
3. 本地编译测试：
   ```bash
   make build-arm64  # 或 build-amd64
   ./dist/darwin-arm64/mtool temp
   ```
4. 提交 PR

### 添加新功能

- 添加新传感器 key：见 `CLAUDE.md` 中的技术说明
- 修改代码后记得更新 `CHANGELOG.md`
- 在 M 系列 Mac（M1-M4）上充分测试

## 自动化发布

mtool 使用 GitHub Actions 自动化编译、打包和发布：

### 发布流程

**每次 push 到 main 分支时：**
1. ✅ 自动编译 ARM64 版本
2. ✅ 自动编译 x86_64 版本
3. ✅ 打包为 `.tar.gz`
4. ✅ 从 `CHANGELOG.md` 读取版本变更
5. ✅ 自动创建或更新 GitHub Release

### 如何发布新版本

**1. 更新版本号**
```bash
# 编辑 Makefile，修改 VERSION
VERSION := 0.2.0
```

**2. 更新 CHANGELOG.md**
```markdown
## [0.2.0] - 2026-03-22

### Added
- 新功能 1
- 新功能 2

### Fixed
- 修复的 bug
```

**3. 提交代码**
```bash
git add .
git commit -m "chore: bump version to 0.2.0"
git push origin main
```

**4. 自动发布**
GitHub Actions 会自动：
- 编译两个架构的二进制
- 读取 CHANGELOG.md 中对应版本的内容
- 创建 Release（标签为 `v0.2.0`）
- 上传编译好的压缩包

### Release 自动生成规则

- Release 标签：自动从 `Makefile` 中的 `VERSION` 生成（如 `v0.2.0`）
- Release 标题：`Release v0.2.0`
- Release 描述：从 `CHANGELOG.md` 中该版本的内容自动提取
- 发布文件：ARM64 和 x86_64 的 `.tar.gz` 压缩包

### 文件说明

- **CHANGELOG.md** - 记录每个版本的变更，GitHub Actions 会自动读取此文件
- **.github/workflows/build.yml** - GitHub Actions 工作流配置
- **scripts/extract-changelog.sh** - 用于从 CHANGELOG.md 提取版本内容的脚本

## 许可证

MIT 许可证 - 详见 LICENSE 文件

## 作者

为 M 系列 Mac 监控而创建。针对 M4 优化。

## 另见

- Apple 系统管理控制器（SMC）- IOKit 文档
- `powermetrics` - Apple 官方功耗指标工具
- `/usr/sbin/system_profiler SPPowerDataType` - 系统分析器
