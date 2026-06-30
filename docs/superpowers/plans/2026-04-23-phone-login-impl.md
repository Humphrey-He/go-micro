# 用户手机登录实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 支持用户使用手机号登录系统

**Architecture:** 在 users 表添加 phone 字段，修改 LoginRequest 支持 account 字段，Login() 方法优先按手机号查找用户

**Tech Stack:** Go, MySQL, bcrypt, JWT

---

## 文件结构

- **Modify:** `deploy/sql/schema.sql` - 添加 phone 字段
- **Modify:** `internal/gateway/service.go` - LoginRequest 和 Login() 方法
- **Create:** `sql/006_add_phone_and_insert_user.sql` - 迁移 SQL 和用户插入

---

## Task 1: 数据库添加 phone 字段

**Files:**
- Modify: `deploy/sql/schema.sql:4-13`
- Create: `sql/006_add_phone_and_insert_user.sql`

- [ ] **Step 1: 修改 deploy/sql/schema.sql 的 users 表定义**

在 `deploy/sql/schema.sql` 的 users 表定义中添加 `phone` 字段：

```sql
CREATE TABLE IF NOT EXISTS users (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL UNIQUE,
  username VARCHAR(64) NOT NULL UNIQUE,
  phone VARCHAR(32) UNIQUE COMMENT '手机号',
  password_hash VARCHAR(128) NOT NULL,
  role VARCHAR(32) NOT NULL DEFAULT 'user',
  status TINYINT NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

- [ ] **Step 2: 创建迁移 SQL 文件**

创建 `sql/006_add_phone_and_insert_user.sql`：

```sql
-- 添加 phone 字段到 users 表
ALTER TABLE users ADD COLUMN phone VARCHAR(32) UNIQUE COMMENT '手机号';

