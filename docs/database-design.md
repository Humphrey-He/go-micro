# 数据库表结构设计文档

> 参考淘宝/京东/拼多多等主流电商平台成熟实践，结合项目微服务架构设计

---

## 一、核心实体关系总览

```
用户 (users)
    ├── 社交账号绑定 (user_social_bindings)
    ├── 收货地址 (user_addresses)
    ├── 会员等级 (user_levels)
    ├── 收藏/足迹 (user_collects, user_footprints)
    ├── 价格监控 (user_price_watches)
    └── 通知配置 (notification_configs)

商品 (products/skus)
    ├── SPU (product_spus)
    ├── SKU (product_skus)
    ├── 分类 (product_categories)
    ├── 属性 (product_attributes)
    ├── 品牌 (product_brands)
    └── 价格快照 (product_price_snapshots)

订单 (orders)
    ├── 订单主表 (orders)
    ├── 订单明细 (order_items)
    ├── 订单状态日志 (order_status_logs)
    └── 退换货 (order_returns)

库存 (inventory)
    ├── 库存主表 (inventory)
    ├── 库存流水 (inventory_logs)
    └── 预占记录 (inventory_reservations)

支付 (payments)
    ├── 支付记录 (payments)
    └── 支付日志 (payment_logs)

营销 (activities)
    ├── 优惠券 (coupons)
    ├── 秒杀 (seckill_activities)
    ├── 秒杀订单 (seckill_orders)
    └── 优惠券发放记录 (coupon_issues)

推荐 (recommendations)
    ├── 用户行为日志 (user_behavior_logs)
    ├── 商品相似度 (product_similarity)
    ├── 用户分类偏好 (user_category_preference)
    ├── 关联规则 (association_rules)
    ├── 类目畅销榜 (category_bestsellers)
    └── 全局畅销榜 (global_bestsellers)

任务 (tasks)
    ├── 任务表 (tasks)
    ├── SAGA记录 (sagas)
    └── SAGA步骤 (saga_steps)
```

---

## 二、用户服务 (User Service)

### 2.1 用户主表 (users)

参考淘宝用户表设计，支持手机号、社交账号多入口。

```sql
CREATE TABLE users (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         VARCHAR(32) NOT NULL COMMENT '用户唯一标识(UUID)',
    username        VARCHAR(64) NOT NULL UNIQUE COMMENT '用户名(昵称)',
    password_hash   VARCHAR(256) NOT NULL COMMENT '密码哈希(bcrypt)',
    phone           VARCHAR(20) UNIQUE COMMENT '手机号',
    email           VARCHAR(128) COMMENT '邮箱',
    avatar          VARCHAR(512) COMMENT '头像URL',
    gender          TINYINT DEFAULT 0 COMMENT '性别: 0未知 1男 2女',
    birthday        DATE COMMENT '生日',
    role            ENUM('customer', 'merchant', 'admin') DEFAULT 'customer' COMMENT '角色',
    status          TINYINT DEFAULT 1 COMMENT '状态: 1正常 2禁用 3注销',
    user_level      INT DEFAULT 1 COMMENT '会员等级',
    user_points     BIGINT DEFAULT 0 COMMENT '用户积分',
    total_orders    INT DEFAULT 0 COMMENT '累计订单数',
    total_amount    DECIMAL(15,2) DEFAULT 0.00 COMMENT '累计消费金额',

    -- 社交登录字段
    social_provider ENUM('none','wechat','google','apple') DEFAULT 'none' COMMENT '主要社交登录平台',
    social_uid      VARCHAR(128) COMMENT '主要社交平台用户ID',
    is_phone_bound  TINYINT DEFAULT 0 COMMENT '是否已绑定手机',
    created_source  ENUM('phone','wechat','google','apple','admin') COMMENT '注册来源',

    -- 活跃字段
    last_login_at   DATETIME COMMENT '最后登录时间',
    last_login_ip   VARCHAR(45) COMMENT '最后登录IP',
    login_count     INT DEFAULT 0 COMMENT '登录次数',

    -- 安全字段
    email_verified  TINYINT DEFAULT 0 COMMENT '邮箱是否验证',
    phone_verified  TINYINT DEFAULT 0 COMMENT '手机是否验证',

    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      DATETIME COMMENT '软删除时间',

    INDEX idx_phone (phone),
    INDEX idx_status (status),
    INDEX idx_created_source (created_source),
    INDEX idx_user_level (user_level),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
```

### 2.2 用户社交账号绑定表 (user_social_bindings)

```sql
CREATE TABLE user_social_bindings (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    provider        ENUM('wechat', 'google', 'apple') NOT NULL COMMENT '社交平台',
    provider_user_id VARCHAR(128) NOT NULL COMMENT '第三方平台用户ID',
    union_id        VARCHAR(128) COMMENT '微信UnionID(跨应用识别)',
    nickname        VARCHAR(64) COMMENT '第三方昵称',
    avatar          VARCHAR(512) COMMENT '第三方头像URL',
    access_token    TEXT COMMENT '第三方access_token',
    refresh_token   TEXT COMMENT '第三方refresh_token',
    token_expires_at DATETIME COMMENT 'access_token过期时间',
    bind_status     TINYINT DEFAULT 1 COMMENT '绑定状态: 1正常 0解绑',
    bind_source     VARCHAR(32) COMMENT '绑定场景: login/register/bind',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_provider_user (provider, provider_user_id),
    KEY idx_user_id (user_id),
    KEY idx_union_id (union_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户社交账号绑定表';
```

### 2.3 用户收货地址表 (user_addresses)

```sql
CREATE TABLE user_addresses (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL,
    address_id      VARCHAR(32) NOT NULL COMMENT '地址唯一标识',
    receiver_name   VARCHAR(64) NOT NULL COMMENT '收货人姓名',
    receiver_phone  VARCHAR(20) NOT NULL COMMENT '收货人手机',
    province_code   VARCHAR(10) NOT NULL COMMENT '省编码',
    city_code       VARCHAR(10) NOT NULL COMMENT '市编码',
    district_code   VARCHAR(10) NOT NULL COMMENT '区编码',
    province_name   VARCHAR(32) NOT NULL COMMENT '省名称',
    city_name       VARCHAR(32) NOT NULL COMMENT '市名称',
    district_name   VARCHAR(32) NOT NULL COMMENT '区名称',
    detail_address  VARCHAR(256) NOT NULL COMMENT '详细地址',
    postal_code     VARCHAR(10) COMMENT '邮政编码',
    lat             DECIMAL(10,6) COMMENT '纬度',
    lng             DECIMAL(10,6) COMMENT '经度',
    is_default      TINYINT DEFAULT 0 COMMENT '是否默认地址',
    is_deleted      TINYINT DEFAULT 0 COMMENT '是否删除',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_user_address (user_id, address_id),
    INDEX idx_user_default (user_id, is_default),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户收货地址表';
```

