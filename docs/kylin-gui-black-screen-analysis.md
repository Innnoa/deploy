# Kylin V10 GUI 黑屏原因分析

## 系统环境

| 项目 | 值 |
|---|---|
| OS | Kylin Linux Advanced Server V10 (Halberd) |
| 内核 | 4.19.90-89.11.v2401.ky10.x86_64 |
| glibc | 2.28 |
| GCC | 7.3.0 |
| Firefox | 79.0 |
| webkit2gtk3 | 2.22.2-12.p01.ky10 |
| webkit2gtk3-devel | 2.22.2-12.p01.ky10 |
| webkit2gtk3-jsc | 2.22.2-12.p01.ky10 |

## 根本原因

**webkit2gtk3 2.22.2 (2018) 不兼容 Wails v2。**

- Wails v2.0.0 最低要求 webkit2gtk 2.24
- Wails v2.4.0（当前使用）建议 2.28+
- Wails v2.10.1 建议 2.36+

前端使用 Ant Design 5 + antd-style 的 CSS-in-JS 运行时，依赖以下现代 Web API：

| API | WebKitGTK 2.22 支持 | 说明 |
|---|---|---|
| CSSStyleSheet 构造函数 | ❌ | antd-style CSS-in-JS 核心依赖 |
| ResizeObserver | ❌ | 组件尺寸监听 |
| IntersectionObserver | ⚠️ 部分 | 懒加载/可见性检测 |
| ES2020+ 语法 | ❌ | 可选链 `?.`、空值合并 `??` 等 |

WebKitGTK 2.22.2 对应 Safari 12 (2018)，以上 API 在该引擎中不存在或行为不正确，导致 JS 执行报错、React 无法渲染、页面白屏。

## 为什么 Firefox 能显示

| 对比维度 | Firefox 79 | WebKitGTK 2.22.2 |
|---|---|---|
| 发布时间 | 2020 | 2018 |
| CSSStyleSheet 构造 | ✅ | ❌ |
| ResizeObserver | ✅ | ❌ |
| IntersectionObserver | ✅ | ⚠️ 部分 |
| ES2020 语法 | ✅ | ❌ |

Firefox 79 是独立的浏览器引擎（Gecko），比同时期的 WebKit 对现代 Web 标准支持更好。同一份 HTML/JS/CSS 在 Firefox 79 中完整执行，在 WebKitGTK 2.22.2 中因 API 缺失直接报错停止。

## 为什么 Arch Linux 能跑

开发环境为 WSL Arch Linux (内核 5.15)，通过 WSLg 提供 Wayland 图形支持。

| 项目 | Arch (WSL) | Kylin V10 |
|---|---|---|
| webkit2gtk | **2.52.4** (4.1 API) | 2.22.2 (4.0 API) |
| 发布年份 | 2025 | 2018 |
| CSSStyleSheet 构造 | ✅ | ❌ |
| ResizeObserver | ✅ | ❌ |
| IntersectionObserver | ✅ | ⚠️ |
| ES2020+ 语法 | ✅ | ❌ |
| ES2022+ (顶层 await 等) | ✅ | ❌ |
| Wails 版本 | v2.10.1 | v2.4.0 |
| Wails 编译标签 | `-tags webkit2_41` | (无，使用 4.0 API) |

**Arch 的 webkit2gtk 2.52.4（2025 年）涵盖了前端所需的所有现代 Web API**。版本差距跨越 7 年（2018 → 2025），WebKit 在此期间新增了：

- CSS Typed OM
- CSSStyleSheet.replace/replaceSync
- ResizeObserver (Safari 13.1 / WebKit 608+)
- ES2020-ES2022 全套语法支持
- Web Locks API
- 数十个性能/安全改进

**三层对比：**

| API/特性 | Arch (2.52.4) | Kylin (2.22.2) | Firefox 79 |
|---|---|---|---|
| CSSStyleSheet 构造 | ✅ | ❌ | ✅ |
| ResizeObserver | ✅ | ❌ | ✅ |
| ES2020 可选链 `?.` | ✅ | ❌ | ✅ |
| ES2022 顶层 await | ✅ | ❌ | ❌ |
| wails dev | ✅ | ❌（编译 ok） | N/A |
| wails build | ✅ (-tags webkit2_41) | ✅ (webkit2gtk 4.0) | N/A |

**结论**：代码没变，同一仓库，同一分支。Arch 能跑纯粹因为 webkit2gtk 版本覆盖了前端所需的全部 Web API。Kylin 不能跑纯粹因为 webkit2gtk 版本太老。这不是代码兼容性问题，是运行时环境版本差异。

## V10 SP3 的 webkit2gtk 版本

Kylin V10 Halberd 全系列（含 SP1/SP2/SP3）的 webkit2gtk3 版本均为 2.22.2-12.p01.ky10。

这是 Kylin 基于 RHEL/CentOS 8 fork 后锁定的版本。RHEL 8 主线最终版本为 webkit2gtk 2.28 — 但仍低于 Wails v2 建议的最低版本。Kylin 没有后续跟进 webkit2gtk 的升级补丁。

## 升级方案与成本分析

### 方案一：Rocky Linux 8 的 webkit2gtk3 2.42.5

Rocky Linux 8.10 AppStream 提供 webkit2gtk3 2.42.5-1.el8 和 2.46.3-1.el8_10。

**缺失依赖：**

| 包 | 说明 |
|---|---|
| libwpe-1.0 | WPE 渲染后端库 |
| libWPEBackend-fdo-1.0 | WPE FreeDesktop 后端 |
| webkit2gtk3-jsc = 2.42.5 | JavaScriptCore 引擎（配套版本） |
| libicu*.so.60 | ICU 国际化库（版本可能不匹配） |

**成本评估：**

- 需要从 Rocky 8 仓库同步下载 4-6 个 RPM
- ICU 库版本冲突风险：Kylin 的 ICU 与 Rocky 8 的 ICU 版本可能不一致
- 如果 ICU 不兼容，需要连带升级 ICU 及其反向依赖（可能波及系统其他组件）
- 测试成本：升级后需验证系统桌面、浏览器等是否正常
- 回滚成本：需保留旧 RPM 备份，出问题 `dnf downgrade` 恢复

**风险评估：中高。** 虽然 Kylin 和 Rocky 8 同为 RHEL 8 系，但 Kylin 对部分库有自定义 patch，强行替换可能导致系统不稳定。

### 方案二：源码编译 webkit2gtk

需要：
- CMake ≥ 3.16（Kylin 可能不满足,可安装）
- GCC ≥ 9（Kylin 只有 7.3.0,可安装）
- 构建空间 ≥ 20GB（Kylin 虚拟机剩余约 5GB,可分配更大空间）
- 编译时间 ≥ 4-8 小时（6.5GB RAM）


## 证据（日志）

Wails 日志显示 HTML/JS/CSS 文件加载正常，Go 后端无异常：

```
DEBUG | [AssetHandler] Handling request '/' (file='.')
DEBUG | [AssetHandler] Handling request '/assets/index.06a370c4.js'
DEBUG | [AssetHandler] Handling request '/assets/index.8ddcce21.css'
```

无 Go 层 panic/error 日志。白屏完全由 WebKit 渲染层静默失败导致（JS 引擎遇到不支持的 API 抛出异常但未被 Go 层捕获）。

## 结论

非代码问题，是 Kylin V10 操作系统自带的 webkit2gtk 版本过低导致的兼容性问题。Kylin V10 所有子版本（含 SP3）的 webkit2gtk 均为 2.22.2，官方仓库无更新版本。

