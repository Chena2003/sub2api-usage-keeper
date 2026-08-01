# Models.dev 官方模型价格接入设计

**日期：** 2026-08-01  
**项目：** `sub2api-usage-keeper`  
**状态：** 已获用户批准

## 1. 背景与目标

设置页当前的 `ModelPricingReferenceCard` 在前端硬编码少量模型价格。该数据会随厂商调价和模型迭代而过时，也不能表达 models.dev 已支持的长上下文分层价格。

本次改造以 models.dev 的 `https://models.dev/api.json` 为上游数据源，通过 Usage Keeper 后端获取、校验、筛选和缓存，仅向前端提供当前项目关心的官方一方厂商价格。

本功能是“官方价格参考”，不尝试识别某次 Sub2API 请求实际经过的托管或转售供应商，也不用于重算已有 usage 账单。

## 2. 范围

### 2.1 纳入的官方 Provider

仅纳入以下 models.dev Provider ID：

| 厂商 | Provider ID |
|---|---|
| Anthropic | `anthropic` |
| OpenAI | `openai` |
| Google Gemini | `google` |
| xAI | `xai` |
| DeepSeek | `deepseek` |
| Mistral | `mistral` |
| Moonshot / Kimi | `moonshotai` |
| Alibaba / Qwen | `alibaba` |

官方 Provider 白名单由 Usage Keeper 维护，以保证选择规则明确、可测试，不依赖模型名称猜测。

### 2.2 明确排除

- OpenRouter
- Amazon Bedrock
- Azure OpenAI
- Google Vertex AI
- Google Vertex Anthropic
- 各类 Coding Plan、Token Plan、地区镜像或第三方推理平台
- 非上述白名单中的其他 models.dev Provider
- 根据最低价、平均价或随机规则选择第三方价格
- 根据 `usage_logs` 推断实际供应商
- 使用该价格重新计算 Sub2API 已记录的 `total_cost` 或 `actual_cost`

## 3. 上游数据选择

使用 models.dev 的 `/api.json`，因为价格属于 Provider 的服务属性。

不使用 `/models.json` 作为价格来源；该接口提供与供应商无关的模型元数据，不承载本功能需要的官方服务价格。

models.dev 的成本字段单位为美元/百万 Token，主要字段为：

- `cost.input`
- `cost.output`
- `cost.reasoning`（可选）
- `cost.cache_read`（可选）
- `cost.cache_write`（可选）
- `cost.input_audio`（可选）
- `cost.output_audio`（可选）
- `cost.tiers`（可选）

第一版表格继续聚焦当前已有的文本 Token 价格列：输入、输出、缓存读取和缓存写入。后端模型保留可扩展空间，但不把缺失的可选价格解释为零。

## 4. 系统架构

```text
https://models.dev/api.json
        |
        v
Models.dev HTTP Client
  - timeout
  - response size limit
  - JSON validation
        |
        v
Official Pricing Service
  - official provider allowlist
  - model/cost normalization
  - deterministic sorting
  - 24h cache
  - stale-on-error fallback
        |
        v
GET /api/v1/sub2api/model-pricing
        |
        v
ModelPricingReferenceCard
  - loading/error/stale states
  - provider filter
  - search
  - base and tiered prices
```

浏览器不直接请求 models.dev，以避免 CORS、第三方可用性和响应结构直接泄漏到 UI。

## 5. 后端设计

### 5.1 上游客户端

新增一个边界清晰的 models.dev 客户端：

- 默认 URL：`https://models.dev/api.json`
- 使用 10 秒超时的 `http.Client`
- 只接受成功 HTTP 状态
- 将响应体限制为 16 MiB
- 解码本功能所需的最小字段，不依赖完整上游 schema
- 上游 URL 允许通过配置覆盖，便于测试和自托管镜像

客户端不记录完整响应体，避免无意义的大日志。

### 5.2 归一化模型

向浏览器返回的 Provider：

```go
type OfficialPricingProvider struct {
    ID     string                 `json:"id"`
    Name   string                 `json:"name"`
    Models []OfficialPricingModel `json:"models"`
}
```

模型结构：

```go
type OfficialPricingModel struct {
    ID     string              `json:"id"`
    Name   string              `json:"name"`
    Status string              `json:"status,omitempty"`
    Cost   OfficialModelCost   `json:"cost"`
    Tiers  []OfficialCostTier  `json:"tiers"`
}
```

价格使用指针或可空字段表达“上游未提供”，不得把缺失值转换为 `0`：

```go
type OfficialModelCost struct {
    Input      *float64 `json:"input"`
    Output     *float64 `json:"output"`
    CacheRead  *float64 `json:"cacheRead"`
    CacheWrite *float64 `json:"cacheWrite"`
}
```

只保留：

- Provider 在官方白名单内；
- Model 存在 `cost`；
- Model ID 非空；
- 输入、输出或缓存价格中至少有一个可展示值。

模型版本不做别名合并；例如带日期的模型 ID 与无日期别名分别展示，以忠实反映 models.dev 数据。

### 5.3 分层价格

支持 models.dev 的 `cost.tiers`：

- 记录 tier 类型和上下文阈值；
- 保存该 tier 的输入、输出、缓存读取和缓存写入价格；
- 基础价格仍显示在模型主行；
- UI 在主行下展示长上下文 tier；
- 不重复消费兼容字段 `context_over_200k`，优先以 `tiers` 为准，避免同一 tier 重复。

### 5.4 缓存

服务使用并发安全的进程内缓存：

