import { useState, useEffect } from 'react'
import { Card, Tag, Avatar, Image, Rate, Empty } from 'antd-mobile'
import { useParams } from 'react-router-dom'

interface MockReview {
  review_id: string
  user: { user_id: string; nickname: string; avatar: string }
  rating: number
  content: string
  images: string[]
  sku_info?: string
  like_count: number
  created_at: string
  seller_reply?: string
  is_anonymous: boolean
  is_append?: boolean
}

export default function ProductReviews() {
  const { skuId } = useParams<{ skuId: string }>()
  const [reviews, setReviews] = useState<MockReview[]>([])
  const [filter, setFilter] = useState<'all' | 'with_images' | '追加'>('all')

  useEffect(() => {
    loadReviews()
  }, [skuId])

  const loadReviews = async () => {
    setReviews([
      {
        review_id: '1',
        user: { user_id: '1', nickname: '用户1234', avatar: '' },
        rating: 5,
        content: '非常满意的一次购物体验！包装很好，物流也很快，商品质量超出预期，会再次购买。',
        images: ['https://picsum.photos/100?random=1', 'https://picsum.photos/100?random=2'],
        created_at: '2024-01-15',
        is_anonymous: false,
        sku_info: '颜色: 黑色; 内存: 256GB',
        like_count: 0,
      },
      {
        review_id: '2',
        user: { user_id: '2', nickname: '购物达人', avatar: '' },
        rating: 4,
        content: '整体不错，就是发货稍微慢了点。',
        images: [],
        created_at: '2024-01-14',
        is_anonymous: false,
        like_count: 0,
      },
      {
        review_id: '3',
        user: { user_id: '3', nickname: '匿名用户', avatar: '' },
        rating: 5,
        content: '追评：使用了一周，感觉非常好！推荐购买！',
        images: ['https://picsum.photos/100?random=3'],
        created_at: '2024-01-10',
        is_anonymous: true,
        is_append: true,
        like_count: 0,
      },
    ])
  }

  const filteredReviews = reviews.filter((r) => {
    if (filter === 'with_images') return r.images && r.images.length > 0
    if (filter === '追加') return r.is_append
    return true
  })

  const ratingCounts = {
    5: reviews.filter((r) => r.rating === 5).length,
    4: reviews.filter((r) => r.rating === 4).length,
    3: reviews.filter((r) => r.rating === 3).length,
    2: reviews.filter((r) => r.rating === 2).length,
    1: reviews.filter((r) => r.rating === 1).length,
  }

  const averageRating = reviews.length
    ? (reviews.reduce((sum, r) => sum + r.rating, 0) / reviews.length).toFixed(1)
    : '0.0'

  return (
    <div className="min-h-screen bg-gray-50">
      {/* 评分概览 */}
      <Card className="m-2">
        <div className="flex items-center">
          <div className="text-center w-24">
            <div className="text-3xl font-bold text-orange-500">{averageRating}</div>
            <div className="text-sm text-gray-500 mt-1">综合评分</div>
          </div>
          <div className="flex-1 pl-4 border-l border-gray-200">
            {[5, 4, 3, 2, 1].map((star) => (
              <div key={star} className="flex items-center text-sm">
                <span className="w-8">{star}星</span>
                <div className="flex-1 mx-2 h-2 bg-gray-100 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-orange-500"
                    style={{
                      width: `${reviews.length ? (ratingCounts[star as keyof typeof ratingCounts] / reviews.length) * 100 : 0}%`,
                    }}
                  />
                </div>
                <span className="text-gray-400 w-8">{ratingCounts[star as keyof typeof ratingCounts]}</span>
              </div>
            ))}
          </div>
        </div>
      </Card>

      {/* 筛选 Tab */}
      <div className="bg-white flex gap-2 p-2 overflow-x-auto">
        <Tag
          color={filter === 'all' ? 'primary' : 'gray'}
          onClick={() => setFilter('all')}
        >
          全部 ({reviews.length})
        </Tag>
        <Tag
          color={filter === 'with_images' ? 'primary' : 'gray'}
          onClick={() => setFilter('with_images')}
        >
          带图评价 ({reviews.filter((r) => r.images?.length > 0).length})
        </Tag>
        <Tag
          color={filter === '追加' ? 'primary' : 'gray'}
          onClick={() => setFilter('追加')}
        >
          追评 ({reviews.filter((r) => r.is_append).length})
        </Tag>
      </div>

      {/* 评价列表 */}
      <div className="p-2 space-y-3">
        {filteredReviews.length === 0 ? (
          <Empty description="暂无评价" />
        ) : (
          filteredReviews.map((review) => (
            <Card key={review.review_id}>
              <div className="flex items-start gap-3">
                <Avatar src={review.user.avatar || ''} className="bg-gray-200" />
                <div className="flex-1">
                  <div className="flex items-center justify-between">
                    <span className="font-medium">
                      {review.is_anonymous ? '匿名用户' : review.user.nickname}
                    </span>
                    <span className="text-xs text-gray-400">{review.created_at}</span>
                  </div>
                  <Rate readOnly value={review.rating} className="text-xs mt-1" />
                  {review.sku_info && (
                    <div className="text-xs text-gray-400 mt-1">{review.sku_info}</div>
                  )}
                  <div className="mt-2 text-sm">{review.content}</div>
                  {review.images && review.images.length > 0 && (
                    <div className="flex gap-1 mt-2">
                      {review.images.map((img, idx) => (
                        <Image
                          key={idx}
                          src={img}
                          className="w-20 h-20 object-cover rounded"
                        />
                      ))}
                    </div>
                  )}
                  {review.is_append && (
                    <div className="mt-2 text-sm text-orange-500 bg-orange-50 p-2 rounded">
                      追评: {review.content}
                    </div>
                  )}
                </div>
              </div>
            </Card>
          ))
        )}
      </div>
    </div>
  )
}