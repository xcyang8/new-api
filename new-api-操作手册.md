# new-api 操作手册

生产站点：https://www.xiaoyule.com.cn/ （数羊 CountSheep）

---

## 1. 渠道会话保持（多账号负载均衡必配）

### 1.1 背景与原理

new-api 对**同优先级、同权重**的多个渠道按**每请求随机加权**分发（`model/channel_cache.go` 的 `GetRandomSatisfiedChannel`，权重全为 0 时等效等权 100:100）。这意味着同一段对话的每一轮请求都会重新随机，约 50% 概率落到另一个渠道。

当多个渠道是**同一厂商的不同账号**（如两个 Kimi 账号）时，随机分发会导致：

- 上游的**服务端上下文缓存（prompt/context cache）按账号（API key）隔离**。请求轮流打到两个账号，每个账号只有约一半请求能命中自己缓存的对话前缀，另一半全量 prefill 长上下文；
- 冷缓存 → 首 token 延迟（TTFT）大幅上升 → t/s（输出 token ÷ 总耗时）被明显拉低；
- 账号侧的会话连续性同时断裂。

**典型症状**：使用日志中同一用户的请求在两个渠道间高频交替，t/s 剧烈波动（如 10 ~ 45 t/s 交替出现）。

### 1.2 配置规则：新增渠道/模型时必查

> **只要满足以下两个条件，就必须配置会话保持规则：**
> 1. 同一模型挂了 ≥ 2 个同优先级渠道（负载均衡 / 轮询）；
> 2. 这些渠道是不同账号/不同密钥（上游缓存不共享）。

配置入口：**系统设置 → 请求策略 → 会话与重试 → 会话规则 → 添加规则**

| 配置项 | 值 | 说明 |
|---|---|---|
| 名称 | `<厂商> session`，如 `kimi session` | |
| 模型正则 | 覆盖该模型的所有客户端可见名称，如 `^(k3.*\|kimi-.*)$` | 注意含模型映射前的请求名（如 `k3`、`k3[1m]`、`kimi-k3`） |
| 路径正则 | `/v1/messages`、`/v1/responses`、`/v1/chat/completions` 每行一个 | 分别覆盖 Claude 协议、Codex Responses、OpenAI 协议客户端 |
| Key 来源 | 按顺序配置多个，第一个取到值的生效（见下表） | 取不到任何 Key 时规则惰性跳过，无副作用 |
| 会话保持 | 继承全局默认（优先原渠道，失败允许切换） | 兼顾缓存命中与故障转移 |
| TTL | 0（用全局默认 3600s） | |

**各客户端工具的会话标识（Key 来源按此顺序配置）：**

| 顺序 | 类型 | 值 | 覆盖的客户端 |
|---|---|---|---|
| 1 | gjson | `metadata.user_id` | Claude Code（`/v1/messages`） |
| 2 | gjson | `prompt_cache_key` | Codex CLI（`/v1/responses`）、OpenCode OpenAI 协议（`/v1/chat/completions`） |
| 3 | request_header | `x-opencode-session` | OpenCode Anthropic 协议（不带 `metadata.user_id`，用此 header 标识会话） |
| 4 | request_header | `x-session-id` | OpenRouter 风格客户端兜底 |

添加后必须点击页面顶部 **保存更改**，并**刷新页面确认规则仍在**。

效果：同一会话（同一 `metadata.user_id`）固定打到同一账号，缓存保持温热；不同会话仍随机分散到各账号，负载均衡不丢失。

### 1.3 当前生产已配置的会话规则

| 规则名 | 模型正则 | Key 来源 | 备注 |
|---|---|---|---|
| codex cli trace | `^gpt-.*$` | gjson `prompt_cache_key` | 系统默认 |
| claude cli trace | `^claude-.*$` | gjson `metadata.user_id` | 系统默认 |
| kimi session | `^(k3.*\|kimi-.*)$` | gjson `metadata.user_id` → gjson `prompt_cache_key` → header `x-opencode-session` → header `x-session-id` | 2026-09-23 添加，覆盖 k3 / k3[1m] / k3-256k / kimi-k3 / kimi-for-coding(-highspeed)；路径含 /v1/messages、/v1/responses、/v1/chat/completions，兼容 Claude Code / Codex / OpenCode |

### 1.4 验证方法

1. **规则生效**：会话规则表格中该规则的"缓存"列出现 > 0 的条目数，说明已有会话被绑定到固定渠道。
2. **请求落点**：控制台 → 使用日志，观察同一用户同一段时间内的请求是否固定在同一渠道（不再两个渠道交替）。
3. **速度**：数据看板 / 使用日志中 t/s 应趋于稳定在高位（缓存命中态），不再高低交替。

### 1.5 常见注意事项

