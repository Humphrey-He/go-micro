import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Button, Empty, SwipeAction } from 'antd-mobile'
import { DeleteOutline } from 'antd-mobile-icons'

interface FootprintItem {
  id: string
  sku_id: string
  title: string
  image: string
  price: number
  browseTime: string
}

export default function Footprint() {
  const navigate = useNavigate()
  const [footprints, setFootprints] = useState<FootprintItem[]>([])

  useEffect(() => {
    const saved = localStorage.getItem('footprints')
    if (saved) {
      setFootprints(JSON.parse(saved))
    }
  }, [])

  const handleDelete = (id: string) => {
    const newFootprints = footprints.filter((f) => f.id !== id)
    setFootprints(newFootprints)
    localStorage.setItem('footprints', JSON.stringify(newFootprints))
  }

  const handleClearAll = () => {
    setFootprints([])
    localStorage.removeItem('footprints')
  }

  const handleAddToCart = (item: FootprintItem) => {
    // 添加到购物车逻辑
    console.log('添加到购物车:', item)
  }

  if (footprints.length === 0) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Empty
          image="https://gw.alipayobjects.com/zos/antfincdn/Hr3NSeY%24j/production-illustration.svg"
          description="暂无浏览足迹"
        />
      </div>
    )
  }

  // 按日期分组
  const groupedByDate: Record<string, FootprintItem[]> = {}
  footprints.forEach((item) => {
    const date = item.browseTime.split('T')[0]
    if (!groupedByDate[date]) {
      groupedByDate[date] = []
    }
    groupedByDate[date].push(item)
  })

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      {/* 操作栏 */}
      <div className="bg-white p-3 flex items-center justify-between">
        <span className="text-sm text-gray-500">
          共 {footprints.length} 件商品
        </span>
        <Button size="small" color="danger" onClick={handleClearAll}>
          清空足迹
        </Button>
      </div>

      {/* 分组列表 */}
      {Object.entries(groupedByDate).map(([date, items]) => (
        <div key={date} className="mt-2">
          <div className="bg-gray-100 px-4 py-2 text-sm text-gray-500">
            {date === new Date().toISOString().split('T')[0]
              ? '今天'
              : date === new Date(Date.now() - 86400000).toISOString().split('T')[0]
              ? '昨天'
              : date}
          </div>
          <div className="p-2 space-y-2">
            {items.map((item) => (
              <SwipeAction
                key={item.id}
                rightActions={[
                  {
                    key: 'delete',
                    text: <DeleteOutline />,
                    color: 'danger',
                    onClick: () => handleDelete(item.id),
                  },
                ]}
              >
                <Card
                  className="cursor-pointer"
                  onClick={() => navigate(`/product/${item.sku_id}`)}
                >
                  <div className="flex gap-3">
                    <img
                      src={item.image}
                      alt={item.title}
                      className="w-20 h-20 rounded object-cover bg-gray-100"
                    />
                    <div className="flex-1 min-w-0">
                      <div className="text-sm line-clamp-2">{item.title}</div>
                      <div className="text-primary-500 font-bold mt-1">
                        ¥{(item.price / 100).toFixed(2)}
                      </div>
                      <div className="flex gap-2 mt-2">
                        <Button
                          size="small"
                          onClick={(e) => {
                            e.stopPropagation()
                            handleAddToCart(item)
                          }}
                        >
                          加入购物车
                        </Button>
                      </div>
                    </div>
                  </div>
                </Card>
              </SwipeAction>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}