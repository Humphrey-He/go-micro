import { useNavigate } from 'react-router-dom'
import { Swiper } from 'antd-mobile'
import { useState } from 'react'
import RecommendationFeed from '@/components/RecommendationFeed'
import { getHomeRecommendations } from '@/api/recommendation'

// 简化商品类型
interface SimpleProduct {
  sku_id: string
  title: string
  price: number
  original_price: number
  images: string[]
  sales: number
  rating: number
}

// 模拟数据
const banners = [
  { id: 1, image: 'https://picsum.photos/750/300?random=1', link: '' },
  { id: 2, image: 'https://picsum.photos/750/300?random=2', link: '' },
  { id: 3, image: 'https://picsum.photos/750/300?random=3', link: '' },
]

const categories = [
  { id: '1', name: '服装', icon: '👔' },
  { id: '2', name: '鞋靴', icon: '👟' },
  { id: '3', name: '数码', icon: '📱' },
  { id: '4', name: '美妆', icon: '💄' },
  { id: '5', name: '食品', icon: '🍎' },
  { id: '6', name: '家电', icon: '📺' },
  { id: '7', name: '母婴', icon: '🍼' },
  { id: '8', name: '家居', icon: '🏠' },
]

const activities = [
  { id: 'seckill', name: '限时秒杀', icon: '⚡', color: 'from-red-500 to-orange-500', desc: '精选好物等你抢' },
  { id: 'coupon-center', name: '领优惠券', icon: '🎫', color: 'from-pink-500 to-purple-500', desc: '新人专享福利' },
  { id: 'groupbuy', name: '拼团', icon: '👥', color: 'from-green-500 to-teal-500', desc: '邀请好友拼团' },
  { id: 'checkin', name: '每日签到', icon: '📅', color: 'from-blue-500 to-cyan-500', desc: '签到领积分' },
]

const mockProducts: SimpleProduct[] = [
  {
    sku_id: '1',
    title: 'iPhone 15 Pro Max 256GB 深空黑',
    price: 899900,
    original_price: 999900,
    images: ['https://picsum.photos/200?random=10'],
    sales: 1000,
    rating: 4.8,
  },
  {
    sku_id: '2',
    title: 'AirPods Pro 2 无线蓝牙耳机',
    price: 179900,
    original_price: 199900,
    images: ['https://picsum.photos/200?random=11'],
    sales: 500,
    rating: 4.9,
  },
  {
    sku_id: '3',
    title: 'MacBook Pro 14英寸 M3芯片',
    price: 1299900,
    original_price: 1499900,
    images: ['https://picsum.photos/200?random=12'],
    sales: 300,
    rating: 4.7,
  },
  {
    sku_id: '4',
    title: '小米手环8 Pro 健康运动手环',
    price: 39900,
    original_price: 49900,
    images: ['https://picsum.photos/200?random=13'],
    sales: 2000,
    rating: 4.6,
  },
]

export default function Home() {
  const navigate = useNavigate()
  const [products] = useState<SimpleProduct[]>(mockProducts)

  return (
    <div className="animate-fade-in">
      {/* Banner */}
      <Swiper autoplay loop>
        {banners.map((banner) => (
          <Swiper.Item key={banner.id}>
            <img
              src={banner.image}
              alt=""
              className="w-full h-40 object-cover"
              onClick={() => banner.link && navigate(banner.link)}
            />
          </Swiper.Item>
        ))}
      </Swiper>

      {/* 分类入口 */}
      <div className="bg-white py-3">
        <div className="grid grid-cols-4 gap-4 px-4">
          {categories.map((cat) => (
            <div
              key={cat.id}
              className="flex flex-col items-center"
              onClick={() => navigate(`/product/list?category_id=${cat.id}`)}
            >
              <span className="text-2xl">{cat.icon}</span>
              <span className="text-xs text-gray-600 mt-1">{cat.name}</span>
            </div>
          ))}
        </div>
      </div>

      {/* 活动入口 */}
      <div className="bg-white mt-2 p-4">
        <div className="grid grid-cols-4 gap-2">
          {activities.map((act) => (
            <div
              key={act.id}
              className={`bg-gradient-to-br ${act.color} rounded-lg p-3 text-white text-center`}
              onClick={() => navigate(`/${act.id}`)}
            >
              <div className="text-xl">{act.icon}</div>
              <div className="text-xs font-bold mt-1">{act.name}</div>
              <div className="text-xs opacity-80 mt-0.5 truncate">{act.desc}</div>
            </div>
          ))}
        </div>
      </div>

      {/* 为你推荐 */}
<RecommendationFeed
  api={() => getHomeRecommendations({ page: 1, page_size: 6 })}
  layout="grid"
  title="为你推荐"
  emptyText="暂无推荐，看看热销商品"
/>

      {/* 热销榜单 */}
      <div className="mt-2 bg-white p-4">
        <div className="flex items-center justify-between mb-3">
          <span className="font-bold text-lg">热销榜单</span>
          <span
            className="text-sm text-primary-500"
            onClick={() => navigate('/product/list?sort_by=sales')}
          >
            查看更多 &gt;
          </span>
        </div>
        <div className="flex gap-3 overflow-x-auto hide-scrollbar pb-2">
          {products.slice(0, 6).map((product: SimpleProduct, index: number) => (
            <div
              key={product.sku_id}
              className="flex-shrink-0 w-28"
              onClick={() => navigate(`/product/${product.sku_id}`)}
            >
              <div className="relative">
                <img
                  src={product.images[0]}
                  alt={product.title}
                  className="w-28 h-28 rounded-lg object-cover bg-gray-100"
                />
                <span className="absolute top-1 left-1 bg-orange-500 text-white text-xs px-1 rounded">
                  #{index + 1}
                </span>
              </div>
              <div className="mt-1 text-sm line-clamp-1">{product.title}</div>
              <div className="text-primary-500 font-bold">
                ¥{(product.price / 100).toFixed(2)}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}