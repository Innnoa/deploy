# 自定义变量（根据你的项目调整）
GITLAB_HOST = http://git.deepi.tech:888
PROJECT_ID  = 702
PACKAGE_NAME = Deploy
VERSION = 0.9.0.2607060
BUILD_DIR = build/bin
BASE_URL_DEV = https://deploy.ru.com/api-system
BASE_URL = https://ru.hpf.gov.hk/api-system
BINARY_NAME_WIN = $(PACKAGE_NAME)-windows-$(VERSION)-amd64.exe
BINARY_NAME_LINUX = $(PACKAGE_NAME)-linux-$(VERSION)-amd64
BINARY_NAME_WIN_DEV = $(PACKAGE_NAME)-windows-$(VERSION)-amd64-uat.exe
BINARY_NAME_LINUX_DEV = $(PACKAGE_NAME)-linux-$(VERSION)-amd64-uat

# 安全提示：GITLAB_TOKEN 必须通过环境变量传入，不要硬编码 set GITLAB_TOKEN=xxxxxxxxxx！
# 生成方式：GitLab 账号 → Settings → Access Tokens → 勾选 api 权限
GITLAB_TOKEN    ?= ""

.PHONY: build upload clean

# 默认任务：依次执行打包、上传
all: build upload

# 构建 Wails 应用
build-dev:
	wails build -o $(BINARY_NAME_WIN_DEV) -platform windows/amd64 -webview2 Embed -clean -ldflags "-s -w -X main.Version=$(VERSION) -X main.BaseUrl=$(BASE_URL_DEV)"

build:
	wails build -o $(BINARY_NAME_WIN) -platform windows/amd64 -webview2 Embed -clean -ldflags "-s -w -X main.Version=$(VERSION) -X main.BaseUrl=$(BASE_URL)"

build-dev-linux:
	wails build -o $(BINARY_NAME_LINUX_DEV) -platform linux/amd64 -clean -ldflags "-s -w -X main.Version=$(VERSION) -X main.BaseUrl=$(BASE_URL_DEV)"

build-linux:
	wails build -o $(BINARY_NAME_LINUX) -platform linux/amd64 -clean -ldflags "-s -w -X main.Version=$(VERSION) -X main.BaseUrl=$(BASE_URL)"

# 上传到 GitLab Generic Package Registry
upload-dev:
	@test $(GITLAB_TOKEN) || (echo "错误：必须设置 GITLAB_TOKEN 环境变量"; exit 1)
	curl --header "PRIVATE-TOKEN: $(GITLAB_TOKEN)" --upload-file $(BUILD_DIR)/$(BINARY_NAME_WIN_DEV) "$(GITLAB_HOST)/api/v4/projects/$(PROJECT_ID)/packages/generic/$(PACKAGE_NAME)/$(VERSION)/$(BINARY_NAME_WIN_DEV)"

upload:
	@test $(GITLAB_TOKEN) || (echo "错误：必须设置 GITLAB_TOKEN 环境变量"; exit 1)
	curl --header "PRIVATE-TOKEN: $(GITLAB_TOKEN)" --upload-file $(BUILD_DIR)/$(BINARY_NAME_WIN) "$(GITLAB_HOST)/api/v4/projects/$(PROJECT_ID)/packages/generic/$(PACKAGE_NAME)/$(VERSION)/$(BINARY_NAME_WIN)"

upload-linux-dev:
	@test $(GITLAB_TOKEN) || (echo "错误：必须设置 GITLAB_TOKEN 环境变量"; exit 1)
	curl --header "PRIVATE-TOKEN: $(GITLAB_TOKEN)" --upload-file $(BUILD_DIR)/$(BINARY_NAME_LINUX_DEV) "$(GITLAB_HOST)/api/v4/projects/$(PROJECT_ID)/packages/generic/$(PACKAGE_NAME)/$(VERSION)/$(BINARY_NAME_LINUX_DEV)"

upload-linux:
	@test $(GITLAB_TOKEN) || (echo "错误：必须设置 GITLAB_TOKEN 环境变量"; exit 1)
	curl --header "PRIVATE-TOKEN: $(GITLAB_TOKEN)" --upload-file $(BUILD_DIR)/$(BINARY_NAME_LINUX) "$(GITLAB_HOST)/api/v4/projects/$(PROJECT_ID)/packages/generic/$(PACKAGE_NAME)/$(VERSION)/$(BINARY_NAME_LINUX)"
# 清理构建产物
clean:
	rm -rf build

# 下载 WebView2 Fixed Version 运行时 CAB
# 下载后放入 webview2/ 目录，构建时会自动嵌入 exe
prepare-webview2:
	bash build/windows/webview2/download.sh
	@echo "=== 准备就绪 ==="
	@echo "放置 CAB 到 webview2/ 后，运行 make build 即可打包含运行时的单个 exe"
