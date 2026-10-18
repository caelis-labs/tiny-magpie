# tiny-magpie

保留 Magpie 多 Provider 模型网关、原版协议转换、订阅/Provider 插件，以及桌面和 Web GUI 的精简 fork。

裁剪 Agent 管理、内容库、会话浏览、同步、意图路由及附加媒体/搜索/MCP 服务。沿用原版实现，以删减和兼容验证为主线。配置目录独立，不自动迁移原版数据。

当前为实验性版本。构建使用 `GOWORK=off make cli`，随后运行 `./tiny-magpie serve` 或 `./tiny-magpie web --no-open`。macOS 可用 `make app`；Windows/Linux 沿用原生构建流程。

桌面与 CLI 下载见 [Releases](https://github.com/caelis-labs/tiny-magpie/releases)，三平台均提供 amd64/arm64。macOS 使用组织 Developer ID 签名并经 Apple 公证，Windows/Linux 为 unsigned。使用前校验 `SHA256SUMS`。发布与平台要求见 [发布说明](docs/releasing.md)。

完整范围见 [裁剪契约](docs/tiny-scope.md)，上游同步流程见 [UPSTREAM.md](UPSTREAM.md)，验证结果和未完成验收见 [兼容证据](docs/compatibility.md)。主 [README](README.md) 提供命令及测试方法。
