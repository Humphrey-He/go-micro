# 运营看板时间周期筛选 & 通知模块设计文档

**日期：** 2026/04/22
**状态：** 已批准
**版本：** v1.0

---

## 1. 概述

本文档描述两个独立模块的设计方案：

1. **运营看板时间周期筛选** — 支持快速选择（今日/本周/本月/本年）和时间维度切换（天/周/月）
2. **通知模块** — 实时 WebSocket 推送 + 邮件通知，支持关键指标告警和定时报告

两个模块独立开发，可分别上线。

---

## 2. 模块一：运营看板时间周期筛选

### 2.1 需求

- 支持快速时间选择：今日、本周、本月、本年
- 支持时间维度切换：天（day）、周（week）、月（month）
- 用户可自由筛选统计数据的周期

### 2.2 架构改动

```
前端 (DashboardPage)
├── 时间范围选择器
│   ├── 快捷按钮组：今日 | 本周 | 本月 | 本年
│   └── Segmented 维度切换：天 | 周 | 月
└── API 调用携带参数：
    ├── start_time: 开始时间戳（秒）
    ├── end_time: 结束时间戳（秒）
    └── period: day | week | month

后端
└── GET /api/v1/admin/dashboard/stats
    ├── Query 参数：start_time, end_time, period
    └── 返回对应周期的统计数据
```

### 2.3 改动文件清单

| 层级 | 文件路径 | 改动内容 |
|------|----------|----------|
| 前端 | `order-admin/src/features/dashboard/dashboardApi.ts` | `getDashboardStats(params)` 增加可选参数 |
| 前端 | `order-admin/src/features/dashboard/hooks/useDashboardStats.ts` | 支持传入时间参数 |
| 前端 | `order-admin/src/pages/Dashboard/index.tsx` | 增加时间选择 UI（快捷按钮 + Segmented） |
| 后端 | `internal/gateway/handler.go` | `dashboardStats` 解析 start_time, end_time, period 参数 |
| 后端 | `internal/gateway/service.go` | `GetDashboardStats` 接收参数并传给 order.List |

### 2.4 API 设计

**请求：**
```
GET /api/v1/admin/dashboard/stats
Query Parameters:
  - start_time: int64 (可选，默认今日0点时间戳)
  - end_time: int64   (可选，默认当前时间戳)
  - period: string    (可选，day/week/month，默认 day)
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "today_order_count": 123,
    "today_order_amount": 4567890,
    "pending_refund_count": 5,
    "payment_success_rate": 98.5,
    "low_stock_sku_count": 3
  }
}
```

### 2.5 前端组件设计

**时间选择快捷按钮组：**
- 使用 Ant Design `Space` + `Button` 组件
- 当前选中状态高亮（primary color）
- 支持键盘快捷切换

**维度切换：**
- 使用 Ant Design `Segmented` 组件
- 选项：天 / 周 / 月
- 切换后自动刷新统计数据

**数据展示区域：**
- 保持现有统计卡片布局
- 增加「数据周期」标签显示当前查询范围
- 加载状态使用 Spin 组件

### 2.6 后端逻辑

```go
// service.go GetDashboardStats 伪代码
func (s *Service) GetDashboardStats(startTime, endTime int64, period string) {
    // 1. 如果未提供时间，使用默认值（今日0点 ~ 现在）
    // 2. 根据 period 确定时间分桶策略
    // 3. 调用 order.List 获取订单数据
    // 4. 按 period 聚合统计
    // 5. 补充退款数量、库存预警等数据
    // 6. 返回聚合结果
}
```

---

## 3. 模块二：通知模块

### 3.1 需求

- 关键指标告警通知：退款告警、库存告警、支付失败告警
- 定时报告：每日报告、每周报告
- 通知方式：系统内通知（WebSocket）+ 邮件

### 3.2 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                      通知模块架构                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐    ┌──────────┐    ┌──────────────────────┐ │
│  │  触发源   │───▶│ 通知服务  │───▶│ 1. WebSocket 实时推送  │ │
│  │          │    │          │    │ 2. 邮件发送           │ │
│  │ - 退款   │    │          │    │ 3. 存储到数据库        │ │
│  │ - 库存   │    │          │    │                      │ │
│  │ - 支付   │    │          │    │                      │ │
│  └──────────┘    └──────────┘    └──────────────────────┘ │
│                                                             │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                   前端通知中心                            │ │
│  │  ┌─────┐  ┌─────────────┐  ┌──────────────────────┐   │ │
│  │  │ 🔔  │  │ 通知列表     │  │ 通知设置              │   │ │
│  │  │ 徽标 │  │ - 历史记录   │  │ - 邮件开关            │   │ │
│  │  └─────┘  │ - 已读/未读  │  │ - 告警阈值            │   │ │
│  │           └─────────────┘  └──────────────────────┘   │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 3.3 数据库设计