### 2.4 会员等级表 (user_levels)

参考京东会员体系设计。

```sql
CREATE TABLE user_levels (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    level           INT NOT NULL COMMENT '等级编号',
    level_name      VARCHAR(32) NOT NULL COMMENT '等级名称: 注册会员/铜牌/银牌/金牌/钻石',
    min_points      BIGINT DEFAULT 0 COMMENT '最低积分门槛',
    max_points      BIGINT DEFAULT 0 COMMENT '最高积分门槛',
    discount_rate   DECIMAL(4,2) DEFAULT 100.00 COMMENT '折扣率(百分比)',
    points_rate     INT DEFAULT 1 COMMENT '消费1元获得积分',
    exclusive_rights TEXT COMMENT '专属权益(JSON)',
    icon_url        VARCHAR(256) COMMENT '等级图标',
    bg_color        VARCHAR(16) COMMENT '等级背景色',
    sort_order      INT DEFAULT 0 COMMENT '排序',
    status          TINYINT DEFAULT 1 COMMENT '状态',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_level (level),
    INDEX idx_points (min_points, max_points)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会员等级表';
```

### 2.5 用户收藏表 (user_collects)

```sql
CREATE TABLE user_collects (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL,
    sku_id          BIGINT UNSIGNED NOT NULL COMMENT '商品SKU ID',
    collect_type    ENUM('product','shop') DEFAULT 'product' COMMENT '收藏类型',
    source          VARCHAR(32) COMMENT '来源页面',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uk_user_sku_type (user_id, sku_id, collect_type),
    INDEX idx_user_id (user_id),
    INDEX idx_sku_id (sku_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户收藏表';
```

### 2.6 用户足迹表 (user_footprints)

```sql
CREATE TABLE user_footprints (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id         BIGINT UNSIGNED NOT NULL,
    anonymous_id    VARCHAR(64) COMMENT '匿名用户ID',
    sku_id          BIGINT UNSIGNED NOT NULL,
    source          VARCHAR(32) COMMENT '来源: search/recommend/category',
    stay_duration   INT DEFAULT 0 COMMENT '停留时长(秒)',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_created (user_id, created_at),
    INDEX idx_sku_id (sku_id),
    INDEX idx_anonymous (anonymous_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户足迹表';
```

---

## 三、商品服务 (Product Service)

### 3.1 SPU主表 (product_spus)

参考淘宝SPU设计。

```sql
CREATE TABLE product_spus (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    spu_id          VARCHAR(32) NOT NULL COMMENT 'SPU唯一标识',
    name            VARCHAR(256) NOT NULL COMMENT '商品名称',
    subtitle        VARCHAR(512) COMMENT '副标题',
    brand_id        BIGINT UNSIGNED COMMENT '品牌ID',
    category_id     BIGINT UNSIGNED NOT NULL COMMENT '类目ID',
    category_path   VARCHAR(128) COMMENT '类目路径: 1级/2级/3级',
    description     TEXT COMMENT '商品描述(富文本)',
    detail_images   JSON COMMENT '详情图片列表',
    video_url       VARCHAR(512) COMMENT '商品视频URL',

    -- 销售属性
    sales_status    ENUM('on_sale','off_sale','sold_out','deleted') DEFAULT 'off_sale' COMMENT '销售状态',
    review_count    INT DEFAULT 0 COMMENT '评价数量',
    avg_rating      DECIMAL(3,2) DEFAULT 5.00 COMMENT '平均评分',
    collect_count   INT DEFAULT 0 COMMENT '收藏数量',

    -- SEO相关
    keywords        VARCHAR(256) COMMENT '搜索关键词',
    seo_title       VARCHAR(128) COMMENT 'SEO标题',
    seo_description VARCHAR(512) COMMENT 'SEO描述',

    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      DATETIME COMMENT '软删除时间',

    UNIQUE KEY uk_spu_id (spu_id),
    INDEX idx_brand (brand_id),
    INDEX idx_category (category_id),
    INDEX idx_sales_status (sales_status),
    INDEX idx_created_at (created_at),
    FULLTEXT idx_name (name, keywords)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品SPU表';
```

### 3.2 SKU表 (product_skus)

```sql
CREATE TABLE product_skus (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id          VARCHAR(32) NOT NULL COMMENT 'SKU唯一标识',
    spu_id          VARCHAR(32) NOT NULL COMMENT '所属SPU ID',
    name            VARCHAR(256) NOT NULL COMMENT 'SKU名称(显示规格信息)',
    bar_code        VARCHAR(64) COMMENT '条形码',
    price           DECIMAL(12,2) NOT NULL COMMENT '售价(分)',
    cost_price      DECIMAL(12,2) COMMENT '成本价',
    original_price  DECIMAL(12,2) COMMENT '划线价/原价',

    -- 库存
    stock           INT DEFAULT 0 COMMENT '总库存',
    low_stock_threshold INT DEFAULT 10 COMMENT '低库存阈值',
    sold_count      INT DEFAULT 0 COMMENT '销量',

    -- 规格属性
    spec_json       JSON COMMENT '规格属性JSON: [{"attr_id":"1","attr_name":"颜色","value_id":"1","value_name":"红色"}]',
    sku_images      JSON COMMENT 'SKU图片列表',

    -- 物流
    weight          DECIMAL(10,3) COMMENT '重量(kg)',
    volume          DECIMAL(10,3) COMMENT '体积(m³)',
    freight_template_id BIGINT UNSIGNED COMMENT '运费模板ID',

    -- 状态
    status          TINYINT DEFAULT 1 COMMENT '状态: 1上架 0下架',

    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_sku_id (sku_id),
    INDEX idx_spu_id (spu_id),
    INDEX idx_price (price),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品SKU表';
```

### 3.3 商品类目表 (product_categories)

```sql
CREATE TABLE product_categories (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    category_id     BIGINT UNSIGNED NOT NULL COMMENT '类目ID',
    parent_id       BIGINT UNSIGNED DEFAULT 0 COMMENT '父类目ID',
    name            VARCHAR(64) NOT NULL COMMENT '类目名称',
    icon_url        VARCHAR(256) COMMENT '类目图标',
    level           TINYINT DEFAULT 1 COMMENT '层级: 1/2/3',
    sort_order      INT DEFAULT 0 COMMENT '排序',
    is_leaf         TINYINT DEFAULT 0 COMMENT '是否叶子节点',
    status          TINYINT DEFAULT 1 COMMENT '状态',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_category_id (category_id),
    INDEX idx_parent (parent_id),
    INDEX idx_sort (sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品类目表';
```

### 3.4 商品属性表 (product_attributes)

