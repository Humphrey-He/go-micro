# Order Admin Frontend

订单管理后台前端项目，基于 React 18 + Vite + TypeScript + Ant Design 5。

## 技术栈

| 类别 | 技术 |
|------|------|
| 框架 | React 18 + Vite |
| 语言 | TypeScript |
| UI 组件库 | Ant Design 5.x |
| 状态管理 | Zustand |
| 路由 | React Router 6 |
| HTTP 客户端 | Axios |
| 日期处理 | Day.js |

## 项目结构

```
src/
├── api/                    # API 层
│   ├── request.ts        # Axios 实例配置
│   └── order.ts          # 订单相关 API
├── components/           # 通用组件
│   ├── StatusTag/        # 状态标签
│   ├── BusinessTable/    # 通用表格
│   ├── DetailDrawer/     # 详情抽屉
│   └── ConfirmModal/     # 确认弹窗
├── features/             # 功能模块
│   └── orders/           # 订单模块
│       ├── components/  # 订单组件
│       ├── hooks/       # 订单 hooks
│       └── types/       # 订单类型
├── hooks/                # 通用 Hooks
│   ├── usePagination.ts # 分页 hook
│   └── useConfirm.ts    # 确认操作 hook
├── layouts/             # 布局组件
│   ├── BasicLayout/     # 基础侧边栏布局
│   └── BlankLayout/     # 空白布局
├── pages/              # 页面
│   ├── Login/         # 登录页
│   └── Orders/        # 订单页面
│       ├── OrderList.tsx   # 订单列表
│       └── OrderDetail.tsx # 订单详情
├── routes/            # 路由配置
├── stores/            # Zustand 状态库
├── types/             # 全局类型
└── utils/            # 工具函数
```

## 开发

```bash
# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev

# 构建生产版本
pnpm build
```

## 接口说明

项目通过 Vite 代理连接后端 API Gateway：

- `GET /api/v1/admin/orders` - 订单列表（分页、筛选）
- `GET /api/v1/order-views/:order_no` - 订单详情
- `POST /api/v1/orders/:id/cancel` - 取消订单

详见 [API 文档](./docs/order-admin/)。