- TTL：24 小时；
- 缓存有效时直接返回；
- 缓存过期后触发一次刷新；
- 并发请求共享同一次刷新，避免请求风暴；
- 刷新成功后原子替换缓存；
- 刷新失败且有旧缓存时返回旧缓存，并设置 `stale: true`；
- 冷启动且上游不可用时返回 `503 Service Unavailable`；
- 不返回伪造或当前前端硬编码的价格作为降级数据。

第一版不引入 SQLite 持久化缓存。服务重启后重新从 models.dev 获取；后续如果生产网络稳定性证明有需要，再单独增加持久化快照。

### 5.5 API

新增：

```text
GET /api/v1/sub2api/model-pricing
```

响应：

```json
{
  "providers": [
    {
      "id": "anthropic",
      "name": "Anthropic",
      "models": [
        {
          "id": "claude-sonnet-4-6",
          "name": "Claude Sonnet 4.6",
          "cost": {
            "input": 3,
            "output": 15,
            "cacheRead": 0.3,
            "cacheWrite": 3.75
          },
          "tiers": []
        }
      ]
    }
  ],
  "fetchedAt": "2026-08-01T00:00:00Z",
  "stale": false
}
```

API 只暴露 models.dev 的公开模型元数据和价格，不包含 Sub2API 凭据、账号、用户或请求信息。

## 6. 前端设计

`ModelPricingReferenceCard` 删除硬编码 `PRICING_DATA`，改为调用本地 API。

### 6.1 交互

- 默认展示全部 8 个官方厂商；
- 提供 Provider 筛选器，但只包含这 8 个官方厂商；
- 搜索匹配模型 ID、模型名称和官方 Provider 名称；
- 保留价格单位提示：美元/百万 Token；
- 展示数据获取时间；
- stale 数据显示非阻断式警告；
- 请求失败且无数据时显示错误状态和重试按钮；
- 搜索无结果时显示空状态；
- Deprecated 模型保留并显示状态标识，支持历史模型查询。

### 6.2 表格

列：

1. Provider
2. Model ID
3. Display Name
4. Input
5. Output
6. Cache Read
7. Cache Write

缺失价格显示 `—`。真实的零价格显示 `$0`，避免与缺失数据混淆。

存在分层价格时，在对应模型下显示上下文阈值和 tier 价格，不把 tier 价格覆盖到基础价格上。

### 6.3 响应式和可访问性

- 窄屏允许表格横向滚动；
- 搜索、Provider 筛选和重试按钮具备可访问名称；
- Loading、错误和 stale 状态不能只依赖颜色表达；
- 保持现有 Panel 和设计 token，不在本任务内重做整个 Settings 页面布局。

## 7. 配置

新增可选配置：

- `MODELS_DEV_API_URL`，默认 `https://models.dev/api.json`
- `MODELS_DEV_CACHE_TTL`，默认 `24h`
- `MODELS_DEV_HTTP_TIMEOUT`，默认 `10s`

配置会补充到 `.env.example` 和相关维护文档，但不记录任何密钥；models.dev 公共 API 不需要认证。

## 8. 错误处理和可观测性

后端日志记录：

- 刷新成功及归一化后的 Provider/Model 数量；
- 上游 HTTP、解码和校验失败的摘要；
- stale fallback 是否启用。

不记录完整上游 payload。

前端区分：

- 首次加载；
- 成功；
- 成功但 stale；
- 无缓存且加载失败；
- 搜索或筛选无结果。

## 9. 测试策略

遵循 TDD，先写失败测试再实现。

### 9.1 Go

- 官方 Provider 被保留；
- 非官方 Provider 被排除；
- 无 cost 模型被排除；
- 缺失价格保持为空；
- 真实零价格保持为零；
- tier 正确解析且不重复；
- Provider 和 Model 排序稳定；
- 缓存命中不重复请求；
- TTL 过期后刷新；
- 并发刷新去重；
- 刷新失败时返回 stale cache；
- 冷启动失败返回服务不可用；
- API handler 状态码和响应结构正确。

### 9.2 React / TypeScript

- API URL 和响应解析；
- 加载成功后展示官方价格；
- Provider 筛选；
- 模型搜索；
- 缺失值与真实零价格格式不同；
- tier 显示；
- stale 提示；
- 错误和重试；
- 不再包含硬编码价格表。

## 10. 验证

实现后执行：

```bash
go -C /Users/laijiachen.1/Documents/sub2api/sub2api-usage-keeper test ./cmd/... ./internal/...
npm --prefix /Users/laijiachen.1/Documents/sub2api/sub2api-usage-keeper/web test -- --run
npm --prefix /Users/laijiachen.1/Documents/sub2api/sub2api-usage-keeper/web run lint
npm --prefix /Users/laijiachen.1/Documents/sub2api/sub2api-usage-keeper/web run typecheck
npm --prefix /Users/laijiachen.1/Documents/sub2api/sub2api-usage-keeper/web run build
make -C /Users/laijiachen.1/Documents/sub2api/sub2api-usage-keeper verify
```

并在前端开发服务器中检查 Settings 页桌面和窄屏布局。

## 11. 非目标

- 修改 `models.dev` 仓库；
- 修改 `cpa-usage-keeper`；
- 修改 Sub2API 的计费逻辑；
- 纠正或覆盖 Sub2API 数据库内已有成本；
- 支持用户手工编辑价格；
- 自动选择第三方最低价；
- 在本任务中引入登录或权限系统。
