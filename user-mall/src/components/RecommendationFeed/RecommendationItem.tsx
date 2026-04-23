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