```sql
CREATE TABLE product_attributes (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    attr_id         BIGINT UNSIGNED NOT NULL COMMENT '属性ID',
    category_id     BIGINT UNSIGNED COMMENT '所属类目',
    name            VARCHAR(64) NOT NULL COMMENT '属性名称: 颜色/尺码/版本',
    attr_type       ENUM('spec','sale','base') DEFAULT 'spec' COMMENT '属性类型: 规格属性/销售属性/基础属性',
    is_required     TINYINT DEFAULT 0 COMMENT '是否必填',
    is_filterable   TINYINT DEFAULT 0 COMMENT '是否可筛选',
    is_searchable   TINYINT DEFAULT 0 COMMENT '是否可搜索',
    sort_order      INT DEFAULT 0 COMMENT '排序',
    status          TINYINT DEFAULT 1 COMMENT '状态',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_attr_id (attr_id),
    INDEX idx_category (category_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品属性表';

-- 属性值表
CREATE TABLE product_attribute_values (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    attr_id         BIGINT UNSIGNED NOT NULL COMMENT '属性ID',
    value_id        BIGINT UNSIGNED NOT NULL COMMENT '属性值ID',
    value_name      VARCHAR(128) NOT NULL COMMENT '属性值名称',
    attr_value_alias VARCHAR(128) COMMENT '别名',
    image_url       VARCHAR(256) COMMENT '颜色对应图片(颜色属性用)',
    sort_order      INT DEFAULT 0 COMMENT '排序',
    status          TINYINT DEFAULT 1 COMMENT '状态',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_attr_value (attr_id, value_id),
    INDEX idx_value_id (value_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品属性值表';
```

### 3.5 商品品牌表 (product_brands)

```sql
CREATE TABLE product_brands (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    brand_id        BIGINT UNSIGNED NOT NULL COMMENT '品牌ID',
    name            VARCHAR(128) NOT NULL COMMENT '品牌名称',
    logo_url        VARCHAR(512) COMMENT '品牌Logo',
    banner_url      VARCHAR(512) COMMENT '品牌横幅',
    description     TEXT COMMENT '品牌描述',
    country         VARCHAR(32) COMMENT '产地/国家',
    sort_order      INT DEFAULT 0 COMMENT '排序',
    status          TINYINT DEFAULT 1 COMMENT '状态',
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_brand_id (brand_id),
    INDEX idx_name (name),
    INDEX idx_status_sort (status, sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品品牌表';
```

---

## 四、订单服务 (Order Service)

### 4.1 订单主表 (orders)

参考淘宝/京东订单表设计。

```sql
CREATE TABLE orders (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id            VARCHAR(32) NOT NULL COMMENT '订单唯一标识',
    biz_no              VARCHAR(32) NOT NULL COMMENT '业务编号(用于外部显示)',
    user_id             BIGINT UNSIGNED NOT NULL COMMENT '用户ID',

    -- 金额相关
    total_amount        DECIMAL(12,2) NOT NULL COMMENT '商品总金额',
    discount_amount     DECIMAL(12,2) DEFAULT 0.00 COMMENT '整单优惠金额',
    coupon_amount       DECIMAL(12,2) DEFAULT 0.00 COMMENT '优惠券抵扣',
    points_amount       DECIMAL(12,2) DEFAULT 0.00 COMMENT '积分抵扣',
    freight_amount      DECIMAL(12,2) DEFAULT 0.00 COMMENT '运费',
    pay_amount          DECIMAL(12,2) NOT NULL COMMENT '应付金额(最终支付)',
    actual_amount       DECIMAL(12,2) DEFAULT 0.00 COMMENT '实际支付金额',

    -- 优惠信息
    promotion_ids       JSON COMMENT '参与的优惠活动ID列表',
    discount_detail     JSON COMMENT '优惠明细JSON',

    -- 地址信息(冗余存储,订单完成后不依赖地址表)
    receiver_name       VARCHAR(64) NOT NULL COMMENT '收货人',
    receiver_phone      VARCHAR(20) NOT NULL COMMENT '收货手机',
    province_code       VARCHAR(10) NOT NULL COMMENT '省编码',
    city_code           VARCHAR(10) NOT NULL COMMENT '市编码',
    district_code       VARCHAR(10) NOT NULL COMMENT '区编码',
    detail_address      VARCHAR(256) NOT NULL COMMENT '详细地址',

    -- 发票信息
    invoice_type        TINYINT DEFAULT 0 COMMENT '发票类型: 0不要 1个人 2企业',
    invoice_title       VARCHAR(128) COMMENT '发票抬头',
    invoice_content     VARCHAR(256) COMMENT '发票内容',
    tax_no              VARCHAR(32) COMMENT '税号',

    -- 订单状态
    status              ENUM('pending','reserved','paid','shipped','delivered','completed','cancelled','refunding','refunded') DEFAULT 'pending' COMMENT '订单状态',
    sub_status          VARCHAR(32) COMMENT '子状态',

    -- 时间节点
    reserved_at         DATETIME COMMENT '库存预占时间',
    paid_at             DATETIME COMMENT '支付时间',
    shipped_at          DATETIME COMMENT '发货时间',
    delivered_at        DATETIME COMMENT '收货时间',
    completed_at        DATETIME COMMENT '完成时间',
    cancelled_at        DATETIME COMMENT '取消时间',

    -- 备注
    buyer_remark        VARCHAR(512) COMMENT '买家备注',
    seller_remark       VARCHAR(512) COMMENT '卖家备注',

    -- 关联
    reserved_id         VARCHAR(64) COMMENT '库存预占ID',
    payment_id          VARCHAR(64) COMMENT '支付记录ID',
    coupon_id           VARCHAR(64) COMMENT '使用优惠券ID',

    -- 幂等
    idempotent_key      VARCHAR(64) UNIQUE COMMENT '幂等键',

    -- 版本控制(乐观锁)
    version             INT DEFAULT 1 COMMENT '版本号',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at          DATETIME COMMENT '软删除时间',

    UNIQUE KEY uk_order_id (order_id),
    UNIQUE KEY uk_biz_no (biz_no),
    INDEX idx_user_status (user_id, status),
    INDEX idx_created_at (created_at),
    INDEX idx_paid_at (paid_at),
    INDEX idx_status_substatus (status, sub_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单主表';
```

### 4.2 订单明细表 (order_items)

