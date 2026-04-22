import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Button, Empty, Tag } from 'antd-mobile'

interface GroupProduct {
  id: string
  sku_id: string
  title: string
  image: string
  price: number
  originalPrice: number
  groupSize: number
  joined: number
  left: number
  endTime: number
}

interface MyGroup {
  id: string
  product: GroupProduct
  status: 'PENDING' | 'SUCCESS' | 'FAILED'
  leftTime: number
}

const mockProducts: GroupProduct[] = [
  {
    id: '1',
    sku_id: '1',
    title: 'iPhone 15 128GB 蓝色',
    image: 'https://picsum.photos/200?random=10',
    price: 499900,
    originalPrice: 599900,
    groupSize: 2,
    joined: 156,
    left: 44,
    endTime: Date.now() + 12 * 60 * 60 * 1000,
  },
  {
    id: '2',
    sku_id: '2',
    title: 'AirPods Pro 2 代',
    image: 'https://picsum.photos/200?random=11',
    price: 149900,
    originalPrice: 189900,
    groupSize: 3,
    joined: 89,
    left: 11,
    endTime: Date.now() + 6 * 60 * 60 * 1000,
  },
]

export default function GroupBuy() {
  const navigate = useNavigate()
  const [products, setProducts] = useState<GroupProduct[]>([])
  const [myGroups, setMyGroups] = useState<MyGroup[]>([])

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    setProducts(mockProducts)
    setMyGroups([
      {
        id: 'g1',
        product: mockProducts[0],
        status: 'PENDING',
        leftTime: 3600000,
      },
    ])
  }

  const handleJoinGroup = (product: GroupProduct) => {
    navigate(`/product/${product.sku_id}?activity=groupbuy`)
  }

  const handleShare = (product: GroupProduct) => {
    console.log('分享拼团:', product)
  }

  const formatTime = (ms: number) => {
    const hours = Math.floor(ms / (60 * 60 * 1000))
    const minutes = Math.floor((ms % (60 * 60 * 1000)) / (60 * 1000))
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}`
  }

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      {/* 我的拼团 */}
      {myGroups.length > 0 && (
        <div className="bg-primary-500 text-white p-4">
          <div className="flex items-center gap-2 mb-3">
            <span className="text-xl">👥</span>
            <span className="font-bold">我的拼团</span>
          </div>
          {myGroups.map((group) => (
            <Card key={group.id} className="bg-white bg-opacity-90">
              <div className="flex gap-3">
                <img
                  src={group.product.image}
                  alt={group.product.title}
                  className="w-16 h-16 rounded object-cover"
                />
                <div className="flex-1">
                  <div className="text-sm line-clamp-1">{group.product.title}</div>
                  <div className="text-xs text-gray-500 mt-1">
                    剩余 {formatTime(group.leftTime)} 自动解散
                  </div>
                  <div className="flex gap-2 mt-2">
                    <Button size="small" onClick={() => handleShare(group.product)}>
                      分享邀请
                    </Button>
                    {group.status === 'SUCCESS' && (
                      <Tag color="green">拼团成功</Tag>
                    )}
                  </div>
                </div>
              </div>
            </Card>
          ))}
        </div>
      )}

      {/* 拼团商品 */}
      <div className="p-2">
        <div className="font-bold mb-3">更多拼团</div>
        {products.length === 0 ? (
          <Empty description="暂无拼团活动" />
        ) : (
          <div className="space-y-3">
            {products.map((product) => (
              <Card key={product.id}>
                <div className="flex gap-3">
                  <img
                    src={product.image}
                    alt={product.title}
                    className="w-28 h-28 rounded object-cover bg-gray-100"
                  />
                  <div className="flex-1">
                    <div className="text-sm line-clamp-2">{product.title}</div>
                    <div className="flex items-baseline gap-2 mt-1">
                      <span className="text-xl text-red-500 font-bold">
                        ¥{(product.price / 100).toFixed(2)}
                      </span>
                      <span className="text-xs text-gray-400 line-through">
                        ¥{(product.originalPrice / 100).toFixed(2)}
                      </span>
                    </div>
                    <div className="flex items-center gap-2 mt-1">
                      <Tag color="danger">{product.groupSize}人拼团</Tag>
                      <span className="text-xs text-gray-400">
                        已拼 {product.joined} 件
                      </span>
                    </div>
                    <div className="flex items-center justify-between mt-2">
                      <span className="text-xs text-gray-500">
                        {formatTime(product.endTime - Date.now())} 后结束
                      </span>
                      <Button
                        size="small"
                        color="primary"
                        onClick={() => handleJoinGroup(product)}
                      >
                        去拼团
                      </Button>
                    </div>
                  </div>
                </div>

                {/* 已参与用户 */}
                <div className="mt-3 pt-3 border-t border-gray-100">
                  <div className="flex items-center gap-2 mb-2">
                    <span>👤</span>
                    <span className="text-sm text-gray-500">正在拼团</span>
                  </div>
                  <div className="flex -space-x-2">
                    {Array.from({ length: Math.min(5, product.groupSize) }).map((_, i) => (
                      <div
                        key={i}
                        className="w-8 h-8 rounded-full bg-gray-200 border-2 border-white flex items-center justify-center text-xs"
                      >
                        {i + 1}
                      </div>
                    ))}
                    <div className="w-8 h-8 rounded-full bg-primary-500 border-2 border-white flex items-center justify-center text-xs text-white">
                      +{product.left}
                    </div>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}