import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Swiper, Toast, Button, Tag } from 'antd-mobile'
import { HeartOutlined, ShoppingCartOutlined, StarFilled } from '@ant-design/icons'
import { getProductDetail, getProductReviews, addFavorite, removeFavorite } from '@/api/product'
import { getSimilarRecommendations } from '@/api/recommendation'
import { useCartStore } from '@/stores/cartStore'
import type { ProductDetailResponse, Review } from '@/api/product'
import type { RecommendationItem } from '@/api/recommendation'

export default function ProductDetail() {
  const { skuId } = useParams<{ skuId: string }>()
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [product, setProduct] = useState<ProductDetailResponse | null>(null)
  const [reviews, setReviews] = useState<Review[]>([])
  const [selectedSku, setSelectedSku] = useState<{ sku_id: string; attributes: Record<string, string> } | null>(null)
  const [quantity] = useState(1)
  const [activeTab, setActiveTab] = useState<'detail' | 'review'>('detail')
  const [similarScene, setSimilarScene] = useState<'view' | 'purchase'>('view')
  const [similarProducts, setSimilarProducts] = useState<RecommendationItem[]>([])

  const { addItem } = useCartStore()

  useEffect(() => {
    if (skuId) {
      loadData()
    }
  }, [skuId])

  useEffect(() => {
    const fetchSimilar = async () => {
      try {
        const res = await getSimilarRecommendations(skuId!, { scene: similarScene, limit: 6 })
        setSimilarProducts(res.items || [])
      } catch (err) {
        console.error('Failed to load similar products:', err)
      }
    }
    if (skuId) {
      fetchSimilar()
    }
  }, [skuId, similarScene])

  const loadData = async () => {
    if (!skuId) return
    try {
      const [productRes, reviewsRes] = await Promise.all([
        getProductDetail(skuId),
        getProductReviews(skuId, { page: 1 }),
      ])
      setProduct(productRes)
      setReviews(reviewsRes.reviews)
      // 默认选中第一个 SKU
      if (productRes.skus?.length > 0) {
        const firstSku = productRes.skus.find((s) => s.stock > 0) || productRes.skus[0]
        setSelectedSku({ sku_id: firstSku.sku_id, attributes: firstSku.attributes })
      }
    } catch (error) {
      Toast.show('加载失败')
    } finally {
      setLoading(false)
    }
  }

  const handleToggleFavorite = async () => {
    if (!product) return
    try {
      if (product.is_favorite) {
        await removeFavorite(product.sku_id)
        setProduct({ ...product, is_favorite: false, favorite_count: product.favorite_count - 1 })
        Toast.show('已取消收藏')
      } else {
        await addFavorite(product.sku_id)
        setProduct({ ...product, is_favorite: true, favorite_count: product.favorite_count + 1 })
        Toast.show('已添加收藏')
      }
    } catch (error) {
      Toast.show('操作失败')
    }
  }

  const handleAddToCart = () => {
    if (!product || !selectedSku) return
    addItem({
      sku_id: selectedSku.sku_id,
      title: product.title,
      image: product.images[0],
      attributes: Object.values(selectedSku.attributes),
      price: product.price,
      quantity,
      stock: product.stock,
      shop_id: product.shop.shop_id,
      shop_name: product.shop.name,
    })
    Toast.show('已加入购物车')
  }

  const handleBuyNow = () => {
    handleAddToCart()
    navigate('/checkout')
  }

  if (loading || !product) {
    return (
      <div className="p-4 text-center text-gray-500">
        加载中...
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      {/* 商品图片 */}
      <Swiper autoplay loop className="bg-white">
        {product.images.map((img, index) => (
          <Swiper.Item key={index}>
            <img src={img} alt="" className="w-full aspect-square object-cover" />
          </Swiper.Item>
        ))}
      </Swiper>

      {/* 价格信息 */}
      <div className="bg-white p-4">
        <div className="flex items-baseline">
          <span className="text-primary-500 text-2xl font-bold">
            ¥{(product.price / 100).toFixed(2)}
          </span>
          {product.original_price > product.price && (
            <span className="text-gray-400 text-sm line-through ml-2">
              ¥{(product.original_price / 100).toFixed(2)}
            </span>
          )}
        </div>
        <div className="mt-2">
          <span className="text-gray-600">{product.title}</span>
        </div>
        <div className="flex items-center gap-4 mt-2 text-sm text-gray-500">
          <span>销量 {product.sales}</span>
          <span>收藏 {product.favorite_count}</span>
          <span>库存 {product.stock}</span>
        </div>
      </div>

      {/* SKU 选择 */}
      {product.attributes?.length > 0 && (
        <div className="bg-white mt-2 p-4">
          <div className="font-bold mb-2">选择规格</div>
          {product.attributes.map((attr) => (
            <div key={attr.name} className="mb-2">
              <div className="text-sm text-gray-500 mb-1">{attr.name}</div>
              <div className="flex flex-wrap gap-2">
                {attr.values.map((val) => {
                  const isSelected = selectedSku?.attributes[attr.name] === val.value
                  return (
                    <Tag
                      key={val.value}
                      color={isSelected ? 'primary' : 'default'}
                      onClick={() => {
                        setSelectedSku((prev) =>
                          prev ? { ...prev, attributes: { ...prev.attributes, [attr.name]: val.value } } : null
                        )
                      }}
                    >
                      {val.value}
                    </Tag>
                  )
                })}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* 看了又看 / 买了还买 */}
      <div className="mt-2 bg-white">
        {/* Tab 切换 */}
        <div className="flex border-b">
          <div
            className={`flex-1 py-3 text-center ${
              similarScene === 'view'
                ? 'text-primary-500 border-b-2 border-primary-500'
                : 'text-gray-500'
            }`}
            onClick={() => setSimilarScene('view')}
          >
            看了又看
          </div>
          <div
            className={`flex-1 py-3 text-center ${
              similarScene === 'purchase'
                ? 'text-primary-500 border-b-2 border-primary-500'
                : 'text-gray-500'
            }`}
            onClick={() => setSimilarScene('purchase')}
          >
            买了还买
          </div>
        </div>

        {/* 推荐列表 */}
        {similarProducts.length > 0 ? (
          <div className="p-3">
            <div className="flex gap-3 overflow-x-auto hide-scrollbar">
              {similarProducts.map((item) => (
                <div
                  key={item.sku_id}
                  className="flex-shrink-0 w-32"
                  onClick={() => navigate(`/product/${item.sku_id}`)}
                >
                  <img
                    src={item.image}
                    alt={item.title}
                    className="w-32 h-32 rounded object-cover"
                  />
                  <div className="mt-1 text-sm line-clamp-2">{item.title}</div>
                  <div className="text-primary-500 font-bold">
                    ¥{(item.price / 100).toFixed(2)}
                  </div>
                </div>
              ))}
            </div>
          </div>
        ) : (
          <div className="p-4 text-center text-gray-400 text-sm">暂无推荐</div>
        )}
      </div>

      {/* 评价预览 */}
      <div
        className="bg-white mt-2 p-4"
        onClick={() => setActiveTab('review')}
      >
        <div className="flex items-center justify-between">
          <span className="font-bold">评价</span>
          <span className="text-primary-500 text-sm">
            查看全部 &gt;
          </span>
        </div>
        {reviews.length > 0 && (
          <div className="mt-2">
            <div className="flex items-center text-sm text-gray-500">
              <StarFilled className="text-yellow-500 mr-1" />
              <span>{reviews[0].rating}分</span>
              <span className="mx-2">|</span>
              <span>{reviews[0].content.substring(0, 50)}...</span>
            </div>
          </div>
        )}
      </div>

      {/* 商品详情 Tab */}
      <div className="mt-2 bg-white">
        <div className="flex border-b">
          <div
            className={`flex-1 py-3 text-center ${
              activeTab === 'detail' ? 'text-primary-500 border-b-2 border-primary-500' : 'text-gray-500'
            }`}
            onClick={() => setActiveTab('detail')}
          >
            商品详情
          </div>
          <div
            className={`flex-1 py-3 text-center ${
              activeTab === 'review' ? 'text-primary-500 border-b-2 border-primary-500' : 'text-gray-500'
            }`}
            onClick={() => setActiveTab('review')}
          >
            评价 ({product.comment_count})
          </div>
        </div>

        {activeTab === 'detail' ? (
          <div className="p-4">
            <div
              className="prose max-w-none"
              dangerouslySetInnerHTML={{ __html: product.details.description || '<p>暂无详情</p>' }}
            />
          </div>
        ) : (
          <div className="p-4">
            {reviews.slice(0, 3).map((review) => (
              <div key={review.review_id} className="mb-4 border-b border-gray-100 pb-4 last:border-0">
                <div className="flex items-center gap-2">
                  <span className="font-medium">{review.user.nickname}</span>
                  <div className="flex text-yellow-500 text-xs">
                    {Array.from({ length: review.rating }).map((_, i) => (
                      <StarFilled key={i} />
                    ))}
                  </div>
                </div>
                <div className="mt-1 text-sm text-gray-600">{review.content}</div>
                {review.images?.length > 0 && (
                  <div className="flex gap-1 mt-2">
                    {review.images.map((img, i) => (
                      <img key={i} src={img} alt="" className="w-16 h-16 object-cover rounded" />
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 底部操作栏 */}
      <div className="fixed bottom-0 left-0 right-0 bg-white border-t flex items-center justify-around p-3 safe-area-bottom">
        <div className="flex items-center gap-4">
          <div className="flex flex-col items-center" onClick={handleToggleFavorite}>
            <HeartOutlined className={`text-xl ${product.is_favorite ? 'text-red-500' : 'text-gray-400'}`} />
            <span className="text-xs text-gray-500">收藏</span>
          </div>
          <div className="flex flex-col items-center" onClick={() => navigate('/cart')}>
            <ShoppingCartOutlined className="text-xl text-gray-600" />
            <span className="text-xs text-gray-500">购物车</span>
          </div>
        </div>
        <div className="flex gap-2">
          <Button color="warning" onClick={handleAddToCart}>
            加入购物车
          </Button>
          <Button color="primary" onClick={handleBuyNow}>
            立即购买
          </Button>
        </div>
      </div>
    </div>
  )
}