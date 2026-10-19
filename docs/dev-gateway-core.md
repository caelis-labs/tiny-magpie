# Gateway-core Dev acceptance

本地 Dev 已获维护者确认，进入 0.2.0 正式版发布验证。功能边界见
[确认清单](feature-review-2026-10-19.md)。

## 打开与隔离

本工作区的 `.tiny/core-dev/tiny-magpie Dev.app` 是临时的 ad-hoc 签名版本，
不是正式分发包。其本地构建的原生入口在进入 CLI 前设置固定环境；从 Finder 打开同样
使用 `.tiny/core-dev/profile/`，不读取已安装 Tiny/Magpie 的账号或 Agent 配置。
订阅账号读取通过 `MAGPIE_NO_ACCOUNTS=1` 硬禁用，自动更新关闭；该验收包不会因误点开关而读取系统钥匙串。

- GUI/Web：`http://127.0.0.1:13427`
- 模型网关：`http://127.0.0.1:13426`
- 两个本地示例 Provider 的 fixture 服务：`127.0.0.1:13430`

示例 Key、模型、token、速度和价格都是测试数据，不代表厂商实测或真实费用。
没有已知价格的调用显示「未计价」，不会当作免费。关闭后可双击上述 app 或运行
`.tiny/core-dev/launch.sh`。示例服务退出后，可单独运行
`/usr/bin/python3 .tiny/core-dev/vendor.py`；实际使用时可在这个 Dev 中手动添加
API Key 测试 Provider。订阅实测需要另外明确测试账号与隔离方案，不在此示例包中开启。关闭 Dev 请使用自己的菜单/退出，不操作已安装版本。

## 人工验收点

1. 供应商：添加面板保留订阅、API、自定义、本地和 TypeSafe；插件仍可追加。
   示例 API 的 Embedding/Rerank 与聊天模型分开展示，可复制完整模型 ID。
2. 网关：System One 示例使用 `dev-typesafe/jev-latest` 和 `/v1/systemone`。
   Embedding/Rerank 入口给出各自模型与端点；它们不参加聊天模型混选。
3. 路由：普通组、轮询/顺序/权重/额度及失败切换保留；没有智能/意图分类路由。
   多个 ChatGPT 订阅账号沿用上游 Codex Provider 的账号池，普通组引用其模型；
   各账号的额度、暂停状态和模型权益分别生效。此示例包不开放真实订阅接入。
   同一组也可混合订阅、官方 API 和 OpenRouter 等渠道的模型；成员按
   Provider 与模型 ID 区分，保留各渠道的计价和协议能力。真实混合渠道待单独验收。
4. 用量：概览及请求可见 token、费用估算、输出速度和首字延迟；未知值为破折号。
   订阅额度卡保留，重置卡/签到/自动积分消费不再出现。
5. 设置/插件：无隐私、自定义字体、内置功能开关等独立边缘产品；GUI 主导航保留
   供应商、路由、网关、用量和插件。主窗口与托盘鸟图标不超出标题栏。

## 证据边界

自动测试与 native Dev 均只使用 fixture。没有验证真实 TypeSafe/订阅账号、第三方
插件的当前发布包，也没有在 Windows/Linux 打开原生窗口。跨平台 vet/构建不能
替代这些验收。当前已集成的上游基线保持不变，未在本轮混入新的上游区间。
