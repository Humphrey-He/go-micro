# 价格监控前端实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现价格监控前端功能，包括降价提醒设置、价格走势展示、监控列表管理、降价通知展示。

**Architecture:**
- 创建独立的 API 模块 `src/api/priceWatch.ts`
- 创建价格走势图组件 `PriceTrendChart.tsx`（使用简单 SVG 折线图）
- 创建降价提醒按钮组件 `PriceWatchButton.tsx`
- 在商品详情页接入降价提醒和价格走势
- 创建监控列表页面 `PriceWatchList`
- 在消息中心增加"价格提醒"Tab

**Tech Stack:** React 18, TypeScript, Ant Design Mobile 5, Zustand, React Router 6

---

## File Structure

```
user-mall/src/
├── api/
│   └── priceWatch.ts              # 创建：价格监控 API
├── components/
│   ├── PriceTrendChart/          # 创建：价格走势图
│   │   ├── index.tsx
│   │   └── PriceTrendChart.tsx
│   └── PriceWatchButton/         # 创建：降价提醒按钮
│       └── index.tsx
├── pages/
│   ├── PriceWatchList/           # 创建：价格监控列表页
│   │   └── index.tsx
│   ├── ProductDetail/index.tsx   # 修改：接入降价提醒和价格走势
│   └── User/Messages.tsx         # 修改：增加价格提醒 Tab
```

---

## Task 1: 创建价格监控 API 模块

**Files:**
- Create: `user-mall/src/api/priceWatch.ts`

- [ ] **Step 1: 创建价格监控 API

创建 `user-mall/src/api/priceWatch.ts`：

```typescript
import { get, post, del, put } from '@/api/request'

// ============ 价格监控 API ============

// 设置价格监控
export interface SetPriceWatchParams {
  sku_id: string
  target_price?: number  // 可选，NULL表示任意降幅提醒
}

export interface SetPriceWatchResponse {
  watch_id: number
  current_price: number
  target_price?: number
}

export const setPriceWatch = (params: SetPriceWatchParams) =>
  post<SetPriceWatchResponse>('/price-watch', params)

// 取消价格监控
export const cancelPriceWatch = (skuId: string) =>
  del<void>(`/price-watch/${skuId}`)

// 获取我的价格监控列表
export interface PriceWatchItem {
  watch_id: number
  sku_id: string
  product_name: string
  image: string
  current_price: number
  original_price: number
  target_price?: number
  notify_enabled: boolean
  lowest_price: number
  lowest_price_date: string
  created_at: string
  price_trend: 'down' | 'stable' | 'up'
}

export interface PriceWatchListResponse {
  items: PriceWatchItem[]
  page: number
  page_size: number
  total: number
}

export interface GetPriceWatchListParams {
  page?: number
  page_size?: number
  status?: 'active' | 'all'
}

export const getPriceWatchList = (params?: GetPriceWatchListParams) =>
  get<PriceWatchListResponse>('/price-watch/list', { params })

// 更新监控设置
export interface UpdatePriceWatchParams {
  target_price?: number
  notify_enabled?: boolean
}

export const updatePriceWatch = (skuId: string, params: UpdatePriceWatchParams) =>
  put<void>(`/price-watch/${skuId}`, params)

// 获取商品价格走势
export interface PricePoint {
  date: string
  price: number
}

export interface PriceHistoryResponse {
  sku_id: string
  current_price: number
  lowest_price: number
  lowest_date: string
  highest_price: number
  highest_date: string
  average_price: number
  price_points: PricePoint[]
}

export type PriceHistoryPeriod = '30d' | '90d' | '1y'

export const getPriceHistory = (skuId: string, period: PriceHistoryPeriod = '30d') =>
  get<PriceHistoryResponse>(`/products/${skuId}/price-history`, { params: { period } })

// 价格提醒通知列表
export interface PriceWatchNotification {
  notification_id: number
  sku_id: string
  product_name: string
  image: string
  old_price: number
  new_price: number
  discount_amount: number
  discount_rate: number
  is_read: boolean
  created_at: string
}

export interface PriceWatchNotificationResponse {
  items: PriceWatchNotification[]
  unread_count: number
  page: number
  page_size: number
  total: number
}

export interface GetPriceWatchNotificationsParams {
  page?: number
  page_size?: number
}

export const getPriceWatchNotifications = (params?: GetPriceWatchNotificationsParams) =>
  get<PriceWatchNotificationResponse>('/notifications/price-watch', { params })
```