```sql
CREATE TABLE order_items (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id            VARCHAR(32) NOT NULL COMMENT '订单ID',
    item_id             VARCHAR(32) NOT NULL COMMENT '订单明细ID',
    sku_id              VARCHAR(32) NOT NULL COMMENT 'SKU ID',
    spu_id              VARCHAR(32) NOT NULL COMMENT 'SPU ID',
    sku_name            VARCHAR(256) NOT NULL COMMENT 'SKU名称(冗余)',
    sku_image           VARCHAR(512) COMMENT 'SKU图片(冗余)',

    -- 商品信息(下单时快照)
    quantity            INT NOT NULL COMMENT '购买数量',
    unit_price          DECIMAL(12,2) NOT NULL COMMENT '单价(下单时)',
    original_price      DECIMAL(12,2) COMMENT '原价(冗余)',
    cost_price          DECIMAL(12,2) COMMENT '成本价(冗余)',

    -- 优惠分摊
    discount_amount     DECIMAL(12,2) DEFAULT 0.00 COMMENT '优惠分摊金额',
    coupon_amount       DECIMAL(12,2) DEFAULT 0.00 COMMENT '优惠券分摊',

    -- 退款相关
    refund_status       ENUM('none','partial','full') DEFAULT 'none' COMMENT '退款状态',
    refund_quantity      INT DEFAULT 0 COMMENT '已退款数量',
    refund_amount       DECIMAL(12,2) DEFAULT 0.00 COMMENT '已退款金额',

    -- 状态
    status              TINYINT DEFAULT 1 COMMENT '状态',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_order_id (order_id),
    INDEX idx_sku_id (sku_id),
    INDEX idx_refund_status (refund_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单明细表';
```

### 4.3 订单状态日志表 (order_status_logs)

参考京东订单日志设计，便于问题追溯。

```sql
CREATE TABLE order_status_logs (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id            VARCHAR(32) NOT NULL COMMENT '订单ID',
    from_status         VARCHAR(32) COMMENT '变更前状态',
    to_status           VARCHAR(32) NOT NULL COMMENT '变更后状态',
    operator            VARCHAR(64) COMMENT '操作人: user_id/system/admin',
    operator_type       ENUM('user','system','admin') DEFAULT 'system' COMMENT '操作类型',
    remark              VARCHAR(512) COMMENT '变更说明',
    client_ip           VARCHAR(45) COMMENT '客户端IP',
    request_id          VARCHAR(64) COMMENT '请求追踪ID',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_order_id (order_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单状态日志表';
```

### 4.4 退换货表 (order_returns)

```sql
CREATE TABLE order_returns (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    return_id           VARCHAR(32) NOT NULL COMMENT '退换货单号',
    order_id            VARCHAR(32) NOT NULL COMMENT '原订单ID',
    order_item_id       VARCHAR(32) NOT NULL COMMENT '订单明细ID',
    user_id             BIGINT UNSIGNED NOT NULL COMMENT '用户ID',

    -- 类型
    return_type         ENUM('return','exchange','refund') NOT NULL COMMENT '退换货类型: 退货/换货/仅退款',
    return_reason       VARCHAR(64) NOT NULL COMMENT '退换货原因',
    return_desc         VARCHAR(512) COMMENT '问题描述',
    images              JSON COMMENT '凭证图片',

    -- 金额
    refund_amount       DECIMAL(12,2) NOT NULL COMMENT '退款金额',
    actual_refund_amount DECIMAL(12,2) COMMENT '实际退款金额(含运费)',
    compensation_amount DECIMAL(12,2) DEFAULT 0.00 COMMENT '补偿金额',

    -- 物流
    express_company     VARCHAR(32) COMMENT '快递公司',
    express_no          VARCHAR(64) COMMENT '快递单号',

    -- 状态
    status              ENUM('pending','approved','rejected','received','refunding','completed','closed') DEFAULT 'pending' COMMENT '处理状态',
    admin_remark        VARCHAR(512) COMMENT '管理员备注',

    -- 时间
    approved_at         DATETIME COMMENT '审核通过时间',
    received_at        DATETIME COMMENT '收货时间',
    completed_at       DATETIME COMMENT '完成时间',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_return_id (return_id),
    INDEX idx_order_id (order_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='退换货表';
```

---

## 五、库存服务 (Inventory Service)

### 5.1 库存主表 (inventory)

```sql
CREATE TABLE inventory (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id              VARCHAR(32) NOT NULL COMMENT 'SKU ID',
    warehouse_id        BIGINT UNSIGNED DEFAULT 1 COMMENT '仓库ID',

    -- 库存数量
    total_stock         INT NOT NULL DEFAULT 0 COMMENT '总库存',
    available_stock     INT NOT NULL DEFAULT 0 COMMENT '可用库存',
    reserved_stock      INT NOT NULL DEFAULT 0 COMMENT '预占库存(待支付)',
    locked_stock        INT NOT NULL DEFAULT 0 COMMENT '锁定库存(发货中)',
    damaged_stock       INT DEFAULT 0 COMMENT '残次库存',

    -- 库存预警
    low_stock_threshold INT DEFAULT 10 COMMENT '低库存预警阈值',
    overstock_threshold INT COMMENT '超库存阈值',

    -- 状态
    status              TINYINT DEFAULT 1 COMMENT '状态',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_sku_warehouse (sku_id, warehouse_id),
    INDEX idx_available (available_stock)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存主表';
```

### 5.2 库存预占表 (inventory_reservations)

参考京东库存预占设计，支持订单超时释放。

```sql
CREATE TABLE inventory_reservations (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    reserved_id         VARCHAR(32) NOT NULL COMMENT '预占唯一标识',
    order_id            VARCHAR(32) COMMENT '关联订单ID',
    warehouse_id        BIGINT UNSIGNED DEFAULT 1 COMMENT '仓库ID',
    status              ENUM('active','confirmed','released','expired') DEFAULT 'active' COMMENT '状态',
    expire_at           DATETIME NOT NULL COMMENT '预占过期时间(默认30分钟)',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    confirmed_at        DATETIME COMMENT '确认时间',
    released_at         DATETIME COMMENT '释放时间',

    INDEX idx_order_id (order_id),
    INDEX idx_status_expire (status, expire_at),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存预占表';

-- 预占明细
CREATE TABLE inventory_reservation_items (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    reserved_id         VARCHAR(32) NOT NULL COMMENT '预占ID',
    sku_id              VARCHAR(32) NOT NULL COMMENT 'SKU ID',
    quantity            INT NOT NULL COMMENT '预占数量',

    UNIQUE KEY uk_reserved_sku (reserved_id, sku_id),
    INDEX idx_reserved_id (reserved_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存预占明细表';
```

### 5.3 库存流水表 (inventory_logs)

参考淘宝库存流水设计，全链路追踪。

```sql
CREATE TABLE inventory_logs (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id              VARCHAR(32) NOT NULL COMMENT 'SKU ID',
    warehouse_id        BIGINT UNSIGNED DEFAULT 1 COMMENT '仓库ID',

    -- 变更前后
    before_available    INT NOT NULL COMMENT '变更前可用',
    before_reserved     INT NOT NULL COMMENT '变更前预占',
    after_available     INT NOT NULL COMMENT '变更后可用',
    after_reserved      INT NOT NULL COMMENT '变更后预占',

    change_type         ENUM('reserve','confirm','release','lock','unlock','adjust','return_in','damage','transfer') NOT NULL COMMENT '变动类型',
    change_quantity     INT NOT NULL COMMENT '变动数量(正数增加/负数减少)',

    -- 关联业务
    order_id            VARCHAR(32) COMMENT '关联订单',
    reserved_id         VARCHAR(32) COMMENT '关联预占',
    return_id           VARCHAR(32) COMMENT '关联退货单',

    operator            VARCHAR(64) COMMENT '操作人',
    operator_type       ENUM('user','system','admin') DEFAULT 'system' COMMENT '操作类型',
    remark              VARCHAR(256) COMMENT '备注',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_sku_created (sku_id, created_at),
    INDEX idx_order_id (order_id),
    INDEX idx_change_type (change_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存流水表';
```

