# 侧边栏优化实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现侧边栏的拖拽排序、折叠/悬浮模式切换和 UI 优化，顺序和状态持久化到 localStorage

**Architecture:** 使用 zustand + persist 管理侧边栏状态（展开/折叠/悬浮模式、菜单顺序），基于 @dnd-kit/sortable 实现拖拽排序，BasicLayout 作为主容器集成所有组件

**Tech Stack:** React 18, Ant Design 5, zustand (已有), @dnd-kit/sortable, @dnd-kit/core

---

## 文件结构

```
order-admin/src/
  stores/
    sidebarStore.ts          # 侧边栏状态管理 (新增)
  components/
    SortableMenu/           # 可排序菜单组件 (新增)
      index.tsx
    FloatButton.tsx         # 悬浮按钮 (新增)
  layouts/
    BasicLayout/
      index.tsx             # 修改: 集成新组件
```

---

## Task 1: 安装 dnd-kit 依赖

**Files:**
- Modify: `order-admin/package.json`

- [ ] **Step 1: 添加 @dnd-kit 依赖**

Run: `cd /e/awesomeProject/go-micro/order-admin && npm install @dnd-kit/sortable @dnd-kit/core`

---

## Task 2: 创建 sidebarStore

**Files:**
- Create: `order-admin/src/stores/sidebarStore.ts`

```typescript
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface MenuItem {
  key: string
  icon: React.ReactNode
  label: string
}

export interface SidebarState {
  collapsed: boolean
  floatMode: boolean
  menuOrder: string[]
  menuItems: MenuItem[]
  toggleCollapse: () => void
  toggleFloat: () => void
  setMenuOrder: (order: string[]) => void
  setMenuItems: (items: MenuItem[]) => void
}

const defaultMenuItems: MenuItem[] = [
  { key: '/dashboard', label: '运营看板', icon: 'DashboardOutlined' },
  { key: '/orders', label: '订单管理', icon: 'ShoppingOutlined' },
  { key: '/payments', label: '支付管理', icon: 'CreditCardOutlined' },
  { key: '/refunds', label: '退款管理', icon: 'RollbackOutlined' },
  { key: '/inventory', label: '库存管理', icon: 'AppstoreOutlined' },
]

export const useSidebarStore = create<SidebarState>()(
  persist(
    (set) => ({
      collapsed: false,
      floatMode: false,
      menuOrder: defaultMenuItems.map((i) => i.key),
      menuItems: defaultMenuItems,
      toggleCollapse: () => set((s) => ({ collapsed: !s.collapsed, floatMode: false })),
      toggleFloat: () => set((s) => ({ floatMode: !s.floatMode, collapsed: false })),
      setMenuOrder: (order) => set({ menuOrder: order }),
      setMenuItems: (items) => set({ menuItems: items }),
    }),
    { name: 'sidebar-config' }
  )
)
```

- [ ] **Step 1: 创建 sidebarStore**

Create file `order-admin/src/stores/sidebarStore.ts` with the code above.

- [ ] **Step 2: 提交**

```bash
git add order-admin/src/stores/sidebarStore.ts
git commit -m "feat(sidebar): add sidebar state management with zustand"
```

---

## Task 3: 创建 SortableMenu 组件

**Files:**
- Create: `order-admin/src/components/SortableMenu/index.tsx`

```tsx
import React, { useMemo } from 'react'
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
} from '@dnd-kit/core'
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Menu } from 'antd'
import {
  DashboardOutlined,
  ShoppingOutlined,
  CreditCardOutlined,
  RollbackOutlined,
  AppstoreOutlined,
  HolderOutlined,
} from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { useSidebarStore } from '@/stores/sidebarStore'

const iconMap: Record<string, React.ReactNode> = {
  DashboardOutlined: <DashboardOutlined />,
  ShoppingOutlined: <ShoppingOutlined />,
  CreditCardOutlined: <CreditCardOutlined />,
  RollbackOutlined: <RollbackOutlined />,
  AppstoreOutlined: <AppstoreOutlined />,
}

interface SortableItemProps {
  id: string
  icon: React.ReactNode
  label: string
  collapsed: boolean
}

const SortableItem: React.FC<SortableItemProps> = ({ id, icon, label, collapsed }) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.8 : 1,
    boxShadow: isDragging ? '0 4px 12px rgba(0,0,0,0.15)' : 'none',
    cursor: 'grab',
  }

  return (
    <div ref={setNodeRef} style={style} {...attributes}>
      <Menu.Item key={id} icon={collapsed ? icon : <><HolderOutlined style={{ fontSize: 12, opacity: 0.5 }} />{icon}</>}>
        {!collapsed && label}
      </Menu.Item>
    </div>
  )
}

interface SortableMenuProps {
  collapsed: boolean
}

export const SortableMenu: React.FC<SortableMenuProps> = ({ collapsed }) => {
  const navigate = useNavigate()
  const { menuItems, menuOrder, setMenuOrder } = useSidebarStore()
  const location = useLocation()

  const sensors = useSensors(
    useSensor(PointerSensor),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  )

  const orderedItems = useMemo(() => {
    return menuOrder
      .map((key) => menuItems.find((item) => item.key === key))
      .filter((item): item is NonNullable<typeof item> => item !== undefined)
  }, [menuItems, menuOrder])

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (over && active.id !== over.id) {
      const oldIndex = menuOrder.indexOf(active.id as string)
      const newIndex = menuOrder.indexOf(over.id as string)
      setMenuOrder(arrayMove(menuOrder, oldIndex, newIndex))
    }
  }

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key)
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragEnd={handleDragEnd}
    >
      <SortableContext items={menuOrder} strategy={verticalListSortingStrategy}>
        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          items={orderedItems.map((item) => ({
            key: item.key,
            icon: iconMap[item.icon] || <AppstoreOutlined />,
            label: item.label,
          }))}
          onClick={handleMenuClick}
          style={{
            borderRight: 0,
            marginTop: 8,
            padding: collapsed ? '0 4px' : '0 8px',
          }}
        />
      </SortableContext>
    </DndContext>
  )
}
```