```sql
-- 通知记录表
CREATE TABLE notifications (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(64) NOT NULL,
    type VARCHAR(32) NOT NULL,
    title VARCHAR(128) NOT NULL,
    content TEXT,
    is_read TINYINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_read (user_id, is_read),
    INDEX idx_created (created_at)
);

-- 通知订阅配置表
CREATE TABLE notification_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(64) NOT NULL,
    type VARCHAR(32) NOT NULL,
    email_enabled TINYINT DEFAULT 1,
    push_enabled TINYINT DEFAULT 1,
    threshold INT DEFAULT 0,
    UNIQUE INDEX idx_user_type (user_id, type)
);
```

### 3.4 通知类型定义

| type | 名称 | 触发条件 | 通知方式 |
|------|------|----------|----------|
| `refund_pending` | 退款待处理告警 | 待处理退款 > threshold | WebSocket + 邮件 |
| `low_stock` | 库存告警 | SKU 库存 < threshold | WebSocket + 邮件 |
| `payment_failed` | 支付失败告警 | 支付失败率 > threshold | WebSocket + 邮件 |
| `daily_report` | 每日报告 | 每天早上 9:00 | 邮件 |
| `weekly_report` | 每周报告 | 每周一早上 9:00 | 邮件 |

### 3.5 后端服务设计

**新增文件：**
- `internal/notification/service.go` — 通知服务主逻辑
- `internal/notification/handler.go` — HTTP handler
- `internal/notification/websocket.go` — WebSocket 处理
- `internal/notification/email.go` — 邮件发送
- `internal/notification/cron.go` — 定时任务

**API 接口：**
| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/notifications` | 获取通知列表 |
| GET | `/api/v1/notifications/unread-count` | 获取未读数 |
| PUT | `/api/v1/notifications/:id/read` | 标记已读 |
| PUT | `/api/v1/notifications/read-all` | 全部已读 |
| GET | `/api/v1/notification/configs` | 获取通知配置 |
| PUT | `/api/v1/notification/configs` | 更新通知配置 |
| WS | `/ws/notifications` | WebSocket 连接 |

**WebSocket 消息格式：**
```json
{
  "type": "notification",
  "data": {
    "id": 1,
    "title": "库存告警",
    "content": "SKU-001 库存不足 10 件",
    "created_at": "2026-04-22T09:00:00Z"
  }
}
```

### 3.6 前端组件设计

**通知铃铛组件 (NotificationBell)：**
- 位置：页面右上角 Header 区域
- 功能：显示未读通知数量徽标（Badge），点击展开通知列表
- 技术：Ant Design Badge + Popover + List

**通知列表页面 (NotificationCenter)：**
- 路由：`/notifications`
- 功能：展示所有通知，支持筛选（全部/未读/类型）、分页
- 技术：Ant Design List + Tabs + Pagination

**通知设置页面 (NotificationSettings)：**
- 路由：`/notifications/settings`
- 功能：配置各类型通知的开关、阈值
- 技术：Ant Design Form + Switch + InputNumber

### 3.7 邮件模板

**每日/每周报告邮件：**
```
标题：【运营报告】2026年4月22日 数据汇总

内容：
- 今日/本周订单数：XXX
- 今日/本周成交额：¥XXX
- 待处理退款：X 件
- 支付成功率：XX%
- 库存预警：X 个 SKU
```

### 3.8 定时任务设计

| 任务 | 表达式 | 说明 |
|------|--------|------|
| 检查退款告警 | `*/5 * * * *` | 每5分钟检查一次 |
| 检查库存告警 | `*/5 * * * *` | 每5分钟检查一次 |
| 检查支付失败告警 | `*/5 * * * *` | 每5分钟检查一次 |
| 发送每日报告 | `0 9 * * *` | 每天早上9点 |
| 发送每周报告 | `0 9 * * 1` | 每周一早上9点 |

---

## 4. 实施计划

### 阶段一：看板时间筛选（独立上线）
1. 前端：添加时间选择 UI 组件
2. 后端：扩展 dashboardStats 接口
3. 联调测试

### 阶段二：通知模块后端（独立上线）
1. 数据库表创建
2. 通知服务 CRUD
3. WebSocket 服务
4. 邮件服务
5. 定时任务

### 阶段三：通知模块前端（独立上线）
1. NotificationBell 组件
2. 通知列表页面
3. 通知设置页面
4. 与 WebSocket 对接

---

## 5. 风险与注意事项

1. **WebSocket 连接管理**：需要处理断线重连、心跳检测
2. **邮件发送失败**：需要重试机制和失败日志
3. **通知性能**：高并发下注意通知队列处理
4. **向后兼容**：dashboardStats 旧接口参数兼容

---

## 6. 验收标准

### 模块一验收：
- [ ] 可选择今日/本周/本月/本年
- [ ] 可切换天/周/月维度
- [ ] 统计数据根据选择的时间范围返回正确数据

### 模块二验收：
- [ ] 铃铛图标显示未读数量
- [ ] 点击可展开通知列表
- [ ] WebSocket 实时接收新通知
- [ ] 可配置各类型通知的邮件开关
- [ ] 定时任务正确触发并发送邮件
