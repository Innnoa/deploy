# WebView2 CAB 放置目录

把下载好的 WebView2 Fixed Version x64 `.cab` 文件放到这个目录。

推荐文件名格式：

```text
Microsoft.WebView2.FixedVersionRuntime.130.x.x.x.x64.cab
```

来源：

```text
https://developer.microsoft.com/en-us/microsoft-edge/webview2/
```

操作路径：

```text
Fixed Version -> x64 -> Download .cab
```

说明：

- 当前仓库会跟踪这个 `README.md`，确保 Windows 构建时 `//go:embed webview2/*` 始终有可匹配文件。
- 真正的 `.cab` 文件不会提交到仓库，`.gitignore` 已忽略 `webview2/*.cab`。
- 构建 Windows 包前，只需要把 `.cab` 放进当前目录即可。
