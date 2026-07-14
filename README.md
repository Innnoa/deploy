# Deploy

Deploy 是一个基于 Wails 的桌面部署工具，前端使用 React 18 + Ant Design 5，后端使用 Go 1.25。当前仓库已包含 Windows 和 Linux 两套安装流程，以及 Windows 的 WebView2 内嵌兜底逻辑。

## 环境要求

### Windows 开发 / 构建

- Go `1.25+`
- Node.js `18+` 和 npm
- Wails CLI v2
- WebView2 Runtime

安装 Wails CLI：

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

说明：

- Windows 11 通常自带 WebView2。
- Windows 10 若没有系统级 WebView2，可在构建时内嵌 Fixed Version CAB 作为兜底。

### Linux 开发 / 构建

- Go `1.25+`
- Node.js `18+` 和 npm
- Wails CLI v2
- GTK3 + WebKit2GTK

Arch:

```bash
sudo pacman -S gtk3 webkit2gtk-4.1
```

Debian / Ubuntu:

```bash
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev
```

说明：

- 当前 Linux 版本启动时会通过 `pkexec` 提权；开发和调试时需要系统具备对应图形权限提升能力。

### Linux 交叉编译 Windows

- `mingw-w64`

Arch:

```bash
sudo pacman -S mingw-w64-gcc
```

## 首次拉取

```bash
git clone <repo-url> deploy
cd deploy
cd frontend
npm install
cd ..
go mod download
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

当前仓库关键版本：

- Go：`go.mod` 中为 `1.25.0`
- Wails：`github.com/wailsapp/wails/v2 v2.10.1`
- 前端：React 18、Ant Design 5、Vite 3

## WebView2 CAB 准备

构建 Windows 包前，如需内嵌 Fixed Version 运行时：

1. 打开 `https://developer.microsoft.com/en-us/microsoft-edge/webview2/`
2. 在 `Fixed Version` 中下载 x64 的 `.cab`
3. 放到项目根目录的 `webview2/` 下

示例文件名：

```text
webview2/Microsoft.WebView2.FixedVersionRuntime.130.x.x.x.x64.cab
```

说明：

- `webview2/` 目录内容不会提交到仓库。
- 仓库已提供 `webview2/README.md` 作为占位文件，因此即使还没放 CAB，Windows 侧的 `//go:embed webview2/*` 也不会因为目录缺失直接报错。
- Windows 构建时 `//go:embed webview2/*` 会把该目录下的 `.cab` 打进可执行文件。
- 运行时会优先检测系统 WebView2；若不存在，则解压内嵌或外置 CAB 到 `WebView2Runtime/`。
- 仓库里的 `make prepare-webview2` 实际执行 `build/windows/webview2/download.sh`，它只输出手动下载指引，不会自动下载 CAB。

## 开发调试

```bash
wails dev
```

效果：

- 前端走 Vite 热重载
- Go 改动会触发重编译

## 构建打包

构建前先更新 `Makefile` 中的 `VERSION`，当前示例值为：

```makefile
VERSION = 0.9.0.2607140
```

产物默认输出到 `build/bin/`。

### Windows 生产包

```bash
wails build -o Deploy-windows-{VERSION}-amd64.exe -platform windows/amd64 -webview2 Embed -clean -ldflags "-s -w -X main.Version={VERSION} -X main.BaseUrl=https://ru.hpf.gov.hk/api-system"
```

### Windows UAT 包

```bash
wails build -o Deploy-windows-{VERSION}-amd64-uat.exe -platform windows/amd64 -webview2 Embed -clean -ldflags "-s -w -X main.Version={VERSION} -X main.BaseUrl=https://deploy.ru.com/api-system"
```

### Linux 原生包

生产：

```bash
wails build -o Deploy-linux-{VERSION}-amd64 -platform linux/amd64 -clean -ldflags "-s -w -X main.Version={VERSION} -X main.BaseUrl=https://ru.hpf.gov.hk/api-system"
```

UAT：

```bash
wails build -o Deploy-linux-{VERSION}-amd64-uat -platform linux/amd64 -clean -ldflags "-s -w -X main.Version={VERSION} -X main.BaseUrl=https://deploy.ru.com/api-system"
```

