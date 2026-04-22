# 侧边栏优化设计方案

## 概述

优化订单管理系统的侧边栏，实现板块顺序可调节、折叠隐藏和 UI 效果提升。

## 需求

1. **拖拽排序** - 用户可通过拖拽自由调整菜单顺序
2. **双模式隐藏** - 图标折叠模式 + 悬浮抽屉模式
3. **UI 优化** - 动画、图标、布局细节
4. **本地持久化** - 顺序和折叠状态保存在 localStorage

## 设计方案

### 布局结构

```
┌──────┬──────────────────────────────────────────────┐
│ Sider │                Header                       │
│ 220px │  [浮动按钮*]              [通知] [用户菜单]  │
│       ├──────────────────────────────────────────────┤
│ Logo  │                                              │
│ ────  │              Content                         │
│ 看板  │                                              │
│ 订单  │                                              │
│ 支付  │                                              │
│ 退款  │                                              │
│ 库存  │                                              │
│       │                                              │
│ ────  │                                              │
│ [展开] │                                              │
└──────┴──────────────────────────────────────────────┘
      ↑ 悬浮按钮（侧边栏隐藏时显示）
```

### 交互模式

| 操作 | 响应 |
|------|------|
| 点击折叠按钮 | 侧边栏收缩为 64px 图标模式 |
| 点击悬浮按钮 | 侧边栏从左侧滑出（悬浮抽屉模式）|
| 拖拽菜单项 | 显示卡片阴影效果，其他项让出位置 |
| 释放菜单项 | 保存到 localStorage |

### 视觉规范

#### 侧边栏状态
- **展开态**: 宽度 220px，显示 Logo + 菜单图标 + 文字
- **折叠态**: 宽度 64px，仅显示图标，悬浮提示菜单名称

#### 拖拽反馈
- 被拖动项: 卡片阴影效果 + 0.8 透明度
- 目标位置: 蓝色虚线引导 (2px dashed #1677ff)
- 其他项: 平滑让出位置动画 (200ms ease-out)

#### 动画规范
- 展开/收起: `width 200ms ease-in-out`
- 悬浮效果: `background 150ms ease`
- 拖拽项: `box-shadow 150ms ease, opacity 150ms ease`

#### 颜色规范
- Logo 渐变: `linear-gradient(135deg, #1677ff 0%, #0958d9 100%)`
- 激活态: `#1677ff` 背景 + `#fff` 文字
- 悬停态: `#f0f5ff` 背景
- 分割线: `#e5e7eb`

### 组件设计

#### SidebarContext
- 管理侧边栏状态 (展开/折叠/悬浮)
- 提供 `toggleCollapse`, `toggleFloat` 方法
- 持久化到 localStorage

#### SortableMenu
- 基于 `@dnd-kit/sortable` 实现
- 支持拖拽排序和动画反馈
- 排序变化时触发 `onReorder` 回调

#### CollapsedTrigger
- 折叠态底部按钮，点击展开
- 显示展开箭头图标

#### FloatButton
- 绝对定位悬浮按钮
- 仅在侧边栏隐藏时显示
- 左侧 16px 位置，点击展开侧边栏

## 技术实现

### 依赖
- `@dnd-kit/sortable` - 拖拽排序
- `@dnd-kit/core` - DnD 核心

### 数据存储

```typescript
interface SidebarConfig {
  menuOrder: string[]      // 菜单 key 顺序
  collapsed: boolean       // 是否折叠
  floatMode: boolean       // 是否悬浮模式
}

// localStorage key: 'sidebar-config'
```

### 文件结构

```
src/
  layouts/BasicLayout/index.tsx     # 主布局
  components/SortableMenu/          # 可排序菜单组件
    index.tsx
    SortableItem.tsx
    DragHandle.tsx
  context/SidebarContext.tsx        # 侧边栏状态管理
  hooks/useSidebarConfig.ts         # localStorage 持久化
```

## 实现步骤

1. 创建 SidebarContext 管理状态
2. 创建 useSidebarConfig hook 处理持久化
3. 创建 SortableMenu 组件实现拖拽
4. 修改 BasicLayout 集成新组件
5. 添加 FloatButton 悬浮按钮
6. 添加动画和 UI 优化样式
7. 测试拖拽、折叠、悬浮功能

## 验收标准

- [ ] 菜单可通过拖拽调整顺序
- [ ] 拖拽时显示卡片阴影效果
- [ ] 点击折叠按钮，侧边栏收缩为 64px 图标模式
- [ ] 悬浮按钮在侧边栏隐藏时显示
- [ ] 点击悬浮按钮，侧边栏从左侧滑出
- [ ] 刷新页面后保持用户设置的顺序和折叠状态
- [ ] 动画流畅，无卡顿