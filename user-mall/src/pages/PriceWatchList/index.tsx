import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Skeleton, Switch, Empty, PullToRefresh } from 'antd-mobile'
import { getPriceWatchList, updatePriceWatch, cancelPriceWatch, type PriceWatchItem } from '@/api/priceWatch'

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
          className="mt-4 text-[#00C853]"
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
                  <span className="text-[#00C853] font-bold">
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
