import { get, post, del } from '@/api/request'

// 商品列表
export interface ProductListParams {
  keyword?: string
  category_id?: string
  page?: number
  page_size?: number
  sort_by?: 'sales' | 'new' | 'price_asc' | 'price_desc'
  price_min?: number
  price_max?: number
  has_stock?: boolean
}

export interface Product {
  sku_id: string
  title: string
  subtitle: string
  images: string[]
  price: number
  original_price: number
  sales: number
  stock: number
  tags: string[]
  rating: number
  comment_count: number
  shop: {
    shop_id: string
    name: string
    logo: string
  }
}

export interface ProductListResponse {
  products: Product[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export const getProductList = (params: ProductListParams) =>
  get<ProductListResponse>('/products', { params })

// 商品详情
export interface ProductDetailResponse extends Product {
  video_url?: string
  skus: Array<{
    sku_id: string
    attributes: Record<string, string>
    stock: number
    price: number
    image?: string
  }>
  attributes: Array<{
    name: string
    values: Array<{ value: string; image?: string }>
  }>
  details: {
    description: string
    specifications: Array<{ name: string; value: string }>
  }
  is_favorite: boolean
  favorite_count: number
}

export const getProductDetail = (skuId: string) =>
  get<ProductDetailResponse>(`/products/${skuId}`)

// 收藏商品
export const addFavorite = (skuId: string) =>
  post<void>('/user/collect', { type: 'product', id: skuId })

export const removeFavorite = (skuId: string) =>
  del<void>('/user/collect', { data: { type: 'product', id: skuId } })

// 评价列表
export interface Review {
  review_id: string
  user: {
    user_id: string
    nickname: string
    avatar: string
  }
  rating: number
  content: string
  images: string[]
  sku_info: string
  like_count: number
  created_at: string
  seller_reply?: string
}

export interface ReviewListResponse {
  reviews: Review[]
  summary: {
    total: number
    rating: number
    distribution: Record<string, number>
  }
  pagination: {
    page: number
    page_size: number
    total: number
  }
}

export const getProductReviews = (skuId: string, params?: { page?: number; filter?: string }) =>
  get<ReviewListResponse>(`/products/${skuId}/reviews`, { params })