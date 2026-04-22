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
import { useNavigate, useLocation } from 'react-router-dom'
import { useSidebarStore } from '@/stores/sidebarStore'

const iconMap: Record<string, React.ReactNode> = {
  DashboardOutlined: <DashboardOutlined />,
  ShoppingOutlined: <ShoppingOutlined />,
  CreditCardOutlined: <CreditCardOutlined />,
  RollbackOutlined: <RollbackOutlined />,
  AppstoreOutlined: <AppstoreOutlined />,
}

interface SortableMenuProps {
  collapsed: boolean
}

export const SortableMenu: React.FC<SortableMenuProps> = ({ collapsed }) => {
  const navigate = useNavigate()
  const location = useLocation()
  const { menuItems, menuOrder, setMenuOrder } = useSidebarStore()

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

  // Build menu items with drag handles
  const menuItemsWithDrag = orderedItems.map((item) => ({
    key: item.key,
    label: item.label,
    icon: collapsed ? iconMap[item.icon] : (
      <span style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <HolderOutlined style={{ fontSize: 12, opacity: 0.5, cursor: 'grab' }} />
        {iconMap[item.icon]}
      </span>
    ),
  }))

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
          items={menuItemsWithDrag}
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