---

## Task 2: 创建价格走势图组件

**Files:**
- Create: `user-mall/src/components/PriceTrendChart/PriceTrendChart.tsx`
- Create: `user-mall/src/components/PriceTrendChart/index.tsx`

- [ ] **Step 1: 创建 PriceTrendChart.tsx

Create `user-mall/src/components/PriceTrendChart/PriceTrendChart.tsx`:

```tsx
import { useMemo } from 'react'

interface PricePoint {
  date: string
  price: number
}

interface Props {
  pricePoints: PricePoint[]
  width?: number
  height?: number
  currentPrice?: number
}

export default function PriceTrendChart({
  pricePoints,
  width = 300,
  height = 150,
  currentPrice
}: Props) {
  const { minPrice, maxPrice, points, path } = useMemo(() => {
    if (!pricePoints || pricePoints.length === 0) {
      return { minPrice: 0, maxPrice: 100, points: [], path: '' }
    }

    const prices = pricePoints.map(p => p.price)
    const min = Math.min(...prices)
    const max = Math.max(...prices)
    const padding = (max - min) * 0.1 || 10

    const chartWidth = width - 40
    const chartHeight = height - 30
    const chartLeft = 30
    const chartTop = 10

    const pts = pricePoints.map((point, index) => {
      const x = chartLeft + (index / (pricePoints.length - 1 || 1)) * chartWidth
      const y = chartTop + chartHeight - ((point.price - min + padding) / (max - min + padding * 2)) * chartHeight
      return { x, y, price: point.price, date: point.date }
    })

    // Generate SVG path
    let pathD = ''
    if (pts.length > 0) {
      pathD = `M ${pts[0].x} ${pts[0].y}`
      for (let i = 1; i < pts.length; i++) {
        pathD += ` L ${pts[i].x} ${pts[i].y}`
      }
    }

    return {
      minPrice: min,
      maxPrice: max,
      points: pts,
      path: pathD
    }
  }, [pricePoints, width, height])

  const formatPrice = (price: number) => `¥${(price / 100).toFixed(2)}`

  return (
    <div className="relative" style={{ width, height }}>
      <svg width={width} height={height} className="overflow-visible">
        {/* Y-axis labels */}
        <text x={25} y={15} className="text-xs fill-gray-400">{formatPrice(maxPrice)}</text>
        <text x={25} y={height - 15} className="text-xs fill-gray-400">{formatPrice(minPrice)}</text>

        {/* Grid lines */}
        {[0, 0.25, 0.5, 0.75, 1].map((ratio, i) => {
          const y = 10 + (height - 30) * (1 - ratio)
          return (
            <line
              key={i}
              x1={30}
              y1={y}
              x2={width - 10}
              y2={y}
              stroke="#f0f0f0"
              strokeDasharray="2,2"
            />
          )
        })}

        {/* Current price line */}
        {currentPrice && (
          <line
            x1={30}
            y1={10 + (height - 30) * (1 - (currentPrice - minPrice + (maxPrice - minPrice) * 0.1) / ((maxPrice - minPrice) * 1.2))}
            x2={width - 10}
            y2={10 + (height - 30) * (1 - (currentPrice - minPrice + (maxPrice - minPrice) * 0.1) / ((maxPrice - minPrice) * 1.2))}
            stroke="#3b82f6"
            strokeDasharray="4,2"
            strokeOpacity="0.5"
          />
        )}

        {/* Price line */}
        {points.length > 1 && (
          <>
            <path
              d={path}
              fill="none"
              stroke="#ef4444"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            {/* Data points */}
            {points.map((point, index) => (
              <circle
                key={index}
                cx={point.x}
                cy={point.y}
                r={3}
                fill="#ef4444"
              />
            ))}
          </>
        )}
      </svg>

      {/* X-axis labels */}
      {points.length > 0 && (
        <div className="absolute bottom-0 left-8 right-2 flex justify-between text-xs text-gray-400">
          <span>{points[0]?.date?.slice(5) || ''}</span>
          <span>{points[points.length - 1]?.date?.slice(5) || ''}</span>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 2: 创建索引文件

Create `user-mall/src/components/PriceTrendChart/index.tsx`:

```tsx
export { default as PriceTrendChart } from './PriceTrendChart'
```

---

## Task 3: 创建降价提醒按钮组件

**Files:**
- Create: `user-mall/src/components/PriceWatchButton/index.tsx`

- [ ] **Step 1: 创建 PriceWatchButton.tsx

Create `user-mall/src/components/PriceWatchButton/PriceWatchButton.tsx`:

```tsx
import { useState } from 'react'
import { Button, Dialog, Input, Toast } from 'antd-mobile'
import { setPriceWatch, cancelPriceWatch } from '@/api/priceWatch'

