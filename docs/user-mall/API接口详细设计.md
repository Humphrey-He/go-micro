# 用户端商城 - API 接口详细设计

> 本文档定义用户端商城所有 API 接口，包含请求/响应格式、字段说明、错误码。

---

## 1. 接口规范

### 1.1 基础规范

| 项 | 说明 |
|---|---|
| 协议 | HTTPS |
| 数据格式 | JSON |
| 字符编码 | UTF-8 |
| 认证方式 | JWT Bearer Token |
| 签名 | 不签名（HTTPS 传输） |

### 1.2 通用请求头

```
Content-Type: application/json
Authorization: Bearer {jwt_token}
X-Request-ID: {uuid}          # 请求唯一ID
X-App-Version: 1.0.0          # 客户端版本
X-Device-ID: {device_id}      # 设备ID
X-Platform: ios/android/web   # 平台
```

### 1.3 通用响应格式

```json
{
  "code": 0,
  "message": "OK",
  "data": { },
  "request_id": "uuid",
  "timestamp": 1713443200
}
```

### 1.4 通用错误码

| code | 说明 |
|------|------|
| 0 | 成功 |
| 40001 | 参数错误 |
| 40101 | 未授权（未登录） |
| 40301 | 无权限 |
| 40401 | 资源不存在 |
| 40901 | 业务冲突（如库存不足） |
| 42901 | 请求过于频繁 |
| 50001 | 服务端错误 |

---

## 2. 认证模块 `/api/v1/auth`

### 2.1 发送验证码

```
POST /api/v1/auth/sms/send
```

**请求：**
```json
{
  "phone": "13800138000",
  "type": "login"  // login | register | reset_password | bind_phone
}
```

**响应：**
```json
{
  "code": 0,
  "message": "发送成功",
  "data": {
    "expires_in": 300  // 验证码有效期（秒）
  }
}
```

### 2.2 用户注册

```
POST /api/v1/auth/register
```

**请求：**
```json
{
  "phone": "13800138000",
  "code": "123456",
  "password": "Aa123456",
  "invite_code": "ABC123"  // 可选，邀请码
}
```

**响应：**
```json
{
  "code": 0,
  "message": "注册成功",
  "data": {
    "token": "eyJhbGc...",
    "expires_in": 604800,
    "user": {
      "user_id": "U001",
      "phone": "138****8000",
      "nickname": "用户1234",
      "avatar": "https://...",
      "level": 1,
      "points": 100
    }
  }
}
```

### 2.3 用户登录（密码）

```
POST /api/v1/auth/login
```

**请求：**
```json
{
  "account": "13800138000",  // 手机号或用户名
  "password": "Aa123456",
  "remember": true
}
```

**响应：**
```json
{
  "code": 0,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGc...",
    "expires_in": 604800,
    "user": {
      "user_id": "U001",
      "phone": "138****8000",
      "nickname": "买家商城",
      "avatar": "https://...",
      "level": 2,
      "points": 500,
      "member_title": "银卡会员"
    }
  }
}
```

### 2.4 用户登录（验证码）

```
POST /api/v1/auth/login/by-code
```

**请求：**
```json
{
  "phone": "13800138000",
  "code": "123456"
}
```

**响应：** 同 2.3

### 2.5 社交登录

```
POST /api/v1/auth/social/callback
```

**请求：**
```json
{
  "provider": "wechat",  // wechat | google | apple
  "openid": "oxxxxxxxx",
  "unionid": "xxxxxx",   // 微信特有
  "access_token": "xxxxx",
  "nickname": "微信用户",
  "avatar": "https://..."
}
```

**响应：** 同 2.3（首次登录自动创建账号）

### 2.6 退出登录

```
POST /api/v1/auth/logout
```

**响应：**
```json
{
  "code": 0,
  "message": "退出成功"
}
```

---

## 3. 用户模块 `/api/v1/user`

### 3.1 获取用户信息

