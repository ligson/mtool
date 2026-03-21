# Changelog

所有重要的项目变更都记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)，
并且项目遵循 [Semantic Versioning](https://semver.org/spec/v2.0.0.html)。

## [Unreleased]

### Added
- 多架构编译支持（ARM64 和 amd64）
- GitHub Actions 自动构建和发布工作流
- 持续构建（continuous）预发布版本
- 版本标签自动发布流程

### Changed
- 工作流简化为单个构建任务
- 改进错误消息和日志

## [0.1.0] - 2026-03-21

### Added
- 初始版本发布
- CPU 温度传感器读取（无需 sudo）
- GPU 温度传感器读取
- 风扇转速显示
- 多种输出格式支持（表格、JSON、CSV、纯文本）
- 传感器分组和平均值计算
- 传感器类型过滤（CPU、GPU、SoC、电池、环境）
- 诊断命令用于 SMC key 调试
- 完整的 Makefile 和 build.sh 脚本
- 详细的文档和使用示例
- MIT 许可证

### Technical
- CGo + IOKit 框架集成
- IEEE 754 float32 小端序 SMC 数据解码
- Apple Silicon（M1-M4）特定的优化
- 温度范围过滤（10-120°C）

[Unreleased]: https://github.com/ligson/mtool/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ligson/mtool/releases/tag/v0.1.0
