import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Badge } from 'antd'
import type { Product } from '@/api/product'

interface Props {
  product: Product
  viewMode?: 'waterfall' | 'grid'
}

export default function ProductCard({ product, viewMode = 'waterfall' }: Props) {
  const { t } = useTranslation()
  const navigate = useNavigate()

  // 模拟不同图片高度，营造瀑布流效果
  const getImageHeight = (index: number) => {
    const heights = ['aspect-[3/4]', 'aspect-[4/3]', 'aspect-square', 'aspect-[3/5]']
    return heights[index % heights.length]
  }

  const formatSales = (sales: number) => {
    if (sales >= 10000) {
      return `${(sales / 10000).toFixed(1)}万+`
    }
    if (sales >= 1000) {
      return `${(sales / 1000).toFixed(1)}k+`
    }
    return sales.toString()
  }

  if (viewMode === 'waterfall') {
    const index = product.sku_id.charCodeAt(product.sku_id.length - 1) % 4
    const isLongCard = index === 0 || index === 3

    return (
      <div
        className="bg-white rounded-xl overflow-hidden cursor-pointer shadow-sm hover:shadow-md transition-shadow duration-200"
        onClick={() => navigate(`/product/${product.sku_id}`)}
      >
        {/* 图片区域 - 自适应高度 */}
        <div className={`relative ${getImageHeight(product.sku_id.charCodeAt(0))}`}>
          <img
            src={product.images[0]}
            alt={product.title}
            className="w-full h-full object-cover"
            loading="lazy"
          />
          {/* 限时优惠标签 */}
          {product.original_price > product.price && (
            <div className="absolute bottom-0 left-0 bg-gradient-to-r from-primary-500 to-primary-400 text-white text-xs px-2 py-0.5">
              {t('product.tags.discount')}
            </div>
          )}
          {/* 角标 */}
          <div className="absolute top-1.5 left-1.5 flex flex-col gap-1">
            {product.tags?.includes('hot') && (
              <Badge
                count={t('product.tags.hot')}
                style={{ backgroundColor: '#ff4d4f', fontSize: '10px' }}
              />
            )}
            {product.tags?.includes('new') && (
              <Badge
                count={t('product.tags.new')}
                style={{ backgroundColor: '#52c41a', fontSize: '10px' }}
              />
            )}
          </div>
        </div>

        {/* 商品信息 */}
        <div className="p-2">
          {/* 标题 */}
          <div className="text-sm line-clamp-2 leading-snug">
            {product.title}
          </div>

          {/* 副标题 */}
          {product.subtitle && (
            <div className="text-xs text-gray-400 mt-1 line-clamp-1">
              {product.subtitle}
            </div>
          )}

          {/* 价格区域 */}
          <div className="mt-2 flex items-baseline justify-between">
            <div className="flex items-baseline">
              <span className="text-primary-500 font-bold text-lg">
                ¥{(product.price / 100).toFixed(0)}
              </span>
              {product.original_price > product.price && (
                <span className="text-gray-300 text-xs line-through ml-1">
                  ¥{(product.original_price / 100).toFixed(0)}
                </span>
              )}
            </div>
            {product.original_price > product.price && (
              <div className="text-xs text-primary-500 bg-primary-50 px-1 rounded">
                {Math.round((1 - product.price / product.original_price) * 100)}折
              </div>
            )}
          </div>

          {/* 店铺信息 & 销量 */}
          <div className="mt-2 flex items-center justify-between">
            <div className="flex items-center gap-1">
              <img
                src={product.shop?.logo || 'https://via.placeholder.com/16'}
                alt={product.shop?.name || t('product.officialShop')}
                className="w-4 h-4 rounded-full object-cover bg-gray-100"
              />
              <span className="text-xs text-gray-500 truncate max-w-[60px]">
                {product.shop?.name || t('product.officialShop')}
              </span>
            </div>
            <span className="text-xs text-gray-400">
              {formatSales(product.sales)}{t('product.sales')}
            </span>
          </div>

          {/* 评价 */}
          {isLongCard && (
            <div className="mt-2 flex items-center gap-1 text-xs">
              <span className="text-yellow-500">★★★★★</span>
              <span className="text-gray-500">{product.rating.toFixed(1)}</span>
              <span className="text-gray-300">|</span>
              <span className="text-gray-400">{product.comment_count}{t('product.rating')}</span>
            </div>
          )}
        </div>
      </div>
    )
  }

  // 网格视图 - 保持原有风格
  return (
    <div
      className="bg-white rounded-lg overflow-hidden cursor-pointer"
      onClick={() => navigate(`/product/${product.sku_id}`)}
    >
      <div className="relative">
        <img
          src={product.images[0]}
          alt={product.title}
          className="w-full aspect-square object-cover bg-gray-100"
        />
        {(product.tags?.includes('hot')) && (
          <Badge
            count={t('product.tags.hot')}
            style={{ backgroundColor: '#ff4d4f' }}
            className="absolute top-1 left-1"
          />
        )}
        {(product.tags?.includes('new')) && (
          <Badge
            count={t('product.tags.new')}
            style={{ backgroundColor: '#52c41a' }}
            className="absolute top-1 left-1"
          />
        )}
      </div>
      <div className="p-2">
        <div className="text-sm line-clamp-2 h-10">{product.title}</div>
        <div className="flex items-baseline mt-1">
          <span className="text-primary-500 font-bold text-lg">
            ¥{(product.price / 100).toFixed(2)}
          </span>
          {product.original_price > product.price && (
            <span className="text-gray-400 text-xs line-through ml-1">
              ¥{(product.original_price / 100).toFixed(2)}
            </span>
          )}
        </div>
        <div className="flex items-center justify-between mt-1">
          <span className="text-xs text-gray-400">{product.sales}{t('product.sales')}</span>
          <div className="flex items-center">
            <span className="text-yellow-500 text-xs">★</span>
            <span className="text-xs text-gray-500 ml-0.5">{product.rating}</span>
          </div>
        </div>
      </div>
    </div>
  )
}