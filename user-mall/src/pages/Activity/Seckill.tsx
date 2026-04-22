import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Badge } from 'antd-mobile'

interface SeckillProduct {
  id: string
  title: string
  image: string
  price: number
  originalPrice: number
  stock: number
  sold: number
  startTime: string
}

const mockSeckillProducts: SeckillProduct[] = [
  {
    id: '1',
    title: 'iPhone 15 Pro 128GB',
    image: 'https://picsum.photos/200?random=1',
    price: 599900,
    originalPrice: 799900,
    stock: 10,
    sold: 90,
    startTime: '10:00',
  },
  {
    id: '2',
    title: '小米手环8',
    image: 'https://picsum.photos/200?random=2',
    price: 19900,
    originalPrice: 29900,
    stock: 50,
    sold: 150,
    startTime: '14:00',
  },
  {
    id: '3',
    title: 'AirPods Pro 2',
    image: 'https://picsum.photos/200?random=3',
    price: 149900,
    originalPrice: 189900,
    stock: 30,
    sold: 70,
    startTime: '20:00',
  },
]

export default function Seckill() {
  const navigate = useNavigate()
  const [products, setProducts] = useState<SeckillProduct[]>([])
  const [loading, setLoading] = useState(true)
  const [countdown, setCountdown] = useState('')

  useEffect(() => {
    loadData()
  }, [])

  useEffect(() => {
    const timer = setInterval(() => {
      const now = new Date()
      const hours = 23 - now.getHours()
      const minutes = 59 - now.getMinutes()
      const seconds = 59 - now.getSeconds()
      setCountdown(
        `${hours.toString().padStart(2, '0')}:${minutes
          .toString()
          .padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
      )
    }, 1000)
    return () => clearInterval(timer)
  }, [])

  const loadData = async () => {
    try {
      setLoading(true)
      setProducts(mockSeckillProducts)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* 秒杀头部 */}
      <div className="bg-gradient-to-r from-red-500 to-red-400 text-white p-4">
        <div className="flex items-center gap-2 mb-2">
          <span className="text-2xl">⚡</span>
          <span className="font-bold text-lg">限时秒杀</span>
        </div>
        <div className="text-sm opacity-90">距离结束还有</div>
        <div className="text-2xl font-bold mt-1 font-mono">{countdown}</div>
      </div>

      {/* 秒杀商品 */}
      {loading ? (
        <div className="p-4 text-center text-gray-500">加载中...</div>
      ) : (
        <div className="p-2 space-y-3">
          {products.map((product) => (
            <Card
              key={product.id}
              className="overflow-hidden"
              onClick={() => navigate(`/product/${product.id}`)}
            >
              <div className="flex gap-3 p-3">
                <img
                  src={product.image}
                  alt={product.title}
                  className="w-28 h-28 rounded object-cover bg-gray-100"
                />
                <div className="flex-1 min-w-0">
                  <div className="text-sm line-clamp-2">{product.title}</div>
                  <div className="flex items-baseline gap-2 mt-2">
                    <span className="text-xl text-red-500 font-bold">
                      ¥{(product.price / 100).toFixed(2)}
                    </span>
                    <span className="text-sm text-gray-400 line-through">
                      ¥{(product.originalPrice / 100).toFixed(2)}
                    </span>
                  </div>
                  <div className="flex items-center justify-between mt-2">
                    <span className="text-xs text-gray-500">
                      {product.startTime} 开抢
                    </span>
                    <Badge
                      content={`仅剩 ${product.stock}`}
                      color="var(--adm-color-danger)"
                    />
                  </div>
                  {/* 进度条 */}
                  <div className="mt-2 h-2 bg-gray-200 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-red-500"
                      style={{
                        width: `${(product.sold / (product.sold + product.stock)) * 100}%`,
                      }}
                    />
                  </div>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}