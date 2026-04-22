import { useState, useEffect, useRef, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { SearchBar, Empty, DotLoading } from 'antd-mobile'
import { useTranslation } from 'react-i18next'
import { getProductList, type Product } from '@/api/product'
import ProductCard from '@/components/ProductCard'

export default function ProductList() {
  const { t } = useTranslation()
  const [searchParams, setSearchParams] = useSearchParams()
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)
  const [viewMode, setViewMode] = useState<'waterfall' | 'grid'>('waterfall')
  const loaderRef = useRef<HTMLDivElement>(null)

  const keyword = searchParams.get('keyword') || ''
  const categoryId = searchParams.get('category_id') || ''
  const sortBy = (searchParams.get('sort_by') as any) || ''

  const sortOptions = [
    { key: '', label: t('product.sort.comprehensive') },
    { key: 'sales', label: t('product.sort.sales') },
    { key: 'new', label: t('product.sort.new') },
    { key: 'price_asc', label: t('product.sort.priceAsc') },
    { key: 'price_desc', label: t('product.sort.priceDesc') },
  ]

  useEffect(() => {
    loadProducts(1)
  }, [keyword, categoryId, sortBy])

  const loadProducts = async (pageNum: number) => {
    try {
      setLoading(true)
      const res = await getProductList({
        keyword,
        category_id: categoryId,
        sort_by: sortBy || undefined,
        page: pageNum,
        page_size: 20,
      })
      if (pageNum === 1) {
        setProducts(res.products)
      } else {
        setProducts((prev) => [...prev, ...res.products])
      }
      setHasMore(res.pagination.page < res.pagination.total_pages)
      setPage(pageNum)
    } catch (error) {
      console.error(error)
    } finally {
      setLoading(false)
    }
  }

  const handleSort = (key: string) => {
    setSearchParams((prev) => {
      if (key) {
        prev.set('sort_by', key)
      } else {
        prev.delete('sort_by')
      }
      return prev
    })
  }

  const handleSearch = (value: string) => {
    setSearchParams((prev) => {
      if (value) {
        prev.set('keyword', value)
      } else {
        prev.delete('keyword')
      }
      return prev
    })
  }

  const handleLoadMore = () => {
    if (!loading && hasMore) {
      loadProducts(page + 1)
    }
  }

  // 无限滚动 Intersection Observer
  const handleObserver = useCallback((entries: IntersectionObserverEntry[]) => {
    const target = entries[0]
    if (target.isIntersecting && hasMore && !loading) {
      handleLoadMore()
    }
  }, [hasMore, loading, page])

  useEffect(() => {
    const observer = new IntersectionObserver(handleObserver, {
      root: null,
      rootMargin: '100px',
      threshold: 0.1,
    })
    if (loaderRef.current) {
      observer.observe(loaderRef.current)
    }
    return () => observer.disconnect()
  }, [handleObserver])

  return (
    <div className="min-h-screen bg-gray-50">
      {/* 搜索栏 */}
      <div className="sticky top-0 z-50 bg-white shadow-sm">
        <div className="px-3 pt-2">
          <SearchBar
            placeholder={t('product.searchPlaceholder')}
            defaultValue={keyword}
            onSearch={handleSearch}
            onClear={() => handleSearch('')}
          />
        </div>

        {/* 排序 + 视图切换 */}
        <div className="flex items-center justify-between px-3 py-2">
          <div className="flex gap-4">
            {sortOptions.map((opt) => (
              <span
                key={opt.key}
                className={`text-sm cursor-pointer py-1 ${
                  sortBy === opt.key ? 'text-primary-500 font-bold border-b-2 border-primary-500' : 'text-gray-600'
                }`}
                onClick={() => handleSort(opt.key)}
              >
                {opt.label}
              </span>
            ))}
          </div>
          <div className="flex items-center gap-2">
            <span
              className={`cursor-pointer p-1 ${viewMode === 'waterfall' ? 'text-primary-500' : 'text-gray-400'}`}
              onClick={() => setViewMode('waterfall')}
              title={t('product.viewMode.waterfall')}
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path d="M2 4.5A2.5 2.5 0 012 2a2.5 2.5 0 012 2v11a2.5 2.5 0 01-2 2H2a2.5 2.5 0 01-2-2v-11A2.5 2.5 0 012 4.5zM7 2a2.5 2.5 0 012.5 2.5v11A2.5 2.5 0 017 18a2.5 2.5 0 01-2.5-2.5v-11A2.5 2.5 0 017 2zM12 2a2.5 2.5 0 012.5 2.5v11A2.5 2.5 0 0112 18a2.5 2.5 0 01-2.5-2.5v-11A2.5 2.5 0 0112 2z" />
              </svg>
            </span>
            <span
              className={`cursor-pointer p-1 ${viewMode === 'grid' ? 'text-primary-500' : 'text-gray-400'}`}
              onClick={() => setViewMode('grid')}
              title={t('product.viewMode.grid')}
            >
              <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clipRule="evenodd" />
              </svg>
            </span>
          </div>
        </div>
      </div>

      {/* 商品列表 - 瀑布流视图 */}
      {products.length === 0 && !loading ? (
        <Empty description={t('common.noData')} />
      ) : viewMode === 'waterfall' ? (
        <div className="px-2 py-2 columns-2 gap-2">
          {products.map((product, index) => (
            <div
              key={product.sku_id}
              className="animate-fade-in mb-2 break-inside-avoid"
              style={{ animationDelay: `${(index % 10) * 50}ms` }}
            >
              <ProductCard product={product} viewMode="waterfall" />
            </div>
          ))}
        </div>
      ) : (
        <div className="p-2 grid grid-cols-2 gap-2">
          {products.map((product) => (
            <ProductCard key={product.sku_id} product={product} viewMode="grid" />
          ))}
        </div>
      )}

      {/* 加载指示 */}
      <div ref={loaderRef} className="h-20 flex items-center justify-center">
        {loading && <DotLoading />}
      </div>

      {!loading && !hasMore && products.length > 0 && (
        <div className="text-center py-6 text-sm text-gray-400">
          — {t('common.loadedAll')} —
        </div>
      )}
    </div>
  )
}