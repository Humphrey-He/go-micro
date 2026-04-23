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
            <Skeleton key={i} animated style={{ height: layout === 'grid' ? '180px' : '80px' }} />
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
