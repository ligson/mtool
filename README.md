# mtool - Mac 传感器监控工具

[![Build Status](https://github.com/ligson/mtool/actions/workflows/build.yml/badge.svg)](https://github.com/ligson/mtool/actions/workflows/build.yml)
[![Latest Release](https://img.shields.io/github/v/release/ligson/mtool?color=blue)](https://github.com/ligson/mtool/releases)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

一个轻量级 CLI 工具，用于监控 Apple Silicon（M1-M4）的传感器数据：CPU/GPU 温度、风扇转速和热学指标。

## 快速下载

### 最新构建
- **Continuous Build**：[GitHub Release - continuous](https://github.com/ligson/mtool/releases/tag/continuous) （每次 main 分支更新）
- **稳定版本**：[GitHub Releases](https://github.com/ligson/mtool/releases) （版本标签）

## 功能特性

✨ **无需 sudo** - 直接通过 IOKit 读取 SMC
🌡️ **CPU 温度** - 按核心和平均值读取
💨 **风扇转速** - 实际 RPM、最小/最大/目标值
📊 **多种输出格式** - 表格、JSON、CSV、纯文本
🎯 **灵活过滤** - 按类型（CPU/GPU）、key 或分组
⚡ **速度快** - 每次查询约 50ms，开销最小

## 安装

### 从源代码构建
```bash
git clone https://github.com/ligson/mtool.git
cd mtool
go build -o mtool .
sudo cp mtool /usr/local/bin/
```

### 快速测试（不安装）
```bash
cd mtool
go build -o mtool .
./mtool temp
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

## 从源代码构建

### 要求
- Go 1.21+
- macOS 13.0+（M1-M4）

### 构建
```bash
go build -o mtool .
```

### 运行测试
```bash
go test ./...
```

## 贡献

要添加对新传感器 key 或格式的支持：
1. 参见 `CLAUDE.md` 获取开发指南
2. 在你的 M 系列 Mac 上测试
3. 提交发现/PR

## GitHub Actions 自动化

mtool 使用 GitHub Actions 自动化编译、打包和发布：

### 工作流
- **build.yml** - 在每次 push 和 tag 发布时自动编译和发布
- **continuous-release.yml** - 在每次 main 分支更新时生成最新构建

### 自动发布
- 每次 push 到 main 分支时，自动更新 [continuous Release](https://github.com/ligson/mtool/releases/tag/continuous)
- 创建版本标签（如 `v0.2.0`）时，自动发布正式版本到 Releases
- 所有架构（ARM64 和 x86_64）自动编译和打包

详见 [.github/workflows/README.md](.github/workflows/README.md)

## 许可证

MIT 许可证 - 详见 LICENSE 文件

## 作者

为 M 系列 Mac 监控而创建。针对 M4 优化。

## 另见

- Apple 系统管理控制器（SMC）- IOKit 文档
- `powermetrics` - Apple 官方功耗指标工具
- `/usr/sbin/system_profiler SPPowerDataType` - 系统分析器
