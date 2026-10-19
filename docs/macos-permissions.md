# macOS permission attribution

用户报告：使用原版 Magpie 时，客户端申请辅助功能、录屏或自动化权限，系统显示
的申请者有时是 Magpie。本轮未取得该次弹窗的进程链，不能确认具体触发来源。

## 系统归属与 Tiny 的边界

macOS 的 TCC 按 responsible code 识别申请者；由一个 app 启动的 helper 触发
受保护操作时，系统可能把 app 作为权限主体。HTTP 模型协议没有转移本地 TCC
授权主体的机制。参见 Apple DTS 的
[On File System Permissions](https://developer.apple.com/forums/thread/678819)，
尤其 responsible code、稳定签名和 native main executable 的说明。

Tiny 的模型网关应只转发客户端声明的工具调用，交回客户端执行。它不为客户端
启动桌面 Agent，不提供公共计算机/屏幕工具服务，不申请辅助功能、屏幕录制或
Apple Events 自动化授权。现有源码扫描未发现这些主动授权 API；正式 bundle 模板及签名流程
使用自己的 `dev.caelis.tiny-magpie` 身份，没有新增相关 entitlement。

保留的 Claude 订阅通过真正的 Claude Code CLI 提交模型请求，因而仍有必要的
子进程。其工作目录是独立临时目录，`claudeCLIArgs` 显式传入空 `--tools`；仅当
客户端请求供应商搜索时允许 `WebSearch`，并传 `--strict-mcp-config` 与空
`--setting-sources`。客户端的工具通过私有 MCP callback 返回模型响应，不在
网关执行。这里的 `--dangerously-skip-permissions` 是上游私有桥接的既有约定，
不能取消禁用本地工具的条件，也不能当作 macOS 权限绕过手段。

读取订阅登录材料可能需要 Tiny 自己的钥匙串/文件授权，这与辅助功能、录屏、
自动化是不同的操作；关闭供应商页的订阅账号后沿用上游的禁止发现守卫。
Provider 插件仍是用户安装并信任的可执行代码，上游插件宿主不构成权限沙箱；
不能承诺任意第三方插件永远不触发自己的子进程或系统权限。

## 本轮修复与验证

- Tiny 自身更新提升权限的提示残留了 `magpie is updating itself.`，已改为
  `tiny-magpie is updating itself.`。这是管理员更新提示的身份修复，不代表修复
  了用户此前的录屏/辅助功能弹窗。
- `TestTinyAutomationToolCallsStayWithClient` 经完整 HTTP handler 把一个
  `screencapture` 工具调用交还客户端；OS 命令由哨兵替代，未被执行。该测试还
  确认 Claude 订阅不会启用本地工具。启用本地 Read 的 overlay 会使其失败。
- `TestTinyManagementSurface` 覆盖已移除的 Agent 连接、CLI 管理、会话恢复等
  API，返回 404；不恢复原版的应用管理链路。
- `TestTinyAdminPromptIdentity` 替换管理员运行器，不弹真实授权窗口；恢复旧提示
  的 overlay 会因错误产品名称失败。

没有读取或重置用户 TCC 数据库，没有实际申请录屏/辅助功能权限，也没有用真实
Agent 或账号重现原版事件。若今后在 Tiny 再次发生，需要按当时的客户端、插件和
具体弹窗核对进程归属，不能通过改名、私有 API 或扩大 entitlement 掩盖它。
