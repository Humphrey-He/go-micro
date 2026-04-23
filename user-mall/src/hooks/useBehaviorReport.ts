import { useCallback, useRef, useEffect } from 'react'
import { reportBehavior, BehaviorType, RecommendationSource } from '@/api/recommendation'

interface UseBehaviorReportOptions {
  skuId: string
  source: RecommendationSource
  enabled?: boolean
}

export function useBehaviorReport({ skuId, source, enabled = true }: UseBehaviorReportOptions) {
  const viewStartTime = useRef<number>(0)
  const hasReported = useRef(false)

  // 页面可见性变化时上报浏览时长
  useEffect(() => {
    if (!enabled) return

    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        viewStartTime.current = Date.now()
      } else if (document.visibilityState === 'hidden' && !hasReported.current) {
        const stayDuration = Math.round((Date.now() - viewStartTime.current) / 1000)
        if (stayDuration >= 2) {
          // 停留2秒以上才上报
          reportBehavior({
            sku_id: skuId,
            behavior_type: 'view',
            stay_duration: stayDuration,
            source,
          })
          hasReported.current = true
        }
      }
    }

    viewStartTime.current = Date.now()
    document.addEventListener('visibilitychange', handleVisibilityChange)

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange)
    }
  }, [skuId, source, enabled])

  // 上报收藏
  const reportFavorite = useCallback(
    async (action: 'add' | 'remove') => {
      if (!enabled) return
      try {
        await reportBehavior({
          sku_id: skuId,
          behavior_type: 'favorite',
          source,
        })
      } catch (err) {
        console.error('Failed to report favorite:', err)
      }
    },
    [skuId, source, enabled]
  )

  // 上报加购
  const reportCart = useCallback(
    async (action: 'add' | 'remove', quantity: number = 1) => {
      if (!enabled) return
      try {
        await reportBehavior({
          sku_id: skuId,
          behavior_type: 'cart',
          source,
        })
      } catch (err) {
        console.error('Failed to report cart:', err)
      }
    },
    [skuId, source, enabled]
  )

  // 上报购买
  const reportPurchase = useCallback(
    async (orderId: string) => {
      if (!enabled) return
      try {
        await reportBehavior({
          sku_id: skuId,
          behavior_type: 'purchase',
          source,
        })
      } catch (err) {
        console.error('Failed to report purchase:', err)
      }
    },
    [skuId, source, enabled]
  )

  return {
    reportFavorite,
    reportCart,
    reportPurchase,
  }
}