- **规则不匹配时完全惰性**：模型正则、路径正则、Key 提取三个条件任一不满足即跳过，不绑定也不应用参数覆盖，对其他客户端零副作用。
- **客户端不带任何已知会话标识**时规则不产生绑定，退回随机分发。此时可改用 Key 来源 `context_int`（如 `id` 用户 ID）做用户级粘性。
- **只挂了一个渠道、或多个渠道共用同一上游账号**时无需配置。
- **主备模式（不同优先级）不需要会话规则**：流量本就固定在最高优先级渠道。
- 新增模型时若走了**模型映射**（如 `kimi-k3 → k3`），模型正则要匹配**映射前的客户端请求名**。
- 修改规则后如行为异常，可用"清空全部缓存"按钮清除旧的会话绑定。

---

## 2. Kimi For Coding 渠道余额查询 404（已修复）

### 2.1 问题现象

渠道管理页面对 kimi-xcy（ID 3）、kimi-wyj（ID 4）点击"余额查询"返回 **status code: 404**。

### 2.2 根因

- 两个渠道类型是 **58（AdvancedCustom 高级自定义）**，base_url 为 `https://api.kimi.com/coding`（Kimi For Coding 套餐端，非按量计费的 `api.moonshot.cn`）。
- 高级自定义渠道的余额查询走 `advanced_routes` 中入站路径为 `/v1/dashboard/billing/credit_grants` 的路由（`relaykit/dto/channel_settings.go` 的 `AdvancedCustomBalancePath`，控制器在 `controller/channel-billing.go` 的 `fetchAdvancedCustomBalance`）。
- 渠道当时把该余额路由的上游路径也配成了 `/v1/dashboard/billing/credit_grants`，而 Kimi 套餐端**不实现** OpenAI 计费接口 → `https://api.kimi.com/coding/v1/dashboard/billing/credit_grants` 返回 404。

### 2.3 正确的套餐用量端点

Kimi For Coding 套餐用量查询端点（参考 cc-switch 项目 `src-tauri/src/services/coding_plan.rs`，2026-09-23 用生产真实 key 实测 HTTP 200）：

```
GET https://api.kimi.com/coding/v1/usages
Authorization: Bearer <key>
```

返回内容：`usage`（周限额 limit/used/remaining/resetTime）、`limits[].detail`（5 小时窗口）、`usages`（limit_5h / limit_7d 已用比例）、`booster_wallet`（充值钱包，含 topupLimit 等）。

### 2.4 修复（2026-09-23 已执行）

直接改 prod 数据库 `channels.settings`，把渠道 3、4 的余额路由上游路径改为 `/v1/usages`：

```
incoming: /v1/dashboard/billing/credit_grants  →  upstream: /v1/usages
```

备份文件在服务器 `/app/new-api/backup/channels-3-4-settings-20260923-151332.bak`。prod 未开启 `MEMORY_CACHE_ENABLED`（默认 false，`common/init.go`），渠道读取直连 DB，**无需重启即刻生效**。

### 2.5 效果与注意

- 余额查询现在返回 200。
- **2026-09-23 已加原生支持**（`controller/channel-billing.go` 的 `getKimiCodingPlanBalance`，自定义部署）：识别 Kimi 套餐响应形状（`usage.remaining` + `usages.limit_7d`），把**周限额剩余量**写入渠道数字余额；余额弹窗直接显示数字而非原始 JSON。
- **周额度耗尽（remaining ≤ 0）时定时任务会自动禁用渠道**（"余额不足"）；5 小时窗口不参与余额语义（重置太快，避免误禁用）。
- **auto-ban 不会自动恢复**：周额度重置后需在渠道页面手动重新启用（定时任务会跳过已禁用渠道）。
- 非 Kimi 形状的响应仍走原始 JSON 展示逻辑，行为不变。
- **余额显示为百分比**（2026-09-23 前端定制，`web/src/features/channels/lib/channel-utils.ts` 的 `formatChannelBalance`）：type=58 且 base_url 指向 `api.kimi.com/coding` 的渠道，余额是周配额点数（0-100），前端直接显示 `63%`；其他渠道仍按货币显示（USD→CNY 汇率换算）。判断函数 `isKimiCodingPlanChannel`，新增同类套餐渠道时 base_url 必须包含该域名才会走百分比显示。
- **余额弹窗显示双配额窗口**（2026-09-23，commit ed94986ba）：点"更新余额"后，后端除数字余额外还返回 `plan_usage`（`usages.limit_5h` / `limit_7d` 的 used_ratio 与 reset_time），弹窗渲染"5 小时窗口 / 周窗口"两张卡片（已用 % 进度条 + 重置时间）。Kimi 套餐只有这两个窗口，**接口无月额度数据**；`booster_wallet.monthlyChargeLimit` 是充值钱包的月充值上限（当前钱包禁用），不是套餐配额。
- 注意：DeepSeek 渠道（ID 5，type=43）余额查询原生可用，实测返回 CNY 余额并按美元汇率设置换算；按量计费 Kimi（`api.moonshot.cn`）需用 **Moonshot 渠道类型**，其余额端点 `/v1/users/me/balance` 与套餐端完全不同。
