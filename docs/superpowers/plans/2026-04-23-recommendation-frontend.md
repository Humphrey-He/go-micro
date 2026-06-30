# 个性化推荐前端实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现个性化推荐前端功能，包括行为上报、推荐接口对接、首页/商品详情/购物车/支付完成等场景的推荐展示。

**Architecture:**
- 创建独立的推荐 API 模块 `src/api/recommendation.ts`
- 创建通用推荐流组件 `RecommendationFeed.tsx`，支持多种推荐场景
- 扩展现有页面接入推荐：首页、商品详情、购物车、支付完成页
- 行为上报在用户浏览、收藏、加购、购买时自动触发

**Tech Stack:** React 18, TypeScript, Ant Design Mobile 5, Zustand, React Router 6

---

## File Structure

```
user-mall/src/
├── api/
│   └── recommendation.ts           # 创建：推荐 API 模块
├── components/
│   └── RecommendationFeed/         # 创建：推荐流组件
│       ├── index.tsx
│       └── RecommendationItem.tsx
├── hooks/
│   └── useRecommendation.ts        # 创建：推荐数据 hook
├── stores/
│   └── behaviorStore.ts            # 创建：行为上报状态（可选，用于批量上报）
└── pages/
    ├── Home/index.tsx              # 修改：接入推荐
    ├── ProductDetail/index.tsx     # 修改：接入相似推荐
    ├── Cart/index.tsx             # 修改：接入加价购
    └── Payment/Result.tsx         # 修改：接入猜你喜欢
```

---

## Task 1: 创建推荐 API 模块

**Files:**
- Create: `user-mall/src/api/recommendation.ts`

- [ ] **Step 1: 创建推荐 API 类型定义和接口**

创建 `user-mall/src/api/recommendation.ts`：

```typescript
import { get, post } from '@/api/request'

// ============ 行为上报 ============

export type BehaviorType = 'view' | 'favorite' | 'cart' | 'purchase'
export type RecommendationSource = 'home' | 'detail' | 'search' | 'cart' | 'recommendation'

export interface ReportBehaviorParams {
  sku_id: string
  behavior_type: BehaviorType
  stay_duration?: number    // 停留时长(秒)，仅浏览时
  source: RecommendationSource
  anonymous_id?: string    // 未登录用户匿名ID
}

export const reportBehavior = (params: ReportBehaviorParams) =>
  post<void>('/rec/report', params)

// ============ 推荐接口 ============

// 推荐商品项
export interface RecommendationItem {
  sku_id: string
  title: string
  price: number
  original_price: number
  image: string
  sales_count: number
  score?: number           // 推荐分数，仅debug模式返回
  similarity?: number      // 相似度分数
  match_reason?: string    // 推荐理由
}

// 首页推荐
export interface HomeRecommendResponse {
  items: RecommendationItem[]
  page: number
  page_size: number
  total: number
}

export const getHomeRecommendations = (params?: { page?: number; page_size?: number }) =>
  get<HomeRecommendResponse>('/rec/home', { params })

// 商品相似推荐
export type SimilarScene = 'view' | 'purchase'

export interface SimilarRecommendResponse {
  scene: SimilarScene
  items: RecommendationItem[]
}

export const getSimilarRecommendations = (skuId: string, params?: { scene?: SimilarScene; limit?: number }) =>
  get<SimilarRecommendResponse>(`/rec/similar/${skuId}`, { params })

// 购物车加价购
export interface CartAddonParams {
  cart_sku_ids: string[]
}

export interface CartAddonResponse {
  items: (RecommendationItem & { addon_price: number })[]
}

export const getCartAddons = (params: CartAddonParams) =>
  post<CartAddonResponse>('/rec/cart-addon', params)

// 支付完成推荐
export interface PayCompleteRecommendParams {
  purchased_sku_ids: string[]
  limit?: number
}

export interface PayCompleteRecommendResponse {
  items: RecommendationItem[]
}

export const getPayCompleteRecommendations = (params: PayCompleteRecommendParams) =>
  get<PayCompleteRecommendResponse>('/rec/pay-complete', { params })

// 冷启动推荐
export interface ColdStartResponse {
  hot_items: RecommendationItem[]
  category_prefs: Array<{
    category_id: string
    name: string
    image: string
  }>
}

export const getColdStartRecommendations = () =>
  get<ColdStartResponse>('/rec/cold-start')
```

