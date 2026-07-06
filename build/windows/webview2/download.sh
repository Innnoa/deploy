#!/bin/bash
# 下载 WebView2 Fixed Version 运行时 CAB 文件
# 使用方法: bash build/windows/webview2/download.sh

set -e

ARCH="x64"
DEST_DIR="webview2"

# 检查是否已存在 CAB 文件
if ls "${DEST_DIR}"/*.cab 2>/dev/null; then
    echo "已存在 CAB 文件，跳过下载。"
    echo "如需重新下载，请先删除 webview2/*.cab"
    exit 0
fi

echo "=== 正在获取最新 WebView2 Fixed Version 下载链接 ==="

# 从 Microsoft 下载页面获取最新 Fixed Version 的 URL
# 首先获取下载页面
DOWNLOAD_PAGE="https://developer.microsoft.com/en-us/microsoft-edge/webview2/"
echo "访问: ${DOWNLOAD_PAGE}"
echo ""
echo "请手动操作："
echo "1. 打开浏览器访问: ${DOWNLOAD_PAGE}"
echo "2. 在页面中找到 'Fixed Version' 部分"
echo "3. 选择架构: ${ARCH}"
echo "4. 点击下载 .cab 文件"
echo "5. 将下载的 .cab 文件放入: ${DEST_DIR}/"
echo ""
echo "或者如果你知道具体版本号，可以手动拼接 URL："
echo "https://msedge.sf.dl.delivery.mp.microsoft.com/filestreamingservice/files/<content-id>/Microsoft.WebView2.FixedVersionRuntime.<version>.${ARCH}.cab"
echo ""
echo "推荐使用较新版本 (如 130.x 或更新)，确保兼容性。"
