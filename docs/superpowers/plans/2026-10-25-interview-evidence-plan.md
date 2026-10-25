# go-micro 面试佐证改造计划（现状核对与落地清单）

> 依据：`D:\面试用\26面试\04_元信息与手册\go-micro差距需求文档_20261025.md`（下称"差距文档"）
> 本文档对差距文档的"现状盘点"逐项核实后，给出修订后的改造清单、困难点与执行顺序。
> 核实时间：2026-10-25（基于 dev 分支 2c8fcac）；P0 执行进度见同目录《2026-10-25-下一轮开发指引.md》

---

## 一、总体结论

**项目完全支持差距文档描述的全部扩展（R1–R9），无架构性障碍。**
服务拆分、pkg 公共库、CI 五工作流、测试基建（sqlmock/miniredis 模式已覆盖全部业务包）都是现成的扩展点。
且差距文档的现状盘点有 **3 处过时**（Outbox、幂等快照、metrics 端点实际已存在）——真实工作量比差距文档预估少约 1/3。

---

## 二、现状核对表（差距文档判断 vs 仓库实际）

| 差距文档缺口项 | 仓库实际状态 | 对改造的影响 |
|---|---|---|
| 本地消息表 Outbox「无实现」 | **已有**：`pkg/outbox/publisher.go`（PENDING/SENT/FAILED/DLQ 状态机 + Prometheus 指标）+ order-service 的 `order_outbox` 表与 `publishOutboxBatch` | R3 从"从零实现"降为"补深度"：CAS 认领/租约/owner 校验细节 + 杀投递线程演示脚本（≈1 晚） |
| Prometheus /metrics + RED「无」 | **部分有**：各服务 `/metrics` 端点（promhttp）+ outbox 四项指标已有；但 HTTP RED（请求数/错误/延迟）记录在自建 `MetricsCollector`（内存 JSON），未导出为 Prometheus 指标 | R7 从"迁移模式"降为"改造 middleware"：把 `pkg/middleware/metrics.go` 接到 prometheus CounterVec/HistogramVec（≈半天） |
| 双层幂等「雏形，语义未定义」 | **请求级已实现大半**：`pkg/middleware/idempotency.go` 已有响应快照 + `X-Idempotency-Replayed` + pending 防并发；业务级 `idempotent_key`/`order_no` 唯一索引已有 | R2 聚焦三件事：幂等键从 `X-Idempotency-Key` 统一到 `X-Request-Id`、下游消费幂等表、覆盖/累加语义文档化 + 重放测试 |
| Redis Lua 原子预扣 + 异步落库 + 流水对账 | 确实无。现有库存是 **DB 事务 TCC 预占**（Reserve/Confirm/Release + reserved_id 追踪） | R1 全量开发（≈2 晚），见困难点 D1 的模型取舍 |
| 对账服务 | 确实无（task-service 的 compensation 是补偿不是对账） | R4 全量开发 |
| 渠道映射 + 按店铺限频分发 | 确实无 | R5 全量开发（纯新增服务，无历史包袱） |
| 性能 benchmark 叙事 | 确实无 | R6 全量开发 |

## 三、前置清理核对（差距文档第一节）

| # | 差距文档要求 | 核实结果 | 动作 |
|---|---|---|---|
| 0-1 | README 删红旗句 | 属实（"可直接作为个人面试核心项目"等句在第 7、27 行附近） | 改口径为"学习型实战项目：电商下单履约全链路" |
| 0-2 | English 死链 | 属实（中英链接都指向 ./README.md，README_EN.md 存在未被链接） | 链接改 `./README_EN.md` |
| 0-3 | 补 LICENSE | 属实（无 LICENSE 文件） | 加 MIT（署名与 git user 一致） |
| 0-4 | 模块改名 `ecomm-micro` | 属实（go.mod module 名 `go-micro` 与 asim/go-micro 撞名） | module 名 + 全仓库 import 路径替换（sed 一次完成，见困难点 D3） |
| 0-5 | 本地目录 git 化 | **已过时**：仓库已 clone 至 `D:\GO_Repo\go-micro\go-micro`（真 git 仓库，dev 分支已推远端）；外层空目录无害 | 可选：把内层上移为 `D:\GO_Repo\go-micro` 根；不做也不影响 |

---

## 四、困难点（按真实阻塞程度排序）

**D1. R1 与现有 TCC 预占模型的关系（设计取舍，最需要先想清楚）**
现有 inventory-service 是 DB 事务三段式（Reserve 预占 → Confirm/Release）。R1 要加 Redis Lua 预扣层。
建议**叠加而非替换**：Redis 作为"读判扣记"四步原子的加速层（Lua + hash 流水），DB 保持真账；
异步落库 worker 幂等（流水唯一键），对账 job 做 Redis vs DB 差集。
这样既不动既有 TCC 语义（订单链路回归风险低），又完整复刻经历话术"预扣是加速层不是真账"。
**在 docs 里写明"真实经历 vs 练习仓简化"的差异**（差距文档红线 4）。

**D2. 压测/演示环境（操作性困难）**
R1 验收（1000 并发扣 100 零超卖）、R6 基线压测都需要本机 Docker 起 MySQL/Redis（docker-compose.yml 已有）。
当前本机 C 盘紧张（<500MB），需先清理；另 Go 工具链正在升 1.26（进行中，见第八节）。

