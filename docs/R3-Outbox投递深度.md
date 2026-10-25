# R3 Outbox 投递深度 · 认领/租约/owner · 实现与验收文档

> 需求出处：差距文档 R3（本地消息表 + 投递任务，RocketMQ 事务消息的 RabbitMQ 等价实现）
> 代码：`pkg/outbox/publisher.go` · 测试：`publisher_lease_test.go` · 迁移：`000005_outbox_lease`

---

## 一、为什么需要认领/租约/owner（追问八全套）

原有投递循环「SELECT PENDING → 逐条发送 → UPDATE SENT」在多实例部署下有两个致命窗口：

1. **重复发送**：两个实例同一时刻 SELECT 到同一行，各自发送一次（无认领机制）。
2. **状态卡死**：发送成功后进程崩溃、未及标记 SENT，该行永远是 PENDING/CLAIMED，
   若只靠 `FOR UPDATE` 长事务又会把网络调用（Publish）包进 DB 事务——锁持有放大。

补齐后的投递状态机：

```
PENDING ──CAS 认领（status=PENDING→CLAIMED, owner=me, lease_until=+30s）──► CLAIMED
   ▲                                                                          │
   │                                                          Publish 成功    │ Publish 失败
   │◄── 回退（owner 校验：retry+1, owner=''）──────────────────────┘◄────────────┘
   │                                                                          ▼
   └──租约过期（lease_until < NOW，持有者疑似崩溃）──重新进入候选──► 最终 SENT（owner 校验）
```

三件套语义：

| 机制 | SQL | 防的问题 |
|------|-----|---------|
| CAS 认领 | `UPDATE ... SET status='CLAIMED', owner=?, lease_until=NOW()+30s WHERE id=? AND (status='PENDING' OR (status='CLAIMED' AND lease_until<NOW()))` | 多实例重复认领同一行；affected=0 即被抢先，直接跳过 |
| 租约 | `lease_until` 过期的 CLAIMED 重新进入候选 | 持有者崩溃后消息不再卡死（接受重发，交给消费幂等吸收） |
| owner 校验 | 标记 SENT / 回退 PENDING 均带 `AND owner=? AND status='CLAIMED'` | 慢实例在租约过期被抢后，不得覆盖新持有者的状态 |

## 二、与 R2 消费幂等的组合（核心演示）

**杀掉投递线程制造「发送成功未标记」窗口 → 重启重投 → 下游幂等吸收：**

单测 `TestProcessPending_CrashAfterPublish_RedeliveryAbsorbed` 完整重放该链路：

1. 轮 1：instance-A 认领 → Publish 成功 → 标记 SENT 前进程被杀（mock 返回 kill）；
2. 轮 2（重启）：该行 `CLAIMED + lease_until 已过期` → 重新命中候选 → 重投成功；
3. 断言 fake publisher 共收到 **2 次**同一条消息——MQ 侧 at-least-once 成立；
4. 下游第二次收到时由 `consumed_messages(consumer_group, message_id)` 唯一键吸收
   （R2 的 `pkg/idemp.ConsumeRecorder`，见其测试）。

**结论**：at-least-once 发送 + 幂等消费 = 业务恰好一次（限边界：消费处理到一半崩溃时由下游业务级唯一索引兜底）。

## 三、与现有 order-service 投递的关系（诚实红线）

| 真实经历（RocketMQ 事务消息） | 本练习仓 |
|---|---|
| 半消息 + 回查，事务语义由 MQ 中间件保证 | RabbitMQ 无事务语义 → 本地消息表（Outbox）等价实现——「按既有栈选方案」 |
| order-service 现役投递循环 | `FOR UPDATE` 批处理版（`internal/order/service.go publishOutboxBatch`），单实例安全 |
| 本组件（pkg/outbox） | 认领/租约/owner 完整体，可多实例部署；order-service 迁移到本组件属后续工作 |

## 四、测试矩阵

- `TestProcessPending_ClaimAndSendWithOwnerCheck`：认领 → 发送 → SENT 带 owner 校验
- `TestProcessPending_LostCASRace`：CAS affected=0（他实例抢先）→ 不发送
- `TestProcessPending_PublishFailRevertsToPending`：失败回退 PENDING（retry+1 + owner 校验）
- `TestProcessPending_CrashAfterPublish_RedeliveryAbsorbed`：R3-2 崩溃重投演示（两轮状态断言）

## 五、面试话术锚

「本地消息表把发送动作变成可重投记录；投递任务用 CAS 认领 + 租约 + owner 校验
保证多实例不重不漏；发送侧 at-least-once，配合消费端幂等表，业务效果恰好一次（限边界）。」
