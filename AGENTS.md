# mtool - 开发指南

## 项目概述
**mtool** 是一个 Go CLI 工具，用于监控 Mac M 系列芯片（M4）的传感器数据，无需 sudo。它通过 Apple 的系统管理控制器（SMC）读取 CPU 温度、GPU 温度、风扇转速等热学指标。

## 项目结构

```
mtool/
├── go.mod              # Go 模块定义
├── main.go             # CLI 入口、命令路由、输出格式化
├── smc.go              # CGo IOKit 接口，用于 SMC（温度、风扇读取）
├── powermetrics.go     # powermetrics 子进程封装（可选）
├── mtool               # 编译后的二进制文件（由 `go build` 生成）
├── AGENTS.md           # 本文件 - 开发说明
└── README.md           # 面向用户的文档
```

## 文件说明

### `main.go` (~600 行)
- **用途**：命令行界面和输出格式化
- **关键函数**：
  - `cmdTemp()` - 温度传感器显示，支持过滤/分组
  - `cmdFan()` - 风扇转速显示
  - `cmdAll()` - 合并传感器输出
  - `outputSensors()`, `outputGrouped()` - 格式输出（表格/json/纯文本/csv）
  - `groupSensors()` - 按传感器类型计算分组平均值
- **标志**：
  - `-f, --format` - 输出格式（table/json/plain/csv）
  - `-g, --group` - 按传感器类型分组
  - `-k, --key` - 过滤单个 SMC key
  - `-t, --type` - 按传感器类型过滤（cpu/gpu/soc/battery/ambient）

### `smc.go` (~200 行，CGo)
- **用途**：通过 IOKit 的底层 SMC 驱动
- **关键 C 函数**：
  - `smc_open()` - 初始化 AppleSMC IOService 连接
  - `smc_read_key_info()` - 获取 SMC key 元数据（大小、数据类型）
  - `smc_read_key()` - 从 SMC key 读取原始字节
  - `smc_temp()` - 解码温度（支持 sp78、flt、sp5a 格式）
  - `smc_fan_rpm()` - 解码风扇转速（支持 fpe2、flt 格式）
- **Go 包装方法**：
  - `SMC.Open()`, `SMC.Close()` - 生命周期管理
  - `SMC.TemperatureSensors()` - 扫描已知温度 key
  - `SMC.FanCount()`, `SMC.FanRPM()` - 风扇数据
  - `SMC.Temp()` - 查询单个温度 key
- **M 系列特性**：
  - 温度 key 使用 `flt`（IEEE 754 float32）类型，存储为**小端序**
  - 应用过滤：仅保留 10-120°C 之间的值，排除返回 ~0-2°C 的断电核心
  - 支持读取 10 个 CPU 核心 + 芯片总体温度

### `powermetrics.go` (~120 行)
- **用途**：通过 `powermetrics` 命令获取可选的系统功耗/频率数据
- **关键类型**：
  - `PowermetricsResult` - 解析后的输出（风扇、CPU 集群、能耗）
  - `RunPowermetrics()` - 执行 `powermetrics` 并解析 JSON
- **行为**：
  - 无需 sudo 运行（如果不可用则静默失败）
  - 由 `mtool all` 和 `mtool power` 命令使用

## 关键技术决策

### 1. Apple Silicon SMC 数据格式
**M 系列使用不同于 Intel Mac 的 SMC 编码：**
- 温度 key：`flt`（0x666c7420）- IEEE 754 float32，**小端序**
- 风扇转速 key：`flt`（0x666c7420）- 与老旧 `fpe2` 格式相同
- 解码方式：4 字节 → 解释为 LE uint32 → 转换为 float32

**示例：**
```
字节：00 20 20 42
LE uint32：0x42202000 = 40.0（IEEE 754 编码）
```

### 2. 温度过滤策略
- 读取所有已知的 Tp/Tg/Te/TB/Ta key（~50+ 个总计）
- 过滤：仅保留 10.0 °C < 值 < 120.0 °C
  - 避免返回 0-2°C 占位符值的断电 CPU 核心
  - 避免无效/未初始化的 key
- 结果：每次扫描约 20-25 个有效传感器

### 3. 传感器类型分类
按 key 前缀分类：
- `Tp*` → CPU（10 核 + 芯片总体）
- `Tg*` → GPU（2-4 核）
- `Te*` → SoC/能耗（2-4 区域）
- `TB*` → 电池（3-4 个电池单元）
- `Ta*` → 环境温度（2-3 个传感器）
- `Ts*, TW*, T*` → 其他传感器

### 4. 输出格式设计
四种模式适应不同用例：
- **table**（默认）：人类可读，含彩色温度条
- **plain**：一行一个数值（适合 shell 管道）
- **json**：完整结构含详细信息
- **csv**：电子表格兼容格式

## 构建和开发

### 构建
```bash
cd /Users/ligson/workspace/work-org/github/mtool
go build -o mtool .
```

### 系统范围安装
```bash
sudo cp mtool /usr/local/bin/
```

### 依赖
- Go 1.21+
- macOS 13.0+（M1-M4 的 Ventura 或更新版本）
- IOKit 框架（通过 CGo 链接）

### 测试 SMC key
```bash
./mtool diag    # 显示原始 key 类型和值
```

## 常见开发任务

### 添加新的温度传感器 key
1. 查找 SMC key 名称（例如 `Tp15` 用于新的 CPU 核心）
2. 添加到 `smc.go:TemperatureSensors()` 中的 `knownKeys` 切片
3. 测试：`./mtool diag | grep Tp15`

### 添加新的输出格式
1. 在 `main.go` 的 `outputSensors()` 或 `outputGrouped()` 中添加 case
2. 遍历数据并格式化
3. 示例：format=xml、format=yaml

### 调整温度范围过滤
- 编辑 `smc.go:TemperatureSensors()` 第 ~100 行
- 当前：`v > 10.0 && v < 120.0`
- 权衡：更严格 = 更少虚假读数，更宽松 = 可能包含无效数据

## 兼容性说明

### M1-M4 测试状态
- ✅ M4：已全面测试（所有 20+ 个传感器读取正确）
- ⚠️ M1-M3：未测试（key 名称可能略有不同）

### 已知限制
- 仅读取 SMC 数据（不读取内核性能计数器）
- `powermetrics` 要求二进制文件可用（通常存在）
- 被动冷却模式下风扇转速显示 0 RPM
- 没有历史数据/趋势（仅快照）

## 内存和性能
- SMC 读取：每次查询 ~50ms
- 无缓存（每次调用都获取新鲜数据）
- 内存占用：<5MB
- 二进制大小：~6.5MB

## 代码风格
- 无外部依赖（仅 CGo）
- 最小化标志解析（使用内置 flag 包）
- 表格格式化通过 `text/tabwriter`
- JSON 序列化通过内置 `encoding/json`