```
GET /api/v1/user/profile
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "user_id": "U001",
    "phone": "138****8000",
    "nickname": "买家商城",
    "avatar": "https://cdn.example.com/avatar/001.jpg",
    "email": "user@example.com",
    "gender": "female",  // male | female | unknown
    "birthday": "1990-01-01",
    "level": 3,
    "member_title": "金卡会员",
    "points": 1250,
    "growth_value": 2580,
    "next_level_points": 3000,
    "is_verified": true,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

### 3.2 更新用户信息

```
PUT /api/v1/user/profile
```

**请求：**
```json
{
  "nickname": "新昵称",
  "avatar": "https://...",
  "email": "new@example.com",
  "gender": "male",
  "birthday": "1995-06-20"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "更新成功",
  "data": {
    "user_id": "U001",
    "nickname": "新昵称",
    "avatar": "https://...",
    "email": "new@example.com",
    "gender": "male",
    "birthday": "1995-06-20"
  }
}
```

### 3.3 修改密码

```
POST /api/v1/user/password
```

**请求：**
```json
{
  "old_password": "OldPass123",
  "new_password": "NewPass123"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "密码修改成功"
}
```

---

## 4. 收货地址模块 `/api/v1/user/address`

### 4.1 地址列表

```
GET /api/v1/user/address
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "addresses": [
      {
        "id": "A001",
        "receiver": "张三",
        "phone": "13800138000",
        "province": "广东省",
        "city": "深圳市",
        "district": "南山区",
        "detail": "科技园南路XX号",
        "postal_code": "518000",
        "tag": "home",  // home | company | school | other
        "is_default": true,
        "latitude": 22.543,
        "longitude": 114.057
      },
      {
        "id": "A002",
        "receiver": "李四",
        "phone": "13900139000",
        "province": "广东省",
        "city": "广州市",
        "district": "天河区",
        "detail": "体育西路XX号",
        "postal_code": "510000",
        "tag": "company",
        "is_default": false
      }
    ]
  }
}
```

### 4.2 新增地址

```
POST /api/v1/user/address
```

**请求：**
```json
{
  "receiver": "张三",
  "phone": "13800138000",
  "province": "广东省",
  "city": "深圳市",
  "district": "南山区",
  "detail": "科技园南路XX号",
  "postal_code": "518000",
  "tag": "home",
  "is_default": true,
  "latitude": 22.543,
  "longitude": 114.057
}
```

**响应：**
```json
{
  "code": 0,
  "message": "添加成功",
  "data": {
    "id": "A003",
    "receiver": "张三",
    "phone": "13800138000",
    "province": "广东省",
    "city": "深圳市",
    "district": "南山区",
    "detail": "科技园南路XX号",
    "postal_code": "518000",
    "tag": "home",
    "is_default": true
  }
}
```

### 4.3 更新地址

```
PUT /api/v1/user/address/:id
```

**请求：** 同 4.2

**响应：** 同 4.2

### 4.4 删除地址

```
DELETE /api/v1/user/address/:id
```

**响应：**
```json
{
  "code": 0,
  "message": "删除成功"
}
```

### 4.5 设置默认地址

```
PUT /api/v1/user/address/:id/default
```

**响应：**
```json
{
  "code": 0,
  "message": "设置成功"
}
```

---

## 5. 商品模块 `/api/v1/products`

### 5.1 商品列表

```
GET /api/v1/products
```

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| keyword | string | 否 | 搜索关键词 |
| category_id | string | 否 | 分类ID |
| page | int | 否 | 页码，默认1 |
| page_size | int | 否 | 每页条数，默认20，最大50 |
| sort_by | string | 否 | 排序：sales/new/price_asc/price_desc |
| price_min | int | 否 | 最低价格（分） |
| price_max | int | 否 | 最高价格（分） |
| has_stock | bool | 否 | 仅显示有货 |
| tags | string[] | 否 | 标签，如 ["seckill","new"] |

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "products": [
      {
        "sku_id": "SKU001",
        "title": "2024新款男士运动鞋",
        "subtitle": "轻便透气跑步鞋",
        "images": [
          "https://cdn.example.com/product/001_1.jpg",
          "https://cdn.example.com/product/001_2.jpg"
        ],
        "price": 29900,  // 现价（分）
        "original_price": 49900,  // 原价（分）
        "sales": 12580,
        "stock": 500,
        "tags": ["hot", "new"],
        "rating": 4.8,
        "comment_count": 2340,
        "shop": {
          "shop_id": "S001",
          "name": "运动旗舰店",
          "logo": "https://..."
        }
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 156,
      "total_pages": 8
    },
    "filters": {
      "brands": ["Nike", "Adidas", "安踏"],
      "price_ranges": [
        {"min": 0, "max": 10000},
        {"min": 10000, "max": 30000},
        {"min": 30000, "max": null}
      ],
      "attributes": {
        "color": ["黑色", "白色", "蓝色"],
        "size": ["38", "39", "40", "41", "42"]
      }
    }
  }
}
```