### Linux 交叉编译 Windows 包

Go API 有变动时，先在 Linux 原生环境生成 bindings：

```bash
wails generate module
```

生产：

```bash
wails build \
  -o Deploy-windows-{VERSION}-amd64.exe \
  -platform windows/amd64 \
  -webview2 Embed \
  -skipbindings \
  -clean \
  -ldflags "-s -w -X main.Version={VERSION} -X main.BaseUrl=https://ru.hpf.gov.hk/api-system"
```

UAT：

```bash
wails build \
  -o Deploy-windows-{VERSION}-amd64-uat.exe \
  -platform windows/amd64 \
  -webview2 Embed \
  -skipbindings \
  -clean \
  -ldflags "-s -w -X main.Version={VERSION} -X main.BaseUrl=https://deploy.ru.com/api-system"
```

`-skipbindings` 的用途：

- 避免交叉编译时重新生成 bindings
- 避免 Windows 临时二进制在 Linux 上执行导致 `exec format error`

## Makefile 简写

```bash
make build              # Windows 生产
make build-dev          # Windows UAT
make build-linux        # Linux 生产
make build-dev-linux    # Linux UAT
make clean              # 清空构建产物
make prepare-webview2   # 输出 WebView2 CAB 下载说明
```

说明：

- `all` 默认执行 `build` 和 `upload`
- `upload*` 任务要求环境变量 `GITLAB_TOKEN`

## 关键配置

| 配置 | 位置 | 当前值 |
| --- | --- | --- |
| 应用版本 | `Makefile` | `VERSION = 0.9.0.2607140` |
| API 生产 | `Makefile` / `ldflags main.BaseUrl` | `https://ru.hpf.gov.hk/api-system` |
| API UAT | `Makefile` / `ldflags main.BaseUrl` | `https://deploy.ru.com/api-system` |
| 默认 BaseUrl | `main.go` | `https://deploy.ru.com/api-system` |
| HMAC Key | `service/api/api.go` | `ACCESS_KEY` / `ACCESS_SECRET` |
| RU HMAC Key | `service/api/api.go` | `ACCESS_KEY_RU` / `ACCESS_SECRET_RU` |

## 项目结构

```text
.
├── main.go
├── app.go
├── Makefile
├── wails.json
├── webview2_embed.go
├── webview2_stub.go
├── frontend/
│   ├── src/App.tsx
│   └── src/components/
│       ├── info/
│       ├── welcome/
│       ├── configuration/
│       └── deploy/
├── service/
│   ├── api/api.go
│   ├── common/
│   └── deploy/
│       ├── installer.go
│       ├── installer_win.go
│       ├── installer_linux.go
│       ├── reboot_win.go
│       ├── reboot_linux.go
│       ├── computer_win.go
│       ├── computer_linux.go
│       └── package.go
├── build/
│   └── windows/webview2/download.sh
└── webview2/
```

## 核心流程

前端页面流：

```text
Info + Welcome -> Configuration -> Deploy
```

启动与恢复逻辑：

```text
启动
  -> 单实例检查
  -> Linux 下通过 pkexec 提权
  -> 初始化 API Client
  -> 若带 -restart 参数则从 Deploy 页恢复
  -> Windows 下检测系统 WebView2，不满足时解压内嵌/外置 CAB
```

部署流程：

```text
Welcome
  -> 获取 OA Server
  -> 获取并校验 Seed Label
  -> Configuration 选择本地/网络打印机
  -> Deploy.GetInstallPackages 聚合安装项
  -> Deploy.DoInstall 执行安装
  -> 遇到 "Restart Machine" 时保存状态并重启
  -> 重启后通过 -restart + LoadTemporaryInfo 继续
  -> 全部完成后删除自启动/计划任务
```

## 上传产物

Makefile 已内置 GitLab Generic Package Registry 上传命令，依赖：

```bash
export GITLAB_TOKEN=<token>
```

可用目标：

```bash
make upload
make upload-dev
make upload-linux
make upload-linux-dev
```
