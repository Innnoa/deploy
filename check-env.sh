#!/bin/bash
# Deploy 部署依赖检查脚本
# 用于检查目标机器是否满足 Deploy 项目的运行/编译要求

LOG_FILE="deploy-env-check-$(date +%Y%m%d-%H%M%S).log"
exec > >(tee -a "$LOG_FILE") 2>&1

echo "=============================================="
echo "  Deploy 项目部署环境检查"
echo "  检查时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo "  日志文件: $LOG_FILE"
echo "=============================================="
echo ""

# ========== 系统信息 ==========
echo "【1/8】系统基本信息"
echo "----------------------------------------------"
cat /etc/os-release 2>/dev/null | grep -E "^(NAME|VERSION|ID)=" | while read line; do
    echo "  $line"
done
echo "  内核版本: $(uname -r)"
echo "  架构:     $(uname -m)"
echo "  主机名:   $(hostname 2>/dev/null || echo 'unknown')"
echo "  当前用户: $(whoami)"
echo ""

# ========== glibc ==========
echo "【2/8】glibc 版本"
echo "----------------------------------------------"
GLIBC_VER=$(ldd --version 2>/dev/null | head -1 | awk '{print $NF}')
echo "  glibc: $GLIBC_VER"
if [ -n "$GLIBC_VER" ]; then
    MAJOR=$(echo $GLIBC_VER | cut -d. -f1)
    MINOR=$(echo $GLIBC_VER | cut -d. -f2)
    if [ "$MAJOR" -gt 2 ] || ([ "$MAJOR" -eq 2 ] && [ "$MINOR" -ge 28 ]); then
        echo "  ✓ glibc 版本满足要求 (>=2.28)"
    else
        echo "  ✗ glibc 版本过低 (需要 >=2.28)"
    fi
fi
echo ""

# ========== 编译工具 ==========
echo "【3/8】编译工具链"
echo "----------------------------------------------"

# GCC
GCC_VER=$(gcc --version 2>/dev/null | head -1 | awk '{print $NF}')
if [ -z "$GCC_VER" ]; then
    echo "  GCC:  未安装"
    echo "    安装: yum install -y gcc"
else
    echo "  GCC:  $GCC_VER"
    echo "  ✓ GCC 已安装"
fi

# Go
GO_VER=$(go version 2>/dev/null | awk '{print $3}' | sed 's/go//')
if [ -z "$GO_VER" ]; then
    echo "  Go:   未安装"
    echo "    安装: 下载 https://go.dev/dl/ 解压到 /usr/local/go"
    echo "    环境: export PATH=\$PATH:/usr/local/go/bin"
else
    echo "  Go:   $GO_VER"
    # Wails v2 需要 Go 1.18+
    MAJOR=$(echo $GO_VER | cut -d. -f1)
    MINOR=$(echo $GO_VER | cut -d. -f2)
    if [ "$MAJOR" -gt 1 ] || ([ "$MAJOR" -eq 1 ] && [ "$MINOR" -ge 18 ]); then
        echo "  ✓ Go 版本满足要求 (>=1.18)"
    else
        echo "  ✗ Go 版本过低 (需要 >=1.18)"
    fi
fi

# Node.js
NODE_VER=$(node -v 2>/dev/null | sed 's/v//')
if [ -z "$NODE_VER" ]; then
    echo "  Node: 未安装"
    echo "    安装: 下载 https://nodejs.org/ 解压到 /usr/local/node"
else
    echo "  Node: $NODE_VER"
    # Wails 需要 Node 14+
    MAJOR=$(echo $NODE_VER | cut -d. -f1)
    if [ "$MAJOR" -ge 14 ]; then
        echo "  ✓ Node 版本满足要求 (>=14)"
    else
        echo "  ✗ Node 版本过低 (需要 >=14)"
    fi
fi

# npm
NPM_VER=$(npm -v 2>/dev/null)
if [ -z "$NPM_VER" ]; then
    echo "  npm:  未安装"
else
    echo "  npm:  $NPM_VER"
fi

# Wails CLI
WAILS_VER=$(wails version 2>/dev/null | head -1)
if [ -z "$WAILS_VER" ]; then
    echo "  Wails CLI: 未安装"
    echo "    安装: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
