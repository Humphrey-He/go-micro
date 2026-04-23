# 用户手机登录方案设计

## 1. 背景

前端登录接口使用 `account` 字段（可代表手机号或用户名），需要支持手机号直接登录。

## 2. 数据库变更

**users 表添加 phone 字段：**

```sql
ALTER TABLE users ADD COLUMN phone VARCHAR(32) UNIQUE COMMENT '手机号';
```

## 3. 后端变更

### LoginRequest 结构

```go
type LoginRequest struct {
    Account  string `json:"account"`   // 支持手机号或用户名
    Password string `json:"password"`
}
```

### Login() 逻辑

1. 优先按手机号查找用户（`phone = account`）
2. 如果没找到，按用户名查找（`username = account`）
3. 使用 bcrypt 验证密码
4. 验证成功后生成 JWT token

## 4. 用户创建

插入测试用户：
- phone: `17673796081`
- password: `260011`（bcrypt 加密存储）

## 5. API 变更

```
POST /api/v1/auth/login
{
  "account": "17673796081",
  "password": "260011"
}

Response:
{
  "code": 0,
  "message": "OK",
  "data": {
    "token": "<jwt_token>",
    "user": {
      "user_id": "...",
      "username": "...",
      "role": "user",
      "status": 1
    }
  }
}
```

## 6. 影响范围

- `internal/gateway/service.go` - LoginRequest 和 Login() 方法
- 数据库 - users 表新增 phone 字段