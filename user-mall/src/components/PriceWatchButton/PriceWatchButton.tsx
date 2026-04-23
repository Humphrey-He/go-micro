import { useState } from 'react'
import { Button, Dialog, Input, Toast } from 'antd-mobile'
import { setPriceWatch, cancelPriceWatch } from '@/api/priceWatch'

interface Props {
  skuId: string
  productName: string
  currentPrice: number
  isWatching?: boolean
  onWatchChange?: (isWatching: boolean) => void
}

export default function PriceWatchButton({
  skuId,
  productName,
  currentPrice,
  isWatching = false,
  onWatchChange
}: Props) {
  const [showDialog, setShowDialog] = useState(false)
  const [targetPrice, setTargetPrice] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSetWatch = async () => {
    try {
      setLoading(true)
      const target = targetPrice ? Math.round(parseFloat(targetPrice) * 100) : undefined
      await setPriceWatch({ sku_id: skuId, target_price: target })
      Toast.show('降价提醒设置成功')
      setShowDialog(false)
      onWatchChange?.(true)
    } catch (error: any) {
      Toast.show(error?.message || '设置失败')
    } finally {
      setLoading(false)
    }
  }

  const handleCancelWatch = async () => {
    try {
      setLoading(true)
      await cancelPriceWatch(skuId)
      Toast.show('已取消降价提醒')
      onWatchChange?.(false)
    } catch (error: any) {
      Toast.show(error?.message || '取消失败')
    } finally {
      setLoading(false)
    }
  }

  const handleClick = () => {
    if (isWatching) {
      Dialog.confirm({
        content: '确定要取消该商品的降价提醒吗？',
        onConfirm: handleCancelWatch,
      })
    } else {
      setTargetPrice('')
      setShowDialog(true)
    }
  }

  const formatPrice = (price: number) => `¥${(price / 100).toFixed(2)}`

  return (
    <>
      <Button
        size="small"
        color={isWatching ? 'success' : 'default'}
        fill={isWatching ? 'solid' : 'outline'}
        onClick={handleClick}
      >
        <span className="flex items-center gap-1">
          <span>🔔</span>
          <span>{isWatching ? '已设置提醒' : '降价提醒'}</span>
        </span>
      </Button>

      {/* 设置提醒弹窗 */}
      <Dialog
        visible={showDialog}
        title="设置降价提醒"
        onClose={() => setShowDialog(false)}
        content={
          <div className="py-2">
            <div className="text-sm text-gray-600 mb-3">
              商品当前价格：<span className="text-[#00C853] font-bold">{formatPrice(currentPrice)}</span>
            </div>
            <div className="mb-2 text-sm text-gray-500">目标价格（选填）</div>
            <Input
              type="number"
              placeholder="输入目标价格，降价到该价格时提醒"
              value={targetPrice}
              onChange={setTargetPrice}
              prefix="¥"
            />
            <div className="text-xs text-gray-400 mt-2">
              不填目标价格表示：价格下降任意幅度都提醒
            </div>
          </div>
        }
        actions={[
          { text: '取消', key: 'cancel', onClick: () => setShowDialog(false) },
          { text: '确认', key: 'confirm', onClick: handleSetWatch, loading: loading },
        ]}
      />
    </>
  )
}