else
    echo "  Wails CLI: $WAILS_VER"
fi
echo ""

# ========== GTK3 ==========
echo "【4/8】GTK3 (Wails GUI 依赖)"
echo "----------------------------------------------"
GTK3_VER=$(pkg-config --modversion gtk+-3.0 2>/dev/null)
if [ -z "$GTK3_VER" ]; then
    echo "  GTK3:       未找到 (pkg-config)"
    echo "  RPM 检查:"
    rpm -q gtk3 2>/dev/null && echo "    gtk3 已安装" || echo "    gtk3 未安装"
    rpm -q gtk3-devel 2>/dev/null && echo "    gtk3-devel 已安装" || echo "    gtk3-devel 未安装 (编译需要)"
    echo "    安装: yum install -y gtk3 gtk3-devel"
else
    echo "  GTK3:       $GTK3_VER"
    echo "  ✓ GTK3 已安装"
    rpm -q gtk3-devel 2>/dev/null >/dev/null && echo "  ✓ gtk3-devel 已安装" || echo "  ✗ gtk3-devel 未安装 (编译需要)"
fi
echo ""

# ========== WebKitGTK ==========
echo "【5/8】WebKitGTK (Wails 渲染引擎) — 关键！"
echo "----------------------------------------------"

# 尝试 pkg-config
WEBKIT_PC=""
for pc in webkit2gtk-4.0 webkit2gtk-4.1 webkit2gtk-6.0; do
    VER=$(pkg-config --modversion $pc 2>/dev/null)
    if [ -n "$VER" ]; then
        echo "  $pc:         $VER (pkg-config)"
        WEBKIT_PC="$pc"
        break
    fi
done
if [ -z "$WEBKIT_PC" ]; then
    echo "  pkg-config: 未找到 webkit2gtk (编译需要)"
fi

# RPM 版本
WEBKIT_RPM=$(rpm -q webkit2gtk3 2>/dev/null)
if [ -z "$WEBKIT_RPM" ]; then
    echo "  RPM:        webkit2gtk3 未安装"
else
    echo "  RPM:        $WEBKIT_RPM"
fi

WEBKIT_DEVEL=$(rpm -q webkit2gtk3-devel 2>/dev/null)
if [ -z "$WEBKIT_DEVEL" ]; then
    echo "  RPM:        webkit2gtk3-devel 未安装"
    echo "    安装: yum install -y webkit2gtk3-devel"
else
    echo "  RPM:        $WEBKIT_DEVEL"
fi

WEBKIT_JSC=$(rpm -q webkit2gtk3-jsc 2>/dev/null)
if [ -z "$WEBKIT_JSC" ]; then
    echo "  RPM:        webkit2gtk3-jsc 未安装"
else
    echo "  RPM:        $WEBKIT_JSC"
fi

# WebKit 兼容性判断
WEBKIT_NUM=$(rpm -q webkit2gtk3 2>/dev/null --qf "%{VERSION}" | cut -d. -f2)
echo ""
echo "  --- Wails GUI 兼容性检查 ---"
echo "  Wails v2.0  最低要求: webkit2gtk >= 2.24"
echo "  Wails v2.4  建议:      webkit2gtk >= 2.28"
echo "  Wails v2.10 建议:      webkit2gtk >= 2.36"
echo ""
if [ -n "$WEBKIT_NUM" ]; then
    echo "  当前版本: 2.${WEBKIT_NUM}"
    if [ "$WEBKIT_NUM" -ge 36 ]; then
        echo "  ✓ 满足 Wails v2.10 要求 — GUI 正常"
    elif [ "$WEBKIT_NUM" -ge 28 ]; then
        echo "  ⚠ 满足 Wails v2.4 要求 — GUI 可用"
    elif [ "$WEBKIT_NUM" -ge 24 ]; then
        echo "  ⚠ 满足 Wails v2.0 最低要求 — 建议降 Wails 版本"
    else
        echo "  ✗ 版本过低 (<2.24) — GUI 无法工作！"
        echo "    Kylin V10 已知问题: webkit2gtk=2.22.2 不兼容 Ant Design 5 CSS-in-JS"
        echo "    解决方案: 使用 HTTP 模式 + 系统浏览器"
    fi