### 5.2 商品详情

```
GET /api/v1/products/:sku_id
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "sku_id": "SKU001",
    "title": "2024新款男士运动鞋",
    "subtitle": "轻便透气跑步鞋",
    "images": [
      "https://cdn.example.com/product/001_1.jpg",
      "https://cdn.example.com/product/001_2.jpg",
      "https://cdn.example.com/product/001_3.jpg"
    ],
    "video_url": "https://cdn.example.com/product/001.mp4",
    "price": 29900,
    "original_price": 49900,
    "stock": 500,
    "sales": 12580,
    "rating": 4.8,
    "comment_count": 2340,
    "views": 45600,
    "favorite_count": 1200,
    "is_favorite": true,
    "tags": ["hot", "new"],
    "min_limit": 1,
    "max_limit": 5,  // 每人限购数量

    "skus": [
      {
        "sku_id": "SKU001-BLK-40",
        "attributes": {"color": "黑色", "size": "40"},
        "stock": 50,
        "price": 29900,
        "image": "https://..."
      },
      {
        "sku_id": "SKU001-BLK-41",
        "attributes": {"color": "黑色", "size": "41"},
        "stock": 0,  // 无库存，不可选
        "price": 29900,
        "image": "https://..."
      }
    ],

    "attributes": [
      {
        "name": "color",
        "values": [
          {"value": "黑色", "image": "https://color_black.jpg"},
          {"value": "白色", "image": "https://color_white.jpg"}
        ]
      },
      {
        "name": "size",
        "values": [
          {"value": "38"},
          {"value": "39"},
          {"value": "40"},
          {"value": "41"},
          {"value": "42"}
        ]
      }
    ],

    "shop": {
      "shop_id": "S001",
      "name": "运动旗舰店",
      "logo": "https://...",
      "rating": 4.9,
      "followers": 125000,
      "is_followed": false,
      "description": "专注运动鞋10年"
    },

    "freight": {
      "template_id": "T001",
      "rule": "首件10元，续件2元",
      "estimated_days": "3-5天"
    },

    "promotion": {
      "type": "seckill",
      "start_time": "2024-03-20T10:00:00Z",
      "end_time": "2024-03-20T12:00:00Z",
      "seckill_price": 19900,
      "seckill_stock": 100,
      "seckill_limit": 1
    },

    "details": {
      "description": "<p>商品图文详情...</p>",
      "specifications": [
        {"name": "品牌", "value": "运动星"},
        {"name": "型号", "value": "A2024"}
      ],
      "aftersale": {
        "type": "7天无理由退换",
        "content": "支持7天无理由退换货"
      }
    }
  }
}
```

### 5.3 商品评价列表

