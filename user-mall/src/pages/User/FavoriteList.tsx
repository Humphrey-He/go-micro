import { useState, useEffect } from 'react'
import { Card, Empty, Skeleton, Button } from 'antd-mobile'
import { StarOutline } from 'antd-mobile-icons'
import type { Product } from '@/api/product'

const mockFavorites: Product[] = []

export default function FavoriteList() {
  const [favorites, setFavorites] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadFavorites()
  }, [])

  const loadFavorites = async () => {
    try {
      setLoading(true)
      setFavorites(mockFavorites)
    } finally {
      setLoading(false)
    }
  }

  const handleRemove = (skuId: string) => {
    setFavorites((prev) => prev.filter((p) => p.sku_id !== skuId))
  }

  if (loading) {
    return (
      <div className="p-4 space-y-4">
        <Skeleton animated />
        <Skeleton animated />
      </div>
    )
  }

  if (favorites.length === 0) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Empty
          image={<StarOutline style={{ fontSize: 48 }} />}
          description="暂无收藏"
        />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      <div className="p-2 space-y-2">
        {favorites.map((product) => (
          <Card key={product.sku_id} className="overflow-hidden">
            <div className="flex gap-3 p-3">
              <img
                src={product.images[0]}
                alt={product.title}
                className="w-24 h-24 rounded object-cover bg-gray-100"
              />
              <div className="flex-1 min-w-0">
                <div className="text-sm line-clamp-2">{product.title}</div>
                <div className="text-primary-500 font-bold mt-1">
                  ¥{(product.price / 100).toFixed(2)}
                </div>
                <Button
                  size="small"
                  className="mt-2"
                  onClick={() => handleRemove(product.sku_id)}
                >
                  取消收藏
                </Button>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  )
}