- [ ] **Step 1: 创建 SortableMenu 组件**

Create file `order-admin/src/components/SortableMenu/index.tsx`

- [ ] **Step 2: 提交**

```bash
git add order-admin/src/components/SortableMenu/index.tsx
git commit -m "feat(sidebar): add SortableMenu component with dnd-kit"
```

---

## Task 4: 创建 FloatButton 组件

**Files:**
- Create: `order-admin/src/components/FloatButton.tsx`

```tsx
import React from 'react'
import { Button } from 'antd'
import { LeftOutlined, MenuOutlined } from '@ant-design/icons'
import { useSidebarStore } from '@/stores/sidebarStore'

export const FloatButton: React.FC = () => {
  const { floatMode, toggleFloat, collapsed } = useSidebarStore()

  if (!floatMode && !collapsed) return null

  return (
    <Button
      type="primary"
      shape="circle"
      size="large"
      icon={collapsed ? <MenuOutlined /> : <LeftOutlined />}
      onClick={toggleFloat}
      style={{
        position: 'fixed',
        left: collapsed ? 16 : 224,
        top: 80,
        zIndex: 99,
        boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
        transition: 'left 200ms ease-in-out',
      }}
    />
  )
}
```

- [ ] **Step 1: 创建 FloatButton 组件**

Create file `order-admin/src/components/FloatButton.tsx`

- [ ] **Step 2: 提交**

```bash
git add order-admin/src/components/FloatButton.tsx
git commit -m "feat(sidebar): add FloatButton component"
```

---

## Task 5: 修改 BasicLayout 集成新组件

**Files:**
- Modify: `order-admin/src/layouts/BasicLayout/index.tsx`