```
GET /api/v1/products/:sku_id/reviews
```

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页条数 |
| filter | string | 否 | 评价筛选：all/good/medium/bad/with_images |

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "reviews": [
      {
        "review_id": "R001",
        "user": {
          "user_id": "U002",
          "nickname": "运动达人",
          "avatar": "https://..."
        },
        "rating": 5,
        "content": "鞋子很轻便，穿着舒适，尺码标准。",
        "images": [
          "https://cdn.example.com/review/001_1.jpg",
          "https://cdn.example.com/review/001_2.jpg"
        ],
        "video_url": null,
        "sku_info": "黑色/41码",
        "is_anonymous": false,
        "like_count": 45,
        "is_liked": false,
        "created_at": "2024-03-15T14:30:00Z",
        "seller_reply": "感谢您的支持，欢迎下次光临！"
      }
    ],
    "summary": {
      "total": 2340,
      "rating": 4.8,
      "distribution": {
        "5": 2000,
        "4": 200,
        "3": 80,
        "2": 30,
        "1": 30
      },
      "with_images_count": 560,
      "with_video_count": 45
    },
    "pagination": {
      "page": 1,
      "page_size": 10,
      "total": 2340
    }
  }
}
```

### 5.4 商品推荐

```
GET /api/v1/products/:sku_id/recommendations
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "also_bought": [
      // 同 5.1 products 格式
    ],
    "also_viewed": [
      // 同 5.1 products 格式
    ],
    "similar": [
      // 同 5.1 products 格式
    ]
  }
}
```

### 5.5 商品搜索建议

```
GET /api/v1/search/suggestions
```

**请求参数：** `q` - 搜索词

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "suggestions": [
      "运动鞋 男",
      "运动鞋 女",
      "运动鞋 跑步"
    ],
    "hot_keywords": [
      "运动鞋",
      "T恤",
      "牛仔裤",
      "连衣裙"
    ],
    "history": [
      "运动鞋",
      "篮球"
    ]
  }
}
```

---

## 6. 分类模块 `/api/v1/categories`

### 6.1 分类列表

```
GET /api/v1/categories
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "categories": [
      {
        "id": "C001",
        "name": "服装",
        "icon": "https://...",
        "children": [
          {
            "id": "C001-1",
            "name": "男装",
            "icon": "https://...",
            "children": [
              {"id": "C001-1-1", "name": "T恤"},
              {"id": "C001-1-2", "name": "裤子"}
            ]
          },
          {
            "id": "C001-2",
            "name": "女装",
            "children": [
              {"id": "C001-2-1", "name": "连衣裙"},
              {"id": "C001-2-2", "name": "牛仔裤"}
            ]
          }
        ]
      },
      {
        "id": "C002",
        "name": "鞋靴",
        "children": [
          {"id": "C002-1", "name": "运动鞋"},
          {"id": "C002-2", "name": "皮鞋"}
        ]
      }
    ]
  }
}
```

---

## 7. 购物车模块 `/api/v1/cart`

### 7.1 获取购物车

```
GET /api/v1/cart
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "items": [
      {
        "id": "CI001",
        "sku_id": "SKU001",
        "title": "2024新款男士运动鞋",
        "image": "https://cdn.example.com/product/001_1.jpg",
        "attributes": ["黑色", "41码"],
        "price": 29900,
        "quantity": 2,
        "stock": 500,
        "is_selected": true,
        "is_valid": true,
        "invalid_reason": null,
        "shop_id": "S001",
        "shop_name": "运动旗舰店"
      },
      {
        "id": "CI002",
        "sku_id": "SKU002",
        "title": "运动T恤",
        "image": "https://cdn.example.com/product/002_1.jpg",
        "attributes": ["白色", "L码"],
        "price": 9900,
        "quantity": 1,
        "stock": 0,  // 库存不足
        "is_selected": false,
        "is_valid": false,
        "invalid_reason": "库存不足",
        "shop_id": "S001",
        "shop_name": "运动旗舰店"
      }
    ],
    "selected_count": 2,
    "total_amount": 69700,  // 已选商品总价（分）
    "freight_amount": 1200,  // 运费（分）
    "discount_amount": 0      // 优惠金额（分）
  }
}
```

### 7.2 添加购物车

```
POST /api/v1/cart
```

**请求：**
```json
{
  "sku_id": "SKU001",
  "quantity": 2,
  "attributes": ["黑色", "41码"]
}
```

**响应：**
```json
{
  "code": 0,
  "message": "添加成功",
  "data": {
    "cart_count": 5
  }
}
```

### 7.3 更新购物车商品数量

```
PUT /api/v1/cart/:item_id
```

**请求：**
```json
{
  "quantity": 3
}
```