---

## 六、支付服务 (Payment Service)

### 6.1 支付记录表 (payments)

参考支付宝/微信支付设计。

```sql
CREATE TABLE payments (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    payment_id          VARCHAR(32) NOT NULL COMMENT '支付单号',
    order_id            VARCHAR(32) NOT NULL COMMENT '关联订单号',

    -- 金额
    amount              DECIMAL(12,2) NOT NULL COMMENT '支付金额(分)',
    actual_amount       DECIMAL(12,2) COMMENT '实际支付金额',

    -- 支付渠道
    channel             ENUM('wechat','alipay','bank_card','points','cod') NOT NULL COMMENT '支付渠道',
    channel_order_no    VARCHAR(64) COMMENT '渠道侧订单号',
    channel_account     VARCHAR(64) COMMENT '支付账号(脱敏)',

    -- 状态机
    status              ENUM('pending','processing','success','failed','cancelled','refunded','timeout') DEFAULT 'pending' COMMENT '支付状态',
    pay_url             VARCHAR(512) COMMENT '二维码/跳转链接',
    qr_code             VARCHAR(128) COMMENT '二维码凭证',

    -- 时间
    paid_at             DATETIME COMMENT '支付成功时间',
    expired_at          DATETIME NOT NULL COMMENT '支付过期时间',
    timeout_seconds     INT DEFAULT 1800 COMMENT '超时时间(秒)',

    -- 幂等
    request_id          VARCHAR(64) UNIQUE COMMENT '请求ID(幂等键)',

    -- 错误信息
    error_code          VARCHAR(32) COMMENT '错误码',
    error_msg           VARCHAR(256) COMMENT '错误信息',

    -- 退款关联
    refund_id           VARCHAR(32) COMMENT '关联退款单号',
    refunded_amount     DECIMAL(12,2) DEFAULT 0.00 COMMENT '已退款金额',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_payment_id (payment_id),
    INDEX idx_order_id (order_id),
    INDEX idx_channel_order (channel, channel_order_no),
    INDEX idx_status_expired (status, expired_at),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付记录表';
```

### 6.2 支付日志表 (payment_logs)

```sql
CREATE TABLE payment_logs (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    payment_id          VARCHAR(32) NOT NULL COMMENT '支付单号',
    event_type          VARCHAR(32) NOT NULL COMMENT '事件类型: create/callback/notify/refund',
    request_data        TEXT COMMENT '请求数据(JSON)',
    response_data       TEXT COMMENT '响应数据(JSON)',
    client_ip           VARCHAR(45) COMMENT '客户端IP',
    request_id          VARCHAR(64) COMMENT '请求追踪ID',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_payment_event (payment_id, event_type),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付日志表';
```

---

## 七、退款服务 (Refund Service)

### 7.1 退款记录表 (refunds)

参考淘宝退款流程设计。

```sql
CREATE TABLE refunds (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    refund_id           VARCHAR(32) NOT NULL COMMENT '退款单号',
    order_id            VARCHAR(32) NOT NULL COMMENT '原订单号',
    payment_id          VARCHAR(32) COMMENT '原支付单号',
    user_id             BIGINT UNSIGNED NOT NULL COMMENT '用户ID',

    -- 金额
    refund_amount       DECIMAL(12,2) NOT NULL COMMENT '申请退款金额',
    actual_refund_amount DECIMAL(12,2) COMMENT '实际退款金额',
    refund_freight      DECIMAL(12,2) DEFAULT 0.00 COMMENT '退还运费',

    -- 类型
    refund_type         ENUM('only_refund','return_refund') NOT NULL COMMENT '退款类型: 仅退款/退货退款',
    return_logistics   TINYINT DEFAULT 0 COMMENT '是否需要退货',

    -- 原因
    reason_id          VARCHAR(32) COMMENT '退款原因ID',
    reason_desc        VARCHAR(256) COMMENT '退款说明',
    images             JSON COMMENT '凭证图片',

    -- 状态
    status              ENUM('pending','approved','rejected','received','refunding','success','failed','closed') DEFAULT 'pending' COMMENT '退款状态',
    audit_remark       VARCHAR(512) COMMENT '审核备注',
    reject_reason      VARCHAR(256) COMMENT '拒绝原因',

    -- 退货物流
    return_express_company VARCHAR(32) COMMENT '退货快递公司',
    return_express_no      VARCHAR(64) COMMENT '退货快递单号',

    -- 重试机制
    retry_count         INT DEFAULT 0 COMMENT '重试次数',
    max_retry           INT DEFAULT 3 COMMENT '最大重试次数',
    next_retry_at       DATETIME COMMENT '下次重试时间',
    last_error          VARCHAR(512) COMMENT '最后错误信息',

    -- 时间
    approved_at         DATETIME COMMENT '审核通过时间',
    received_at         DATETIME COMMENT '确认收货时间',
    success_at          DATETIME COMMENT '退款成功时间',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_refund_id (refund_id),
    INDEX idx_order_id (order_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_next_retry (next_retry_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='退款记录表';
```

---

## 八、营销活动服务 (Activity Service)

### 8.1 优惠券表 (coupons)

参考淘宝优惠券设计。