interface Props {
  skuId: string
  productName: string
  currentPrice: number
  isWatching?: boolean
  onWatchChange?: (isWatching: boolean) => void
}

export default function PriceWatchButton({
  skuId,
  productName,
  currentPrice,
  isWatching = false,
  onWatchChange
}: Props) {
  const [showDialog, setShowDialog] = useState(false)
  const [targetPrice, setTargetPrice] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSetWatch = async () => {
    try {
      setLoading(true)
      const target = targetPrice ? Math.round(parseFloat(targetPrice) * 100) : undefined
      await setPriceWatch({ sku_id: skuId, target_price: target })
      Toast.show('降价提醒设置成功')
      setShowDialog(false)
      onWatchChange?.(true)
    } catch (error: any) {
      Toast.show(error?.message || '设置失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCancelWatch = async () => {
    try {
      setLoading(true)
      await cancelPriceWatch(skuId)
      Toast.show('已取消降价提醒')
      onWatchChange?.(false)
    } catch (error: any) {
      Toast.show(error?.message || '取消失败')
    } finally {
      setLoading(false)
    }
  }

  const handleClick = () => {
    if (isWatching) {
      Dialog.confirm({
        content: '确定要取消该商品的降价提醒吗？',
        onConfirm: handleCancelWatch,
      })
    } else {
      setTargetPrice('')
      setShowDialog(true)
    }
  }

  const formatPrice = (price: number) => `¥${(price / 100).toFixed(2)}`

  return (
    <>
      <Button
        size="small"
        color={isWatching ? 'success' : 'default'}
        fill={isWatching ? 'solid' : 'outline'}
        onClick={handleClick}
      >
        <span className="flex items-center gap-1">
          <span>🔔</span>
          <span>{isWatching ? '已设置提醒' : '降价提醒'}</span>
        </span>
      </Button>

      {/* 设置提醒弹窗 */}
      <Dialog
        visible={showDialog}
        title="设置降价提醒"
        onClose={() => setShowDialog(false)}
        content={
          <div className="py-2">
            <div className="text-sm text-gray-600 mb-3">
              商品当前价格：<span className="text-primary-500 font-bold">{formatPrice(currentPrice)}</span>
            </div>
            <div className="mb-2 text-sm text-gray-500">目标价格（选填）</div>
            <Input
              type="number"
              placeholder="输入目标价格，降价到该价格时提醒"
              value={targetPrice}
              onChange={setTargetPrice}
              prefix="¥"
            />
            <div className="text-xs text-gray-400 mt-2">
              不填目标价格表示：价格下降任意幅度都提醒
            </div>
          </div>
        }
        actions={[
          { text: '取消', key: 'cancel', onClick: () => setShowDialog(false) },
          { text: '确认', key: 'confirm', onClick: handleSetWatch, loading: loading },
        ]}
      />
    </>
  )
}
```

- [ ] **Step 2: 创建索引文件

Create `user-mall/src/components/PriceWatchButton/index.tsx`:

```tsx
export { default as PriceWatchButton } from './PriceWatchButton'
```

---

## Task 4: 创建价格监控列表页面

**Files:**
- Create: `user-mall/src/pages/PriceWatchList/index.tsx`
- Modify: `user-mall/src/App.tsx` (添加路由)

- [ ] **Step 1: 创建价格监控列表页面

Create `user-mall/src/pages/PriceWatchList/index.tsx`:

```tsx
import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Skeleton, Switch, Empty, PullToRefresh } from 'antd-mobile'
import { getPriceWatchList, updatePriceWatch, cancelPriceWatch, type PriceWatchItem } from '@/api/priceWatch'
import PriceTrendChart from '@/components/PriceTrendChart/PriceTrendChart'

export default function PriceWatchList() {
  const navigate = useNavigate()
  const [items, setItems] = useState<PriceWatchItem[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  const fetchList = async () => {
    try {
      const res = await getPriceWatchList({ page: 1, page_size: 50, status: 'all' })
      setItems(res.items || [])
    } catch (err) {
      console.error('Failed to fetch price watch list:', err)
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }

  useEffect(() => {
    fetchList()
  }, [])

  const handleRefresh = async () => {
    setRefreshing(true)
    await fetchList()
  }

  const handleToggleNotify = async (item: PriceWatchItem, checked: boolean) => {
    try {
      await updatePriceWatch(item.sku_id, { notify_enabled: checked })
      setItems(prev => prev.map(i =>
        i.sku_id === item.sku_id ? { ...i, notify_enabled: checked } : i
      ))
    } catch (err) {
      console.error('Failed to update notify setting:', err)
    }
  }

  const handleCancel = async (item: PriceWatchItem) => {
    try {
      await cancelPriceWatch(item.sku_id)
      setItems(prev => prev.filter(i => i.sku_id !== item.sku_id))
    } catch (err) {
      console.error('Failed to cancel watch:', err)
    }
  }

  const formatPrice = (price: number) => `¥${(price / 100).toFixed(2)}`

  const getTrendIcon = (trend: string) => {
    if (trend === 'down') return '📉'
    if (trend === 'up') return '📈'
    return '➡️'
  }

  const getTrendColor = (trend: string) => {
    if (trend === 'down') return 'text-green-500'
    if (trend === 'up') return 'text-red-500'
    return 'text-gray-500'
  }

  if (loading) {
    return (
      <div className="p-4 space-y-3">
        {[1, 2, 3].map(i => (
          <Skeleton key={i} animated height={100} className="rounded-lg" />
        ))}
      </div>
    )
  }

  if (items.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <Empty description="暂无价格监控" />
        <button
          className="mt-4 text-primary-500"
          onClick={() => navigate('/')}
        >
          去逛逛商品
        </button>
      </div>
    )
  }

  return (
    <PullToRefresh onRefresh={handleRefresh}>
      <div className="p-4 space-y-3">
        {items.map(item => (
          <div key={item.watch_id} className="bg-white rounded-lg p-3 shadow-sm">
            <div className="flex gap-3">
              {/* 商品图片 */}
              <img
                src={item.image}
                alt={item.product_name}
                className="w-20 h-20 rounded object-cover bg-gray-100"
                onClick={() => navigate(`/product/${item.sku_id}`)}
              />

              {/* 商品信息 */}
              <div className="flex-1">
                <div
                  className="text-sm font-medium line-clamp-2"
                  onClick={() => navigate(`/product/${item.sku_id}`)}
                >
                  {item.product_name}
                </div>

                {/* 价格信息 */}
                <div className="mt-1 flex items-baseline gap-2">
                  <span className="text-primary-500 font-bold">
                    {formatPrice(item.current_price)}
                  </span>
                  {item.original_price > item.current_price && (
                    <span className="text-gray-400 text-xs line-through">
                      {formatPrice(item.original_price)}
                    </span>
                  )}
                  {item.target_price && (
                    <span className="text-xs text-gray-500">
                      目标: {formatPrice(item.target_price)}
                    </span>
                  )}
                </div>

                {/* 趋势 */}
                <div className="mt-1 flex items-center gap-2">
                  <span className={`text-sm ${getTrendColor(item.price_trend)}`}>
                    {getTrendIcon(item.price_trend)} 历史最低: {formatPrice(item.lowest_price)}
                  </span>
                </div>

                {/* 操作 */}
                <div className="mt-2 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Switch
                      checked={item.notify_enabled}
                      onChange={(checked) => handleToggleNotify(item, checked)}
                      size="small"
                    />
                    <span className="text-xs text-gray-500">提醒</span>
                  </div>
                  <button
                    className="text-xs text-gray-400"
                    onClick={() => handleCancel(item)}
                  >
                    删除
                  </button>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </PullToRefresh>
  )
}
```

- [ ] **Step 2: 添加路由

Modify `user-mall/src/App.tsx`:

1. Add import:
```tsx
const PriceWatchList = lazy(() => import('./pages/PriceWatchList'))
```

2. Add route:
```tsx
{ path: '/price-watch', component: PriceWatchList, layout: BasicLayout, requiresAuth: true },
```

---

## Task 5: 更新商品详情页接入降价提醒和价格走势

**Files:**
- Modify: `user-mall/src/pages/ProductDetail/index.tsx`

- [ ] **Step 1: 阅读商品详情页

Read `user-mall/src/pages/ProductDetail/index.tsx` to understand:
- Where the product info section is
- Where action buttons are (收藏/分享)
- How to add new buttons

- [ ] **Step 2: 添加导入

```tsx
import { useState, useEffect } from 'react'
import { PriceWatchButton } from '@/components/PriceWatchButton'
import { PriceTrendChart } from '@/components/PriceTrendChart'
import { getPriceHistory, getPriceWatchList, type PriceHistoryResponse } from '@/api/priceWatch'
```

- [ ] **Step 3: 添加状态

```tsx
const [isWatching, setIsWatching] = useState(false)
const [priceHistory, setPriceHistory] = useState<PriceHistoryResponse | null>(null)
const [showPriceHistory, setShowPriceHistory] = useState(false)
```

- [ ] **Step 4: 添加数据获取

```tsx
useEffect(() => {
  // 检查是否已在监控中
  const checkWatchStatus = async () => {
    try {
      const res = await getPriceWatchList({ page: 1, page_size: 50, status: 'active' })
      const watching = res.items?.some(item => item.sku_id === skuId)
      setIsWatching(watching)
    } catch (err) {
      console.error('Failed to check watch status:', err)
    }
  }

  // 获取价格走势
  const fetchPriceHistory = async () => {
    try {
      const res = await getPriceHistory(skuId, '30d')
      setPriceHistory(res)
    } catch (err) {
      console.error('Failed to fetch price history:', err)
    }
  }

  checkWatchStatus()
  fetchPriceHistory()
}, [skuId])
```

- [ ] **Step 5: 添加降价提醒按钮

在收藏按钮旁边添加降价提醒按钮：

```tsx
<PriceWatchButton
  skuId={skuId}
  productName={product.title}
  currentPrice={product.price}
  isWatching={isWatching}
  onWatchChange={setIsWatching}
/>
```

- [ ] **Step 6: 添加价格走势展示

在商品详情底部（评价Tab之前或之后）添加可折叠的价格走势区域：

```tsx
{/* 价格走势 */}
{priceHistory && priceHistory.price_points.length > 0 && (
  <div className="mt-2 bg-white">
    <div
      className="p-3 flex items-center justify-between"
      onClick={() => setShowPriceHistory(!showPriceHistory)}
    >
      <span className="font-bold">价格走势</span>
      <span className="text-gray-400">
        {showPriceHistory ? '收起' : '查看'}
      </span>
    </div>
    {showPriceHistory && (
      <div className="px-3 pb-3">
        <div className="flex justify-between text-sm mb-2">
          <span className="text-gray-500">
            最高: <span className="text-red-500">{formatPrice(priceHistory.highest_price)}</span>
          </span>
          <span className="text-gray-500">
            最低: <span className="text-green-500">{formatPrice(priceHistory.lowest_price)}</span>
          </span>
          <span className="text-gray-500">
            平均: <span>{formatPrice(priceHistory.average_price)}</span>
          </span>
        </div>
        <PriceTrendChart
          pricePoints={priceHistory.price_points}
          currentPrice={priceHistory.current_price}
          width={320}
          height={160}
        />
        <div className="mt-2 text-xs text-gray-400 text-center">
          数据来源：近30天价格记录
        </div>
      </div>
    )}
  </div>
)}
```

---

## Task 6: 更新消息中心增加价格提醒 Tab

**Files:**
- Modify: `user-mall/src/pages/User/Messages.tsx`

- [ ] **Step 1: 阅读当前 Messages.tsx

Read `user-mall/src/pages/User/Messages.tsx` to understand:
- Current tab structure
- How to add a new tab for price notifications

- [ ] **Step 2: 添加导入

```tsx
import { getPriceWatchNotifications, type PriceWatchNotification } from '@/api/priceWatch'
```

- [ ] **Step 3: 添加状态

```tsx
const [priceNotifications, setPriceNotifications] = useState<PriceWatchNotification[]>([])
const [priceUnreadCount, setPriceUnreadCount] = useState(0)
```

- [ ] **Step 4: 添加数据获取

```tsx
useEffect(() => {
  if (activeTab === 'price') {
    fetchPriceNotifications()
  }
}, [activeTab])

const fetchPriceNotifications = async () => {
  try {
    const res = await getPriceWatchNotifications({ page: 1, page_size: 50 })
    setPriceNotifications(res.items || [])
    setPriceUnreadCount(res.unread_count)
  } catch (err) {
    console.error('Failed to fetch price notifications:', err)
  }
}
```

- [ ] **Step 5: 在 Tabs 中添加价格提醒 Tab

找到现有的 Tabs 定义，添加新的 Tab：

```tsx
<Tabs.Tab title={`价格提醒${priceUnreadCount > 0 ? ` (${priceUnreadCount})` : ''}`} key="price">
```

- [ ] **Step 6: 添加价格提醒内容

在 Tabs.Tab content 中添加：

```tsx
<Tabs.Tab title={`价格提醒${priceUnreadCount > 0 ? ` (${priceUnreadCount})` : ''}`} key="price">
  {priceNotifications.length === 0 ? (
    <Empty description="暂无价格提醒" />
  ) : (
    <div className="p-4 space-y-3">
      {priceNotifications.map(item => (
        <div
          key={item.notification_id}
          className="bg-white rounded-lg p-3 flex gap-3"
          onClick={() => navigate(`/product/${item.sku_id}`)}
        >
          <img
            src={item.image}
            alt={item.product_name}
            className="w-16 h-16 rounded object-cover bg-gray-100"
          />
          <div className="flex-1">
            <div className="text-sm font-medium line-clamp-1">{item.product_name}</div>
            <div className="mt-1 text-sm">
              <span className="text-gray-400 line-through">{formatPrice(item.old_price)}</span>
              <span className="text-primary-500 font-bold ml-2">{formatPrice(item.new_price)}</span>
            </div>
            <div className="mt-1 text-xs text-green-500">
              降价 {formatPrice(item.discount_amount)} ({item.discount_rate.toFixed(1)}%)
            </div>
            <div className="mt-1 text-xs text-gray-400">
              {item.created_at}
            </div>
          </div>
          {!item.is_read && (
            <div className="w-2 h-2 bg-primary-500 rounded-full" />
          )}
        </div>
      ))}
    </div>
  )}
</Tabs.Tab>
```

---

## Task 7: 验证与提交

- [ ] **Step 1: 运行构建验证**

```bash
cd user-mall && npm run build 2>&1 | head -50
```

- [ ] **Step 2: 修复 TypeScript 错误**

- [ ] **Step 3: 提交代码**

```bash
git add -A
git commit -m "feat(price-watch): add price monitoring frontend

- Add price watch API module
- Add PriceTrendChart component for price history visualization
- Add PriceWatchButton component for setting price alerts
- Add PriceWatchList page for managing price watches
- Integrate price watch into ProductDetail page
- Add price watch notifications tab to Messages page

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Spec Coverage Checklist

| Spec Section | Task | Status |
|-------------|------|--------|
| 设置价格监控 API | Task 1 | ✅ |
| 取消价格监控 API | Task 1 | ✅ |
| 监控列表 API | Task 1 | ✅ |
| 更新监控设置 API | Task 1 | ✅ |
| 价格走势 API | Task 1 | ✅ |
| 提醒通知 API | Task 1 | ✅ |
| 价格走势图组件 | Task 2 | ✅ |
| 降价提醒按钮组件 | Task 3 | ✅ |
| 价格监控列表页 | Task 4 | ✅ |
| 商品详情页降价提醒 | Task 5 | ✅ |
| 商品详情页价格走势 | Task 5 | ✅ |
| 消息中心价格提醒 Tab | Task 6 | ✅ |

---

## Execution Options

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