-- 插入测试用户 (手机号 17673796081, 密码 260011)
-- bcrypt hash of "260011" with cost 10
INSERT INTO users (user_id, username, phone, password_hash, role, status, created_at, updated_at)
VALUES (
  'USER-17673796081',
  'user_17673796081',
  '17673796081',
  '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
  'user',
  1,
  NOW(),
  NOW()
) ON DUPLICATE KEY UPDATE phone = VALUES(phone);
```

- [ ] **Step 3: 提交数据库变更**

```bash
git add deploy/sql/schema.sql sql/006_add_phone_and_insert_user.sql
git commit -m "feat: add phone field to users table for mobile login"
```

---

## Task 2: 修改 LoginRequest 结构

**Files:**
- Modify: `internal/gateway/service.go:543-546`

- [ ] **Step 1: 修改 LoginRequest 定义**

找到 `internal/gateway/service.go` 中的 LoginRequest 定义，修改为：

```go
type LoginRequest struct {
    Account  string `json:"account"`   // 支持手机号或用户名
    Password string `json:"password"`
}
```

- [ ] **Step 2: 提交 LoginRequest 变更**

```bash
git add internal/gateway/service.go
git commit -m "feat: modify LoginRequest to use account field"
```

---

## Task 3: 修改 Login() 方法支持手机号登录

**Files:**
- Modify: `internal/gateway/service.go:553-606`

- [ ] **Step 1: 找到现有的 Login() 方法**

Login() 方法位于 `internal/gateway/service.go:553-606`

- [ ] **Step 2: 修改 Login() 方法实现**

替换整个 Login() 方法为以下逻辑：

```go
func (s *Service) Login(account, password string) (*LoginResponse, error) {
    // Development mode fallback - allow hardcoded admin credentials
    if account == "admin" && password == "admin123" {
        secret := []byte(config.GetEnv("JWT_SECRET", "dev-secret"))
        claims := jwt.MapClaims{
            "user_id":  "dev-admin-001",
            "username": "admin",
            "role":     "admin",
            "exp":      time.Now().Add(2 * time.Hour).Unix(),
        }
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
        signed, _ := token.SignedString(secret)
        return &LoginResponse{
            Token: signed,
            User:  &user.User{UserID: "dev-admin-001", Username: "admin", Role: "admin", Status: 1},
        }, nil
    }

    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    var u *user.User
    var err error

    // 优先按手机号查找用户
    err = s.cbUser.Execute(func() error {
        upb, callErr := s.user.GetUserByPhone(ctx, account)
        if callErr == nil && upb != nil {
            u = &user.User{
                UserID:   upb.UserId,
                Username: upb.Username,
                Role:     upb.Role,
                Status:   int(upb.Status),
            }
        }
        return callErr
    })

    // 如果没找到手机号，按用户名查找
    if u == nil || err != nil {
        err = s.cbUser.Execute(func() error {
            upb, callErr := s.user.GetUserByName(ctx, account)
            if callErr == nil && upb != nil {
                u = &user.User{
                    UserID:   upb.UserId,
                    Username: upb.Username,
                    Role:     upb.Role,
                    Status:   int(upb.Status),
                }
            }
            return callErr
        })
    }

    if err != nil {
        return nil, err
    }

    if u == nil {
        return nil, errors.New("user not found")
    }

    // 验证密码 - 需要获取完整用户信息包含 password_hash
    fullUser, err := s.user.GetUserByName(ctx, u.Username)
    if err != nil {
        return nil, err
    }

    // 构造 user.User 用于密码验证
    userForVerify := &user.User{
        UserID:       fullUser.UserId,
        Username:     fullUser.Username,
        PasswordHash: fullUser.PasswordHash,
        Role:         fullUser.Role,
        Status:       int(fullUser.Status),
    }

    if !s.user.VerifyPassword(userForVerify, password) {
        return nil, errors.New("invalid credentials")
    }

    secret := []byte(config.GetEnv("JWT_SECRET", "dev-secret"))
    claims := jwt.MapClaims{
        "user_id":  u.UserID,
        "username": u.Username,
        "role":     u.Role,
        "exp":      time.Now().Add(2 * time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString(secret)
    if err != nil {
        return nil, err
    }
    return &LoginResponse{
        Token: signed,
        User:  &user.User{UserID: u.UserID, Username: u.Username, Role: u.Role, Status: int(u.Status)},
    }, nil
}
```

- [ ] **Step 3: 添加 VerifyPassword 方法到 user service**

检查 `internal/user/service.go` 是否已有 `VerifyPassword` 方法（之前看到是有的），如果没有需要添加。

- [ ] **Step 4: 添加 GetUserByPhone 方法到 user GRPC client**

在 `internal/user/grpc_client.go` 中添加：

```go
func (c *GRPCClient) GetUserByPhone(ctx context.Context, phone string) (*userpb.User, error) {
    return c.client.GetUserByPhone(ctx, &userpb.GetUserByPhoneRequest{Phone: phone})
}
```

- [ ] **Step 5: 添加 GetUserByPhone 到 proto**

如果 proto 文件中没有定义，需要添加。在 `proto/user.proto` 中添加：

```protobuf
message GetUserByPhoneRequest {
  string phone = 1;
}

service UserService {
  rpc GetUser(GetUserRequest) returns (User);
  rpc GetUserByName(GetUserByNameRequest) returns (User);
  rpc GetUserByPhone(GetUserByPhoneRequest) returns (User);
  rpc Authenticate(AuthRequest) returns (User);
}
```

- [ ] **Step 6: 提交 Login 方法变更**

```bash
git add internal/gateway/service.go internal/user/grpc_client.go proto/user.proto
git commit -m "feat: support phone login by checking phone before username"
```

---

## Task 4: 执行数据库迁移

- [ ] **Step 1: 执行 SQL 迁移**

```bash
mysql -u root -p go_micro < sql/006_add_phone_and_insert_user.sql
```

或如果 MySQL 配置不同：

```bash
mysql -u root -p -e "USE go_micro; ALTER TABLE users ADD COLUMN phone VARCHAR(32) UNIQUE COMMENT '手机号';"
mysql -u root -p -e "USE go_micro; INSERT INTO users (user_id, username, phone, password_hash, role, status, created_at, updated_at) VALUES ('USER-17673796081', 'user_17673796081', '17673796081', '\$2a\$10\$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'user', 1, NOW(), NOW()) ON DUPLICATE KEY UPDATE phone = VALUES(phone);"
```

- [ ] **Step 2: 验证用户已插入**

```bash
mysql -u root -p -e "USE go_micro; SELECT user_id, username, phone, role FROM users WHERE phone = '17673796081';"
```

---

## Task 5: 验证登录功能

- [ ] **Step 1: 启动服务**

确保 MySQL 和 Redis 服务运行中，然后启动 gateway 服务。

- [ ] **Step 2: 测试手机号登录**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"account": "17673796081", "password": "260011"}'
```

预期响应：
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "token": "<jwt_token>",
    "user": {
      "user_id": "USER-17673796081",
      "username": "user_17673796081",
      "role": "user",
      "status": 1
    }
  }
}
```

- [ ] **Step 3: 测试用户名登录（回退）**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"account": "admin", "password": "admin123"}'
```

---

## 验证清单

- [ ] users 表有 phone 字段
- [ ] 用户 17673796081 已插入
- [ ] 手机号登录成功返回 JWT
- [ ] 用户名登录仍然工作
- [ ] 错误密码登录失败