```sql
CREATE TABLE coupons (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    coupon_id           VARCHAR(32) NOT NULL COMMENT '优惠券模板ID',
    name                VARCHAR(64) NOT NULL COMMENT '优惠券名称',
    description         VARCHAR(256) COMMENT '使用说明',

    -- 面额
    face_amount         DECIMAL(12,2) NOT NULL COMMENT '面额',
    min_order_amount    DECIMAL(12,2) DEFAULT 0.00 COMMENT '最低订单金额',
    max_discount_amount DECIMAL(12,2) COMMENT '最高抵扣金额',

    -- 发放
    total_count         INT NOT NULL COMMENT '发放总数量',
    remain_count        INT NOT NULL COMMENT '剩余数量',
    per_user_limit      INT DEFAULT 1 COMMENT '每人限领数量',

    -- 使用条件
    applicable_type     ENUM('all','category','sku') DEFAULT 'all' COMMENT '适用范围类型',
    applicable_ids      JSON COMMENT '适用商品/类目ID列表',
    start_time          DATETIME NOT NULL COMMENT '开始时间',
    end_time            DATETIME NOT NULL COMMENT '结束时间',
    valid_days          INT COMMENT '领取后有效天数(与时间二选一)',

    -- 状态
    status              ENUM('draft','active','paused','expired') DEFAULT 'draft' COMMENT '状态',
    is_transferable     TINYINT DEFAULT 0 COMMENT '是否可转赠',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_coupon_id (coupon_id),
    INDEX idx_status_time (status, start_time, end_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券模板表';

-- 用户已领取优惠券
CREATE TABLE coupon_issues (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    issue_id            VARCHAR(32) NOT NULL COMMENT '发放记录ID',
    coupon_id           VARCHAR(32) NOT NULL COMMENT '优惠券模板ID',
    user_id             BIGINT UNSIGNED NOT NULL COMMENT '用户ID',

    -- 面额(冗余)
    face_amount         DECIMAL(12,2) NOT NULL COMMENT '面额',
    min_order_amount    DECIMAL(12,2) DEFAULT 0.00 COMMENT '最低订单金额',

    -- 状态
    status              ENUM('unused','used','expired','cancelled') DEFAULT 'unused' COMMENT '状态',

    -- 时间
    received_at         DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '领取时间',
    used_at             DATETIME COMMENT '使用时间',
    expired_at          DATETIME COMMENT '过期时间',

    -- 关联订单
    used_order_id       VARCHAR(32) COMMENT '使用的订单ID',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_issue_id (issue_id),
    UNIQUE KEY uk_user_coupon (user_id, coupon_id, issue_id),
    INDEX idx_user_status (user_id, status),
    INDEX idx_expired (expired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户优惠券表';
```

### 8.2 秒杀活动表 (seckill_activities)

参考京东秒杀设计。

```sql
CREATE TABLE seckill_activities (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    activity_id         VARCHAR(32) NOT NULL COMMENT '活动ID',
    name                VARCHAR(128) NOT NULL COMMENT '活动名称',
    description         VARCHAR(512) COMMENT '活动描述',

    -- 时间
    start_time          DATETIME NOT NULL COMMENT '开始时间',
    end_time            DATETIME NOT NULL COMMENT '结束时间',
    -- 秒杀场次
    session_time        DATETIME COMMENT '场次时间(整天秒杀时分段)',

    -- 库存
    total_stock         INT NOT NULL COMMENT '总库存',
    remain_stock        INT NOT NULL COMMENT '剩余库存',

    -- 限制
    per_user_limit      INT DEFAULT 1 COMMENT '每人限购数量',
    min_interval        INT DEFAULT 0 COMMENT '购买间隔(秒)',

    -- 状态
    status              ENUM('pending','active','paused','sold_out','ended') DEFAULT 'pending' COMMENT '状态',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_activity_id (activity_id),
    INDEX idx_time_status (start_time, end_time, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀活动表';

-- 秒杀商品表
CREATE TABLE seckill_skus (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    activity_id         VARCHAR(32) NOT NULL COMMENT '活动ID',
    sku_id              VARCHAR(32) NOT NULL COMMENT 'SKU ID',

    -- 价格
    seckill_price       DECIMAL(12,2) NOT NULL COMMENT '秒杀价',
    original_price      DECIMAL(12,2) NOT NULL COMMENT '原价',

    -- 库存
    total_stock         INT NOT NULL COMMENT '总数量',
    remain_stock        INT NOT NULL COMMENT '剩余数量',
    locked_stock        INT DEFAULT 0 COMMENT '锁定数量',

    -- 每人限购
    per_user_limit      INT DEFAULT 1 COMMENT '每人限购',

    -- 状态
    status              TINYINT DEFAULT 1 COMMENT '状态',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_activity_sku (activity_id, sku_id),
    INDEX idx_remain (remain_stock)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀商品表';

-- 秒杀订单表
CREATE TABLE seckill_orders (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_id            VARCHAR(32) NOT NULL COMMENT '秒杀订单号',
    activity_id         VARCHAR(32) NOT NULL COMMENT '活动ID',
    sku_id              VARCHAR(32) NOT NULL COMMENT 'SKU ID',
    user_id             BIGINT UNSIGNED NOT NULL COMMENT '用户ID',

    quantity            INT NOT NULL DEFAULT 1 COMMENT '购买数量',
    unit_price          DECIMAL(12,2) NOT NULL COMMENT '秒杀单价',
    total_amount        DECIMAL(12,2) NOT NULL COMMENT '总价',

    -- 状态
    status              ENUM('reserved','paid','cancelled','timeout') DEFAULT 'reserved' COMMENT '状态',
    reserved_at         DATETIME COMMENT '预占时间',
    expired_at          DATETIME NOT NULL COMMENT '订单过期时间',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_order_id (order_id),
    INDEX idx_user_id (user_id),
    INDEX idx_activity_sku (activity_id, sku_id),
    INDEX idx_expired (expired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀订单表';
```

---

## 九、推荐服务 (Recommendation Service)

### 9.1 用户行为日志表 (user_behavior_logs)

参考淘宝用户行为分析设计。

```sql
CREATE TABLE user_behavior_logs (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id             BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户ID(未登录为0)',
    anonymous_id        VARCHAR(64) COMMENT '匿名用户ID(cookie/device)',
    sku_id              BIGINT UNSIGNED NOT NULL COMMENT '商品SKU ID',
    spu_id              BIGINT UNSIGNED COMMENT '商品SPU ID',
    category_id         BIGINT UNSIGNED COMMENT '类目ID',

    behavior_type       ENUM('view','cart','favor','buy','comment','search') NOT NULL COMMENT '行为类型',
    source              VARCHAR(32) COMMENT '来源: home/search/recommend/category/detail',
    stay_duration       INT COMMENT '停留时长(秒)',
    search_keyword      VARCHAR(128) COMMENT '搜索关键词',

    -- 时间分桶(用于离线分析)
    time_bucket         INT NOT NULL COMMENT '5分钟时间分桶',

    -- 商品快照
    product_name        VARCHAR(256) COMMENT '商品名称快照',
    product_price       DECIMAL(12,2) COMMENT '商品价格快照',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_user_type_time (user_id, behavior_type, created_at),
    INDEX idx_sku_id (sku_id),
    INDEX idx_category_time (category_id, created_at),
    INDEX idx_anonymous (anonymous_id),
    INDEX idx_search_keyword (search_keyword),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户行为日志表';
```

### 9.2 商品相似度表 (product_similarity)

基于Item-CF的协同过滤。