else
    echo "  ✗ 未检测到 webkit2gtk 版本"
fi
echo ""

# ========== 磁盘/内存 ==========
echo "【6/8】磁盘与内存"
echo "----------------------------------------------"
echo "  磁盘使用:"
df -h / | tail -1 | awk '{printf "    总量: %s  已用: %s  可用: %s  使用率: %s\n", $2, $3, $4, $5}'
echo ""
echo "  内存:"
free -h | head -2 | tail -1 | awk '{printf "    总量: %s  已用: %s  可用: %s\n", $2, $3, $7}'
echo ""

# ========== 浏览器 ==========
echo "【7/8】浏览器 (HTTP 模式备用)"
echo "----------------------------------------------"
FIREFOX_VER=$(firefox --version 2>/dev/null)
if [ -z "$FIREFOX_VER" ]; then
    echo "  Firefox: 未安装"
else
    echo "  $FIREFOX_VER"
fi

CHROME_VER=$(google-chrome --version 2>/dev/null || chromium-browser --version 2>/dev/null)
if [ -z "$CHROME_VER" ]; then
    echo "  Chrome:  未安装"
else
    echo "  $CHROME_VER"
fi
echo ""

# ========== 网络 ==========
echo "【8/8】网络连接测试"
echo "----------------------------------------------"
echo "  目标服务器: deploy.ru.com"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 5 https://deploy.ru.com 2>/dev/null)
if [ "$HTTP_CODE" = "000" ]; then
    echo "  ✗ 无法连接 deploy.ru.com"
else
    echo "  ✓ HTTP $HTTP_CODE"
fi
echo ""

# ========== 当前 Wails 版本配置 ==========
echo "=============================================="
echo "  总结"
echo "=============================================="

ISSUES=0

check_or_install() {
    local name="$1"
    local cmd="$2"
    local install_hint="$3"
    if eval "$cmd" >/dev/null 2>&1; then
        return 0
    else
        echo "  ✗ 缺少: $name — $install_hint"
        ISSUES=$((ISSUES + 1))
        return 1
    fi
}

echo "  编译依赖:"
check_or_install "gcc" "gcc --version" "yum install -y gcc"
check_or_install "go >=1.18" "go version" "下载 https://go.dev/dl/ 解压到 /usr/local/go"
check_or_install "node >=14" "node -v" "下载 https://nodejs.org/"
check_or_install "gtk3-devel" "rpm -q gtk3-devel" "yum install -y gtk3-devel"
check_or_install "webkit2gtk3-devel" "rpm -q webkit2gtk3-devel" "yum install -y webkit2gtk3-devel"

echo ""
echo "  运行依赖:"
check_or_install "gtk3" "rpm -q gtk3" "yum install -y gtk3"
check_or_install "webkit2gtk3" "rpm -q webkit2gtk3" "yum install -y webkit2gtk3"
WARN_GUI=0
if [ -n "$WEBKIT_NUM" ] && [ "$WEBKIT_NUM" -lt 24 ]; then
    echo "  ⚠ webkit2gtk 版本过低，GUI 无法工作 — 需 HTTP 模式"
    WARN_GUI=1
fi

echo ""
echo "  配置项:"
echo "  - /etc/oem-info: $( [ -f /etc/oem-info ] && echo '✓ 存在' || echo '✗ 不存在 (需手动创建或设 \$Longlabel)' )"
echo "  - MTU:           $(ip link show ens160 2>/dev/null | grep -oP 'mtu \K[0-9]+' || echo 'N/A')"

echo ""
echo "=============================================="
echo "  检查完成: $(date '+%Y-%m-%d %H:%M:%S')"
echo "  问题数量: $ISSUES 个编译/运行依赖缺失"
if [ "$WARN_GUI" -eq 1 ]; then
    echo "  ⚠ GUI 模式不可用，请使用 HTTP 模式部署"
fi
echo "  日志文件: $LOG_FILE"
echo "=============================================="