```tsx
import React from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu, Avatar, Dropdown, Space, Button, Tooltip } from 'antd'
import {
  DashboardOutlined,
  ShoppingOutlined,
  CreditCardOutlined,
  RollbackOutlined,
  AppstoreOutlined,
  UserOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '@/stores/authStore'
import { NotificationBell } from '@/features/notification/components/NotificationBell'
import { SortableMenu } from '@/components/SortableMenu'
import { FloatButton } from '@/components/FloatButton'
import { useSidebarStore } from '@/stores/sidebarStore'

const { Header, Sider, Content } = Layout

export const BasicLayout: React.FC = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { userInfo, logout } = useAuthStore()
  const { collapsed, floatMode, toggleCollapse } = useSidebarStore()

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key)
  }

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: userInfo?.username || 'Admin',
      disabled: true,
    },
    { type: 'divider' as const },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout,
    },
  ]

  const siderWidth = collapsed ? 64 : 220

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* 悬浮模式下的侧边栏 */}
      <Sider
        trigger={null}
        width={220}
        style={{
          borderRight: '1px solid #e5e7eb',
          background: '#fff',
          position: 'fixed',
          left: floatMode ? 0 : -220,
          top: 0,
          bottom: 0,
          zIndex: 100,
          transition: 'left 200ms ease-in-out',
          overflow: 'auto',
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: '1px solid #e5e7eb',
            background: 'linear-gradient(135deg, #1677ff 0%, #0958d9 100%)',
          }}
        >
          <span style={{ fontSize: 17, fontWeight: 700, color: '#fff', letterSpacing: 0.5 }}>
            订单管理系统
          </span>
        </div>
        <SortableMenu collapsed={false} />
      </Sider>

      {/* 折叠模式的侧边栏 */}
      <Sider
        trigger={null}
        width={siderWidth}
        collapsedWidth={64}
        style={{
          borderRight: '1px solid #e5e7eb',
          background: '#fff',
          position: 'fixed',
          left: 0,
          top: 0,
          bottom: 0,
          zIndex: 100,
          transition: 'width 200ms ease-in-out',
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: collapsed ? 'center' : 'center',
            borderBottom: '1px solid #e5e7eb',
            background: 'linear-gradient(135deg, #1677ff 0%, #0958d9 100%)',
          }}
        >
          {!collapsed && (
            <span style={{ fontSize: 17, fontWeight: 700, color: '#fff', letterSpacing: 0.5 }}>
              订单管理系统
            </span>
          )}
          {collapsed && (
            <span style={{ fontSize: 17, fontWeight: 700, color: '#fff' }}>订</span>
          )}
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', height: 'calc(100vh - 64px)' }}>
          <div style={{ flex: 1, overflow: 'auto' }}>
            <SortableMenu collapsed={collapsed} />
          </div>

          {/* 折叠/展开按钮 */}
          <div
            style={{
              padding: collapsed ? '12px 0' : '12px 8px',
              borderTop: '1px solid #e5e7eb',
              display: 'flex',
              justifyContent: 'center',
            }}
          >
            <Tooltip title={collapsed ? '展开菜单' : '折叠菜单'} placement="right">
              <Button
                type="text"
                icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                onClick={toggleCollapse}
                style={{ color: '#666' }}
              />
            </Tooltip>
          </div>
        </div>
      </Sider>

      {/* 悬浮按钮 */}
      <FloatButton />

      <Layout style={{ marginLeft: collapsed ? 64 : 220, transition: 'margin-left 200ms ease-in-out' }}>
        <Header
          style={{
            background: '#fff',
            padding: '0 24px',
            display: 'flex',
            justifyContent: 'flex-end',
            alignItems: 'center',
            borderBottom: '1px solid #e5e7eb',
            position: 'sticky',
            top: 0,
            zIndex: 99,
          }}
        >
          <Space size={16}>
            <NotificationBell />
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight" trigger={['click']}>
              <Space
                style={{
                  cursor: 'pointer',
                  padding: '4px 8px',
                  borderRadius: 8,
                  transition: 'background 0.2s',
                }}
                onMouseEnter={(e) => (e.currentTarget.style.background = '#f3f4f6')}
                onMouseLeave={(e) => (e.currentTarget.style.background = 'transparent')}
              >
                <Avatar size={32} style={{ background: '#1677ff' }} icon={<UserOutlined />} />
                <span style={{ fontWeight: 500, fontSize: 14 }}>{userInfo?.username || 'Admin'}</span>
              </Space>
            </Dropdown>
          </Space>
        </Header>
        <Content style={{ background: '#f5f5f5', minHeight: 'calc(100vh - 64px)' }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
```

- [ ] **Step 1: 修改 BasicLayout 集成新组件**

Replace content of `order-admin/src/layouts/BasicLayout/index.tsx` with the code above.

- [ ] **Step 2: 提交**

```bash
git add order-admin/src/layouts/BasicLayout/index.tsx
git commit -m "feat(sidebar): integrate sortable menu and collapse functionality"
```

---

## Task 6: 修复导入问题

**Files:**
- Check: `order-admin/src/components/SortableMenu/index.tsx`

- [ ] **Step 1: 检查并修复 SortableMenu 中的 useLocation 导入**

Read `order-admin/src/components/SortableMenu/index.tsx` and add missing import:
```tsx
import { useNavigate, useLocation } from 'react-router-dom'
```

- [ ] **Step 2: 构建验证**

Run: `cd /e/awesomeProject/go-micro/order-admin && npm run build`
Expected: Build succeeds

- [ ] **Step 3: 提交**

```bash
git add order-admin/src/components/SortableMenu/index.tsx
git commit -m "fix(sidebar): add missing useLocation import"
```

---

## Task 7: 测试验收

**Files:**
- Manual testing required

- [ ] **Step 1: 验证拖拽排序**
   - 打开浏览器进入任意页面
   - 尝试拖拽侧边栏菜单项
   - 确认拖拽时有卡片阴影效果
   - 确认释放后顺序改变

- [ ] **Step 2: 验证折叠功能**
   - 点击底部折叠按钮
   - 确认侧边栏收缩为 64px 图标模式
   - 点击展开按钮，确认恢复

- [ ] **Step 3: 验证悬浮模式**
   - 折叠侧边栏后，确认悬浮按钮出现
   - 点击悬浮按钮，确认侧边栏滑出
   - 点击其他区域，确认悬浮模式关闭

- [ ] **Step 4: 验证持久化**
   - 调整菜单顺序或折叠状态
   - 刷新页面
   - 确认设置保持

---

## 验收清单

- [ ] 菜单可通过拖拽调整顺序
- [ ] 拖拽时显示卡片阴影效果
- [ ] 点击折叠按钮，侧边栏收缩为 64px 图标模式
- [ ] 悬浮按钮在侧边栏隐藏时显示
- [ ] 点击悬浮按钮，侧边栏从左侧滑出
- [ ] 刷新页面后保持用户设置的顺序和折叠状态
- [ ] 动画流畅，无卡顿
- [ ] `npm run build` 成功