---

## Task 2: 创建推荐流组件

**Files:**
- Create: `user-mall/src/components/RecommendationFeed/index.tsx`
- Create: `user-mall/src/components/RecommendationFeed/RecommendationItem.tsx`

- [ ] **Step 1: 创建 RecommendationItem 子组件**

创建 `user-mall/src/components/RecommendationFeed/RecommendationItem.tsx`：

```tsx
import { useNavigate } from 'react-router-dom'
import type { RecommendationItem as IRecommendationItem } from '@/api/recommendation'

interface Props {
  item: IRecommendationItem
  layout?: 'vertical' | 'horizontal'
}

export default function RecommendationItem({ item, layout = 'vertical' }: Props) {
  const navigate = useNavigate()

  const formatPrice = (price: number) => `¥${(price / 100).toFixed(2)}`

  if (layout === 'horizontal') {
    return (
      <div
        className="flex bg-white rounded-lg p-2 cursor-pointer hover:bg-gray-50"
        onClick={() => navigate(`/product/${item.sku_id}`)}
      >
        <img
          src={item.image}
          alt={item.title}
          className="w-20 h-20 rounded object-cover bg-gray-100"
        />
        <div className="ml-3 flex-1 flex flex-col justify-between">
          <div>
            <div className="text-sm line-clamp-2">{item.title}</div>
            {item.match_reason && (
              <div className="text-xs text-gray-400 mt-1">{item.match_reason}</div>
            )}
          </div>
          <div className="flex items-center justify-between mt-1">
            <span className="text-primary-500 font-bold">{formatPrice(item.price)}</span>
            {item.original_price > item.price && (
              <span className="text-gray-400 text-xs line-through">
                {formatPrice(item.original_price)}
              </span>
            )}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div
      className="bg-white rounded-lg overflow-hidden cursor-pointer"
      onClick={() => navigate(`/product/${item.sku_id}`)}
    >
      <div className="relative">
        <img
          src={item.image}
          alt={item.title}
          className="w-full aspect-square object-cover bg-gray-100"
        />
        {item.original_price > item.price && (
          <div className="absolute top-1 left-1 bg-primary-500 text-white text-xs px-1 rounded">
            {Math.round((1 - item.price / item.original_price) * 100)}%
          </div>
        )}
      </div>
      <div className="p-2">
        <div className="text-sm line-clamp-2">{item.title}</div>
        {item.match_reason && (
          <div className="text-xs text-gray-400 mt-1 line-clamp-1">{item.match_reason}</div>
        )}
        <div className="mt-2 flex items-baseline">
          <span className="text-primary-500 font-bold">{formatPrice(item.price)}</span>
          {item.original_price > item.price && (
            <span className="text-gray-400 text-xs line-through ml-1">
              {formatPrice(item.original_price)}
            </span>
          )}
        </div>
        <div className="mt-1 text-xs text-gray-400">
          销量 {item.sales_count}
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 创建 RecommendationFeed 主组件**

创建 `user-mall/src/components/RecommendationFeed/index.tsx`：

```tsx
import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Skeleton } from 'antd-mobile'
import RecommendationItem from './RecommendationItem'
import type { RecommendationItem as IRecommendationItem } from '@/api/recommendation'

type LayoutType = 'grid' | 'list' | 'horizontal'

interface Props {
  api: () => Promise<{ items: IRecommendationItem[] }>
  layout?: LayoutType
  title?: string
  emptyText?: string
  onLoad?: (items: IRecommendationItem[]) => void
}