**D3. 0-4 模块改名的影响面（机械但面广）**
`go-micro` → `ecomm-micro`：go.mod module 名 + 全部 Go 文件 import（190+ 文件）+ goctl generated 代码 + CI 脚本里的路径引用。
一条 sed 可完成，但必须单独成一个 commit 且在 CI 绿灯后做，方便回滚。

**D4. 提交节奏（差距文档红线 2）**
"每天几笔自然提交"意味着改造不要单日爆发。建议按本清单的颗粒度拆 commit，与 R1-R9 的开发节奏天然对齐。

**D5. 挂起的进行中工作**
CI 修复（gofmt 41 文件 + notification race 修复 + migration 正则 + 依赖升 Go 1.26）在工作区未提交、未验证完。
必须先收尾并让 CI 绿灯，再开始前置清理——否则改名等大改动会叠在未验证状态上。

---

## 五、修订后任务清单

### 第 0 批：收尾进行中工作（先做，≈半天）
- [x] 完成依赖升级验证（x/net v0.60 + amqp091 v1.14 要求 Go 1.26）：CI `GO_VERSION` 升 `1.26.x`，全量测试通过
- [x] 提交 gofmt 41 文件 + notification race 修复 + migration 正则修复 + 依赖升级（一个 commit：`chore: fix CI ...`）
- [x] push dev → CI 全绿 → auto-merge 到 main

### 第 1 批：前置清理（≈半天）
- [x] 0-1 README 口径改学习型项目（中英三份 README 同步）
- [x] 0-2 英文死链修复
- [ ] 0-3 MIT LICENSE
- [ ] 0-4 模块改名 ecomm-micro（单独 commit）

### 第 2 批：P0（5–6 晚，较差距文档省 1–2 晚）
- [x] R1-1 Redis Lua 预扣（叠加层设计，docs 先写差异说明）——先写失败测试（红线 3）
- [x] R1-2 stock-sync 异步落库 worker（流水唯一键幂等）
- [x] R1-3 Redis/DB 流水对账 job + 差异台账表
- [x] R1 验收：1000 并发扣 100 → DB=0、零超卖零少卖、对账零差异报告（压测脚本进仓）
- [x] R2-1 幂等键统一 X-Request-Id + 业务级消费幂等表（消息 ID 唯一约束）
- [x] R2-2 覆盖/累加语义文档（docs/幂等语义.md）
- [x] R2 验收：同一请求 5 次 → 1 单 5 同响应（幂等重放测试用例）
- [x] R3-1 Outbox 投递补深度：CAS 认领（WHERE status='PENDING'）+ 租约 + owner 校验
- [x] R3-2 演示：杀投递线程 → 重启重投 → 下游幂等吸收（演示脚本进仓）
- [x] R4 reconcile-service 最小版：订单 vs 库存流水 diff + reconcile_diff 状态机台账
- [x] R4 验收：手工造差异 → 台账捕获 → 修复关闭

### 第 3 批：P1（4–6 晚）
- [ ] R5-1 渠道映射表 + seed 数据（几千条）
- [ ] R5-2 distribution-service：按店铺令牌桶限频（x/time/rate per shop）+ 超额排队
- [ ] R5-3 15 分钟轮次差异比对兜底 job（分发结果 vs mock 平台回执）
- [ ] R5 验收：3 个 mock 平台打到限频 → 零超限拒绝；diff job 抓漏发
- [ ] R6-1 基线压测（同步链路）→ wrk/vegeta 脚本进仓
- [ ] R6-2 三个优化（同步收窄 / Redis 预扣 / 并行化）→ 优化后压测
- [ ] R6-3 docs/性能优化记录.md + README 贴 before/after（自测真实数字）

### 第 4 批：P2 工程化收尾（穿插）
- [ ] R7 middleware/metrics.go 接 Prometheus（CounterVec/HistogramVec，RED 三件套）
- [ ] R8 三套门面测试：超卖并发 / 幂等重放 / 对账差异（升级为业务语义级）
- [ ] R9 worker 优雅停机 drain

---

## 六、节奏与里程碑（修订）

| 周 | 内容 | 里程碑 |
|---|---|---|
| 本周 | 第 0 批 + 第 1 批 | CI 全绿 + 仓库口径干净 |
| 第 1 周 | R1 + R2 | 超卖演示 + 幂等重放演示可跑 |
| 第 2 周 | R3 + R4 | Outbox 重投演示 + 差异台账 |
| 第 3 周起 | R5 / R6 / R7–R9 | 分发调度器 + 性能叙事 |

R1–R4 完成即达"佐证"标准（与差距文档一致）。

---

## 七、面试话术锚（沿用差距文档第七节，代码落地时同步记录）

| 需求 | 一句话 |
|---|---|
| R1 | 预扣 Lua 四步原子，真账在 DB，对账兜底——预扣是加速层不是真账 |
| R2 | 请求级防重放、业务级防重复，兜底是唯一索引——重放返回首次响应 |
| R3 | 本地消息表把发送变成可重投记录——at-least-once + 幂等消费 = 业务恰好一次（限边界） |
| R4 | 对账不是监控是兜底——差异进台账状态机直到关闭 |
| R5 | 平台限频是硬约束，超限封号不是报错——按店铺令牌桶削峰 + 轮次差异兜底 |
| R6 | 优化前先测基线——数字是自己压的，口径写在文档 |
