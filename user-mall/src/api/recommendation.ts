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