export default function RecommendationFeed({
  api,
  layout = 'grid',
  title,
  emptyText = '暂无推荐',
  onLoad,
}: Props) {
  const navigate = useNavigate()
  const [items, setItems] = useState<IRecommendationItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchRecommendations = async () => {
      try {
        setLoading(true)
        setError(null)
        const response = await api()
        setItems(response.items || [])
        onLoad?.(response.items || [])
      } catch (err) {
        console.error('Failed to load recommendations:', err)
        setError('加载失败')
      } finally {
        setLoading(false)
      }
    }

    fetchRecommendations()
  }, [api, onLoad])

  if (loading) {
    return (
      <div className="p-4 bg-white">
        {title && <div className="font-bold text-lg mb-3">{title}</div>}
        <div className={layout === 'grid' ? 'grid grid-cols-2 gap-3' : 'space-y-2'}>
          {[1, 2, 3, 4].map((i) => (
            <Skeleton key={i} animated="pulse" height={layout === 'grid' ? '180px' : '80px'} />
          ))}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="p-4 bg-white text-center text-gray-400">
        <div className="text-sm">{error}</div>
      </div>
    )
  }

  if (items.length === 0) {
    return (
      <div className="p-4 bg-white text-center text-gray-400">
        <div className="text-sm">{emptyText}</div>
      </div>
    )
  }

  return (
    <div className="p-4 bg-white">
      {title && (
        <div className="flex items-center justify-between mb-3">
          <span className="font-bold text-lg">{title}</span>
          <span
            className="text-sm text-primary-500"
            onClick={() => navigate('/product/list')}
          >
            查看更多 &gt;
          </span>
        </div>
      )}

      {layout === 'grid' && (
        <div className="grid grid-cols-2 gap-3">
          {items.map((item) => (
            <RecommendationItem key={item.sku_id} item={item} layout="vertical" />
          ))}
        </div>
      )}

      {layout === 'list' && (
        <div className="space-y-3">
          {items.map((item) => (
            <RecommendationItem key={item.sku_id} item={item} layout="horizontal" />
          ))}
        </div>
      )}

      {layout === 'horizontal' && (
        <div className="flex gap-3 overflow-x-auto hide-scrollbar pb-2">
          {items.map((item) => (
            <div key={item.sku_id} className="flex-shrink-0 w-36">
              <RecommendationItem item={item} layout="vertical" />
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
```

---

## Task 3: 创建行为上报 Hook

**Files:**
- Create: `user-mall/src/hooks/useBehaviorReport.ts`

- [ ] **Step 1: 创建 useBehaviorReport Hook**

创建 `user-mall/src/hooks/useBehaviorReport.ts`：

```typescript
import { useCallback, useRef, useEffect } from 'react'
import { reportBehavior, BehaviorType, RecommendationSource } from '@/api/recommendation'
import { useAuthStore } from '@/stores/authStore'

interface UseBehaviorReportOptions {
  skuId: string
  source: RecommendationSource
  enabled?: boolean
}

export function useBehaviorReport({ skuId, source, enabled = true }: UseBehaviorReportOptions) {
  const viewStartTime = useRef<number>(0)
  const hasReported = useRef(false)
  const { isLoggedIn, userInfo } = useAuthStore()

  // 页面可见性变化时上报浏览时长
  useEffect(() => {
    if (!enabled) return

    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        viewStartTime.current = Date.now()
      } else if (document.visibilityState === 'hidden' && !hasReported.current) {
        const stayDuration = Math.round((Date.now() - viewStartTime.current) / 1000)
        if (stayDuration >= 2) {
          // 停留2秒以上才上报
          reportBehavior({
            sku_id: skuId,
            behavior_type: 'view',
            stay_duration: stayDuration,
            source,
          })
          hasReported.current = true
        }
      }
    }

    viewStartTime.current = Date.now()
    document.addEventListener('visibilitychange', handleVisibilityChange)

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [skuId, source, enabled])

  // 上报收藏
  const reportFavorite = useCallback(
    async (action: 'add' | 'remove') => {
      if (!enabled) return
      try {
        await reportBehavior({
          sku_id: skuId,
          behavior_type: 'favorite',
          source,
        })
      } catch (err) {
        console.error('Failed to report favorite:', err)
      }
    },
    [skuId, source, enabled]
  )

  // 上报加购
  const reportCart = useCallback(
    async (action: 'add' | 'remove', quantity: number = 1) => {
      if (!enabled) return
      try {
        await reportBehavior({
          sku_id: skuId,
          behavior_type: 'cart',
          source,
        })
      } catch (err) {
        console.error('Failed to report cart:', err)
      }
    },
    [skuId, source, enabled]
  )

  // 上报购买
  const reportPurchase = useCallback(
    async (orderId: string) => {
      if (!enabled) return
      try {
        await reportBehavior({
          sku_id: skuId,
          behavior_type: 'purchase',
          source,
        })
      } catch (err) {
        console.error('Failed to report purchase:', err)
      }
    },
    [skuId, source, enabled]
  )

  return {
    reportFavorite,
    reportCart,
    reportPurchase,
  }
}
```

---

## Task 4: 更新首页接入推荐

**Files:**
- Modify: `user-mall/src/pages/Home/index.tsx`

- [ ] **Step 1: 更新首页接入推荐 API**

在 `Home/index.tsx` 中：

1. 添加导入：
```tsx
import RecommendationFeed from '@/components/RecommendationFeed'
import { getHomeRecommendations } from '@/api/recommendation'
```

2. 替换"为你推荐"区域，使用 RecommendationFeed 组件：

找到当前的推荐区域（约134-145行）：
```tsx
{/* 商品推荐 */}
<div className="mt-2 bg-white p-4">
  <div className="flex items-center justify-between mb-3">
    <span className="font-bold text-lg">为你推荐</span>
    <span className="text-sm text-gray-400">根据你的喜好</span>
  </div>
  <div className="grid grid-cols-2 gap-3">
    {products.slice(0, 4).map((product) => (
      <ProductCard key={product.sku_id} product={product as any} />
    ))}
  </div>
</div>
```

替换为：
```tsx
{/* 为你推荐 */}
<RecommendationFeed
  api={() => getHomeRecommendations({ page: 1, page_size: 6 })}
  layout="grid"
  title="为你推荐"
  emptyText="暂无推荐，看看热销商品"
/>
```

---

## Task 5: 更新商品详情页接入相似推荐

**Files:**
- Modify: `user-mall/src/pages/ProductDetail/index.tsx`

- [ ] **Step 1: 阅读当前商品详情页结构

找到关键位置：
- 商品信息展示区域
- "看了又看"、"买了还买"推荐区域（如果存在）
- 收藏/加购按钮所在位置

- [ ] **Step 2: 添加相似推荐区域

在商品详情页底部（评价区域之前）添加：

```tsx
import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { getSimilarRecommendations, SimilarScene } from '@/api/recommendation'
// ... existing imports

// 在组件内添加状态：
const [similarProducts, setSimilarProducts] = useState<any[]>([])
const [similarScene, setSimilarScene] = useState<'view' | 'purchase'>('view')

// 在 useEffect 中获取推荐数据：
useEffect(() => {
  const fetchSimilar = async () => {
    try {
      const res = await getSimilarRecommendations(skuId, { scene: 'view', limit: 6 })
      setSimilarProducts(res.items)
    } catch (err) {
      console.error('Failed to load similar products:', err)
    }
  }
  if (skuId) fetchSimilar()
}, [skuId])

// 在 JSX 中添加（商品详情Tab之后、评价Tab之前）：
<div className="mt-2 bg-white">
  {/* Tab 切换 */}
  <div className="flex border-b">
    <div
      className={`flex-1 py-3 text-center ${similarScene === 'view' ? 'text-primary-500 border-b-2 border-primary-500' : 'text-gray-500'}`}
      onClick={() => setSimilarScene('view')}
    >
      看了又看
    </div>
    <div
      className={`flex-1 py-3 text-center ${similarScene === 'purchase' ? 'text-primary-500 border-b-2 border-primary-500' : 'text-gray-500'}`}
      onClick={() => setSimilarScene('purchase')}
    >
      买了还买
    </div>
  </div>

  {/* 推荐列表 */}
  <div className="p-3">
    <div className="flex gap-3 overflow-x-auto hide-scrollbar">
      {similarProducts.map((item) => (
        <div
          key={item.sku_id}
          className="flex-shrink-0 w-32"
          onClick={() => navigate(`/product/${item.sku_id}`)}
        >
          <img src={item.image} alt={item.title} className="w-32 h-32 rounded object-cover" />
          <div className="mt-1 text-sm line-clamp-2">{item.title}</div>
          <div className="text-primary-500 font-bold">¥{(item.price / 100).toFixed(2)}</div>
        </div>
      ))}
    </div>
  </div>
</div>
```

---

## Task 6: 更新购物车页面接入加价购

**Files:**
- Modify: `user-mall/src/pages/Cart/index.tsx`

- [ ] **Step 1: 阅读当前购物车页面结构

找到关键位置：
- 商品列表区域
- 价格汇总区域
- 去结算按钮所在位置

- [ ] **Step 2: 添加加价购推荐区域

在价格汇总和去结算按钮之间添加：

```tsx
import { getCartAddons } from '@/api/recommendation'

// 在组件内添加：
const [cartAddons, setCartAddons] = useState<any[]>([])
const cartSkuIds = selectedItems.map(item => item.sku_id)

// 获取加价购
useEffect(() => {
  if (cartSkuIds.length > 0) {
    const fetchAddons = async () => {
      try {
        const res = await getCartAddons({ cart_sku_ids: cartSkuIds })
        setCartAddons(res.items || [])
      } catch (err) {
        console.error('Failed to load cart addons:', err)
      }
    }
    fetchAddons()
  }
}, [cartSkuIds.join(',')]) // 依赖选中商品变化

// 在 JSX 中（去结算按钮之前）添加：
{cartAddons.length > 0 && (
  <div className="bg-white mt-2 p-3">
    <div className="flex items-center justify-between mb-2">
      <span className="font-bold">加价购</span>
      <span className="text-xs text-gray-400">更优惠</span>
    </div>
    <div className="flex gap-2 overflow-x-auto hide-scrollbar">
      {cartAddons.map((item) => (
        <div
          key={item.sku_id}
          className="flex-shrink-0 w-24 text-center"
          onClick={() => navigate(`/product/${item.sku_id}`)}
        >
          <img src={item.image} alt={item.title} className="w-24 h-24 rounded object-cover" />
          <div className="text-xs line-clamp-1 mt-1">{item.title}</div>
          <div className="text-primary-500 font-bold text-sm">
            ¥{(item.addon_price / 100).toFixed(2)}
          </div>
          <div className="text-xs text-gray-400 line-through">
            ¥{(item.price / 100).toFixed(2)}
          </div>
        </div>
      ))}
    </div>
  </div>
)}
```

---

## Task 7: 更新支付完成页接入猜你喜欢

**Files:**
- Modify: `user-mall/src/pages/Payment/Result.tsx`

- [ ] **Step 1: 阅读当前支付结果页面结构

找到关键位置：
- 支付成功/失败展示区域
- 订单信息展示区域
- 页面底部区域

- [ ] **Step 2: 添加猜你喜欢推荐

在支付结果页面底部（返回首页等按钮之后）添加：

```tsx
import RecommendationFeed from '@/components/RecommendationFeed'
import { getPayCompleteRecommendations } from '@/api/recommendation'

// 获取订单中的商品 ID（从 URL 参数或页面状态获取）
// 假设订单商品 IDs 存储在组件状态中
const orderSkuIds = /* 从订单信息中提取 */ []

// 使用 RecommendationFeed：
<RecommendationFeed
  api={() => getPayCompleteRecommendations({
    purchased_sku_ids: orderSkuIds,
    limit: 6
  })}
  layout="grid"
  title="猜你喜欢"
  emptyText="暂无推荐"
/>
```

---

## Task 8: 验证与提交

- [ ] **Step 1: 运行构建验证**

```bash
cd user-mall && npm run build 2>&1 | head -50
```

- [ ] **Step 2: 修复任何 TypeScript 错误**

常见问题：
- 类型不匹配
- 缺失导入
- 未定义的接口字段

- [ ] **Step 3: 提交代码**

```bash
git add -A
git commit -m "feat(recommendation): add personalized recommendation frontend

- Add recommendation API module with behavior reporting
- Add RecommendationFeed component for displaying recommendations
- Add useBehaviorReport hook for automatic view tracking
- Integrate recommendations into Home, ProductDetail, Cart, PaymentResult pages

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Spec Coverage Checklist

| Spec Section | Task | Status |
|-------------|------|--------|
| 行为上报 API | Task 1 | ✅ |
| 首页推荐 API | Task 1 | ✅ |
| 相似推荐 API | Task 1 | ✅ |
| 加价购 API | Task 1 | ✅ |
| 支付完成推荐 API | Task 1 | ✅ |
| 冷启动 API | Task 1 | ✅ |
| RecommendationFeed 组件 | Task 2 | ✅ |
| useBehaviorReport Hook | Task 3 | ✅ |
| 首页接入推荐 | Task 4 | ✅ |
| 商品详情相似推荐 | Task 5 | ✅ |
| 购物车加价购 | Task 6 | ✅ |
| 支付完成猜你喜欢 | Task 7 | ✅ |

---

## Execution Options

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
