-- ============================================================================
-- 电商网站数据库建表脚本 - 优化商用版 V2
-- 参考文档: 个人电商网站数据库设计文档_优化商用版V2.md
-- 版本: V2.0
-- 适用范围: 中小型电商站、独立站、单商户先行并预留多商户能力的项目
-- ============================================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;
SET SQL_MODE = 'STRICT_TRANS_TABLES';

-- ----------------------------------------------------------------------------
-- 4. 全局基础能力表
-- ----------------------------------------------------------------------------

-- 4.1 商户表（预留多商户能力）
DROP TABLE IF EXISTS merchants;
CREATE TABLE merchants (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    merchant_no         VARCHAR(32) NOT NULL COMMENT '商户编号',
    merchant_name       VARCHAR(128) NOT NULL COMMENT '商户名称',
    merchant_type       TINYINT NOT NULL DEFAULT 1 COMMENT '1个人店 2企业店 3平台自营',
    contact_name        VARCHAR(64) NULL,
    contact_phone       VARCHAR(32) NULL,
    contact_email       VARCHAR(128) NULL,
    status              TINYINT NOT NULL DEFAULT 1 COMMENT '1正常 2冻结 3关闭',
    settled_at          DATETIME(3) NULL,
    remark              VARCHAR(255) NULL,
    ext_json            JSON NULL,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_merchant_no (merchant_no),
    KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商户表';

-- 4.2 店铺表
DROP TABLE IF EXISTS shops;
CREATE TABLE shops (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    shop_no             VARCHAR(32) NOT NULL COMMENT '店铺编号',
    merchant_no         VARCHAR(32) NOT NULL COMMENT '商户编号',
    shop_name           VARCHAR(128) NOT NULL COMMENT '店铺名称',
    shop_logo_url       VARCHAR(512) NULL,
    shop_banner_url     VARCHAR(512) NULL,
    shop_status         TINYINT NOT NULL DEFAULT 1 COMMENT '1营业中 2打烊 3关闭',
    service_phone       VARCHAR(32) NULL,
    country_code        VARCHAR(16) NOT NULL DEFAULT 'CN',
    currency_code       CHAR(3) NOT NULL DEFAULT 'CNY',
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_shop_no (shop_no),
    KEY idx_merchant_no (merchant_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='店铺表';

-- ----------------------------------------------------------------------------
-- 5. 用户域设计
-- ----------------------------------------------------------------------------

-- 5.1 用户主表
DROP TABLE IF EXISTS users;
CREATE TABLE users (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_no             VARCHAR(32) NOT NULL COMMENT '用户编号',
    username            VARCHAR(64) NULL COMMENT '登录名',
    display_name        VARCHAR(64) NULL COMMENT '显示昵称',
    password_hash       VARCHAR(255) NOT NULL COMMENT '密码哈希',
    phone               VARCHAR(20) NULL COMMENT '手机号',
    email               VARCHAR(128) NULL COMMENT '邮箱',
    avatar_url          VARCHAR(512) NULL COMMENT '头像',
    status              TINYINT NOT NULL DEFAULT 1 COMMENT '1正常 2禁用 3注销',
    register_source     VARCHAR(32) NOT NULL DEFAULT 'phone' COMMENT '注册来源',
    last_login_at       DATETIME(3) NULL COMMENT '最后登录时间',
    last_login_ip       VARCHAR(45) NULL COMMENT '最后登录IP',
    password_updated_at DATETIME(3) NULL COMMENT '密码最后变更时间',
    deleted_at          DATETIME(3) NULL COMMENT '软删除时间',
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_user_no (user_no),
    UNIQUE KEY uk_phone (phone),
    UNIQUE KEY uk_email (email),
    UNIQUE KEY uk_username (username),
    KEY idx_status_created (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户主表';

-- 5.2 用户资料表
DROP TABLE IF EXISTS user_profiles;
CREATE TABLE user_profiles (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_no             VARCHAR(32) NOT NULL,
    gender              TINYINT NOT NULL DEFAULT 0 COMMENT '0未知 1男 2女',
    birthday            DATE NULL,
    country_code        VARCHAR(16) NOT NULL DEFAULT 'CN',
    language_code       VARCHAR(16) NOT NULL DEFAULT 'zh-CN',
    timezone            VARCHAR(64) NOT NULL DEFAULT 'Asia/Shanghai',
    bio                 VARCHAR(255) NULL,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_user_no (user_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户资料表';

-- 5.3 用户安全表
DROP TABLE IF EXISTS user_security;
CREATE TABLE user_security (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_no             VARCHAR(32) NOT NULL,
    email_verified      TINYINT NOT NULL DEFAULT 0,
    phone_verified      TINYINT NOT NULL DEFAULT 0,
    two_factor_enabled  TINYINT NOT NULL DEFAULT 0,
    failed_login_count  INT NOT NULL DEFAULT 0,
    lock_until          DATETIME(3) NULL,
    risk_level          TINYINT NOT NULL DEFAULT 1 COMMENT '1低 2中 3高',
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_user_no (user_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户安全表';

-- 5.4 用户社交绑定表
DROP TABLE IF EXISTS user_social_bindings;
CREATE TABLE user_social_bindings (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_no             VARCHAR(32) NOT NULL,
    provider            VARCHAR(32) NOT NULL COMMENT 'wechat/google/apple/github等',
    provider_user_id    VARCHAR(128) NOT NULL COMMENT '第三方用户ID',
    union_id            VARCHAR(128) NULL,
    nickname            VARCHAR(128) NULL,
    avatar_url          VARCHAR(512) NULL,
    is_primary          TINYINT NOT NULL DEFAULT 0,
    bind_status         TINYINT NOT NULL DEFAULT 1 COMMENT '1已绑定 0已解绑',
    bound_at            DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    unbound_at          DATETIME(3) NULL,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_provider_uid (provider, provider_user_id),
    KEY idx_user_no (user_no),
    KEY idx_union_id (union_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户社交绑定表';

-- 5.5 用户地址表
DROP TABLE IF EXISTS user_addresses;
CREATE TABLE user_addresses (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    address_no          VARCHAR(32) NOT NULL COMMENT '地址编号',
    user_no             VARCHAR(32) NOT NULL COMMENT '用户编号',
    receiver_name       VARCHAR(64) NOT NULL COMMENT '收货人姓名',
    receiver_phone      VARCHAR(20) NOT NULL COMMENT '收货人手机号',
    country_code        VARCHAR(16) NOT NULL DEFAULT 'CN',
    province_code       VARCHAR(16) NOT NULL,
    city_code           VARCHAR(16) NOT NULL,
    district_code       VARCHAR(16) NOT NULL,
    province_name       VARCHAR(64) NOT NULL,
    city_name           VARCHAR(64) NOT NULL,
    district_name       VARCHAR(64) NOT NULL,
    detail_address      VARCHAR(255) NOT NULL,
    postal_code         VARCHAR(16) NULL,
    address_tag         VARCHAR(32) NULL COMMENT '家/公司/学校',
    lng                 DECIMAL(11,6) NULL,
    lat                 DECIMAL(11,6) NULL,
    is_default          TINYINT NOT NULL DEFAULT 0,
    deleted_at          DATETIME(3) NULL,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_address_no (address_no),
    KEY idx_user_default (user_no, is_default),
    KEY idx_user_created (user_no, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户地址表';

-- 5.6 用户统计表
DROP TABLE IF EXISTS user_stats;
CREATE TABLE user_stats (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_no             VARCHAR(32) NOT NULL,
    order_count         BIGINT NOT NULL DEFAULT 0,
    paid_order_count    BIGINT NOT NULL DEFAULT 0,
    total_paid_amount   BIGINT NOT NULL DEFAULT 0 COMMENT '单位分',
    refund_count        BIGINT NOT NULL DEFAULT 0,
    total_refund_amount BIGINT NOT NULL DEFAULT 0 COMMENT '单位分',
    points_balance      BIGINT NOT NULL DEFAULT 0,
    collect_count       BIGINT NOT NULL DEFAULT 0,
    last_order_at       DATETIME(3) NULL,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_user_no (user_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户统计表';

-- ----------------------------------------------------------------------------
-- 6. 商品域设计
-- ----------------------------------------------------------------------------

-- 6.1 商品品牌表
DROP TABLE IF EXISTS product_brands;
CREATE TABLE product_brands (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    brand_no            VARCHAR(32) NOT NULL,
    brand_name          VARCHAR(128) NOT NULL,
    brand_logo_url      VARCHAR(512) NULL,
    brand_desc          VARCHAR(255) NULL,
    country_code        VARCHAR(16) NULL,
    status              TINYINT NOT NULL DEFAULT 1,
    sort_no             INT NOT NULL DEFAULT 0,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_brand_no (brand_no),
    KEY idx_status_sort (status, sort_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品品牌表';

-- 6.2 商品类目表
DROP TABLE IF EXISTS product_categories;
CREATE TABLE product_categories (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    category_no         VARCHAR(32) NOT NULL,
    parent_no           VARCHAR(32) NULL,
    category_name       VARCHAR(128) NOT NULL,
    category_level      TINYINT NOT NULL DEFAULT 1,
    category_path       VARCHAR(255) NOT NULL COMMENT '如 001/001001/001001003',
    is_leaf             TINYINT NOT NULL DEFAULT 0,
    sort_no             INT NOT NULL DEFAULT 0,
    status              TINYINT NOT NULL DEFAULT 1,
    icon_url            VARCHAR(512) NULL,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_category_no (category_no),
    KEY idx_parent_sort (parent_no, sort_no),
    KEY idx_path (category_path)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品类目表';

-- 6.3 商品属性表
DROP TABLE IF EXISTS product_attributes;
CREATE TABLE product_attributes (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    attr_no             VARCHAR(32) NOT NULL,
    attr_name           VARCHAR(64) NOT NULL,
    attr_type           TINYINT NOT NULL DEFAULT 1 COMMENT '1基础属性 2销售属性 3筛选属性',
    input_type          TINYINT NOT NULL DEFAULT 1 COMMENT '1单选 2多选 3输入',
    unit                VARCHAR(32) NULL,
    is_required         TINYINT NOT NULL DEFAULT 0,
    status              TINYINT NOT NULL DEFAULT 1,
    sort_no             INT NOT NULL DEFAULT 0,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_attr_no (attr_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品属性表';

-- 6.4 商品属性值表
DROP TABLE IF EXISTS product_attribute_values;
CREATE TABLE product_attribute_values (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    value_no            VARCHAR(32) NOT NULL,
    attr_no             VARCHAR(32) NOT NULL,
    value_name          VARCHAR(128) NOT NULL,
    value_alias         VARCHAR(128) NULL,
    image_url           VARCHAR(512) NULL,
    status              TINYINT NOT NULL DEFAULT 1,
    sort_no             INT NOT NULL DEFAULT 0,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_value_no (value_no),
    KEY idx_attr_sort (attr_no, sort_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品属性值表';

-- 6.5 类目属性映射表
DROP TABLE IF EXISTS category_attributes;
CREATE TABLE category_attributes (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    category_no         VARCHAR(32) NOT NULL,
    attr_no             VARCHAR(32) NOT NULL,
    is_required         TINYINT NOT NULL DEFAULT 0,
    is_filterable       TINYINT NOT NULL DEFAULT 0,
    is_searchable       TINYINT NOT NULL DEFAULT 0,
    is_sale_attr        TINYINT NOT NULL DEFAULT 0,
    sort_no             INT NOT NULL DEFAULT 0,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_category_attr (category_no, attr_no),
    KEY idx_category_sort (category_no, sort_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='类目属性映射表';

-- 6.6 商品 SPU 表
DROP TABLE IF EXISTS product_spus;
CREATE TABLE product_spus (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    spu_no              VARCHAR(32) NOT NULL,
    merchant_no         VARCHAR(32) NOT NULL,
    shop_no             VARCHAR(32) NOT NULL,
    brand_no            VARCHAR(32) NULL,
    category_no         VARCHAR(32) NOT NULL,
    spu_name            VARCHAR(255) NOT NULL,
    spu_subtitle        VARCHAR(255) NULL,
    product_type        TINYINT NOT NULL DEFAULT 1 COMMENT '1普通商品 2虚拟商品 3预售商品',
    audit_status        TINYINT NOT NULL DEFAULT 0 COMMENT '0待提审 1审核中 2通过 3拒绝',
    publish_status      TINYINT NOT NULL DEFAULT 0 COMMENT '0草稿 1上架 2下架 3删除',
    main_image_url      VARCHAR(512) NULL,
    video_url           VARCHAR(512) NULL,
    keywords            VARCHAR(255) NULL,
    seo_title           VARCHAR(255) NULL,
    seo_desc            VARCHAR(512) NULL,
    detail_html         MEDIUMTEXT NULL,
    review_count        BIGINT NOT NULL DEFAULT 0,
    avg_rating          DECIMAL(4,2) NOT NULL DEFAULT 5.00,
    sale_count          BIGINT NOT NULL DEFAULT 0 COMMENT '展示销量，异步汇总',
    collect_count       BIGINT NOT NULL DEFAULT 0 COMMENT '展示收藏量，异步汇总',
    listed_at           DATETIME(3) NULL,
    unlisted_at         DATETIME(3) NULL,
    deleted_at          DATETIME(3) NULL,
    version             INT NOT NULL DEFAULT 1,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_spu_no (spu_no),
    KEY idx_shop_status (shop_no, publish_status),
    KEY idx_category_status (category_no, publish_status),
    KEY idx_brand_status (brand_no, publish_status),
    FULLTEXT KEY ft_spu_name_keywords (spu_name, keywords)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品SPU表';

-- 6.7 商品 SKU 表
DROP TABLE IF EXISTS product_skus;
CREATE TABLE product_skus (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_no              VARCHAR(32) NOT NULL,
    spu_no              VARCHAR(32) NOT NULL,
    merchant_no         VARCHAR(32) NOT NULL,
    shop_no             VARCHAR(32) NOT NULL,
    sku_code            VARCHAR(64) NULL COMMENT '商家内部编码',
    barcode             VARCHAR(64) NULL,
    sku_name            VARCHAR(255) NOT NULL,
    currency_code       CHAR(3) NOT NULL DEFAULT 'CNY',
    sale_price          BIGINT NOT NULL COMMENT '销售价，单位分',
    market_price        BIGINT NOT NULL DEFAULT 0 COMMENT '划线价，单位分',
    cost_price          BIGINT NOT NULL DEFAULT 0 COMMENT '成本价，单位分',
    tax_rate            DECIMAL(6,4) NOT NULL DEFAULT 0.0000 COMMENT '税率，如0.1300',
    weight_gram         INT NOT NULL DEFAULT 0 COMMENT '重量，克',
    volume_cm3          INT NOT NULL DEFAULT 0 COMMENT '体积，立方厘米',
    sale_status         TINYINT NOT NULL DEFAULT 1 COMMENT '1可售 2不可售',
    status              TINYINT NOT NULL DEFAULT 1 COMMENT '1正常 2禁用 3删除',
    is_default          TINYINT NOT NULL DEFAULT 0,
    sold_count          BIGINT NOT NULL DEFAULT 0 COMMENT '展示销量，异步汇总',
    ext_json            JSON NULL,
    version             INT NOT NULL DEFAULT 1,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_sku_no (sku_no),
    KEY idx_spu_status (spu_no, status),
    KEY idx_shop_sale (shop_no, sale_status),
    KEY idx_price (sale_price)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品SKU表';

-- 6.8 SKU 规格值表
DROP TABLE IF EXISTS sku_specs;
CREATE TABLE sku_specs (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_no              VARCHAR(32) NOT NULL,
    attr_no             VARCHAR(32) NOT NULL,
    attr_name           VARCHAR(64) NOT NULL,
    value_no            VARCHAR(32) NULL,
    value_name          VARCHAR(128) NOT NULL,
    sort_no             INT NOT NULL DEFAULT 0,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_sku_attr (sku_no, attr_no),
    KEY idx_sku_sort (sku_no, sort_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SKU规格值表';

-- 6.9 商品图片表
DROP TABLE IF EXISTS product_images;
CREATE TABLE product_images (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    image_no            VARCHAR(32) NOT NULL,
    spu_no              VARCHAR(32) NULL,
    sku_no              VARCHAR(32) NULL,
    image_type          TINYINT NOT NULL DEFAULT 1 COMMENT '1主图 2详情图 3规格图 4白底图',
    image_url           VARCHAR(512) NOT NULL,
    sort_no             INT NOT NULL DEFAULT 0,
    is_main             TINYINT NOT NULL DEFAULT 0,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_image_no (image_no),
    KEY idx_spu_sort (spu_no, sort_no),
    KEY idx_sku_sort (sku_no, sort_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品图片表';

-- 6.10 商品价格快照表
DROP TABLE IF EXISTS product_price_snapshots;
CREATE TABLE product_price_snapshots (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    sku_no              VARCHAR(32) NOT NULL,
    snapshot_date       DATE NOT NULL,
    currency_code       CHAR(3) NOT NULL DEFAULT 'CNY',
    sale_price          BIGINT NOT NULL,
    market_price        BIGINT NOT NULL DEFAULT 0,
    promotion_price     BIGINT NOT NULL DEFAULT 0,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_sku_date (sku_no, snapshot_date),
    KEY idx_snapshot_date (snapshot_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品价格快照表';

-- ----------------------------------------------------------------------------
-- 7. 购物车域设计
-- ----------------------------------------------------------------------------

-- 7.1 购物车主表
DROP TABLE IF EXISTS shopping_carts;
CREATE TABLE shopping_carts (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    cart_no             VARCHAR(32) NOT NULL,
    user_no             VARCHAR(32) NOT NULL,
    cart_status         TINYINT NOT NULL DEFAULT 1 COMMENT '1有效 2失效',
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_cart_no (cart_no),
    UNIQUE KEY uk_user_no (user_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='购物车主表';

-- 7.2 购物车明细表
DROP TABLE IF EXISTS shopping_cart_items;
CREATE TABLE shopping_cart_items (
    id                  BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    cart_item_no        VARCHAR(32) NOT NULL,
    cart_no             VARCHAR(32) NOT NULL,
    user_no             VARCHAR(32) NOT NULL,
    sku_no              VARCHAR(32) NOT NULL,
    quantity            INT NOT NULL,
    checked             TINYINT NOT NULL DEFAULT 1,
    price_snapshot      BIGINT NOT NULL COMMENT '加入购物车时价格快照，单位分',
    promotion_snapshot  JSON NULL COMMENT '加入购物车时促销快照',
    deleted_at          DATETIME(3) NULL,
    created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_cart_item_no (cart_item_no),
    UNIQUE KEY uk_user_sku_deleted (user_no, sku_no, deleted_at),
    KEY idx_cart_no (cart_no),
    KEY idx_user_checked (user_no, checked)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='购物车明细表';

-- ----------------------------------------------------------------------------
-- 8. 订单域设计
-- ----------------------------------------------------------------------------

-- 8.2 订单主表
DROP TABLE IF EXISTS orders;
CREATE TABLE orders (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_no                VARCHAR(32) NOT NULL COMMENT '订单号',
    parent_order_no         VARCHAR(32) NULL COMMENT '父订单号，拆单场景预留',
    merchant_no             VARCHAR(32) NOT NULL,
    shop_no                 VARCHAR(32) NOT NULL,
    user_no                 VARCHAR(32) NOT NULL,
    order_type              VARCHAR(32) NOT NULL DEFAULT 'normal' COMMENT 'normal/seckill/pre_sale',
    order_source            VARCHAR(32) NOT NULL DEFAULT 'pc' COMMENT 'pc/app/h5/mini_program/admin',
    order_status            VARCHAR(32) NOT NULL DEFAULT 'pending_payment',
    after_sale_status       VARCHAR(32) NOT NULL DEFAULT 'none',
    currency_code           CHAR(3) NOT NULL DEFAULT 'CNY',
    item_amount             BIGINT NOT NULL COMMENT '商品原始总额，分',
    discount_amount         BIGINT NOT NULL DEFAULT 0 COMMENT '总优惠金额，分',
    coupon_amount           BIGINT NOT NULL DEFAULT 0 COMMENT '优惠券金额，分',
    promotion_amount        BIGINT NOT NULL DEFAULT 0 COMMENT '活动优惠金额，分',
    points_deduct_amount    BIGINT NOT NULL DEFAULT 0 COMMENT '积分抵扣金额，分',
    freight_amount          BIGINT NOT NULL DEFAULT 0 COMMENT '运费，分',
    tax_amount              BIGINT NOT NULL DEFAULT 0 COMMENT '税费，分',
    payable_amount          BIGINT NOT NULL COMMENT '应付金额，分',
    paid_amount             BIGINT NOT NULL DEFAULT 0 COMMENT '已支付金额，分',
    buyer_message           VARCHAR(255) NULL,
    seller_remark           VARCHAR(255) NULL,
    address_snapshot        JSON NOT NULL COMMENT '收货地址快照',
    invoice_snapshot        JSON NULL COMMENT '发票信息快照',
    promotion_snapshot      JSON NULL COMMENT '整单营销快照',
    device_id               VARCHAR(64) NULL,
    client_ip               VARCHAR(45) NULL,
    trace_id                VARCHAR(64) NULL,
    payment_deadline_at     DATETIME(3) NULL,
    paid_at                 DATETIME(3) NULL,
    completed_at            DATETIME(3) NULL,
    cancelled_at            DATETIME(3) NULL,
    cancel_reason           VARCHAR(255) NULL,
    close_reason            VARCHAR(255) NULL,
    deleted_at              DATETIME(3) NULL,
    version                 INT NOT NULL DEFAULT 1,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_order_no (order_no),
    KEY idx_user_status_time (user_no, order_status, created_at),
    KEY idx_shop_status_time (shop_no, order_status, created_at),
    KEY idx_paid_at (paid_at),
    KEY idx_deadline (payment_deadline_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单主表';

-- 8.3 订单明细表
DROP TABLE IF EXISTS order_items;
CREATE TABLE order_items (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    order_item_no           VARCHAR(32) NOT NULL,
    order_no                VARCHAR(32) NOT NULL,
    merchant_no             VARCHAR(32) NOT NULL,
    shop_no                 VARCHAR(32) NOT NULL,
    user_no                 VARCHAR(32) NOT NULL,
    spu_no                  VARCHAR(32) NOT NULL,
    sku_no                  VARCHAR(32) NOT NULL,
    brand_no                VARCHAR(32) NULL,
    category_no             VARCHAR(32) NOT NULL,
    sku_name                VARCHAR(255) NOT NULL,
    sku_image_url           VARCHAR(512) NULL,
    sku_spec_snapshot       JSON NOT NULL COMMENT 'SKU规格快照',
    quantity                INT NOT NULL,
    unit_price              BIGINT NOT NULL COMMENT '成交单价，分',
    market_price            BIGINT NOT NULL DEFAULT 0 COMMENT '原价，分',
    cost_price              BIGINT NOT NULL DEFAULT 0 COMMENT '成本价，分',
    item_amount             BIGINT NOT NULL COMMENT '明细总额，分',
    discount_amount         BIGINT NOT NULL DEFAULT 0 COMMENT '分摊优惠总额，分',
    coupon_split_amount     BIGINT NOT NULL DEFAULT 0 COMMENT '券分摊，分',
    promotion_split_amount  BIGINT NOT NULL DEFAULT 0 COMMENT '活动分摊，分',
    points_split_amount     BIGINT NOT NULL DEFAULT 0 COMMENT '积分分摊，分',
    freight_split_amount    BIGINT NOT NULL DEFAULT 0 COMMENT '运费分摊，分',
    tax_amount              BIGINT NOT NULL DEFAULT 0 COMMENT '税费分摊，分',
    payable_amount          BIGINT NOT NULL COMMENT '明细应付，分',
    paid_amount             BIGINT NOT NULL DEFAULT 0 COMMENT '明细已支付，分',
    refund_amount           BIGINT NOT NULL DEFAULT 0 COMMENT '累计已退款，分',
    refunded_quantity       INT NOT NULL DEFAULT 0 COMMENT '累计已退款数量',
    item_status             VARCHAR(32) NOT NULL DEFAULT 'normal',
    after_sale_status       VARCHAR(32) NOT NULL DEFAULT 'none',
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_order_item_no (order_item_no),
    KEY idx_order_no (order_no),
    KEY idx_sku_no (sku_no),
    KEY idx_user_created (user_no, created_at),
    KEY idx_after_sale_status (after_sale_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单明细表';

-- 8.4 订单操作日志表
DROP TABLE IF EXISTS order_operates;
CREATE TABLE order_operates (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    operate_no              VARCHAR(32) NOT NULL,
    order_no                VARCHAR(32) NOT NULL,
    action                  VARCHAR(64) NOT NULL COMMENT 'create/pay/cancel/update_address/review/manual_fix',
    from_status             VARCHAR(32) NULL,
    to_status               VARCHAR(32) NULL,
    operator_type           VARCHAR(32) NOT NULL COMMENT 'user/admin/system',
    operator_no             VARCHAR(32) NULL,
    request_id              VARCHAR(64) NULL,
    trace_id                VARCHAR(64) NULL,
    client_ip               VARCHAR(45) NULL,
    operate_desc            VARCHAR(255) NULL,
    ext_json                JSON NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_operate_no (operate_no),
    KEY idx_order_created (order_no, created_at),
    KEY idx_request_id (request_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单操作日志表';

-- 8.5 订单优惠分摊明细表
DROP TABLE IF EXISTS order_discount_details;
CREATE TABLE order_discount_details (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    detail_no               VARCHAR(32) NOT NULL,
    order_no                VARCHAR(32) NOT NULL,
    order_item_no           VARCHAR(32) NULL COMMENT '整单优惠可为空，单品优惠需填写',
    discount_type           VARCHAR(32) NOT NULL COMMENT 'coupon/promotion/member/points/freight_discount',
    discount_ref_no         VARCHAR(32) NULL COMMENT '关联优惠实体编号',
    discount_name           VARCHAR(128) NULL,
    discount_amount         BIGINT NOT NULL COMMENT '优惠金额，分',
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_detail_no (detail_no),
    KEY idx_order_no (order_no),
    KEY idx_item_no (order_item_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单优惠分摊明细表';

-- ----------------------------------------------------------------------------
-- 9. 履约/发货域设计
-- ----------------------------------------------------------------------------

-- 9.1 发货单主表
DROP TABLE IF EXISTS shipments;
CREATE TABLE shipments (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    shipment_no             VARCHAR(32) NOT NULL,
    order_no                VARCHAR(32) NOT NULL,
    merchant_no             VARCHAR(32) NOT NULL,
    shop_no                 VARCHAR(32) NOT NULL,
    warehouse_no            VARCHAR(32) NOT NULL,
    shipment_status         VARCHAR(32) NOT NULL DEFAULT 'created' COMMENT 'created/shipped/in_transit/signed/returned/cancelled',
    logistics_company_code  VARCHAR(32) NULL,
    logistics_company_name  VARCHAR(64) NULL,
    tracking_no             VARCHAR(64) NULL,
    consignee_snapshot      JSON NOT NULL COMMENT '收件信息快照',
    shipped_at              DATETIME(3) NULL,
    delivered_at            DATETIME(3) NULL,
    signed_at               DATETIME(3) NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_shipment_no (shipment_no),
    KEY idx_order_no (order_no),
    KEY idx_tracking_no (tracking_no),
    KEY idx_status_created (shipment_status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='发货单主表';

-- 9.2 发货单明细表
DROP TABLE IF EXISTS shipment_items;
CREATE TABLE shipment_items (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    shipment_item_no        VARCHAR(32) NOT NULL,
    shipment_no             VARCHAR(32) NOT NULL,
    order_item_no           VARCHAR(32) NOT NULL,
    sku_no                  VARCHAR(32) NOT NULL,
    quantity                INT NOT NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_shipment_item_no (shipment_item_no),
    KEY idx_shipment_no (shipment_no),
    KEY idx_order_item_no (order_item_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='发货单明细表';

-- ----------------------------------------------------------------------------
-- 10. 库存域设计
-- ----------------------------------------------------------------------------

-- 10.2 仓库表
DROP TABLE IF EXISTS warehouses;
CREATE TABLE warehouses (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    warehouse_no            VARCHAR(32) NOT NULL,
    warehouse_name          VARCHAR(128) NOT NULL,
    merchant_no             VARCHAR(32) NOT NULL,
    status                  TINYINT NOT NULL DEFAULT 1,
    address_json            JSON NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_warehouse_no (warehouse_no),
    KEY idx_merchant_status (merchant_no, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='仓库表';

-- 10.3 库存主表
DROP TABLE IF EXISTS inventories;
CREATE TABLE inventories (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    inventory_no            VARCHAR(32) NOT NULL,
    sku_no                  VARCHAR(32) NOT NULL,
    warehouse_no            VARCHAR(32) NOT NULL,
    total_stock             INT NOT NULL DEFAULT 0 COMMENT '总库存',
    available_stock         INT NOT NULL DEFAULT 0 COMMENT '可用库存',
    reserved_stock          INT NOT NULL DEFAULT 0 COMMENT '已预占库存',
    locked_stock            INT NOT NULL DEFAULT 0 COMMENT '锁定库存',
    freeze_stock            INT NOT NULL DEFAULT 0 COMMENT '冻结库存',
    damaged_stock           INT NOT NULL DEFAULT 0 COMMENT '残次库存',
    low_stock_threshold     INT NOT NULL DEFAULT 0 COMMENT '低库存预警阈值',
    status                  TINYINT NOT NULL DEFAULT 1,
    version                 INT NOT NULL DEFAULT 1,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_inventory_no (inventory_no),
    UNIQUE KEY uk_sku_warehouse (sku_no, warehouse_no),
    KEY idx_available_stock (available_stock),
    KEY idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存主表';

-- 10.4 库存预占主表
DROP TABLE IF EXISTS inventory_reservations;
CREATE TABLE inventory_reservations (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    reservation_no          VARCHAR(32) NOT NULL,
    order_no                VARCHAR(32) NOT NULL,
    reservation_status      VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT 'active/confirmed/released/expired',
    expire_at               DATETIME(3) NOT NULL,
    confirmed_at            DATETIME(3) NULL,
    released_at             DATETIME(3) NULL,
    request_id              VARCHAR(64) NULL,
    trace_id                VARCHAR(64) NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_reservation_no (reservation_no),
    KEY idx_order_no (order_no),
    KEY idx_status_expire (reservation_status, expire_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存预占主表';

-- 10.5 库存预占明细表
DROP TABLE IF EXISTS inventory_reservation_items;
CREATE TABLE inventory_reservation_items (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    reservation_item_no     VARCHAR(32) NOT NULL,
    reservation_no          VARCHAR(32) NOT NULL,
    sku_no                  VARCHAR(32) NOT NULL,
    warehouse_no            VARCHAR(32) NOT NULL,
    quantity                INT NOT NULL,
    confirmed_quantity      INT NOT NULL DEFAULT 0,
    released_quantity       INT NOT NULL DEFAULT 0,
    item_status             VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_reservation_item_no (reservation_item_no),
    UNIQUE KEY uk_reservation_sku_wh (reservation_no, sku_no, warehouse_no),
    KEY idx_reservation_no (reservation_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存预占明细表';

-- 10.6 库存流水表
DROP TABLE IF EXISTS inventory_logs;
CREATE TABLE inventory_logs (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    log_no                  VARCHAR(32) NOT NULL,
    sku_no                  VARCHAR(32) NOT NULL,
    warehouse_no            VARCHAR(32) NOT NULL,
    biz_type                VARCHAR(32) NOT NULL COMMENT 'order/payment/refund/adjust/return',
    biz_no                  VARCHAR(32) NOT NULL COMMENT '业务单号',
    change_type             VARCHAR(32) NOT NULL COMMENT 'reserve/confirm/release/deduct/return/adjust',
    change_quantity         INT NOT NULL COMMENT '变动数量，正负值均可',
    before_available        INT NOT NULL,
    after_available         INT NOT NULL,
    before_reserved         INT NOT NULL,
    after_reserved         INT NOT NULL,
    before_locked           INT NOT NULL DEFAULT 0,
    after_locked            INT NOT NULL DEFAULT 0,
    operator_type           VARCHAR(32) NOT NULL DEFAULT 'system',
    operator_no             VARCHAR(32) NULL,
    request_id              VARCHAR(64) NULL,
    trace_id                VARCHAR(64) NULL,
    remark                  VARCHAR(255) NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_log_no (log_no),
    KEY idx_sku_created (sku_no, created_at),
    KEY idx_biz_no (biz_type, biz_no),
    KEY idx_request_id (request_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存流水表';

-- ----------------------------------------------------------------------------
-- 11. 支付域设计
-- ----------------------------------------------------------------------------

-- 11.2 支付单表
DROP TABLE IF EXISTS payments;
CREATE TABLE payments (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    payment_no              VARCHAR(32) NOT NULL,
    biz_type                VARCHAR(32) NOT NULL DEFAULT 'order',
    biz_no                  VARCHAR(32) NOT NULL COMMENT '如 order_no',
    order_no                VARCHAR(32) NOT NULL,
    merchant_no             VARCHAR(32) NOT NULL,
    shop_no                 VARCHAR(32) NOT NULL,
    user_no                 VARCHAR(32) NOT NULL,
    currency_code           CHAR(3) NOT NULL DEFAULT 'CNY',
    payable_amount          BIGINT NOT NULL COMMENT '应支付金额，分',
    paid_amount             BIGINT NOT NULL DEFAULT 0 COMMENT '已支付金额，分',
    channel_code            VARCHAR(32) NOT NULL COMMENT 'wechat/alipay/balance/cod',
    payment_status          VARCHAR(32) NOT NULL DEFAULT 'created' COMMENT 'created/paying/success/failed/cancelled/closed',
    channel_status          VARCHAR(32) NULL COMMENT '渠道侧状态',
    channel_trade_no        VARCHAR(64) NULL COMMENT '渠道交易号',
    pay_url                 VARCHAR(512) NULL,
    qr_code                 VARCHAR(512) NULL,
    expired_at              DATETIME(3) NULL,
    paid_at                 DATETIME(3) NULL,
    closed_at               DATETIME(3) NULL,
    request_id              VARCHAR(64) NULL COMMENT '支付请求幂等键',
    trace_id                VARCHAR(64) NULL,
    error_code              VARCHAR(64) NULL,
    error_msg               VARCHAR(255) NULL,
    ext_json                JSON NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_payment_no (payment_no),
    UNIQUE KEY uk_request_id (request_id),
    KEY idx_order_no (order_no),
    KEY idx_user_created (user_no, created_at),
    KEY idx_channel_trade (channel_code, channel_trade_no),
    KEY idx_status_expired (payment_status, expired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付单表';

-- 11.3 支付渠道交易流水表
DROP TABLE IF EXISTS payment_transactions;
CREATE TABLE payment_transactions (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    transaction_no          VARCHAR(32) NOT NULL,
    payment_no              VARCHAR(32) NOT NULL,
    channel_code            VARCHAR(32) NOT NULL,
    trade_type              VARCHAR(32) NOT NULL COMMENT 'unified_order/query/close/refund',
    channel_trade_no        VARCHAR(64) NULL,
    transaction_status      VARCHAR(32) NOT NULL DEFAULT 'created',
    request_payload         JSON NULL,
    response_payload        JSON NULL,
    error_code              VARCHAR(64) NULL,
    error_msg               VARCHAR(255) NULL,
    success_at              DATETIME(3) NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_transaction_no (transaction_no),
    KEY idx_payment_no (payment_no),
    KEY idx_channel_trade_no (channel_code, channel_trade_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付渠道交易流水表';

-- 11.4 支付回调日志表
DROP TABLE IF EXISTS payment_callbacks;
CREATE TABLE payment_callbacks (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    callback_no             VARCHAR(32) NOT NULL,
    payment_no              VARCHAR(32) NULL,
    channel_code            VARCHAR(32) NOT NULL,
    callback_id             VARCHAR(128) NOT NULL COMMENT '渠道通知唯一标识',
    callback_type           VARCHAR(32) NOT NULL COMMENT 'pay/refund/close',
    callback_body           JSON NULL,
    callback_status         VARCHAR(32) NOT NULL DEFAULT 'received',
    processed               TINYINT NOT NULL DEFAULT 0,
    processed_at            DATETIME(3) NULL,
    request_id              VARCHAR(64) NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_callback_no (callback_no),
    UNIQUE KEY uk_channel_callback (channel_code, callback_id),
    KEY idx_payment_no (payment_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付回调日志表';

-- 11.5 支付对账明细表
DROP TABLE IF EXISTS payment_reconciliations;
CREATE TABLE payment_reconciliations (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    reconciliation_no       VARCHAR(32) NOT NULL,
    reconcile_date          DATE NOT NULL,
    channel_code            VARCHAR(32) NOT NULL,
    payment_no              VARCHAR(32) NULL,
    order_no                VARCHAR(32) NULL,
    channel_trade_no        VARCHAR(64) NULL,
    channel_amount          BIGINT NOT NULL DEFAULT 0,
    system_amount           BIGINT NOT NULL DEFAULT 0,
    reconcile_status        VARCHAR(32) NOT NULL DEFAULT 'init' COMMENT 'match/mismatch/missing_channel/missing_system',
    diff_reason             VARCHAR(255) NULL,
    channel_payload         JSON NULL,
    system_payload          JSON NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_reconciliation_no (reconciliation_no),
    KEY idx_date_channel (reconcile_date, channel_code),
    KEY idx_payment_no (payment_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付对账明细表';

-- ----------------------------------------------------------------------------
-- 12. 售后与退款域设计
-- ----------------------------------------------------------------------------

-- 12.2 售后主表
DROP TABLE IF EXISTS after_sale_orders;
CREATE TABLE after_sale_orders (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    after_sale_no           VARCHAR(32) NOT NULL,
    order_no                VARCHAR(32) NOT NULL,
    user_no                 VARCHAR(32) NOT NULL,
    merchant_no             VARCHAR(32) NOT NULL,
    shop_no                 VARCHAR(32) NOT NULL,
    after_sale_type         VARCHAR(32) NOT NULL COMMENT 'only_refund/return_refund/exchange/reissue',
    reason_code             VARCHAR(32) NULL,
    reason_desc             VARCHAR(255) NULL,
    apply_amount            BIGINT NOT NULL DEFAULT 0 COMMENT '申请金额，分',
    approved_amount         BIGINT NOT NULL DEFAULT 0 COMMENT '审批金额，分',
    compensation_amount     BIGINT NOT NULL DEFAULT 0 COMMENT '补偿金额，分',
    status                  VARCHAR(32) NOT NULL DEFAULT 'applied' COMMENT 'applied/reviewing/approved/rejected/waiting_return/returned/refund_processing/completed/closed',
    reject_reason           VARCHAR(255) NULL,
    buyer_remark            VARCHAR(255) NULL,
    seller_remark           VARCHAR(255) NULL,
    evidence_json           JSON NULL COMMENT '图片视频证据',
    applied_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    approved_at             DATETIME(3) NULL,
    completed_at            DATETIME(3) NULL,
    closed_at               DATETIME(3) NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_after_sale_no (after_sale_no),
    KEY idx_order_no (order_no),
    KEY idx_user_created (user_no, created_at),
    KEY idx_status_created (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='售后主表';

-- 12.3 售后明细表
DROP TABLE IF EXISTS after_sale_items;
CREATE TABLE after_sale_items (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    after_sale_item_no      VARCHAR(32) NOT NULL,
    after_sale_no           VARCHAR(32) NOT NULL,
    order_item_no           VARCHAR(32) NOT NULL,
    sku_no                  VARCHAR(32) NOT NULL,
    apply_quantity          INT NOT NULL DEFAULT 0,
    approved_quantity       INT NOT NULL DEFAULT 0,
    refund_amount           BIGINT NOT NULL DEFAULT 0 COMMENT '该明细预期退款金额，分',
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_after_sale_item_no (after_sale_item_no),
    KEY idx_after_sale_no (after_sale_no),
    KEY idx_order_item_no (order_item_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='售后明细表';

-- 12.4 退货物流表
DROP TABLE IF EXISTS return_shipments;
CREATE TABLE return_shipments (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    return_shipment_no      VARCHAR(32) NOT NULL,
    after_sale_no           VARCHAR(32) NOT NULL,
    logistics_company_code  VARCHAR(32) NULL,
    logistics_company_name  VARCHAR(64) NULL,
    tracking_no             VARCHAR(64) NULL,
    sender_name             VARCHAR(64) NULL,
    sender_phone            VARCHAR(20) NULL,
    sender_address_json     JSON NULL,
    sent_at                 DATETIME(3) NULL,
    received_at             DATETIME(3) NULL,
    status                  VARCHAR(32) NOT NULL DEFAULT 'created',
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_return_shipment_no (return_shipment_no),
    KEY idx_after_sale_no (after_sale_no),
    KEY idx_tracking_no (tracking_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='退货物流表';

-- 12.5 退款单表
DROP TABLE IF EXISTS refunds;
CREATE TABLE refunds (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    refund_no               VARCHAR(32) NOT NULL,
    after_sale_no           VARCHAR(32) NOT NULL,
    order_no                VARCHAR(32) NOT NULL,
    payment_no              VARCHAR(32) NOT NULL,
    user_no                 VARCHAR(32) NOT NULL,
    merchant_no             VARCHAR(32) NOT NULL,
    shop_no                 VARCHAR(32) NOT NULL,
    currency_code           CHAR(3) NOT NULL DEFAULT 'CNY',
    refund_amount           BIGINT NOT NULL COMMENT '申请退款金额，分',
    actual_refund_amount    BIGINT NOT NULL DEFAULT 0 COMMENT '实际退款金额，分',
    refund_freight_amount   BIGINT NOT NULL DEFAULT 0 COMMENT '退回运费，分',
    refund_status           VARCHAR(32) NOT NULL DEFAULT 'created' COMMENT 'created/processing/success/failed/closed',
    reason_code             VARCHAR(32) NULL,
    reason_desc             VARCHAR(255) NULL,
    retry_count             INT NOT NULL DEFAULT 0,
    max_retry               INT NOT NULL DEFAULT 3,
    next_retry_at           DATETIME(3) NULL,
    success_at              DATETIME(3) NULL,
    closed_at               DATETIME(3) NULL,
    error_code              VARCHAR(64) NULL,
    error_msg               VARCHAR(255) NULL,
    request_id              VARCHAR(64) NULL,
    trace_id                VARCHAR(64) NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_refund_no (refund_no),
    UNIQUE KEY uk_request_id (request_id),
    KEY idx_payment_no (payment_no),
    KEY idx_status_retry (refund_status, next_retry_at),
    KEY idx_order_no (order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='退款单表';

-- 12.6 退款流水表
DROP TABLE IF EXISTS refund_transactions;
CREATE TABLE refund_transactions (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    refund_txn_no           VARCHAR(32) NOT NULL,
    refund_no               VARCHAR(32) NOT NULL,
    payment_no              VARCHAR(32) NOT NULL,
    channel_code            VARCHAR(32) NOT NULL,
    channel_refund_no       VARCHAR(64) NULL,
    transaction_status      VARCHAR(32) NOT NULL DEFAULT 'created',
    request_payload         JSON NULL,
    response_payload        JSON NULL,
    error_code              VARCHAR(64) NULL,
    error_msg               VARCHAR(255) NULL,
    success_at              DATETIME(3) NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_refund_txn_no (refund_txn_no),
    KEY idx_refund_no (refund_no),
    KEY idx_channel_refund_no (channel_code, channel_refund_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='退款流水表';

-- ----------------------------------------------------------------------------
-- 13. 营销域设计
-- ----------------------------------------------------------------------------

-- 13.1 优惠券模板表
DROP TABLE IF EXISTS coupon_templates;
CREATE TABLE coupon_templates (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    coupon_tpl_no           VARCHAR(32) NOT NULL,
    merchant_no             VARCHAR(32) NOT NULL,
    shop_no                 VARCHAR(32) NOT NULL,
    coupon_name             VARCHAR(128) NOT NULL,
    coupon_type             VARCHAR(32) NOT NULL DEFAULT 'full_reduction' COMMENT 'cash/full_reduction/discount',
    threshold_amount        BIGINT NOT NULL DEFAULT 0 COMMENT '门槛金额，分',
    discount_amount         BIGINT NOT NULL DEFAULT 0 COMMENT '优惠金额，分',
    discount_rate           DECIMAL(6,4) NOT NULL DEFAULT 0.0000 COMMENT '折扣券用',
    max_discount_amount     BIGINT NOT NULL DEFAULT 0 COMMENT '封顶金额，分',
    total_count             INT NOT NULL DEFAULT 0,
    remain_count            INT NOT NULL DEFAULT 0,
    per_user_limit          INT NOT NULL DEFAULT 1,
    applicable_scope        VARCHAR(32) NOT NULL DEFAULT 'all' COMMENT 'all/category/spu/sku',
    applicable_scope_json   JSON NULL,
    receive_start_at        DATETIME(3) NULL,
    receive_end_at          DATETIME(3) NULL,
    valid_start_at          DATETIME(3) NULL,
    valid_end_at            DATETIME(3) NULL,
    valid_days              INT NOT NULL DEFAULT 0 COMMENT '领券后有效天数，与固定有效期二选一',
    status                  VARCHAR(32) NOT NULL DEFAULT 'draft' COMMENT 'draft/active/paused/expired',
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_coupon_tpl_no (coupon_tpl_no),
    KEY idx_shop_status (shop_no, status),
    KEY idx_receive_time (receive_start_at, receive_end_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='优惠券模板表';

-- 13.2 用户优惠券表
DROP TABLE IF EXISTS user_coupons;
CREATE TABLE user_coupons (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    coupon_user_no          VARCHAR(32) NOT NULL,
    coupon_tpl_no           VARCHAR(32) NOT NULL,
    user_no                 VARCHAR(32) NOT NULL,
    coupon_status           VARCHAR(32) NOT NULL DEFAULT 'unused' COMMENT 'unused/locked/used/expired/cancelled',
    threshold_amount        BIGINT NOT NULL DEFAULT 0,
    discount_amount         BIGINT NOT NULL DEFAULT 0,
    received_at             DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    valid_start_at          DATETIME(3) NULL,
    valid_end_at            DATETIME(3) NULL,
    used_order_no           VARCHAR(32) NULL,
    used_at                 DATETIME(3) NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_coupon_user_no (coupon_user_no),
    KEY idx_user_status (user_no, coupon_status),
    KEY idx_valid_end (valid_end_at),
    KEY idx_used_order_no (used_order_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户优惠券表';

-- 13.3 秒杀活动表
DROP TABLE IF EXISTS seckill_activities;
CREATE TABLE seckill_activities (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    activity_no             VARCHAR(32) NOT NULL,
    merchant_no             VARCHAR(32) NOT NULL,
    shop_no                 VARCHAR(32) NOT NULL,
    activity_name           VARCHAR(128) NOT NULL,
    activity_desc           VARCHAR(255) NULL,
    start_at                DATETIME(3) NOT NULL,
    end_at                  DATETIME(3) NOT NULL,
    status                  VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'pending/active/paused/ended',
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_activity_no (activity_no),
    KEY idx_shop_status_time (shop_no, status, start_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀活动表';

-- 13.4 秒杀 SKU 表
DROP TABLE IF EXISTS seckill_skus;
CREATE TABLE seckill_skus (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    seckill_sku_no          VARCHAR(32) NOT NULL,
    activity_no             VARCHAR(32) NOT NULL,
    sku_no                  VARCHAR(32) NOT NULL,
    seckill_price           BIGINT NOT NULL COMMENT '秒杀价，分',
    origin_price            BIGINT NOT NULL COMMENT '原价，分',
    total_stock             INT NOT NULL DEFAULT 0,
    remain_stock            INT NOT NULL DEFAULT 0,
    locked_stock            INT NOT NULL DEFAULT 0,
    per_user_limit          INT NOT NULL DEFAULT 1,
    status                  TINYINT NOT NULL DEFAULT 1,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_seckill_sku_no (seckill_sku_no),
    UNIQUE KEY uk_activity_sku (activity_no, sku_no),
    KEY idx_remain_stock (remain_stock)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='秒杀SKU表';

-- ----------------------------------------------------------------------------
-- 14. 评价域设计
-- ----------------------------------------------------------------------------

-- 14.1 商品评价表
DROP TABLE IF EXISTS product_reviews;
CREATE TABLE product_reviews (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    review_no               VARCHAR(32) NOT NULL,
    order_no                VARCHAR(32) NOT NULL,
    order_item_no           VARCHAR(32) NOT NULL,
    user_no                 VARCHAR(32) NOT NULL,
    spu_no                  VARCHAR(32) NOT NULL,
    sku_no                  VARCHAR(32) NOT NULL,
    score                   TINYINT NOT NULL COMMENT '1-5分',
    review_content          VARCHAR(2000) NULL,
    is_anonymous            TINYINT NOT NULL DEFAULT 0,
    has_images              TINYINT NOT NULL DEFAULT 0,
    review_status           TINYINT NOT NULL DEFAULT 1 COMMENT '1正常 2隐藏',
    replied                 TINYINT NOT NULL DEFAULT 0,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_review_no (review_no),
    UNIQUE KEY uk_order_item_no (order_item_no),
    KEY idx_spu_score_time (spu_no, score, created_at),
    KEY idx_user_time (user_no, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品评价表';

-- 14.2 评价媒体表
DROP TABLE IF EXISTS product_review_media;
CREATE TABLE product_review_media (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    media_no                VARCHAR(32) NOT NULL,
    review_no               VARCHAR(32) NOT NULL,
    media_type              VARCHAR(16) NOT NULL COMMENT 'image/video',
    media_url               VARCHAR(512) NOT NULL,
    sort_no                 INT NOT NULL DEFAULT 0,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_media_no (media_no),
    KEY idx_review_sort (review_no, sort_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='评价媒体表';

-- ----------------------------------------------------------------------------
-- 15. 通知域设计
-- ----------------------------------------------------------------------------

-- 15.1 通知记录表
DROP TABLE IF EXISTS notifications;
CREATE TABLE notifications (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    notification_no         VARCHAR(32) NOT NULL,
    user_no                 VARCHAR(32) NOT NULL,
    notification_type       VARCHAR(32) NOT NULL COMMENT 'order_paid/order_shipped/refund_success等',
    title                   VARCHAR(128) NOT NULL,
    content                 VARCHAR(1000) NULL,
    channel                 VARCHAR(32) NOT NULL COMMENT 'app/sms/email/wechat',
    data_json               JSON NULL,
    push_status             VARCHAR(32) NOT NULL DEFAULT 'init' COMMENT 'init/sent/failed/read',
    is_read                 TINYINT NOT NULL DEFAULT 0,
    read_at                 DATETIME(3) NULL,
    sent_at                 DATETIME(3) NULL,
    error_msg               VARCHAR(255) NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_notification_no (notification_no),
    KEY idx_user_read_time (user_no, is_read, created_at),
    KEY idx_type_status (notification_type, push_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知记录表';

-- 15.2 通知配置表
DROP TABLE IF EXISTS notification_configs;
CREATE TABLE notification_configs (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    config_no               VARCHAR(32) NOT NULL,
    user_no                 VARCHAR(32) NOT NULL,
    notification_type       VARCHAR(32) NOT NULL,
    channel                 VARCHAR(32) NOT NULL,
    enabled                 TINYINT NOT NULL DEFAULT 1,
    threshold_amount        BIGINT NOT NULL DEFAULT 0 COMMENT '如降价通知阈值，单位分',
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_config_no (config_no),
    UNIQUE KEY uk_user_type_channel (user_no, notification_type, channel)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知配置表';

-- ----------------------------------------------------------------------------
-- 16. 推荐与行为分析域设计
-- ----------------------------------------------------------------------------

-- 16.1 用户行为日志表
DROP TABLE IF EXISTS user_behavior_logs;
CREATE TABLE user_behavior_logs (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    log_no                  VARCHAR(32) NOT NULL,
    user_no                 VARCHAR(32) NULL,
    anonymous_id            VARCHAR(64) NULL,
    spu_no                  VARCHAR(32) NULL,
    sku_no                  VARCHAR(32) NULL,
    category_no             VARCHAR(32) NULL,
    behavior_type           VARCHAR(32) NOT NULL COMMENT 'view/search/cart/favor/buy/comment',
    source_page             VARCHAR(32) NULL,
    search_keyword          VARCHAR(128) NULL,
    stay_duration_seconds   INT NOT NULL DEFAULT 0,
    product_name_snapshot   VARCHAR(255) NULL,
    product_price_snapshot  BIGINT NOT NULL DEFAULT 0,
    time_bucket             INT NOT NULL DEFAULT 0 COMMENT '5分钟桶',
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_log_no (log_no),
    KEY idx_user_behavior_time (user_no, behavior_type, created_at),
    KEY idx_sku_time (sku_no, created_at),
    KEY idx_category_time (category_no, created_at),
    KEY idx_anonymous_time (anonymous_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户行为日志表';

-- ----------------------------------------------------------------------------
-- 17. 任务与事件域设计
-- ----------------------------------------------------------------------------

-- 17.1 异步任务表
DROP TABLE IF EXISTS tasks;
CREATE TABLE tasks (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    task_no                 VARCHAR(32) NOT NULL,
    biz_type                VARCHAR(32) NOT NULL COMMENT 'order/payment/inventory/refund',
    biz_no                  VARCHAR(32) NOT NULL,
    task_type               VARCHAR(64) NOT NULL COMMENT 'order_close/payment_query/inventory_confirm/refund_retry',
    priority                INT NOT NULL DEFAULT 5,
    task_status             VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'pending/running/success/failed/retry/cancelled',
    payload                 JSON NULL,
    retry_count             INT NOT NULL DEFAULT 0,
    max_retry               INT NOT NULL DEFAULT 3,
    next_retry_at           DATETIME(3) NULL,
    executed_at             DATETIME(3) NULL,
    finished_at             DATETIME(3) NULL,
    request_id              VARCHAR(64) NULL,
    trace_id                VARCHAR(64) NULL,
    error_msg               VARCHAR(1000) NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_task_no (task_no),
    KEY idx_status_retry (task_status, next_retry_at),
    KEY idx_biz (biz_type, biz_no),
    KEY idx_type_priority (task_type, priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='异步任务表';

-- 17.2 Outbox 事件表
DROP TABLE IF EXISTS outbox_events;
CREATE TABLE outbox_events (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    event_no                VARCHAR(32) NOT NULL,
    aggregate_type          VARCHAR(32) NOT NULL COMMENT 'order/payment/refund/inventory',
    aggregate_no            VARCHAR(32) NOT NULL,
    event_type              VARCHAR(64) NOT NULL COMMENT 'order_created/payment_succeeded/refund_succeeded',
    event_status            VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT 'pending/published/failed',
    payload                 JSON NOT NULL,
    published_at            DATETIME(3) NULL,
    retry_count             INT NOT NULL DEFAULT 0,
    next_retry_at           DATETIME(3) NULL,
    trace_id                VARCHAR(64) NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_event_no (event_no),
    KEY idx_status_retry (event_status, next_retry_at),
    KEY idx_aggregate (aggregate_type, aggregate_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Outbox事件表';

-- ----------------------------------------------------------------------------
-- 18. 审计与系统配置域设计
-- ----------------------------------------------------------------------------

-- 18.1 管理操作日志表
DROP TABLE IF EXISTS admin_operation_logs;
CREATE TABLE admin_operation_logs (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    log_no                  VARCHAR(32) NOT NULL,
    operator_no             VARCHAR(32) NOT NULL,
    operator_name           VARCHAR(64) NULL,
    operator_type           VARCHAR(32) NOT NULL COMMENT 'admin/merchant/system',
    action                  VARCHAR(64) NOT NULL,
    resource_type           VARCHAR(32) NOT NULL,
    resource_no             VARCHAR(32) NULL,
    before_value            JSON NULL,
    after_value             JSON NULL,
    request_id              VARCHAR(64) NULL,
    trace_id                VARCHAR(64) NULL,
    client_ip               VARCHAR(45) NULL,
    user_agent              VARCHAR(512) NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_log_no (log_no),
    KEY idx_operator_time (operator_no, created_at),
    KEY idx_resource (resource_type, resource_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理操作日志表';

-- 18.2 系统配置表
DROP TABLE IF EXISTS system_configs;
CREATE TABLE system_configs (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    config_no               VARCHAR(32) NOT NULL,
    config_key              VARCHAR(128) NOT NULL,
    config_value            TEXT NULL,
    config_type             VARCHAR(32) NOT NULL DEFAULT 'string',
    config_group            VARCHAR(32) NOT NULL DEFAULT 'default',
    is_encrypted            TINYINT NOT NULL DEFAULT 0,
    status                  TINYINT NOT NULL DEFAULT 1,
    config_desc             VARCHAR(255) NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_config_no (config_no),
    UNIQUE KEY uk_config_key (config_key),
    KEY idx_group_status (config_group, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置表';

-- ----------------------------------------------------------------------------
-- 19. 字典与原因表
-- ----------------------------------------------------------------------------

DROP TABLE IF EXISTS biz_dicts;
CREATE TABLE biz_dicts (
    id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    dict_no                 VARCHAR(32) NOT NULL,
    dict_type               VARCHAR(64) NOT NULL,
    dict_code               VARCHAR(64) NOT NULL,
    dict_name               VARCHAR(128) NOT NULL,
    parent_code             VARCHAR(64) NULL,
    sort_no                 INT NOT NULL DEFAULT 0,
    status                  TINYINT NOT NULL DEFAULT 1,
    ext_json                JSON NULL,
    created_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at              DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_dict_no (dict_no),
    UNIQUE KEY uk_type_code (dict_type, dict_code),
    KEY idx_type_sort (dict_type, sort_no)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='业务字典表';

-- ----------------------------------------------------------------------------
-- 初始化字典数据
-- ----------------------------------------------------------------------------

-- 订单取消原因
INSERT INTO biz_dicts (dict_no, dict_type, dict_code, dict_name, sort_no) VALUES
('DICT001', 'order_cancel_reason', 'USER_CANCEL', '用户取消', 1),
('DICT002', 'order_cancel_reason', 'OVERDUE', '支付超时', 2),
('DICT003', 'order_cancel_reason', 'OUT_OF_STOCK', '缺货', 3),
('DICT004', 'order_cancel_reason', 'ADMIN_CANCEL', '管理员取消', 4);

-- 售后原因
INSERT INTO biz_dicts (dict_no, dict_type, dict_code, dict_name, sort_no) VALUES
('DICT005', 'after_sale_reason', 'DEFECTIVE', '商品质量问题', 1),
('DICT006', 'after_sale_reason', 'WRONG_ITEM', '发错货', 2),
('DICT007', 'after_sale_reason', 'NOT_AS_DESC', '与描述不符', 3),
('DICT008', 'after_sale_reason', 'CHANGED_MIND', '改变主意', 4),
('DICT009', 'after_sale_reason', 'DELIVERY_DELAY', '配送延迟', 5);

-- 支付渠道
INSERT INTO biz_dicts (dict_no, dict_type, dict_code, dict_name, sort_no) VALUES
('DICT010', 'payment_channel', 'WECHAT', '微信支付', 1),
('DICT011', 'payment_channel', 'ALIPAY', '支付宝', 2),
('DICT012', 'payment_channel', 'BALANCE', '余额支付', 3),
('DICT013', 'payment_channel', 'COD', '货到付款', 4);

-- 通知类型
INSERT INTO biz_dicts (dict_no, dict_type, dict_code, dict_name, sort_no) VALUES
('DICT014', 'notification_type', 'ORDER_PAID', '订单支付成功', 1),
('DICT015', 'notification_type', 'ORDER_SHIPPED', '订单已发货', 2),
('DICT016', 'notification_type', 'ORDER_DELIVERED', '订单已送达', 3),
('DICT017', 'notification_type', 'REFUND_SUCCESS', '退款成功', 4),
('DICT018', 'notification_type', 'PRICE_DROP', '商品降价', 5),
('DICT019', 'notification_type', 'COUPON_EXPIRE', '优惠券即将过期', 6);

-- ----------------------------------------------------------------------------
-- 初始化测试商户和店铺数据
-- ----------------------------------------------------------------------------

INSERT INTO merchants (merchant_no, merchant_name, merchant_type, contact_name, contact_phone, status) VALUES
('M001', '平台自营', 3, '平台管理员', '13800138000', 1);

INSERT INTO shops (shop_no, merchant_no, shop_name, shop_status, service_phone) VALUES
('S001', 'M001', '官方旗舰店', 1, '400-800-8888');

-- ----------------------------------------------------------------------------
-- 初始化测试仓库数据
-- ----------------------------------------------------------------------------

INSERT INTO warehouses (warehouse_no, warehouse_name, merchant_no, status) VALUES
('W001', '中央仓库', 'M001', 1);

-- ----------------------------------------------------------------------------
-- 初始化测试管理员用户
-- ----------------------------------------------------------------------------

INSERT INTO users (user_no, username, display_name, password_hash, phone, email, status, register_source) VALUES
('U001', 'admin', '管理员', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhKu', '13800138000', 'admin@example.com', 1, 'admin');

INSERT INTO user_profiles (user_no, gender, country_code) VALUES
('U001', 0, 'CN');

INSERT INTO user_security (user_no, phone_verified, email_verified) VALUES
('U001', 1, 1);

SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================================
-- 建表完成
-- 共计 51 张表
-- ============================================================================