**响应：**
```json
{
  "code": 0,
  "message": "更新成功"
}
```

### 7.4 选择/取消选择购物车商品

```
PUT /api/v1/cart/selection
```

**请求：**
```json
{
  "item_ids": ["CI001", "CI002"],  // 要选中的商品，空数组则全部取消
  "selected": true
}
```

**响应：**
```json
{
  "code": 0,
  "message": "更新成功"
}
```

### 7.5 删除购物车商品

```
DELETE /api/v1/cart/:item_id
```

或批量删除：

```
DELETE /api/v1/cart
```

**请求：**
```json
{
  "item_ids": ["CI001", "CI002"]
}
```

**响应：**
```json
{
  "code": 0,
  "message": "删除成功"
}
```

### 7.6 合并购物车（登录时）

```
POST /api/v1/cart/merge
```

**请求：**
```json
{
  "guest_cart_items": [
    {"sku_id": "SKU001", "quantity": 1, "attributes": ["黑色", "41码"]},
    {"sku_id": "SKU003", "quantity": 2, "attributes": ["蓝色", "L码"]}
  ]
}
```

**响应：**
```json
{
  "code": 0,
  "message": "合并成功",
  "data": {
    "merged_count": 2,
    "failed_items": [
      {"sku_id": "SKU003", "reason": "商品已下架"}
    ]
  }
}
```

---

## 8. 订单模块 `/api/v1/orders`

### 8.1 确认订单（计算价格）

```
POST /api/v1/orders/checkout
```

**请求：**
```json
{
  "cart_item_ids": ["CI001", "CI002"],  // 购物车选中的商品
  "address_id": "A001",
  "coupon_id": "CPN001",  // 可选
  "use_points": 100,  // 可选，使用积分数量
  "remark": "请尽快发货"  // 可选
}
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "order_no": null,  // 确认订单不生成订单号
    "items": [
      {
        "sku_id": "SKU001",
        "title": "2024新款男士运动鞋",
        "image": "https://...",
        "attributes": ["黑色", "41码"],
        "price": 29900,
        "quantity": 2,
        "subtotal": 59800
      }
    ],
    "address": {
      "id": "A001",
      "receiver": "张三",
      "phone": "138****8000",
      "detail": "广东省深圳市南山区科技园南路XX号"
    },
    "price_summary": {
      "goods_amount": 69700,      // 商品总额
      "freight_amount": 1200,      // 运费
      "coupon_discount": -2000,   // 优惠券优惠
      "points_discount": -100,    // 积分优惠
      "total_amount": 68800        // 应付总额
    },
    "available_coupons": [
      {
        "coupon_id": "CPN001",
        "name": "满100减20",
        "discount": 2000
      }
    ],
    "points_to_use": {
      "available": 500,
      "max_use": 500,
      "exchange_rate": 100,  // 100积分=1元
      "抵扣金额": 5
    }
  }
}
```

### 8.2 创建订单

```
POST /api/v1/orders
```

**请求：**
```json
{
  "cart_item_ids": ["CI001", "CI002"],
  "address_id": "A001",
  "coupon_id": "CPN001",
  "use_points": 100,
  "remark": "请尽快发货"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "订单创建成功",
  "data": {
    "order_no": "O20240320123456789",
    "total_amount": 68800,
    "pay_amount": 68800,
    "pay_deadline": "2024-03-20T11:00:00Z",  // 30分钟后过期
    "items": [
      {
        "sku_id": "SKU001",
        "title": "2024新款男士运动鞋",
        "quantity": 2,
        "price": 29900
      }
    ]
  }
}
```

### 8.3 订单列表

```
GET /api/v1/user/orders
```

