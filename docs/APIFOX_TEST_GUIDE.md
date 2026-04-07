# Go-Micro 完整测试流程指南

## 项目概述
本文档提供完整的 API 测试流程，涵盖网关、库存、订单、支付服务的端到端测试。

## 环境配置

### 基础 URL
```
http://localhost:8080
```

### 全局变量设置
在 Apifox 中设置以下环境变量：
- `base_url`: http://localhost:8080
- `token`: (登录后自动设置)
- `user_id`: user_001
- `order_id`: (创建订单后自动设置)
- `payment_id`: (创建支付后自动设置)

---

## 测试流程

### 第一步：用户登录

**请求信息**
- 方法: `POST`
- URL: `{{base_url}}/api/v1/auth/login`
- 请求头: `Content-Type: application/json`

**请求体**
```json
{
  "username": "testuser",
  "password": "password123"
}
```

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user_id": "user_001",
    "username": "testuser"
  }
}
```

**后置脚本** (Apifox)
```javascript
// 保存 token 到环境变量
pm.environment.set("token", pm.response.json().data.token);
pm.environment.set("user_id", pm.response.json().data.user_id);
```

---

### 第二步：查询库存

**请求信息**
- 方法: `GET`
- URL: `{{base_url}}/api/v1/inventory/SKU001`
- 请求头: `Authorization: Bearer {{token}}`

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "sku_id": "SKU001",
    "quantity": 100,
    "reserved": 0,
    "available": 100
  }
}
```

---

### 第三步：创建订单

**请求信息**
- 方法: `POST`
- URL: `{{base_url}}/api/v1/orders`
- 请求头:
  - `Authorization: Bearer {{token}}`
  - `Content-Type: application/json`

**请求体**
```json
{
  "request_id": "REQ-{{$timestamp}}",
  "user_id": "{{user_id}}",
  "items": [
    {
      "sku_id": "SKU001",
      "quantity": 5,
      "price": 1000
    }
  ],
  "total_amount": 5000
}
```

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_id": "ORD-20260330-001",
    "biz_no": "BIZ-20260330-001",
    "user_id": "user_001",
    "status": "RESERVED",
    "total_amount": 5000,
    "created_at": "2026-03-30T10:50:00Z"
  }
}
```

**后置脚本**
```javascript
pm.environment.set("order_id", pm.response.json().data.order_id);
pm.environment.set("biz_no", pm.response.json().data.biz_no);
```

---

### 第四步：查询订单详情

**请求信息**
- 方法: `GET`
- URL: `{{base_url}}/api/v1/orders/{{order_id}}`
- 请求头: `Authorization: Bearer {{token}}`

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_id": "ORD-20260330-001",
    "biz_no": "BIZ-20260330-001",
    "user_id": "user_001",
    "status": "RESERVED",
    "total_amount": 5000,
    "items": [
      {
        "sku_id": "SKU001",
        "quantity": 5,
        "price": 1000
      }
    ]
  }
}
```

---

### 第五步：查询订单聚合视图

**请求信息**
- 方法: `GET`
- URL: `{{base_url}}/api/v1/order-views/{{biz_no}}`
- 请求头: `Authorization: Bearer {{token}}`

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_no": "BIZ-20260330-001",
    "view_status": "PENDING",
    "order_status": "RESERVED",
    "task_status": "PENDING",
    "reservation_status": "RESERVED"
  }
}
```

---

### 第六步：创建支付

**请求信息**
- 方法: `POST`
- URL: `{{base_url}}/api/v1/payments`
- 请求头:
  - `Authorization: Bearer {{token}}`
  - `Content-Type: application/json`

**请求体**
```json
{
  "payment_id": "PAY-{{$timestamp}}",
  "order_id": "{{order_id}}",
  "user_id": "{{user_id}}",
  "amount": 5000,
  "payment_method": "credit_card",
  "status": "PENDING"
}
```

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "payment_id": "PAY-1774867800000",
    "order_id": "ORD-20260330-001",
    "amount": 5000,
    "status": "PENDING",
    "created_at": "2026-03-30T10:50:00Z"
  }
}
```

**后置脚本**
```javascript
pm.environment.set("payment_id", pm.response.json().data.payment_id);
```

---

### 第七步：查询支付详情

**请求信息**
- 方法: `GET`
- URL: `{{base_url}}/api/v1/payments/{{payment_id}}`
- 请求头: `Authorization: Bearer {{token}}`

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "payment_id": "PAY-1774867800000",
    "order_id": "ORD-20260330-001",
    "amount": 5000,
    "status": "PENDING",
    "created_at": "2026-03-30T10:50:00Z"
  }
}
```

---

### 第八步：标记支付成功

**请求信息**
- 方法: `POST`
- URL: `{{base_url}}/api/v1/payments/{{payment_id}}/success`
- 请求头:
  - `Authorization: Bearer {{token}}`
  - `Content-Type: application/json`

**请求体**
```json
{
  "transaction_id": "TXN-{{$timestamp}}"
}
```

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "payment_id": "PAY-1774867800000",
    "status": "SUCCESS",
    "updated_at": "2026-03-30T10:50:05Z"
  }
}
```

---

### 第九步：查询订单最终状态

**请求信息**
- 方法: `GET`
- URL: `{{base_url}}/api/v1/order-views/{{biz_no}}`
- 请求头: `Authorization: Bearer {{token}}`

**预期响应**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "order_no": "BIZ-20260330-001",
    "view_status": "SUCCESS",
    "order_status": "SUCCESS",
    "task_status": "SUCCESS",
    "reservation_status": "CONFIRMED"
  }
}
```

---

## 测试场景

### 场景1：正常流程（成功）
1. 登录 → 2. 查询库存 → 3. 创建订单 → 4. 查询订单 → 5. 创建支付 → 6. 标记支付成功 → 7. 查询最终状态

### 场景2：库存不足
- 在第3步创建订单时，修改数量为 200（超过库存 100）
- 预期：返回 409 Conflict，错误信息 "insufficient inventory"

### 场景3：支付失败
- 在第8步，调用 `/payments/{{payment_id}}/failed` 而不是 `/success`
- 预期：订单状态变为 CANCELED

### 场景4：支付超时
- 在第8步，调用 `/payments/{{payment_id}}/timeout`
- 预期：订单状态变为 TIMEOUT

---

## 常见错误处理

| 错误码 | 含义 | 解决方案 |
|--------|------|--------|
| 401 | 未授权 | 检查 token 是否过期，重新登录 |
| 409 | 库存冲突 | 减少订单数量或等待库存补充 |
| 500 | 服务错误 | 检查后端服务是否正常运行 |
| 503 | 服务不可用 | 确保所有微服务已启动 |

---

## 性能测试建议

- 并发用户数：10-50
- 循环次数：100
- 关键指标：平均响应时间 < 500ms，成功率 > 99%