```sql
CREATE TABLE product_similarity (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id_a            BIGINT UNSIGNED NOT NULL COMMENT '商品A',
    sku_id_b            BIGINT UNSIGNED NOT NULL COMMENT '商品B',
    scene               ENUM('cart','favor','buy','view') NOT NULL COMMENT '场景',
    similarity          DECIMAL(10,6) NOT NULL COMMENT '相似度: 0-1',
    weight              INT DEFAULT 1 COMMENT '共现次数',
    co_cart_count       INT DEFAULT 0 COMMENT '同加购次数',
    co_buy_count        INT DEFAULT 0 COMMENT '同购买次数',
    co_view_count       INT DEFAULT 0 COMMENT '同浏览次数',

    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (sku_id_a, sku_id_b, scene),
    INDEX idx_sku_a_scene (sku_id_a, scene),
    INDEX idx_similarity (similarity DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品相似度表(Item-CF)';
```

### 9.3 关联规则表 (association_rules)

基于Apriori算法的"买了还买"推荐。

```sql
CREATE TABLE association_rules (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id_a            BIGINT UNSIGNED NOT NULL COMMENT '前提商品(A)',
    sku_id_b            BIGINT UNSIGNED NOT NULL COMMENT '结果商品(B)',

    -- 关联指标
    support             DECIMAL(10,6) NOT NULL COMMENT '支持度: P(A∩B)',
    confidence          DECIMAL(10,6) NOT NULL COMMENT '置信度: P(B|A)',
    lift                DECIMAL(10,6) NOT NULL COMMENT '提升度',

    -- 统计
    co_purchase_count   INT NOT NULL DEFAULT 0 COMMENT '共购次数',
    a_purchase_count    INT NOT NULL DEFAULT 0 COMMENT 'A购买次数',
    b_purchase_count    INT NOT NULL DEFAULT 0 COMMENT 'B购买次数',

    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_rule (sku_id_a, sku_id_b),
    INDEX idx_sku_a (sku_id_a),
    INDEX idx_sku_b (sku_id_b),
    INDEX idx_lift (lift DESC),
    INDEX idx_confidence (confidence DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='关联规则表(Apriori)';
```

---

## 十、通知服务 (Notification Service)

### 10.1 通知记录表 (notifications)

```sql
CREATE TABLE notifications (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    notification_id     VARCHAR(32) NOT NULL COMMENT '通知ID',
    user_id             VARCHAR(32) NOT NULL COMMENT '用户ID',
    type                VARCHAR(32) NOT NULL COMMENT '通知类型',

    -- 内容
    title               VARCHAR(128) NOT NULL COMMENT '标题',
    content             TEXT COMMENT '内容',
    data_json           JSON COMMENT '扩展数据(跳转等)',

    -- 状态
    is_read             TINYINT DEFAULT 0 COMMENT '是否已读',
    read_at             DATETIME COMMENT '阅读时间',

    -- 渠道
    channels            JSON COMMENT '推送渠道: ["app","sms","email"]',
    push_status         TINYINT DEFAULT 0 COMMENT '推送状态',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uk_notification_id (notification_id),
    INDEX idx_user_read_time (user_id, is_read, created_at),
    INDEX idx_type (type),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知记录表';

-- 通知类型枚举
-- order_paid: 订单支付成功
-- order_shipped: 订单发货
-- order_delivered: 订单送达
-- refund_approved: 退款审核通过
-- refund_success: 退款成功
-- seckill_start: 秒杀开始
-- price_drop: 价格降价
-- coupon_expiring: 优惠券即将过期
```

### 10.2 通知订阅配置表 (notification_configs)

```sql
CREATE TABLE notification_configs (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id             VARCHAR(32) NOT NULL COMMENT '用户ID',
    type                VARCHAR(32) NOT NULL COMMENT '通知类型',
    channel             ENUM('app','sms','email','wechat') NOT NULL COMMENT '推送渠道',

    enabled             TINYINT DEFAULT 1 COMMENT '是否启用',
    threshold           INT DEFAULT 0 COMMENT '阈值(如: 降价超过X元通知)',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_user_type_channel (user_id, type, channel),
    INDEX idx_user_type (user_id, type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知订阅配置表';
```

---

## 十一、价格监控服务 (PriceWatch Service)

### 11.1 价格监控表 (user_price_watches)

```sql
CREATE TABLE user_price_watches (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    watch_id            VARCHAR(32) NOT NULL COMMENT '监控ID',
    user_id             BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    sku_id              VARCHAR(32) NOT NULL COMMENT '商品SKU ID',

    -- 价格设置
    target_price        DECIMAL(12,2) COMMENT '目标价格(NULL表示任意降价)',
    current_price       DECIMAL(12,2) NOT NULL COMMENT '监控时的价格',

    -- 通知设置
    notify_enabled      TINYINT DEFAULT 1 COMMENT '是否启用通知',
    last_notify_at      DATETIME COMMENT '最后通知时间',
    notify_count        INT DEFAULT 0 COMMENT '通知次数',
    notify_interval     INT DEFAULT 24 COMMENT '通知间隔(小时)',

    -- 状态
    status              ENUM('active','stopped','price_reached','expired') DEFAULT 'active' COMMENT '状态',
    expire_at           DATE COMMENT '过期日期(NULL永不过期)',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_user_sku (user_id, sku_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_expire (expire_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='价格监控表';
```

### 11.2 价格历史快照表 (product_price_snapshots)

```sql
CREATE TABLE product_price_snapshots (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_id              VARCHAR(32) NOT NULL COMMENT 'SKU ID',
    price               DECIMAL(12,2) NOT NULL COMMENT '价格',
    original_price      DECIMAL(12,2) COMMENT '原价',
    promotion_price     DECIMAL(12,2) COMMENT '促销价',
    snapshot_date       DATE NOT NULL COMMENT '快照日期',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uk_sku_date (sku_id, snapshot_date),
    INDEX idx_sku_id (sku_id),
    INDEX idx_snapshot_date (snapshot_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品价格快照表';
```

---

## 十二、任务服务 (Task Service)

### 12.1 异步任务表 (tasks)

```sql
CREATE TABLE tasks (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    task_id             VARCHAR(32) NOT NULL COMMENT '任务ID',
    biz_no              VARCHAR(64) COMMENT '业务编号',
    order_id            VARCHAR(32) COMMENT '关联订单',

    -- 任务类型
    type                VARCHAR(32) NOT NULL COMMENT '任务类型: order_fulfill/inventory_confirm/payment_check',
    priority            INT DEFAULT 5 COMMENT '优先级: 1-10',
    payload             JSON COMMENT '任务参数',

    -- 执行状态
    status              ENUM('pending','running','success','failed','retry') DEFAULT 'pending' COMMENT '状态',
    retry_count         INT DEFAULT 0 COMMENT '重试次数',
    max_retry           INT DEFAULT 3 COMMENT '最大重试',
    next_retry_at       DATETIME COMMENT '下次重试时间',
    last_error          TEXT COMMENT '最后错误信息',

    -- 执行记录
    execute_at          DATETIME COMMENT '执行时间',
    execute_duration    INT COMMENT '执行耗时(ms)',
    result              TEXT COMMENT '执行结果',

    -- 时间
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    expired_at          DATETIME COMMENT '任务过期时间',

    UNIQUE KEY uk_task_id (task_id),
    INDEX idx_type_status (type, status),
    INDEX idx_priority_time (priority, created_at),
    INDEX idx_next_retry (next_retry_at),
    INDEX idx_order_id (order_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='异步任务表';

-- 任务类型说明
-- order_fulfill: 订单履约(确认支付->扣减库存->发货)
-- inventory_confirm: 库存确认(支付成功后确认预占)
-- payment_check: 支付状态检查(轮询第三方支付状态)
-- seckill_settle: 秒杀订单结算
-- refund_process: 退款处理
```