**请求参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码 |
| page_size | int | 否 | 每页条数 |
| status | string | 否 | 状态筛选 |

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "orders": [
      {
        "order_no": "O20240320123456789",
        "status": "PAID",
        "status_text": "待发货",
        "total_amount": 68800,
        "pay_amount": 68800,
        "created_at": "2024-03-20T10:30:00Z",
        "shop": {
          "shop_id": "S001",
          "name": "运动旗舰店"
        },
        "items": [
          {
            "sku_id": "SKU001",
            "title": "2024新款男士运动鞋",
            "image": "https://...",
            "attributes": ["黑色", "41码"],
            "price": 29900,
            "quantity": 2
          }
        ],
        "action_buttons": ["cancel", "remind"]
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 5
    },
    "status_counts": {
      "ALL": 5,
      "PENDING_PAYMENT": 1,
      "PAID": 2,
      "SHIPPED": 1,
      "CONFIRMED": 1
    }
  }
}
```

### 8.4 订单详情

```
GET /api/v1/user/orders/:order_no
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "order_no": "O20240320123456789",
    "status": "PAID",
    "status_text": "待发货",
    "total_amount": 68800,
    "pay_amount": 68800,
    "created_at": "2024-03-20T10:30:00Z",
    "pay_time": "2024-03-20T10:35:00Z",
    "ship_time": null,
    "receive_time": null,

    "address": {
      "receiver": "张三",
      "phone": "13800138000",
      "province": "广东省",
      "city": "深圳市",
      "district": "南山区",
      "detail": "科技园南路XX号"
    },

    "shop": {
      "shop_id": "S001",
      "name": "运动旗舰店",
      "phone": "400-800-8888"
    },

    "items": [
      {
        "sku_id": "SKU001",
        "title": "2024新款男士运动鞋",
        "image": "https://...",
        "attributes": ["黑色", "41码"],
        "price": 29900,
        "quantity": 2,
        "subtotal": 59800
      }
    ],

    "price_summary": {
      "goods_amount": 69700,
      "freight_amount": 1200,
      "coupon_discount": -2000,
      "points_discount": -100,
      "total_amount": 68800
    },

    "remark": "请尽快发货",

    "logistics": null,  // 未发货时为空

    "action_buttons": ["cancel", "remind"],
    "action_logs": [
      {"action": "创建订单", "time": "2024-03-20T10:30:00Z"},
      {"action": "支付成功", "time": "2024-03-20T10:35:00Z"}
    ]
  }
}
```

### 8.5 取消订单

```
POST /api/v1/orders/:order_no/cancel
```

**请求：**
```json
{
  "reason": "不想要了"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "订单已取消"
}
```

### 8.6 确认收货

```
POST /api/v1/orders/:order_no/confirm
```

**响应：**
```json
{
  "code": 0,
  "message": "确认收货成功"
}
```

### 8.7 订单评价

```
POST /api/v1/orders/:order_no/reviews
```

**请求：**
```json
{
  "reviews": [
    {
      "sku_id": "SKU001",
      "rating": 5,
      "content": "鞋子很棒，穿着舒适！",
      "images": ["https://...", "https://..."],
      "is_anonymous": false
    }
  ]
}
```

**响应：**
```json
{
  "code": 0,
  "message": "评价成功",
  "data": {
    "points_reward": 50  // 评价奖励积分
  }
}
```

---

## 9. 支付模块 `/api/v1/payments`

### 9.1 获取支付信息

```
GET /api/v1/payments/:order_no
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "order_no": "O20240320123456789",
    "pay_amount": 68800,
    "pay_deadline": "2024-03-20T11:00:00Z",
    "payment_methods": [
      {
        "id": "wechat",
        "name": "微信支付",
        "icon": "https://..."
      },
      {
        "id": "alipay",
        "name": "支付宝",
        "icon": "https://..."
      },
      {
        "id": "balance",
        "name": "账户余额",
        "icon": "https://...",
        "balance": 50000
      }
    ]
  }
}
```

### 9.2 创建支付

```
POST /api/v1/payments
```

**请求：**
```json
{
  "order_no": "O20240320123456789",
  "method": "wechat"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "payment_id": "PAY001",
    "order_no": "O20240320123456789",
    "amount": 68800,
    "status": "PENDING",
    "qr_code": "weixin://wxpay/xxxxx",  // 扫码支付二维码
    "h5_url": "https://wx.tenpay.com/xxxxx"  // H5支付链接
  }
}
```

### 9.3 支付结果查询

```
GET /api/v1/payments/:payment_id/status
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "payment_id": "PAY001",
    "status": "SUCCESS",
    "pay_time": "2024-03-20T10:40:00Z"
  }
}
```

---

## 10. 收藏与足迹模块

### 10.1 我的收藏

```
GET /api/v1/user/collect
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "products": [
      {
        "sku_id": "SKU001",
        "title": "2024新款男士运动鞋",
        "image": "https://...",
        "price": 29900,
        "stock": 500,
        "is_valid": true,
        "added_at": "2024-03-15T10:00:00Z"
      }
    ],
    "shops": [
      {
        "shop_id": "S001",
        "name": "运动旗舰店",
        "logo": "https://...",
        "followed_at": "2024-03-10T10:00:00Z"
      }
    ]
  }
}
```

### 10.2 添加收藏

```
POST /api/v1/user/collect
```

**请求：**
```json
{
  "type": "product",  // product | shop
  "id": "SKU001"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "收藏成功"
}
```

### 10.3 取消收藏

```
DELETE /api/v1/user/collect
```

**请求：**
```json
{
  "type": "product",
  "id": "SKU001"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "已取消收藏"
}
```

### 10.4 我的足迹

```
GET /api/v1/user/footprint
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "footprints": [
      {
        "sku_id": "SKU001",
        "title": "2024新款男士运动鞋",
        "image": "https://...",
        "price": 29900,
        "viewed_at": "2024-03-20T14:30:00Z",
        "date": "2024-03-20"
      }
    ],
    "grouped_by_date": {
      "2024-03-20": [...],
      "2024-03-19": [...]
    }
  }
}
```

---

## 11. 优惠券模块 `/api/v1/coupons`

### 11.1 领券中心

```
GET /api/v1/coupons/center
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "coupons": [
      {
        "coupon_id": "C001",
        "name": "新人专享券",
        "type": "fixed",  // fixed | percent
        "value": 2000,  // 固定金额（分）
        "min_amount": 10000,  // 最低消费（分）
        "scope": "all",  // all | category | product
        "scope_ids": [],
        "total_count": 10000,
        "remaining_count": 5230,
        "per_limit": 1,
        "user_claimed": false,
        "valid_from": "2024-03-01T00:00:00Z",
        "valid_until": "2024-03-31T23:59:59Z",
        "claimed_at": null
      }
    ]
  }
}
```

### 11.2 领取优惠券

```
POST /api/v1/coupons/:coupon_id/claim
```

**响应：**
```json
{
  "code": 0,
  "message": "领取成功",
  "data": {
    "coupon_id": "C001",
    "name": "新人专享券",
    "valid_until": "2024-03-31T23:59:59Z"
  }
}
```

### 11.3 我的优惠券

```
GET /api/v1/user/coupons
```

**请求参数：** `status` - unused | used | expired

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "coupons": [
      {
        "coupon_id": "UC001",
        "name": "满100减20",
        "type": "fixed",
        "value": 2000,
        "min_amount": 10000,
        "status": "unused",
        "valid_from": "2024-03-01T00:00:00Z",
        "valid_until": "2024-03-31T23:59:59Z",
        "order_no": null,
        "used_at": null
      }
    ],
    "counts": {
      "unused": 3,
      "used": 5,
      "expired": 2
    }
  }
}
```

