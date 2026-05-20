# 自定义变量（根据你的项目调整）
GITLAB_HOST = http://git.deepi.tech:888
PROJECT_ID  = 702
PACKAGE_NAME = Deploy
VERSION = 0.9.0.2605200
BUILD_DIR = build/bin
BASE_URL = https://ru.hpf.gov.hk/api-system
BINARY_NAME_WIN = $(PACKAGE_NAME)-windows-$(VERSION)-amd64.exe
BINARY_NAME_LINUX = $(PACKAGE_NAME)-linux-$(VERSION)-amd64

# 安全提示：GITLAB_TOKEN 必须通过环境变量传入，不要硬编码 set GITLAB_TOKEN=xxxxxxxxxx！
# 生成方式：GitLab 账号 → Settings → Access Tokens → 勾选 api 权限
GITLAB_TOKEN    ?= ""

.PHONY: build upload clean

# 默认任务：依次执行打包、上传
all: build upload

# 构建 Wails 应用
build:
	wails build -o $(BINARY_NAME_WIN) -platform windows/amd64 -webview2 Embed -clean -ldflags "-s -w -X main.Version=$(VERSION) -X main.BaseUrl=$(BASE_URL)"

build-linux:
	wails build -o $(BINARY_NAME_LINUX) -platform linux/amd64 -clean -ldflags "-s -w -X main.Version=$(VERSION) -X main.BaseUrl=$(BASE_URL)"

# 上传到 GitLab Generic Package Registry
upload:
	@test $(GITLAB_TOKEN) || (echo "错误：必须设置 GITLAB_TOKEN 环境变量"; exit 1)
	curl --header "PRIVATE-TOKEN: $(GITLAB_TOKEN)" --upload-file $(BUILD_DIR)/$(BINARY_NAME_WIN) "$(GITLAB_HOST)/api/v4/projects/$(PROJECT_ID)/packages/generic/$(PACKAGE_NAME)/$(VERSION)/$(BINARY_NAME_WIN)"

upload-linux:
	@test $(GITLAB_TOKEN) || (echo "错误：必须设置 GITLAB_TOKEN 环境变量"; exit 1)
	curl --header "PRIVATE-TOKEN: $(GITLAB_TOKEN)" --upload-file $(BUILD_DIR)/$(BINARY_NAME_LINUX) "$(GITLAB_HOST)/api/v4/projects/$(PROJECT_ID)/packages/generic/$(PACKAGE_NAME)/$(VERSION)/$(BINARY_NAME_LINUX)"
# 清理构建产物
clean:
	rm -rf build