### 12.2 SAGA事务表 (sagas)

```sql
CREATE TABLE sagas (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    saga_id             VARCHAR(32) NOT NULL COMMENT 'Saga ID',
    biz_no              VARCHAR(64) NOT NULL COMMENT '业务编号',
    type                VARCHAR(32) NOT NULL COMMENT 'Saga类型: order_create/payment_refund',

    -- 状态
    status              ENUM('running','success','failed','compensating','compensated') DEFAULT 'running' COMMENT '状态',
    current_step        VARCHAR(32) COMMENT '当前步骤',
    reason              TEXT COMMENT '失败原因',

    -- 执行统计
    total_steps         INT NOT NULL COMMENT '总步骤数',
    completed_steps     INT DEFAULT 0 COMMENT '已完成步骤',
    compensation_steps INT DEFAULT 0 COMMENT '已补偿步骤',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    finished_at         DATETIME COMMENT '完成时间',

    UNIQUE KEY uk_saga_id (saga_id),
    INDEX idx_biz_no (biz_no),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Saga事务表';

-- Saga步骤表
CREATE TABLE saga_steps (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    step_id             VARCHAR(32) NOT NULL COMMENT '步骤ID',
    saga_id             VARCHAR(32) NOT NULL COMMENT 'Saga ID',
    step_name           VARCHAR(32) NOT NULL COMMENT '步骤名称',
    step_order          INT NOT NULL COMMENT '步骤顺序',

    -- 执行状态
    status              ENUM('pending','executing','success','failed','compensated') DEFAULT 'pending' COMMENT '状态',
    payload             JSON COMMENT '步骤输入参数',
    result              TEXT COMMENT '执行结果',

    -- 补偿
    compensation_type   VARCHAR(32) COMMENT '补偿类型: release/rollback/refund',
    compensation_payload JSON COMMENT '补偿参数',

    -- 时间
    executed_at         DATETIME COMMENT '执行时间',
    compensated_at      DATETIME COMMENT '补偿时间',

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_step_id (step_id),
    INDEX idx_saga_id (saga_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Saga步骤表';
```

---

## 十三、通用组件表

### 13.1 操作日志表 (admin_operation_logs)

```sql
CREATE TABLE admin_operation_logs (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    operator_id         VARCHAR(32) NOT NULL COMMENT '操作人ID',
    operator_name       VARCHAR(64) COMMENT '操作人名称',
    operator_type       ENUM('admin','merchant','system') NOT NULL COMMENT '操作人类型',
    action              VARCHAR(64) NOT NULL COMMENT '操作动作',
    resource_type       VARCHAR(32) NOT NULL COMMENT '资源类型',
    resource_id         VARCHAR(32) COMMENT '资源ID',
    before_value        JSON COMMENT '变更前值',
    after_value         JSON COMMENT '变更后值',
    client_ip           VARCHAR(45) COMMENT '客户端IP',
    user_agent          VARCHAR(512) COMMENT 'User-Agent',
    request_id          VARCHAR(64) COMMENT '请求追踪ID',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_operator_time (operator_id, created_at),
    INDEX idx_resource (resource_type, resource_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志表';
```

### 13.2 系统配置表 (system_configs)

```sql
CREATE TABLE system_configs (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    config_key          VARCHAR(64) NOT NULL COMMENT '配置键',
    config_value        TEXT COMMENT '配置值',
    config_type         VARCHAR(32) DEFAULT 'string' COMMENT '配置类型',
    config_group        VARCHAR(32) COMMENT '配置分组',
    description         VARCHAR(256) COMMENT '配置描述',
    is_encrypted        TINYINT DEFAULT 0 COMMENT '是否加密',
    status              TINYINT DEFAULT 1 COMMENT '状态',
    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_config_key (config_key),
    INDEX idx_config_group (config_group)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置表';

-- 配置分组示例
-- order: 订单配置 (订单超时时间/自动取消等)
-- payment: 支付配置 (渠道密钥/手续费等)
-- inventory: 库存配置 (预占超时/低库存阈值等)
-- seckill: 秒杀配置 (活动时长/限购数量等)
```

---

## 十四、索引设计原则

参考淘宝/京东最佳实践:

1. **主键索引**: 所有表使用BIGINT自增主键 + 业务ID(如order_id)作为逻辑主键
2. **唯一索引**: 关键业务字段添加UNIQUE约束
3. **联合索引**: 根据查询场景设计,遵循最左前缀原则
4. **分页索引**: 列表查询使用 `(status, created_at)` 复合索引
5. **防止重复**: 使用 `UNIQUE KEY uk_user_sku (user_id, sku_id)` 防止重复记录
6. **全文索引**: 商品搜索使用 `FULLTEXT idx_name (name, keywords)`
7. **避免跨域**: 不在低区分度字段上建索引(如status只有几个值)

---

## 十五、分库分表策略(可选)

参考淘宝双十一数据量:

| 表名 | 预估年数据量 | 分库策略 | 分表策略 |
|------|------------|---------|---------|
| orders | 1亿+ | 按用户ID取模 | 按月分表 |
| order_items | 5亿+ | 按订单ID哈希 | 按月分表 |
| payments | 1亿+ | 按用户ID取模 | 按月分表 |
| user_behavior_logs | 10亿+ | 按用户ID取模 | 按天分表 |
| inventory_logs | 5亿+ | 按SKU ID取模 | 按天分表 |

---

## 十六、字段设计规范

1. **金额字段**: 使用 DECIMAL(12,2), 存储单位为"元"
2. **价格字段**: 使用 DECIMAL(12,2), 存储单位为"元"
3. **数量字段**: 使用 INT 或 BIGINT, 根据预估最大值选择
4. **主键ID**: 使用 BIGINT UNSIGNED AUTO_INCREMENT + 业务ID
5. **时间字段**:
   - 精确到秒: DATETIME
   - 精确到日期: DATE
   - 需要时区: 存储UTC, 应用层转换
6. **状态字段**: 使用 ENUM 类型, 便于扩展
7. **JSON字段**: 用于扩展字段, 不常查询的属性
8. **软删除**: 使用 `deleted_at DATETIME` 而非 is_deleted 标志

---

*文档版本: v1.0*
*最后更新: 2026-04-23*
*参考: 淘宝/京东/拼多多等主流电商平台数据库设计*