---

## 12. 会员与积分模块

### 12.1 会员信息

```
GET /api/v1/user/level
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "level": 3,
    "member_title": "金卡会员",
    "growth_value": 2580,
    "next_level": 4,
    "next_level_title": "铂金会员",
    "next_level_growth": 5000,
    "progress": 51.6,
    "privileges": [
      "专享折扣 95 折",
      "每月领取专属优惠券",
      "优先客服",
      "生日礼包"
    ],
    "benefits": [
      {"name": "折扣", "value": "95折"},
      {"name": "免运费门槛", "value": "满99免运"},
      {"name": "专属客服", "value": "1对1"}
    ]
  }
}
```

### 12.2 积分记录

```
GET /api/v1/user/points
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "balance": 1250,
    "records": [
      {
        "id": "P001",
        "type": "earn",  // earn | spend
        "amount": 50,
        "reason": "商品评价奖励",
        "order_no": "O20240320123456789",
        "created_at": "2024-03-20T14:00:00Z"
      },
      {
        "id": "P002",
        "type": "spend",
        "amount": -100,
        "reason": "积分兑换优惠券",
        "order_no": null,
        "created_at": "2024-03-19T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 45
    }
  }
}
```

### 12.3 签到

```
POST /api/v1/user/checkin
```

**响应：**
```json
{
  "code": 0,
  "message": "签到成功",
  "data": {
    "points": 5,
    "total_days": 7,
    "consecutive_days": 3,
    "rewards": {
      "day3": {"points": 10, "claimed": true},
      "day7": {"points": 30, "claimed": false}
    }
  }
}
```

---

## 13. 秒杀模块 `/api/v1/activities/seckill`

### 13.1 秒杀场次列表

```
GET /api/v1/activities/seckill/sessions
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "current_time": "2024-03-20T10:00:00Z",
    "sessions": [
      {
        "session_id": "S001",
        "name": "10:00场",
        "start_time": "2024-03-20T10:00:00Z",
        "end_time": "2024-03-20T12:00:00Z",
        "status": "ongoing",  // upcoming | ongoing | ended
        "products_count": 20
      },
      {
        "session_id": "S002",
        "name": "14:00场",
        "start_time": "2024-03-20T14:00:00Z",
        "end_time": "2024-03-20T16:00:00Z",
        "status": "upcoming",
        "countdown": 14400  // 距离开始的秒数
      }
    ]
  }
}
```

### 13.2 秒杀商品列表

```
GET /api/v1/activities/seckill/sessions/:session_id/products
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "products": [
      {
        "sku_id": "SKU001",
        "title": "运动鞋秒杀款",
        "image": "https://...",
        "original_price": 49900,
        "seckill_price": 19900,
        "seckill_stock": 100,
        "sold_count": 65,
        "sold_percent": 65,
        "limit_per_user": 1,
        "user_bought": 0,
        "can_buy": true
      }
    ]
  }
}
```

### 13.3 秒杀下单

```
POST /api/v1/activities/seckill/orders
```

**请求：**
```json
{
  "sku_id": "SKU001",
  "quantity": 1,
  "address_id": "A001"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "秒杀成功",
  "data": {
    "order_no": "O20240320100012345",
    "total_amount": 19900
  }
}
```

---

## 14. 消息通知模块

### 14.1 消息列表

```
GET /api/v1/notifications
```

**响应：**
```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "notifications": [
      {
        "id": "N001",
        "type": "order",  // order | system | promotion
        "title": "订单已发货",
        "content": "您的订单 O20240320123456789 已发货",
        "data": {"order_no": "O20240320123456789"},
        "is_read": false,
        "created_at": "2024-03-20T10:00:00Z"
      }
    ],
    "unread_count": 3
  }
}
```

### 14.2 标记已读

```
PUT /api/v1/notifications/:id/read
```

---

## 15. 错误码汇总

### 15.1 认证模块错误码

| code | 说明 |
|------|------|
| 40101 | token 过期 |
| 40102 | token 无效 |
| 40103 | 需要登录 |
| 40104 | 验证码错误 |
| 40105 | 验证码过期 |
| 40106 | 验证码发送太频繁 |
| 40107 | 账户被禁用 |
| 40108 | 社交登录失败 |

### 15.2 订单模块错误码

| code | 说明 |
|------|------|
| 40901 | 库存不足 |
| 40902 | 超过限购数量 |
| 40903 | 优惠券不可用 |
| 40904 | 订单已支付 |
| 40905 | 订单已取消 |
| 40906 | 订单状态不允许该操作 |
| 40907 | 支付超时，订单已关闭 |

### 15.3 商品模块错误码

| code | 说明 |
|------|------|
| 40401 | 商品不存在 |
| 40402 | 商品已下架 |
| 40403 | SKU 不存在 |
| 40901 | 库存不足 |